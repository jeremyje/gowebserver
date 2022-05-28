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

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/metric/unit"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var (

	//go:embed upload.html
	uploadHTML []byte
)

const (
	uploadFileFormName = "gowebserveruploadfile[]"
)

type uploadHTTPHandler struct {
	tp                 trace.TracerProvider
	uploadHTTPPath     string
	uploadDirectory    string
	uploadedBytesTotal syncint64.Counter
	uploadedFilesTotal syncint64.Counter
}

type uploadResponse struct {
	Success bool  `json:"success"`
	Error   error `json:"error,omitempty"`
}

const (
	uploadTraceName = "uploadHTTPHandler"
)

func (uh *uploadHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	uploadTracer := uh.tp.Tracer(uploadTraceName)
	ctx, span := uploadTracer.Start(r.Context(), r.Method)
	defer span.End()

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
			ctx, childSpan := uploadTracer.Start(ctx, fileName)
			defer childSpan.End()
			span.AddEvent("create file", trace.WithAttributes(attribute.String("filename", fileName)))

			file, err := files[i].Open()
			if err != nil {
				resp.Error = fmt.Errorf("InternalError: Cannot download file (%s), %s", fileName, err)
				writeUploadResponse(w, resp, logger, childSpan)
				return
			}
			defer file.Close()
			err = os.MkdirAll(uh.uploadDirectory, 0766)
			if err != nil {
				resp.Error = fmt.Errorf("InternalError: Cannot create directory to store file (%s), %s", uh.uploadDirectory, err)
				writeUploadResponse(w, resp, logger, childSpan)
				return
			}

			name := sanitizeFileName(filepath.Base(fileName))
			localPath := filepath.Join(uh.uploadDirectory, name)
			f, err := os.OpenFile(localPath, os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				resp.Error = fmt.Errorf("InternalError: Cannot create file (%s), %s", localPath, err)
				writeUploadResponse(w, resp, logger, childSpan)
				return
			}
			defer f.Close()
			bytesWritten, err := io.Copy(f, file)
			if err != nil {
				resp.Error = fmt.Errorf("InternalError: Cannot write file (%s), %s", localPath, err)
				writeUploadResponse(w, resp, logger, childSpan)
			}
			childSpan.SetAttributes(attribute.Int64("bytesWritten", bytesWritten))
			uh.uploadedBytesTotal.Add(ctx, bytesWritten)
			uh.uploadedFilesTotal.Add(ctx, 1)
			logger.With("fileName", fileName).With("localPath", localPath).Info("Upload Complete")
		}

		resp.Success = true
		writeUploadResponse(w, resp, logger, span)
	}
}

func newUploadHandler(mc *monitoringContext, uploadHTTPPath string, uploadDirectory string) (http.Handler, error) {
	m := mc.getMeterProvider().Meter(uploadDirectory)

	uploadedBytesTotal, err := m.SyncInt64().Counter("uploaded_bytes_total", instrument.WithDescription("Number of bytes uploaded."), instrument.WithUnit(unit.Bytes))
	if err != nil {
		return nil, err
	}
	uploadedFilesTotal, err := m.SyncInt64().Counter("uploaded_files_total", instrument.WithDescription("Number of files uploaded."), instrument.WithUnit(unit.Dimensionless))
	if err != nil {
		return nil, err
	}
	return &uploadHTTPHandler{
		tp:                 mc.getTraceProvider(),
		uploadHTTPPath:     uploadHTTPPath,
		uploadDirectory:    uploadDirectory,
		uploadedBytesTotal: uploadedBytesTotal,
		uploadedFilesTotal: uploadedFilesTotal,
	}, nil
}

func writeUploadResponse(w http.ResponseWriter, resp uploadResponse, logger *zap.SugaredLogger, span trace.Span) {
	if resp.Error != nil {
		logger.With("error", resp.Error).Warn("Upload error")
		w.WriteHeader(http.StatusBadRequest)
		span.RecordError(resp.Error)
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
