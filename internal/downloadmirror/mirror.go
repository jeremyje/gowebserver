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

// Package downloadmirror is a cache for the build's downloads, and the way
// clients reach it.
//
// A mirror serves any public http or https URL under a path naming it:
// https://github.com/a/b is fetched from <mirror>/https/github.com/a/b.
// The first request downloads the file and keeps it on disk; later ones
// are served from there for as long as the response's own caching headers
// say it's fresh, and after that only cost a conditional request to check
// it hasn't changed. When the original server can't be reached, what's
// kept is served anyway. Each client (curl in the Makefiles, Go's module
// proxy, npm, Playwright, the ffmpeg download) has a way to put a prefix
// in front of the URLs it fetches, which is all that's needed: see URL.
//
// Run one on each CI machine (cmd: internal/tools/downloadcache) and set
// DOWNLOAD_MIRROR to its address in the machine's environment; when it's
// unset, or the mirror fails, the build downloads directly.
package downloadmirror

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Env is the environment variable naming the mirror, e.g.
// http://localhost:3142.
const Env = "DOWNLOAD_MIRROR"

// URL returns where the mirror at mirror serves u, an http or https URL.
// With no mirror, or a URL it doesn't serve, it returns u.
func URL(mirror, u string) string {
	if mirror == "" {
		return u
	}
	scheme, rest, ok := strings.Cut(u, "://")
	if !ok || (scheme != "http" && scheme != "https") {
		return u
	}
	return strings.TrimSuffix(mirror, "/") + "/" + scheme + "/" + rest
}

// Upstream is the inverse of URL: the original URL of a request path on
// the mirror ("/https/github.com/a/b") and its query.
func Upstream(p, rawQuery string) (*url.URL, error) {
	scheme, rest, ok := strings.Cut(strings.TrimPrefix(p, "/"), "/")
	if !ok || (scheme != "http" && scheme != "https") {
		return nil, errors.New("path must be /http/HOST/PATH or /https/HOST/PATH")
	}
	u, err := url.Parse(scheme + "://" + rest)
	if err != nil {
		return nil, err
	}
	if u.Host == "" || u.User != nil {
		return nil, errors.New("the URL needs a host and no user")
	}
	u.RawQuery = rawQuery
	return u, nil
}

// Transport sends requests through a mirror, and straight to the original
// server when the mirror can't be reached or fails (a 5xx status), so a
// mirror that's down only costs time. Requests the mirror doesn't serve
// (other schemes, other methods, or with credentials) go straight there.
type Transport struct {
	// Mirror is the mirror's address; empty sends everything straight.
	Mirror string
	// Base makes the requests; nil uses http.DefaultTransport.
	Base http.RoundTripper
}

// FromEnv returns a Transport through the mirror named by Env, if any,
// over base.
func FromEnv(base http.RoundTripper) *Transport {
	return &Transport{Mirror: os.Getenv(Env), Base: base}
}

// RoundTrip implements http.RoundTripper.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	if t.Mirror == "" || (req.Method != http.MethodGet && req.Method != http.MethodHead) ||
		req.Header.Get("Authorization") != "" || req.URL.User != nil {
		return base.RoundTrip(req)
	}
	mirrored := URL(t.Mirror, req.URL.String())
	if mirrored == req.URL.String() {
		return base.RoundTrip(req)
	}
	u, err := url.Parse(mirrored)
	if err != nil {
		return base.RoundTrip(req)
	}
	mreq := req.Clone(req.Context())
	mreq.URL, mreq.Host = u, ""
	resp, err := base.RoundTrip(mreq)
	switch {
	case err == nil && resp.StatusCode < 500:
		return resp, nil
	case err == nil:
		if _, err := io.Copy(io.Discard, resp.Body); err != nil {
			slog.Debug("reading a failed response from the download mirror", "error", err)
		}
		if err := resp.Body.Close(); err != nil {
			slog.Debug("closing a failed response from the download mirror", "error", err)
		}
	case req.Context().Err() != nil:
		return nil, err
	}
	return base.RoundTrip(req)
}
