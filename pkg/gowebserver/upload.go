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

// https://astaxie.gitbooks.io/build-web-application-with-golang/content/en/04.5.html
// http://sanatgersappa.blogspot.com/2013/03/handling-multiple-file-uploads-in-go.html
import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
	"time"

	_ "embed"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var (

	//go:embed upload.html
	uploadHTML []byte

	uploadedBytesTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "uploaded_bytes_total",
			Help: "Number of bytes uploaded.",
		},
	)
	uploadedFilesTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "uploaded_files_total",
			Help: "Number of files uploaded.",
		},
	)
)

func init() {
	prometheus.MustRegister(uploadedBytesTotal)
	prometheus.MustRegister(uploadedFilesTotal)
}

const (
	uploadFileFormName = "gowebserveruploadfile[]"
)

type uploadHTTPHandler struct {
	uploadHTTPPath  string
	uploadDirectory string
}

type uploadResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func (uh *uploadHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := zap.S().With("url", r.URL)
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		tmpl := template.New("")
		t, err := tmpl.Parse(string(uploadHTML))
		if err != nil {
			logger.With("error", err).Error("Error parsing html template")
			w.Write([]byte(err.Error()))
			return
		}
		var params = struct {
			UploadHTTPPath     string
			UploadToken        string
			UploadFileFormName string
		}{uh.uploadHTTPPath, token, uploadFileFormName}
		if err := t.Execute(w, params); err != nil {
			logger.With("error", err).Error("cannot parse upload.html template.")
		}
	} else {
		var resp uploadResponse
		r.ParseMultipartForm(32 << 20)
		m := r.MultipartForm
		files := m.File[uploadFileFormName]
		for i := range files {
			fileName := files[i].Filename
			file, err := files[i].Open()
			if err != nil {
				resp.Error = fmt.Sprintf("InternalError: Cannot download file (%s), %s", fileName, err)
				writeUploadResponse(w, resp, logger)
				return
			}
			defer file.Close()
			err = os.MkdirAll(uh.uploadDirectory, 0766)
			if err != nil {
				resp.Error = fmt.Sprintf("InternalError: Cannot create directory to store file (%s), %s", uh.uploadDirectory, err)
				writeUploadResponse(w, resp, logger)
				return
			}
			localPath := filepath.Join(uh.uploadDirectory, filepath.Base(fileName))
			f, err := os.OpenFile(localPath, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				resp.Error = fmt.Sprintf("InternalError: Cannot create file (%s), %s", localPath, err)
				writeUploadResponse(w, resp, logger)
				return
			}
			defer f.Close()
			bytesWritten, err := io.Copy(f, file)
			if err != nil {
				resp.Error = fmt.Sprintf("InternalError: Cannot write file (%s), %s", localPath, err)
				writeUploadResponse(w, resp, logger)
			}
			uploadedBytesTotal.Add(float64(bytesWritten))
			uploadedFilesTotal.Inc()
			logger.With("fileName", fileName).With("localPath", localPath).Info("Upload Complete")
		}

		resp.Success = true
		writeUploadResponse(w, resp, logger)
	}
}

func newUploadHandler(uploadHTTPPath string, uploadDirectory string) http.Handler {
	return &uploadHTTPHandler{
		uploadHTTPPath:  uploadHTTPPath,
		uploadDirectory: uploadDirectory,
	}
}

func writeUploadResponse(w http.ResponseWriter, resp uploadResponse, logger *zap.SugaredLogger) {
	if len(resp.Error) > 0 {
		logger.With("error", resp.Error).Warn("Upload error")
		w.WriteHeader(http.StatusBadRequest)
	} else {
		logger.Debug("Upload Successful")
	}
	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Malformed server response.", http.StatusInternalServerError)
		logger.With("error", err).Warn("Cannot marshal upload JSON response")
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}
