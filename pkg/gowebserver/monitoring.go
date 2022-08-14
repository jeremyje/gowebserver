package gowebserver

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/contrib/zpages"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
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

	prom, err := newPrometheusExporter(m, r)
	if err != nil {
		mc.shutdown()
		return nil, err
	}
	if prom != nil {
		mc.prom = prom
		if m.Metrics.Path != "" {
			mc.handlers[m.Metrics.Path] = prom
		}

		if err := host.Start(); err != nil {
			mc.shutdown()
			return nil, err
		}

		if err := runtime.Start(
			runtime.WithMeterProvider(prom.MeterProvider()),
			runtime.WithMinimumReadMemStatsInterval(10*time.Second),
		); err != nil {
			mc.shutdown()
			return nil, err
		}
	}

	return mc, nil
}

func newJaegerExporter(m Monitoring, r *resource.Resource) (*sdktrace.TracerProvider, error) {
	if len(m.Trace.URI) > 0 {
		exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(m.Trace.URI)))
		if err != nil {
			return nil, err
		}

		tp := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exp),
			sdktrace.WithResource(r),
		)
		return tp, nil
	}
	return nil, nil
}

func newPrometheusExporter(m Monitoring, r *resource.Resource) (*prometheus.Exporter, error) {
	config := prometheus.Config{
		DefaultHistogramBoundaries: []float64{1, 2, 5, 10, 20, 50},
	}
	c := controller.New(
		processor.NewFactory(
			selector.NewWithHistogramDistribution(
				histogram.WithExplicitBoundaries(config.DefaultHistogramBoundaries),
			),
			aggregation.CumulativeTemporalitySelector(),
			processor.WithMemory(true),
		),
		controller.WithResource(r),
	)

	exporter, err := prometheus.New(config, c)
	if err != nil {
		return nil, err
	}

	global.SetMeterProvider(exporter.MeterProvider())
	return exporter, nil
}

type monitoringContext struct {
	handlers map[string]http.Handler
	prom     *prometheus.Exporter
	tp       *sdktrace.TracerProvider
}

func (m *monitoringContext) getTraceProvider() trace.TracerProvider {
	if m == nil || m.tp == nil {
		return otel.GetTracerProvider()
	}
	return m.tp
}

func (m *monitoringContext) getMeterProvider() metric.MeterProvider {
	if m == nil || m.prom == nil {
		return global.MeterProvider()
	}
	return m.prom.MeterProvider()
}
func (m *monitoringContext) shutdown() {
	if m == nil {
		return
	}
	ctx := context.Background()
	if m.prom != nil {
		runtime.Start()
		m.prom = nil
	}
	if m.tp != nil {
		m.tp.Shutdown(ctx)
	}
}
