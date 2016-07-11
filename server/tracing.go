package server

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"time"
)

var (
	uniformDomain     = flag.Float64("uniform.domain", 200, "The domain for the uniform distribution.")
	normDomain        = flag.Float64("normal.domain", 200, "The domain for the normal distribution.")
	normMean          = flag.Float64("normal.mean", 10, "The mean for the normal distribution.")
	oscillationPeriod = flag.Duration("oscillation-period", 10*time.Minute, "The duration of the rate oscillation period.")
)

var (
	httpRequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_count",
			Help: "RPC latency distributions.",
		},
		[]string{"method"},
	)

	httpRequestByPathCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_count_by_path",
			Help: "RPC latency distributions.",
		},
		[]string{"method", "path"},
	)
)

func init() {
	// Register the summary and the histogram with Prometheus's default registry.
	prometheus.MustRegister(httpRequestCount)
	prometheus.MustRegister(httpRequestByPathCount)
}

func newTracingHttpHandler(handler http.Handler) http.Handler {
	return &tracingHttpHandler{
		handler: handler,
	}
}

type tracingHttpHandler struct {
	handler http.Handler
}

func (this *tracingHttpHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	httpRequestCount.WithLabelValues(request.Method).Inc()
	httpRequestByPathCount.WithLabelValues(request.Method, request.URL.Path).Inc()
	this.handler.ServeHTTP(writer, request)
}