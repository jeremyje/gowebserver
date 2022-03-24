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
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"go.uber.org/zap"
)

// WebServer is a convience wrapper for Go's HTTP/HTTPS Web serving API.
type WebServer interface {
	// Serve starts serving the HTTP/HTTPS server synchronously.
	Serve(<-chan error) error
}

type webServerImpl struct {
	httpAddr            string
	httpsAddr           string
	metricsEnabled      bool
	fileSystemServePath []servePath
	metricsServePath    string
	certificateFilePath string
	privateKeyFilePath  string
	verbose             bool
	uploadPath          string
	uploadHTTPPath      string

	httpListenPort  int
	httpsListenPort int

	sync.RWMutex
}

type servePath struct {
	localPath string
	httpPath  string
}

func expandPath(dir string) (string, error) {
	if len(dir) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return cwd, nil
	}
	return dir, nil
}

func (ws *webServerImpl) addHandler(serverMux *http.ServeMux, servePath string, handler http.Handler) {
	//	if ws.metricsEnabled {
	//		serverMux.HandleFunc(servePath, promhttp.InstrumentHandler(servePath, handler))
	//} else {
	serverMux.Handle(servePath, handler)
	//}
}

func getPort(lis net.Listener) (int, error) {
	addr, ok := lis.Addr().(*net.TCPAddr)
	if ok {
		return addr.Port, nil
	}
	return 0, fmt.Errorf("cannot get port from '%s'", lis)
}

func (ws *webServerImpl) setPorts(httpPort int, httpsPort int) {
	ws.Lock()
	ws.httpListenPort = httpPort
	ws.httpsListenPort = httpsPort
	ws.Unlock()
}

func (ws *webServerImpl) getPorts() (int, int) {
	ws.RLock()
	httpPort := ws.httpListenPort
	httpsPort := ws.httpsListenPort
	ws.RUnlock()
	return httpPort, httpsPort
}

func (ws *webServerImpl) Serve(termCh <-chan error) error {
	displayPath := ""
	for i, paths := range ws.fileSystemServePath {
		if i > 0 {
			displayPath += ","
		}
		displayPath += paths.localPath
	}
	zap.S().With("HTTP", ws.httpAddr, "HTTPS", ws.httpsAddr).Info("Serving")
	serverMux := http.NewServeMux()
	if ws.metricsEnabled {
		serverMux.Handle(ws.metricsServePath, promhttp.Handler())
	}
	allCleanups := []func(){}
	hasIndex := false
	endpoints := []string{}
	for _, paths := range ws.fileSystemServePath {
		zap.S().With("localPath", paths.localPath, "http", paths.httpPath).Info("Endpoint")
		endpoints = append(endpoints, paths.httpPath)
		fsHandler, cleanup, err := newFS(paths.localPath)
		if err != nil {
			return err
		}
		allCleanups = append(allCleanups, cleanup)

		if paths.httpPath == "/" {
			ws.addHandler(serverMux, paths.httpPath, fsHandler)
			hasIndex = true
		} else {
			ws.addHandler(serverMux, paths.httpPath, http.StripPrefix(paths.httpPath, fsHandler))
		}
	}

	if !hasIndex {
		h := &indexHTTPHandler{
			servePaths: endpoints,
		}
		ws.addHandler(serverMux, "/", h)
	}

	defer func() {
		for _, cleanup := range allCleanups {
			cleanup()
		}
	}()

	if len(ws.uploadHTTPPath) > 0 {
		uploadHandler := newUploadHandler(ws.uploadHTTPPath, ws.uploadPath)
		ws.addHandler(serverMux, ws.uploadHTTPPath, uploadHandler)
	}

	corsHandler := cors.Default().Handler(serverMux)
	httpHandler := newTracingHTTPHandler(corsHandler, ws.metricsEnabled)

	httpSocket, err := net.Listen("tcp", ws.httpAddr)
	if err != nil {
		return err
	}

	defer httpSocket.Close()

	httpsSocket, err := net.Listen("tcp", ws.httpsAddr)
	if err != nil {
		return err
	}

	defer httpsSocket.Close()

	httpPort, err := getPort(httpSocket)
	if err != nil {
		zap.S().With("error", err).Error("cannot get port from HTTP listener")
	}
	httpsPort, err := getPort(httpsSocket)
	if err != nil {
		zap.S().With("error", err).Error("cannot get port from HTTPS listener")
	}
	ws.setPorts(httpPort, httpsPort)

	go func() {
		checkError(http.ServeTLS(httpsSocket, httpHandler, ws.certificateFilePath, ws.privateKeyFilePath))
	}()
	go func() {
		checkError(http.Serve(httpSocket, httpHandler))
	}()

	<-termCh
	return nil
}

func New(conf *Config) (WebServer, error) {
	if conf == nil {
		conf = &Config{}
	}
	sp := []servePath{}
	for _, paths := range conf.Serve {
		p, err := expandPath(paths.Source)
		if err != nil {
			return nil, fmt.Errorf("cannot expand path '%s', %s", paths.Source, err)
		}
		sp = append(sp, servePath{
			localPath: p,
			httpPath:  normalizeHTTPPath(paths.HTTPPath),
		})
	}

	uploadPath := ""
	if conf.UploadPath != "" {
		dir, err := expandPath(conf.UploadPath)
		if err != nil {
			return nil, fmt.Errorf("cannot expand path '%s', %s", conf.UploadPath, err)
		}
		uploadPath = dir
	}

	ws := &webServerImpl{
		httpAddr:            ":" + strconv.Itoa(conf.HTTP.Port),
		httpsAddr:           ":" + strconv.Itoa(conf.HTTPS.Port),
		metricsEnabled:      conf.Metrics.Enabled,
		fileSystemServePath: sp,
		metricsServePath:    conf.Metrics.Path,
		certificateFilePath: conf.HTTPS.Certificate.CertificateFilePath,
		privateKeyFilePath:  conf.HTTPS.Certificate.PrivateKeyFilePath,
		uploadPath:          uploadPath,
		uploadHTTPPath:      conf.UploadHTTPPath,
		verbose:             conf.Verbose,
	}

	return ws, nil
}

func normalizeHTTPPath(path string) string {
	return strings.ReplaceAll("/"+strings.Trim(path, "/")+"/", "//", "/")
}
