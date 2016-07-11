package server

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"time"
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

func newTracingHttpHandler(handler http.Handler, metricsEnabled bool, verbose bool) http.Handler {
	return &tracingHttpHandler{
		handler:        handler,
		metricsEnabled: metricsEnabled,
		verbose:        verbose,
	}
}

type tracingHttpHandler struct {
	handler        http.Handler
	metricsEnabled bool
	verbose        bool
}

func (this *tracingHttpHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if this.metricsEnabled {
		httpRequestCount.WithLabelValues(request.Method).Inc()
	}
	if this.verbose {
		log.Printf("%s %s", request.Method, request.URL.Path)
	}
	this.handler.ServeHTTP(writer, request)
}
