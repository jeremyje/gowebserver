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
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// origin is an upstream server whose handler a test sets, counting the
// requests it gets.
type origin struct {
	*httptest.Server
	hits    atomic.Int32
	mu      sync.Mutex
	handler http.HandlerFunc
}

func newOrigin(t *testing.T, h http.HandlerFunc) *origin {
	o := &origin{handler: h}
	o.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		o.hits.Add(1)
		o.mu.Lock()
		h := o.handler
		o.mu.Unlock()
		h(w, r)
	}))
	t.Cleanup(o.Close)
	return o
}

func (o *origin) set(h http.HandlerFunc) {
	o.mu.Lock()
	o.handler = h
	o.mu.Unlock()
}

// clock is a test's time, which it moves.
type clock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *clock) advance(d time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(d)
	c.mu.Unlock()
}

func newCache(t *testing.T) (*Cache, *clock, *httptest.Server) {
	clk := &clock{now: time.Now()}
	c := &Cache{Dir: t.TempDir(), now: clk.Now}
	srv := httptest.NewServer(c)
	t.Cleanup(srv.Close)
	return c, clk, srv
}

// get fetches u (an origin URL) through the mirror at mirror, returning
// the status, body and X-Cache header.
func get(t *testing.T, mirror, u string, header http.Header) (int, string, string) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, URL(mirror, u), nil)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range header {
		req.Header[k] = v
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer closeBody(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, string(body), resp.Header.Get("X-Cache")
}

func expect(t *testing.T, what string, status int, body, how string, wantStatus int, wantBody, wantHow string) {
	t.Helper()
	if status != wantStatus || body != wantBody || how != wantHow {
		t.Errorf("%s: %d %q %s, want %d %q %s", what, status, body, how, wantStatus, wantBody, wantHow)
	}
}

func TestURL(t *testing.T) {
	for _, tc := range []struct{ mirror, u, want string }{
		{"", "https://github.com/a/b", "https://github.com/a/b"},
		{"http://localhost:3142", "https://github.com/a/b?x=1", "http://localhost:3142/https/github.com/a/b?x=1"},
		{"http://localhost:3142/", "http://example.com/", "http://localhost:3142/http/example.com/"},
		{"http://localhost:3142", "ftp://example.com/x", "ftp://example.com/x"},
	} {
		if got := URL(tc.mirror, tc.u); got != tc.want {
			t.Errorf("URL(%q, %q) = %q, want %q", tc.mirror, tc.u, got, tc.want)
		}
	}
	for p, want := range map[string]string{
		"/https/github.com/a/b": "https://github.com/a/b?q",
		"/http/host:8080/x%20y": "http://host:8080/x%20y?q",
		"/ftp/host/x":           "",
		"/https":                "",
		"/https/user@host/x":    "",
	} {
		u, err := Upstream(p, "q")
		switch {
		case want == "" && err == nil:
			t.Errorf("Upstream(%q) = %v, want an error", p, u)
		case want != "" && (err != nil || u.String() != want):
			t.Errorf("Upstream(%q) = %v, %v; want %q", p, u, err, want)
		}
	}
}

func TestCacheKeepsFreshFiles(t *testing.T) {
	o := newOrigin(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=60")
		w.Header().Set("Content-Type", "application/zip")
		reply(w, "v1")
	})
	_, clk, m := newCache(t)
	s, b, how := get(t, m.URL, o.URL+"/tool.zip", nil)
	expect(t, "first", s, b, how, 200, "v1", cacheMiss)
	s, b, how = get(t, m.URL, o.URL+"/tool.zip", nil)
	expect(t, "again", s, b, how, 200, "v1", cacheHit)
	if n := o.hits.Load(); n != 1 {
		t.Errorf("origin hit %d times while fresh, want 1", n)
	}
	// Stale, and changed: fetched again.
	clk.advance(2 * time.Minute)
	o.set(func(w http.ResponseWriter, _ *http.Request) { reply(w, "v2") })
	s, b, how = get(t, m.URL, o.URL+"/tool.zip", nil)
	expect(t, "stale", s, b, how, 200, "v2", cacheMiss)
	// A query is part of the URL.
	s, b, how = get(t, m.URL, o.URL+"/tool.zip?arch=arm64", nil)
	expect(t, "other query", s, b, how, 200, "v2", cacheMiss)
}

func TestCacheRevalidates(t *testing.T) {
	o := newOrigin(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", `"abc"`)
		w.Header().Set("Cache-Control", "no-cache")
		if r.Header.Get("If-None-Match") == `"abc"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		reply(w, "body")
	})
	_, _, m := newCache(t)
	s, b, how := get(t, m.URL, o.URL+"/latest/checksums", nil)
	expect(t, "first", s, b, how, 200, "body", cacheMiss)
	s, b, how = get(t, m.URL, o.URL+"/latest/checksums", nil)
	expect(t, "again", s, b, how, 200, "body", cacheRevalidated)
	if n := o.hits.Load(); n != 2 {
		t.Errorf("origin hit %d times, want 2 (no-cache asks every time)", n)
	}
}

func TestCacheServesStaleWhenOriginFails(t *testing.T) {
	o := newOrigin(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Cache-Control", "max-age=1")
		reply(w, "kept")
	})
	_, clk, m := newCache(t)
	get(t, m.URL, o.URL+"/f", nil)
	clk.advance(time.Minute)
	o.set(func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "down", http.StatusServiceUnavailable) })
	s, b, how := get(t, m.URL, o.URL+"/f", nil)
	expect(t, "origin failing", s, b, how, 200, "kept", cacheStale)
	o.Close()
	s, b, how = get(t, m.URL, o.URL+"/f", nil)
	expect(t, "origin gone", s, b, how, 200, "kept", cacheStale)
	// Nothing kept: the failure.
	if s, _, _ := get(t, m.URL, o.URL+"/other", nil); s != http.StatusBadGateway {
		t.Errorf("unreachable and not kept: status %d, want 502", s)
	}
}

func TestCacheKeepsOnlySuccess(t *testing.T) {
	o := newOrigin(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/missing":
			http.NotFound(w, r)
		case "/private":
			w.Header().Set("Cache-Control", "no-store")
			reply(w, "secret")
		}
	})
	_, _, m := newCache(t)
	for range 2 {
		if s, _, how := get(t, m.URL, o.URL+"/missing", nil); s != http.StatusNotFound || how != cacheMiss {
			t.Errorf("missing: %d %s, want 404 passed on", s, how)
		}
		s, b, how := get(t, m.URL, o.URL+"/private", nil)
		expect(t, "no-store", s, b, how, 200, "secret", cacheMiss)
	}
	if n := o.hits.Load(); n != 4 {
		t.Errorf("origin hit %d times, want 4: nothing kept", n)
	}
}

func TestCacheFollowsRedirects(t *testing.T) {
	var o *origin
	o = newOrigin(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/releases/download/v1/tool" {
			http.Redirect(w, r, o.URL+"/signed/abc", http.StatusFound)
			return
		}
		w.Header().Set("Last-Modified", time.Now().Add(-100*time.Hour).UTC().Format(http.TimeFormat))
		reply(w, "binary")
	})
	_, _, m := newCache(t)
	s, b, how := get(t, m.URL, o.URL+"/releases/download/v1/tool", nil)
	expect(t, "first", s, b, how, 200, "binary", cacheMiss)
	// Fresh for a tenth of its age (10 hours), by heuristic.
	s, b, how = get(t, m.URL, o.URL+"/releases/download/v1/tool", nil)
	expect(t, "again", s, b, how, 200, "binary", cacheHit)
}

func TestCacheKeysByAccept(t *testing.T) {
	o := newOrigin(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=300")
		reply(w, "accept="+r.Header.Get("Accept"))
	})
	_, _, m := newCache(t)
	_, short, _ := get(t, m.URL, o.URL+"/pkg", http.Header{"Accept": {"application/vnd.npm.install-v1+json"}})
	_, full, _ := get(t, m.URL, o.URL+"/pkg", http.Header{"Accept": {"application/json"}})
	if short == full {
		t.Errorf("both Accept headers got %q", short)
	}
}

func TestCacheOneDownloadAtATime(t *testing.T) {
	release := make(chan struct{})
	o := newOrigin(t, func(w http.ResponseWriter, _ *http.Request) {
		<-release
		w.Header().Set("Cache-Control", "max-age=60")
		reply(w, "big")
	})
	_, _, m := newCache(t)
	var wg sync.WaitGroup
	for range 5 {
		wg.Go(func() {
			if _, b, _ := get(t, m.URL, o.URL+"/big", nil); b != "big" {
				t.Errorf("got %q", b)
			}
		})
	}
	time.Sleep(50 * time.Millisecond)
	close(release)
	wg.Wait()
	if n := o.hits.Load(); n != 1 {
		t.Errorf("origin hit %d times by 5 requests at once, want 1", n)
	}
}

func TestCacheRangeAndHead(t *testing.T) {
	o := newOrigin(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Cache-Control", "max-age=60")
		reply(w, "0123456789")
	})
	_, _, m := newCache(t)
	s, b, _ := get(t, m.URL, o.URL+"/f", http.Header{"Range": {"bytes=2-4"}})
	if s != http.StatusPartialContent || b != "234" {
		t.Errorf("range: %d %q, want 206 \"234\"", s, b)
	}
	resp, err := http.Head(URL(m.URL, o.URL+"/f"))
	if err != nil {
		t.Fatal(err)
	}
	closeBody(resp.Body)
	if resp.StatusCode != 200 || resp.ContentLength != 10 {
		t.Errorf("HEAD: %d, length %d", resp.StatusCode, resp.ContentLength)
	}
}

func TestCacheHealth(t *testing.T) {
	_, _, m := newCache(t)
	resp, err := http.Get(m.URL + HealthPath)
	if err != nil {
		t.Fatal(err)
	}
	closeBody(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("health check: %d, want 200", resp.StatusCode)
	}
}

func TestCacheRejects(t *testing.T) {
	c, _, m := newCache(t)
	c.Hosts = []string{"allowed.example"}
	if s, _, _ := get(t, m.URL, "https://other.example/x", nil); s != http.StatusForbidden {
		t.Errorf("other host: %d, want 403", s)
	}
	resp, err := http.Get(m.URL + "/ftp/host/x")
	if err != nil {
		t.Fatal(err)
	}
	closeBody(resp.Body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("bad path: %d, want 400", resp.StatusCode)
	}
	resp, err = http.Post(URL(m.URL, "https://allowed.example/x"), "text/plain", strings.NewReader("x"))
	if err != nil {
		t.Fatal(err)
	}
	closeBody(resp.Body)
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("POST: %d, want 405", resp.StatusCode)
	}
}

func TestCacheTrim(t *testing.T) {
	o := newOrigin(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Cache-Control", "max-age=0")
		reply(w, strings.Repeat("x", 100))
	})
	c, clk, m := newCache(t)
	for _, f := range []string{"/a", "/b", "/c"} {
		get(t, m.URL, o.URL+f, nil)
		clk.advance(time.Minute)
	}
	// Each is downloaded again, leaving the old copies behind.
	for _, f := range []string{"/a", "/b", "/c"} {
		get(t, m.URL, o.URL+f, nil)
	}
	count := func() int {
		n := 0
		err := filepath.WalkDir(c.Dir, func(p string, d os.DirEntry, err error) error {
			if err == nil && !d.IsDir() && !strings.HasSuffix(p, ".json") {
				n++
			}
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
		return n
	}
	clk.advance(2 * time.Hour)
	c.MaxBytes = 250
	if _, err := c.Trim(); err != nil {
		t.Fatal(err)
	}
	if n := count(); n != 2 {
		t.Errorf("%d files after trimming to 250 bytes, want 2", n)
	}
	// What's left is still served.
	served := 0
	for _, f := range []string{"/a", "/b", "/c"} {
		if _, b, _ := get(t, m.URL, o.URL+f, nil); len(b) == 100 {
			served++
		}
	}
	if served != 3 {
		t.Errorf("served %d of 3 after trimming", served)
	}
}

func TestFreshness(t *testing.T) {
	now := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		h     http.Header
		fresh time.Duration
		store bool
	}{
		{http.Header{"Cache-Control": {"public, max-age=300"}}, 5 * time.Minute, true},
		{http.Header{"Cache-Control": {"max-age=300"}, "Age": {"100"}}, 200 * time.Second, true},
		{http.Header{"Cache-Control": {"max-age=10, s-maxage=60"}}, time.Minute, true},
		{http.Header{"Cache-Control": {"no-cache"}}, 0, true},
		{http.Header{"Cache-Control": {"private, no-store"}}, 0, false},
		{http.Header{"Expires": {now.Add(time.Hour).Format(http.TimeFormat)}, "Date": {now.Format(http.TimeFormat)}}, time.Hour, true},
		{http.Header{"Last-Modified": {now.Add(-10 * time.Hour).Format(http.TimeFormat)}}, time.Hour, true},
		{http.Header{"Last-Modified": {now.Add(-1000 * time.Hour).Format(http.TimeFormat)}}, maxHeuristicFresh, true},
		{http.Header{}, 0, true},
	} {
		fresh, store := freshness(tc.h, now)
		if fresh != tc.fresh || store != tc.store {
			t.Errorf("freshness(%v) = %v, %v; want %v, %v", tc.h, fresh, store, tc.fresh, tc.store)
		}
	}
}

// reply writes s as an origin's response.
func reply(w io.Writer, s string) {
	if _, err := io.WriteString(w, s); err != nil {
		slog.Debug("origin reply", "error", err)
	}
}
