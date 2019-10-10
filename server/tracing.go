// Copyright 2019 Jeremy Edwards
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

package server

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestByMethodCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_by_method_count",
			Help: "RPC latency distributions.",
		},
		[]string{"method"},
	)
)

func init() {
	// Register the summary and the histogram with Prometheus's default registry.
	prometheus.MustRegister(httpRequestByMethodCount)
}

func newTracingHTTPHandler(handler http.Handler, metricsEnabled bool, verbose bool) http.Handler {
	return &tracingHTTPHandler{
		handler:        handler,
		metricsEnabled: metricsEnabled,
		verbose:        verbose,
	}
}

type tracingHTTPHandler struct {
	handler        http.Handler
	metricsEnabled bool
	verbose        bool
}

func (thh *tracingHTTPHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if thh.metricsEnabled {
		httpRequestByMethodCount.WithLabelValues(request.Method).Inc()
	}
	if thh.verbose {
		log.Printf("%s %s", request.Method, request.URL.Path)
	}
	thh.handler.ServeHTTP(writer, request)
}
