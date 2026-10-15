// Copyright 2026 Cloudfra
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

package downloadmirror

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTransport(t *testing.T) {
	o := newOrigin(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Cache-Control", "max-age=60")
		reply(w, "from origin")
	})
	_, _, m := newCache(t)
	broken := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "broken", http.StatusBadGateway)
	}))
	t.Cleanup(broken.Close)
	down := httptest.NewServer(http.NotFoundHandler())
	down.Close()

	for _, tc := range []struct {
		name, mirror, method, how string
	}{
		{"through the mirror", m.URL, http.MethodGet, cacheMiss},
		{"kept by the mirror", m.URL, http.MethodGet, cacheHit},
		{"no mirror", "", http.MethodGet, ""},
		{"mirror failing", broken.URL, http.MethodGet, ""},
		{"mirror down", down.URL, http.MethodGet, ""},
		{"not a GET", m.URL, http.MethodPost, ""},
	} {
		client := &http.Client{Transport: &Transport{Mirror: tc.mirror}}
		req, err := http.NewRequestWithContext(t.Context(), tc.method, o.URL+"/f", nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		body, err := io.ReadAll(resp.Body)
		closeBody(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != "from origin" || resp.Header.Get("X-Cache") != tc.how {
			t.Errorf("%s: %q, X-Cache %q; want %q", tc.name, body, resp.Header.Get("X-Cache"), tc.how)
		}
	}
}
