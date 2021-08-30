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
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/jeremyje/gowebserver/filesystem"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
)

// WebServer is a convience wrapper for Go's HTTP/HTTPS Web serving API.
type WebServer interface {
	// SetPorts sets the ports for the server.
	SetPorts(httpPort, httpsPort int) WebServer
	// SetMetricsEnabled enables Prometheus metrics export.
	SetMetricsEnabled(enabled bool) WebServer
	// SetServePath specifies the path to serve the file system.
	SetServePath(fileSystemServePath string, metricsServePath string) WebServer
	// SetPath sets the directory to serve.
	SetPath(path string) error
	// SetCertificateFile sets the certificates that should be used to serve HTTPS traffic.
	SetCertificateFile(certificateFilePath string) WebServer
	// SetPrivateKey sets the private key file path for HTTPS traffic encryption.
	SetPrivateKey(privateKeyFilePath string) WebServer
	// SetVerbose sets verbose logging.
	SetVerbose(verbose bool) WebServer
	// SetUpload sets the upload endpoint and upload directory.
	SetUpload(uploadPath string, uploadServePath string) error
	// Serve starts serving the HTTP/HTTPS server synchronously.
	Serve(<-chan error) error
}

type webServerImpl struct {
	httpPort            string
	httpsPort           string
	metricsEnabled      bool
	fileSystemServePath string
	metricsServePath    string
	certificateFilePath string
	privateKeyFilePath  string
	servingPath         string
	verbose             bool
	uploadPath          string
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

func (ws *webServerImpl) SetPath(path string) error {
	if len(path) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		path = cwd
	}
	ws.servingPath = path
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

func (ws *webServerImpl) SetUpload(uploadPath string, uploadServePath string) error {
	if len(uploadPath) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		uploadPath = cwd
	}
	ws.uploadPath = uploadPath
	ws.uploadServePath = uploadServePath
	return nil
}

func (ws *webServerImpl) addHandler(serverMux *http.ServeMux, servePath string, handler http.Handler) {
	//	if ws.metricsEnabled {
	//		serverMux.HandleFunc(servePath, promhttp.InstrumentHandler(servePath, handler))
	//} else {
	serverMux.Handle(servePath, handler)
	//}
}

func (ws *webServerImpl) Serve(termCh <-chan error) error {
	log.Printf("Serving %s on %s and %s", ws.servingPath, ws.httpPort, ws.httpsPort)
	fsHandler, err := filesystem.New(ws.servingPath)
	if err != nil {
		return err
	}
	serverMux := http.NewServeMux()
	if ws.metricsEnabled {
		serverMux.Handle(ws.metricsServePath, promhttp.Handler())
	}
	ws.addHandler(serverMux, ws.fileSystemServePath, fsHandler)

	if len(ws.uploadServePath) > 0 {
		uploadHandler := newUploadHandler(ws.uploadServePath, ws.uploadPath)
		ws.addHandler(serverMux, ws.uploadServePath, uploadHandler)
	}
	corsHandler := cors.Default().Handler(serverMux)
	httpHandler := newTracingHTTPHandler(corsHandler, ws.metricsEnabled, ws.verbose)

	httpSocket, err := net.Listen("tcp", ws.httpPort)
	if err != nil {
		return err
	}

	defer httpSocket.Close()

	httpsSocket, err := net.Listen("tcp", ws.httpsPort)
	if err != nil {
		return err
	}

	defer httpsSocket.Close()

	go func() {
		checkError(http.ServeTLS(httpsSocket, httpHandler, ws.certificateFilePath, ws.privateKeyFilePath))
	}()
	go func() {
		checkError(http.Serve(httpSocket, httpHandler))
	}()

	<-termCh
	return nil
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
		servingPath:         "",
	}
}
