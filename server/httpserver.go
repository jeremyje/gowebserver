package server

import (
	"github.com/jeremyje/gowebserver/filesystem"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"strconv"
)

// WebServer is a convience wrapper for Go's HTTP/HTTPS Web serving API.
type WebServer interface {
	// SetPorts sets the ports for the server.
	SetPorts(httpPort, httpsPort int) WebServer
	// SetMetricsEnabled enables Prometheus metrics export.
	SetMetricsEnabled(enabled bool) WebServer
	// SetServePath specifies the path to serve the file system.
	SetServePath(fileSystemServePath string, metricsServePath string) WebServer
	// SetDirectory sets the directory to serve.
	SetDirectory(dir string) error
	// SetCertificateFile sets the certificates that should be used to serve HTTPS traffic.
	SetCertificateFile(certificateFilePath string) WebServer
	// SetPrivateKey sets the private key file path for HTTPS traffic encryption.
	SetPrivateKey(privateKeyFilePath string) WebServer
	// SetVerbose sets verbose logging.
	SetVerbose(verbose bool) WebServer
	// SetUpload sets the upload endpoint and upload directory.
	SetUpload(uploadDirectory string, uploadServePath string) error
	// Serve starts serving the HTTP/HTTPS server synchronously.
	Serve()
}

type webServerImpl struct {
	httpPort            string
	httpsPort           string
	metricsEnabled      bool
	fileSystemServePath string
	metricsServePath    string
	certificateFilePath string
	privateKeyFilePath  string
	servingDirectory    string
	verbose             bool
	uploadDirectory     string
	uploadServePath     string
}

func (ws *webServerImpl) SetPorts(httpPort, httpsPort int) WebServer {
	ws.httpPort = ":" + strconv.Itoa(httpPort)
	ws.httpsPort = ":" + strconv.Itoa(httpsPort)
	return ws
}

func (ws *webServerImpl) SetMetricsEnabled(enabled bool) WebServer {
	ws.metricsEnabled = enabled
	return ws
}

func (ws *webServerImpl) SetServePath(fileSystemServePath string, metricsServePath string) WebServer {
	ws.fileSystemServePath = fileSystemServePath
	ws.metricsServePath = metricsServePath
	return ws
}

func (ws *webServerImpl) SetDirectory(dir string) error {
	if len(dir) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		dir = cwd
	}
	ws.servingDirectory = dir
	return nil
}

func (ws *webServerImpl) SetCertificateFile(certificateFilePath string) WebServer {
	ws.certificateFilePath = certificateFilePath
	return ws
}

func (ws *webServerImpl) SetPrivateKey(privateKeyFilePath string) WebServer {
	ws.privateKeyFilePath = privateKeyFilePath
	return ws
}

func (ws *webServerImpl) SetVerbose(verbose bool) WebServer {
	ws.verbose = verbose
	return ws
}

func (ws *webServerImpl) SetUpload(uploadDirectory string, uploadServePath string) error {
	if len(uploadDirectory) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		uploadDirectory = cwd
	}
	ws.uploadDirectory = uploadDirectory
	ws.uploadServePath = uploadServePath
	return nil
}

func (ws *webServerImpl) addHandler(serverMux *http.ServeMux, servePath string, handler http.Handler) {
	if ws.metricsEnabled {
		serverMux.HandleFunc(servePath, prometheus.InstrumentHandler(servePath, handler))
	} else {
		serverMux.Handle(servePath, handler)
	}
}

func (ws *webServerImpl) Serve() {
	log.Printf("Serving %s on %s and %s", ws.servingDirectory, ws.httpPort, ws.httpsPort)
	httpFs, err := filesystem.New(ws.servingDirectory)
	if err != nil {
		log.Fatal(err)
	}
	fsHandler := http.FileServer(httpFs)
	serverMux := http.NewServeMux()
	if ws.metricsEnabled {
		serverMux.Handle(ws.metricsServePath, prometheus.Handler())
		//serverMux.HandleFunc(ws.fileSystemServePath, prometheus.InstrumentHandler(ws.fileSystemServePath, fsHandler))
	} else {
		//serverMux.Handle(ws.fileSystemServePath, fsHandler)
	}
	ws.addHandler(serverMux, ws.fileSystemServePath, fsHandler)

	if len(ws.uploadServePath) > 0 {
		uploadHandler := newUploadHandler(ws.uploadServePath, ws.uploadDirectory)
		ws.addHandler(serverMux, ws.uploadServePath, uploadHandler)
		//serverMux.Handle(ws.uploadServePath, uploadHandler)
	}
	corsHandler := cors.Default().Handler(serverMux)
	httpHandler := newTracingHTTPHandler(corsHandler, ws.metricsEnabled, ws.verbose)
	go func() {
		err := http.ListenAndServeTLS(ws.httpsPort, ws.certificateFilePath, ws.privateKeyFilePath, httpHandler)
		if err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		err := http.ListenAndServe(ws.httpPort, httpHandler)
		if err != nil {
			log.Fatal(err)
		}
	}()
	ch := make(chan bool)
	<-ch
}

// NewWebServer creates a new web server instance.
func NewWebServer() WebServer {
	return &webServerImpl{
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
