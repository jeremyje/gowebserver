package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type WebServer interface {
	SetPorts(httpPort, httpsPort int) WebServer
	SetMetricsEnabled(enabled bool) WebServer
	SetServePath(fileSystemServePath string, metricsServePath string) WebServer
	SetDirectory(dir string) error
	SetCertificateFile(certificateFilePath string) WebServer
	SetPrivateKey(privateKeyFilePath string) WebServer
	SetVerbose(verbose bool) WebServer
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
	verbose             bool
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

func (this *WebServerImpl) SetVerbose(verbose bool) WebServer {
	this.verbose = verbose
	return this
}

func (this *WebServerImpl) Serve() {
	log.Printf("Serving %s on %s and %s", this.servingDirectory, this.httpPort, this.httpsPort)
	fsHandler := http.FileServer(http.Dir(this.servingDirectory + "/"))
	serverMux := http.NewServeMux()
	var httpHandler http.Handler
	if this.metricsEnabled {
		serverMux.Handle(this.metricsServePath, prometheus.Handler())
		serverMux.HandleFunc(this.fileSystemServePath, prometheus.InstrumentHandler(this.fileSystemServePath, fsHandler))
		httpHandler = newTracingHttpHandler(serverMux)
	} else {
		serverMux.Handle(this.fileSystemServePath, fsHandler)
		httpHandler = serverMux
	}
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
