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
	"bytes"
	"net/http"

	_ "embed"
)

var (
	//go:embed template-index.html
	templateIndexHTML []byte
)

type indexHTTPHandler struct {
	page []byte
}

func newIndexHTTPHandler(servePaths []string) (*indexHTTPHandler, error) {
	w := &bytes.Buffer{}
	var params = struct {
		ServePaths []string
	}{servePaths}
	if err := executeTemplate(templateIndexHTML, params, w); err != nil {
		return nil, err
	}

	return &indexHTTPHandler{
		page: w.Bytes(),
	}, nil
}

func (h *indexHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.Write(h.page)
}
