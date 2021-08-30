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

package filesystem

import (
	"net/http"
)

// New creates a filesystem for the HTTP server from an archive file.
func New(path string) (http.Handler, error) {
	if isSupportedZip(path) {
		staged := newZipFs(path)
		return staged.handler, staged.err
	} else if isSupportedTar(path) {
		staged := newTarFs(path)
		return staged.handler, staged.err
	} else if isSupportedGit(path) {
		staged := newGitFs(path)
		return staged.handler, staged.err
	} else if isSupportedHTTP(path) {
		staged := newHTTPReverseProxy(path)
		return staged.handler, staged.err
	}
	return newNative(path)
}
