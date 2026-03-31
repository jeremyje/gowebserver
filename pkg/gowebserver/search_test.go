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
	"testing"

	gowsTesting "github.com/jeremyje/gowebserver/v2/internal/gowebserver/testing"
)

func TestNullSearchFS(t *testing.T) {
	nestedZipPath := gowsTesting.MustNestedZipFilePath(t)

	vFS, err := newRawFSFromURI(nestedZipPath)
	if err != nil {
		t.Error(err)
	}
	defer vFS.Close()
	nFS := newNestedFS(vFS)
	defer nFS.Close()

	sFS := newNullSearchFS(nFS)
	if sFS == nil {
		t.Fatal("searchFS is nil")
	}
}
