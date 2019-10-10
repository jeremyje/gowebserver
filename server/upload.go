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

// https://astaxie.gitbooks.io/build-web-application-with-golang/content/en/04.5.html
// http://sanatgersappa.blogspot.com/2013/03/handling-multiple-file-uploads-in-go.html
import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
	"time"

	"github.com/jeremyje/gowebserver/embedded"
	"github.com/prometheus/client_golang/prometheus"
)

var (
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
	uploadHTMLPage     = "ajaxupload.html"
	uploadFileFormName = "gowebserveruploadfile[]"
)

type uploadHTTPHandler struct {
	uploadServePath string
	uploadDirectory string
}

type uploadResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func (uh *uploadHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		tmpl := template.New("")
		t, err := tmpl.Parse(string(embedded.MustAsset(uploadHTMLPage)))
		if err != nil {
			fmt.Printf("Error parsing html template: %s", err)
			w.Write([]byte(err.Error()))
			return
		}
		var params = struct {
			UploadServePath    string
			UploadToken        string
			UploadFileFormName string
		}{uh.uploadServePath, token, uploadFileFormName}
		t.Execute(w, params)
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
				writeUploadResponse(w, resp)
				return
			}
			defer file.Close()
			err = os.MkdirAll(uh.uploadDirectory, 0766)
			if err != nil {
				resp.Error = fmt.Sprintf("InternalError: Cannot create directory to store file (%s), %s", uh.uploadDirectory, err)
				writeUploadResponse(w, resp)
				return
			}
			localPath := filepath.Join(uh.uploadDirectory, filepath.Base(fileName))
			f, err := os.OpenFile(localPath, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				resp.Error = fmt.Sprintf("InternalError: Cannot create file (%s), %s", localPath, err)
				writeUploadResponse(w, resp)
				return
			}
			defer f.Close()
			bytesWritten, err := io.Copy(f, file)
			if err != nil {
				resp.Error = fmt.Sprintf("InternalError: Cannot write file (%s), %s", localPath, err)
				writeUploadResponse(w, resp)
			}
			uploadedBytesTotal.Add(float64(bytesWritten))
			uploadedFilesTotal.Inc()
			log.Printf("Upload %s complete, wrote %s.", fileName, localPath)
		}

		resp.Success = true
		writeUploadResponse(w, resp)
	}
}

func newUploadHandler(uploadServePath string, uploadDirectory string) http.Handler {
	return &uploadHTTPHandler{
		uploadServePath: uploadServePath,
		uploadDirectory: uploadDirectory,
	}
}

func writeUploadResponse(w http.ResponseWriter, resp uploadResponse) {
	if len(resp.Error) > 0 {
		log.Println(resp.Error)
		w.WriteHeader(http.StatusBadRequest)
	}
	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "InternalError: Malformed server response.", http.StatusInternalServerError)
		log.Printf("Error: Cannot marshal upload JSON response, %s", err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}
