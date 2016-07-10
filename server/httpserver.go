package server

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

type WebServer interface {
	SetPorts(httpPort, httpsPort int) WebServer
	SetMetricsEnabled(enabled bool) WebServer
	SetServePath(fileSystemServePath string, metricsServePath string) WebServer
	SetDirectory(dir string) error
	SetCertificateFile(certificateFilePath string) WebServer
	SetPrivateKey(privateKeyFilePath string) WebServer
	Serve()
}

type WebServerImpl struct {
	httpPort            string
	httpsPort           string
	metricsEnabled      bool
	fileSystemServePath string
	metricsServePath    string
	certificateFilePath string
	privateKeyFilePath  string
	servingDirectory    string
}

func (this *WebServerImpl) SetPorts(httpPort, httpsPort int) WebServer {
	this.httpPort = ":" + strconv.Itoa(httpPort)
	this.httpsPort = ":" + strconv.Itoa(httpsPort)
	return this
}

func (this *WebServerImpl) SetMetricsEnabled(enabled bool) WebServer {
	this.metricsEnabled = enabled
	return this
}

func (this *WebServerImpl) SetServePath(fileSystemServePath string, metricsServePath string) WebServer {
	this.fileSystemServePath = fileSystemServePath
	this.metricsServePath = metricsServePath
	return this
}

func (this *WebServerImpl) SetDirectory(dir string) error {
	if len(dir) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		dir = cwd
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	this.servingDirectory = dir
	return nil
}

func (this *WebServerImpl) SetCertificateFile(certificateFilePath string) WebServer {
	this.certificateFilePath = certificateFilePath
	return this
}

func (this *WebServerImpl) SetPrivateKey(privateKeyFilePath string) WebServer {
	this.privateKeyFilePath = privateKeyFilePath
	return this
}

func (this *WebServerImpl) Serve() {
	log.Printf("Serving %s on %s and %s", this.servingDirectory, this.httpPort, this.httpsPort)
	fsHandler := http.FileServer(http.Dir(this.servingDirectory + "/"))
	serverMux := http.NewServeMux()
	if this.metricsEnabled {
		serverMux.Handle(this.metricsServePath, prometheus.Handler())
		serverMux.HandleFunc(this.fileSystemServePath, prometheus.InstrumentHandler(this.fileSystemServePath, fsHandler))
	} else {
		serverMux.Handle(this.fileSystemServePath, fsHandler)
	}

	httpHandler := newTracingHttpHandler(serverMux)
	go func() {
		err := http.ListenAndServeTLS(this.httpsPort, this.certificateFilePath, this.privateKeyFilePath, httpHandler)
		if err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		err := http.ListenAndServe(this.httpPort, httpHandler)
		if err != nil {
			log.Fatal(err)
		}
	}()
	ch := make(chan bool)
	<-ch
}

func NewWebServer() WebServer {
	return &WebServerImpl{
		httpPort:            "80",
		httpsPort:           "443",
		metricsEnabled:      true,
		fileSystemServePath: "/",
		metricsServePath:    "/metrics",
		certificateFilePath: "",
		privateKeyFilePath:  "",
		servingDirectory:    "",
	}
}
