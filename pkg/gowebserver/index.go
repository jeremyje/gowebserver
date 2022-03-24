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
	"net/http"
	"text/template"

	_ "embed"

	"go.uber.org/zap"
)

var (
	//go:embed index.html
	indexHTML []byte
)

type indexHTTPHandler struct {
	servePaths []string
}

func (h *indexHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := zap.S().With("url", r.URL)

	tmpl := template.New("")
	t, err := tmpl.Parse(string(indexHTML))
	if err != nil {
		logger.With("error", err).Error("Error parsing html template")
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Add("Content-Type", "text/html")
	var params = struct {
		ServePaths []string
	}{h.servePaths}
	if err := t.Execute(w, params); err != nil {
		zap.S().With("error", err).Error("cannot parse index.html template.")
	}
}
