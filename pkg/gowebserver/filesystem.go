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
	"net/http"

	"github.com/jeremyje/gowebserver/v2/pkg/filesystem"
	"go.opentelemetry.io/otel/trace"
)

func newHandlerFromFS(path string, tp trace.TracerProvider, enhancedList bool) (http.Handler, func() error, error) {
	if !filesystem.IsSupportedGit(path) && isSupportedHTTP(path) {
		handler, err := newHTTPReverseProxy(path)
		return handler, nilFuncWithError, err
	}
	vFS, err := filesystem.NewRawFSFromURI(path)
	if err != nil {
		return nil, nilFuncWithError, err
	}
	nFS := filesystem.NewNestedFS(vFS)

	ci, err := newCustomIndex(http.FileServer(http.FS(nFS)), nFS, tp, enhancedList)
	if err != nil {
		nFS.Close()
		return nil, nilFuncWithError, err
	}
	return ci, nFS.Close, nil
}
