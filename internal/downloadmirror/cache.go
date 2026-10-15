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
	"cmp"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

// The X-Cache header of each response says where it came from.
const (
	cacheHit         = "HIT"         // kept, and fresh
	cacheMiss        = "MISS"        // downloaded now
	cacheRevalidated = "REVALIDATED" // kept, and the server said it's unchanged
	cacheStale       = "STALE"       // kept, and the server couldn't be asked
)

// maxHeuristicFresh bounds how long a response with no caching headers but
// a Last-Modified date is taken to be fresh: a tenth of its age, as HTTP
// caches do, but no more than this.
const maxHeuristicFresh = 24 * time.Hour

// Cache is a mirror (see the package doc): an http.Handler serving
// GET /SCHEME/HOST/PATH from a copy on disk of SCHEME://HOST/PATH.
//
// Only successful (200) responses are kept, and nothing whose server says
// not to store it. Requests carry no credentials upstream, so only public
// files are cached. Each file is kept under a hash of its URL and the
// request's Accept header (npm, for one, answers differently by Accept).
type Cache struct {
	// Dir holds the files.
	Dir string
	// Client fetches from the original servers, following redirects; a
	// file reached through a redirect is kept under the URL asked for.
	Client *http.Client
	// Hosts, when set, are the only hosts fetched from.
	Hosts []string
	// MaxBytes, when set, is how much Dir may hold; Trim removes the
	// files used least recently to keep under it.
	MaxBytes int64

	now func() time.Time // time.Now, or a test's clock

	mu       sync.Mutex
	fetching map[string]*fetch // by key
}

// fetch is a download in progress, which other requests for the same file
// wait for.
type fetch struct {
	done chan struct{}
	err  error
}

// entry describes a kept file, stored beside it as JSON.
type entry struct {
	// File is the kept file's name in its directory: each download gets a
	// new one, since Windows can't replace a file being read.
	File         string    `json:"file"`
	URL          string    `json:"url"`
	Accept       string    `json:"accept,omitempty"`
	Fetched      time.Time `json:"fetched"` // when the server last confirmed it
	Fresh        Duration  `json:"fresh"`   // how long after Fetched it's fresh
	ContentType  string    `json:"contentType,omitempty"`
	ETag         string    `json:"etag,omitempty"`
	LastModified string    `json:"lastModified,omitempty"`
}

// Duration is a time.Duration that's a string in JSON.
type Duration time.Duration

// MarshalJSON implements json.Marshaler.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	v, err := time.ParseDuration(s)
	*d = Duration(v)
	return err
}

func (c *Cache) clock() time.Time {
	if c.now != nil {
		return c.now()
	}
	return time.Now()
}

// HealthPath answers 200 while the mirror is up, for clients that can't
// fall back to the original server by themselves to check first.
const HealthPath = "/healthz"

// ServeHTTP implements http.Handler.
func (c *Cache) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == HealthPath {
		_, err := io.WriteString(w, "ok\n")
		if err != nil {
			slog.Debug("answering a health check", "error", err)
		}
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "only GET and HEAD", http.StatusMethodNotAllowed)
		return
	}
	u, err := Upstream(r.URL.EscapedPath(), r.URL.RawQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(c.Hosts) > 0 && !slices.Contains(c.Hosts, u.Hostname()) {
		http.Error(w, "host not allowed: "+u.Hostname(), http.StatusForbidden)
		return
	}
	accept := r.Header.Get("Accept")
	key := c.key(u.String(), accept)

	e, kept := c.load(key)
	how := cacheHit
	if !kept || !c.clock().Before(e.Fetched.Add(time.Duration(e.Fresh))) {
		var pass *http.Response
		how, pass, err = c.refresh(r.Context(), key, u.String(), accept, r.Header.Get("User-Agent"))
		if pass != nil {
			// Not kept: passed on as it is.
			defer closeBody(pass.Body)
			copyHeaders(w.Header(), pass.Header, "Content-Type", "Content-Length", "Location", "Retry-After")
			w.Header().Set("X-Cache", cacheMiss)
			w.WriteHeader(pass.StatusCode)
			if _, err := io.Copy(w, pass.Body); err != nil {
				slog.Debug("passing on a response", "url", u.Redacted(), "error", err)
			}
			return
		}
		if err != nil {
			slog.Warn("download failed", "url", u.Redacted(), "error", err)
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		if e, kept = c.load(key); !kept {
			http.Error(w, "download vanished", http.StatusBadGateway)
			return
		}
	}
	body := filepath.Join(c.dir(key), e.File)
	f, err := os.Open(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer closeBody(f)
	// Used now, for Trim.
	now := c.clock()
	if err := os.Chtimes(body, now, now); err != nil {
		slog.Debug("marking a download used", "error", err)
	}
	h := w.Header()
	h.Set("X-Cache", how)
	if e.ContentType != "" {
		h.Set("Content-Type", e.ContentType)
	} else {
		h["Content-Type"] = nil // don't let ServeContent guess
	}
	if e.ETag != "" {
		h.Set("ETag", e.ETag)
	}
	var mod time.Time
	if t, err := http.ParseTime(e.LastModified); err == nil {
		mod = t
	}
	http.ServeContent(w, r, "", mod, f)
}

// refresh downloads the file at key, or checks the kept one is current,
// sharing the work with any other request for it. how is one of the
// X-Cache values. A response that isn't kept is returned as pass instead,
// for the caller to pass on and close.
func (c *Cache) refresh(ctx context.Context, key, u, accept, agent string) (how string, pass *http.Response, err error) {
	c.mu.Lock()
	if c.fetching == nil {
		c.fetching = map[string]*fetch{}
	}
	if f := c.fetching[key]; f != nil {
		c.mu.Unlock()
		select {
		case <-f.done:
		case <-ctx.Done():
			return "", nil, ctx.Err()
		}
		if f.err != nil {
			return "", nil, f.err
		}
		if _, ok := c.load(key); ok {
			return cacheHit, nil, nil
		}
		// What it fetched wasn't kept; fetch again for this request.
		return c.refresh(ctx, key, u, accept, agent)
	}
	f := &fetch{done: make(chan struct{})}
	c.fetching[key] = f
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		delete(c.fetching, key)
		c.mu.Unlock()
		close(f.done)
	}()
	// Finish the download even if the request that started it goes away:
	// others may be waiting, and it's wanted next time anyway.
	how, pass, err = c.download(context.WithoutCancel(ctx), key, u, accept, agent)
	f.err = err
	return how, pass, err
}

// download fetches u, conditionally when a copy is kept, and keeps it.
func (c *Cache) download(ctx context.Context, key, u, accept, agent string) (string, *http.Response, error) {
	old, kept := c.load(key)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", nil, err
	}
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	if agent != "" {
		req.Header.Set("User-Agent", agent)
	}
	if kept {
		if old.ETag != "" {
			req.Header.Set("If-None-Match", old.ETag)
		}
		if old.LastModified != "" {
			req.Header.Set("If-Modified-Since", old.LastModified)
		}
	}
	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		if kept {
			slog.Warn("serving a kept download; its server can't be reached", "url", u, "error", err)
			return cacheStale, nil, nil
		}
		return "", nil, err
	}
	now := c.clock()
	switch {
	case resp.StatusCode == http.StatusNotModified && kept:
		closeBody(resp.Body)
		old.Fetched = now
		if fresh, _ := freshness(resp.Header, now); fresh > 0 {
			old.Fresh = Duration(fresh)
		}
		return cacheRevalidated, nil, c.save(key, old)
	case resp.StatusCode >= 500 && kept:
		closeBody(resp.Body)
		slog.Warn("serving a kept download; its server failed", "url", u, "status", resp.Status)
		return cacheStale, nil, nil
	case resp.StatusCode != http.StatusOK:
		return "", resp, nil
	}
	fresh, store := freshness(resp.Header, now)
	if !store {
		return "", resp, nil
	}
	defer closeBody(resp.Body)
	e := entry{
		URL:          u,
		Accept:       accept,
		Fetched:      now,
		Fresh:        Duration(fresh),
		ContentType:  resp.Header.Get("Content-Type"),
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
	}
	if err := c.write(key, &e, resp); err != nil {
		return "", nil, fmt.Errorf("keeping %s: %w", u, err)
	}
	slog.Info("downloaded", "url", u, "bytes", resp.ContentLength)
	return cacheMiss, nil, nil
}

// write keeps resp's body at key as a new file, with e describing it,
// and removes the one kept before if it can (Trim removes it later if it
// was being read).
func (c *Cache) write(key string, e *entry, resp *http.Response) error {
	dir := c.dir(key)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".download-*")
	if err != nil {
		return err
	}
	n, err := io.Copy(tmp, resp.Body)
	if err == nil && resp.ContentLength >= 0 && n != resp.ContentLength {
		err = fmt.Errorf("got %d bytes of %d", n, resp.ContentLength)
	}
	err = errors.Join(err, tmp.Close())
	e.File = key + "-" + strconv.FormatInt(e.Fetched.UnixNano(), 36)
	if err == nil {
		err = os.Rename(tmp.Name(), filepath.Join(dir, e.File))
	}
	if err != nil {
		return errors.Join(err, os.Remove(tmp.Name()))
	}
	old, kept := c.load(key)
	if err := c.save(key, *e); err != nil {
		return err
	}
	if kept && old.File != e.File {
		if err := os.Remove(filepath.Join(dir, old.File)); err != nil {
			slog.Debug("removing a replaced download", "error", err)
		}
	}
	return nil
}

// freshness returns how long a response with header h, received at now,
// may be served without asking its server again, and whether it may be
// kept at all.
func freshness(h http.Header, now time.Time) (time.Duration, bool) {
	maxAge, sMaxAge := -1, -1
	for _, d := range strings.Split(strings.Join(h.Values("Cache-Control"), ","), ",") {
		name, value, _ := strings.Cut(strings.TrimSpace(d), "=")
		switch strings.ToLower(name) {
		case "no-store":
			return 0, false
		case "no-cache":
			return 0, true
		case "max-age":
			if n, err := strconv.Atoi(strings.Trim(value, `"`)); err == nil {
				maxAge = n
			}
		case "s-maxage":
			if n, err := strconv.Atoi(strings.Trim(value, `"`)); err == nil {
				sMaxAge = n
			}
		}
	}
	age := time.Duration(0)
	if n, err := strconv.Atoi(h.Get("Age")); err == nil && n > 0 {
		age = time.Duration(n) * time.Second
	}
	switch {
	case sMaxAge >= 0:
		return max(0, time.Duration(sMaxAge)*time.Second-age), true
	case maxAge >= 0:
		return max(0, time.Duration(maxAge)*time.Second-age), true
	}
	date := now
	if t, err := http.ParseTime(h.Get("Date")); err == nil {
		date = t
	}
	if t, err := http.ParseTime(h.Get("Expires")); err == nil {
		return max(0, t.Sub(date)), true
	}
	if t, err := http.ParseTime(h.Get("Last-Modified")); err == nil && t.Before(date) {
		return min(maxHeuristicFresh, date.Sub(t)/10), true
	}
	return 0, true
}

// key names the file kept for u fetched with Accept header accept.
func (c *Cache) key(u, accept string) string {
	sum := sha256.Sum256([]byte(u + "\n" + accept))
	return hex.EncodeToString(sum[:])
}

// dir is where the files for key are kept: a directory by its first two
// characters, so no directory gets too large.
func (c *Cache) dir(key string) string {
	return filepath.Join(c.Dir, key[:2])
}

// meta is the file describing what's kept for key.
func (c *Cache) meta(key string) string {
	return filepath.Join(c.dir(key), key+".json")
}

// load returns what's kept at key, if anything.
func (c *Cache) load(key string) (entry, bool) {
	data, err := os.ReadFile(c.meta(key))
	if err != nil {
		return entry{}, false
	}
	var e entry
	if err := json.Unmarshal(data, &e); err != nil {
		return entry{}, false
	}
	if e.File == "" || filepath.Base(e.File) != e.File {
		return entry{}, false
	}
	if _, err := os.Stat(filepath.Join(c.dir(key), e.File)); err != nil {
		return entry{}, false
	}
	return e, true
}

// save writes e for key, whole or not at all.
func (c *Cache) save(key string, e entry) error {
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(c.meta(key)), ".meta-*")
	if err != nil {
		return err
	}
	_, err = tmp.Write(data)
	err = errors.Join(err, tmp.Close())
	if err == nil {
		err = replace(tmp.Name(), c.meta(key))
	}
	if err != nil {
		return errors.Join(err, os.Remove(tmp.Name()))
	}
	return nil
}

// replace renames src over dst. On Windows that fails while dst is open,
// which for a description is only for a moment, so it's tried a few
// times.
func replace(src, dst string) error {
	var err error
	for i := range 20 {
		if err = os.Rename(src, dst); err == nil {
			return nil
		}
		time.Sleep(time.Duration(i+1) * 5 * time.Millisecond)
	}
	return err
}

// Trim removes the files used least recently until Dir holds no more than
// MaxBytes (when it's set), and files nothing uses: copies replaced by a
// newer download, and ones left half-written. It returns how many bytes
// it removed.
func (c *Cache) Trim() (int64, error) {
	type file struct {
		key, path string
		size      int64
		used      time.Time
	}
	var (
		files          []file
		total, removed int64
		current        = map[string]string{} // key: its current file
	)
	err := filepath.WalkDir(c.Dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if key, ok := strings.CutSuffix(d.Name(), ".json"); ok && !strings.HasPrefix(key, ".") {
			if e, ok := c.load(key); ok {
				current[key] = e.File
			}
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		key, _, _ := strings.Cut(d.Name(), "-")
		files = append(files, file{key, p, info.Size(), info.ModTime()})
		return nil
	})
	if err != nil {
		return 0, err
	}
	var kept []file
	for _, f := range files {
		switch {
		case current[f.key] == filepath.Base(f.path):
			kept = append(kept, f)
			total += f.size
		case c.clock().Sub(f.used) > time.Hour:
			// Replaced, or half-written, an hour ago: nothing's reading it.
			if err := os.Remove(f.path); err != nil {
				slog.Debug("removing an unused download", "error", err)
				continue
			}
			removed += f.size
		}
	}
	if c.MaxBytes <= 0 || total <= c.MaxBytes {
		return removed, nil
	}
	slices.SortFunc(kept, func(a, b file) int { return cmp.Compare(a.used.UnixNano(), b.used.UnixNano()) })
	for _, f := range kept {
		if total <= c.MaxBytes {
			break
		}
		if err := os.Remove(c.meta(f.key)); err != nil {
			return removed, err
		}
		if err := os.Remove(f.path); err != nil {
			// Being read; it goes on a later Trim.
			slog.Debug("removing a download", "error", err)
		}
		total -= f.size
		removed += f.size
	}
	return removed, nil
}

func copyHeaders(dst, src http.Header, names ...string) {
	for _, n := range names {
		if v := src.Values(n); len(v) > 0 {
			dst[http.CanonicalHeaderKey(n)] = v
		}
	}
}

func closeBody(c io.Closer) {
	if err := c.Close(); err != nil {
		slog.Debug("closing a download", "error", err)
	}
}
