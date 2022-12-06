// Copyright 2022 Jeremy Edwards
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gowebserver

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/contrib/zpages"
	"go.opentelemetry.io/otel"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

func setupMonitoring(m Monitoring) (*monitoringContext, error) {
	mc := &monitoringContext{
		handlers: map[string]http.Handler{},
	}

	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("gowebserver"),
			semconv.ServiceVersionKey.String(version),
		),
	)
	if err != nil {
		mc.shutdown()
		return nil, err
	}

	tp, err := newJaegerExporter(m, r)
	if err != nil {
		mc.shutdown()
		return nil, err
	}

	if tp != nil {
		mc.tp = tp
		if len(m.DebugEndpoint) > 0 {
			h := zpages.NewTracezHandler(zpages.NewSpanProcessor())
			mc.handlers[m.DebugEndpoint+"/tracez"] = h
		}
	}

	promExporter, promProvider, prom, err := newPrometheusExporter(m, r)
	if err != nil {
		mc.shutdown()
		return nil, err
	}
	if prom != nil {
		mc.promProvider = promProvider
		mc.promExporter = promExporter
		if m.Metrics.Path != "" {
			mc.handlers[m.Metrics.Path] = prom
		}

		if err := startHostMetrics(); err != nil {
			mc.shutdown()
			return nil, err
		}

		if err := runtime.Start(
			runtime.WithMeterProvider(promProvider),
			runtime.WithMinimumReadMemStatsInterval(10*time.Second),
		); err != nil {
			mc.shutdown()
			return nil, err
		}
	}

	return mc, nil
}

func newPrometheusExporter(m Monitoring, r *resource.Resource) (*otelprom.Exporter, *sdkmetric.MeterProvider, http.Handler, error) {
	registry := prometheus.NewRegistry()
	registry.Register(collectors.NewBuildInfoCollector())
	registry.Register(collectors.NewGoCollector())
	registry.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
		ReportErrors: true,
	}))

	prometheusExporter, err := otelprom.New(otelprom.WithRegisterer(registry))
	if err != nil {
		return nil, nil, nil, err
	}

	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(prometheusExporter))
	global.SetMeterProvider(provider)

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	return prometheusExporter, provider, h, nil
}

type monitoringContext struct {
	handlers     map[string]http.Handler
	promExporter *otelprom.Exporter
	promProvider metric.MeterProvider
	tp           *sdktrace.TracerProvider
}

func (m *monitoringContext) getTraceProvider() trace.TracerProvider {
	if m == nil || m.tp == nil {
		return otel.GetTracerProvider()
	}
	return m.tp
}

func (m *monitoringContext) getMeterProvider() metric.MeterProvider {
	if m == nil || m.promProvider == nil {
		return global.MeterProvider()
	}
	return m.promProvider
}

func (m *monitoringContext) shutdown() {
	if m == nil {
		return
	}
	ctx := context.Background()
	if m.promProvider != nil {
		runtime.Start()
		m.promProvider = nil
	}
	if m.promExporter != nil {
		m.promExporter.Shutdown(ctx)
		m.promExporter = nil
	}
	if m.tp != nil {
		m.tp.Shutdown(ctx)
	}
}
