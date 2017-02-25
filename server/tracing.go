package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
)

var (
	httpRequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_count",
			Help: "RPC latency distributions.",
		},
		[]string{"method"},
	)
)

func init() {
	// Register the summary and the histogram with Prometheus's default registry.
	prometheus.MustRegister(httpRequestCount)
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
		httpRequestCount.WithLabelValues(request.Method).Inc()
	}
	if thh.verbose {
		log.Printf("%s %s", request.Method, request.URL.Path)
	}
	thh.handler.ServeHTTP(writer, request)
}
