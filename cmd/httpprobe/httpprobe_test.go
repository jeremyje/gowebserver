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

package main

import (
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProbe(t *testing.T) {
	insecure := httptest.NewServer(&okHandler{})
	defer insecure.Close()
	secure := httptest.NewTLSServer(&okHandler{})
	defer secure.Close()

	certFile := filepath.Join(t.TempDir(), "test-probe.cert")

	if err := os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: secure.Certificate().Raw,
	}), 0644); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(certFile)
	t.Logf("%s %s", content, err)

	tests := []struct {
		url      string
		certFile string
		wantCode int
	}{
		{
			url:      insecure.URL,
			wantCode: errnoSuccess,
		},
		{
			url:      insecure.URL,
			certFile: "does-not-exist",
			wantCode: errnoFileNotFound,
		},
		{
			url:      "bad",
			wantCode: 503,
		},
		{
			url:      secure.URL,
			certFile: certFile,
			wantCode: errnoSuccess,
		},
		{
			url:      secure.URL,
			wantCode: 503,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("%s_%s", tc.url, tc.certFile), func(t *testing.T) {
			got := run(tc.url, tc.certFile, time.Second)
			if got != tc.wantCode {
				t.Errorf("want: %v, got %v", tc.wantCode, got)
			}
		})
	}
}

type okHandler struct {
}

func (h *okHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
