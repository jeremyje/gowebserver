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
	"io/ioutil"
	"net/http/httptest"
	"testing"

	_ "embed"

	"github.com/google/go-cmp/cmp"
)

var (
	//go:embed testdata/test-index.html
	testIndexHTML []byte
)

func TestTemplateIndexHTML(t *testing.T) {
	if len(templateIndexHTML) < 50 {
		t.Errorf("data/index.html was not stored")
	}
}

func TestIndexHTTPHandlerServeHTTP(t *testing.T) {
	h, err := newIndexHTTPHandler([]string{"/ok", "/abc"})
	if err != nil {
		t.Fatal(err)
	}
	hs := httptest.NewServer(h)
	defer hs.Close()
	resp, err := hs.Client().Get(hs.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(string(testIndexHTML), string(data)); diff != "" {
		t.Errorf("index mismatch (-want +got):\n%s", diff)
	}
}
