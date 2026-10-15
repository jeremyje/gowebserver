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

// Command downloadcache is a cache for the build's downloads, run on each
// CI machine (see package internal/downloadmirror). Set DOWNLOAD_MIRROR to
// its address (by default http://localhost:3142) in the machine's
// environment and the Makefiles send their downloads through it.
//
//	go install github.com/cloudfra/gowebserver/internal/tools/downloadcache@latest
//	downloadcache -dir /var/cache/downloadcache -max-gb 50
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudfra/gowebserver/internal/downloadmirror"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:3142", "Address to listen on. Anything but loopback lets other machines fetch any URL through it; use -hosts then.")
	dir := flag.String("dir", "", "Where downloads are kept. Empty means <user cache dir>/downloadcache.")
	maxGB := flag.Float64("max-gb", 50, "Most gigabytes to keep; the least recently used go first. 0 keeps everything.")
	hosts := flag.String("hosts", "", "Comma-separated hosts it may fetch from. Empty allows any.")
	flag.Parse()
	if err := run(*addr, *dir, *maxGB, *hosts); err != nil {
		slog.Error("downloadcache", "error", err)
		os.Exit(1)
	}
}

func run(addr, dir string, maxGB float64, hosts string) error {
	if dir == "" {
		base, err := os.UserCacheDir()
		if err != nil {
			return err
		}
		dir = filepath.Join(base, "downloadcache")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	c := &downloadmirror.Cache{
		Dir: dir,
		Client: &http.Client{Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           (&net.Dialer{Timeout: 30 * time.Second}).DialContext,
			TLSHandshakeTimeout:   30 * time.Second,
			ResponseHeaderTimeout: 2 * time.Minute,
			MaxIdleConnsPerHost:   8,
		}},
		MaxBytes: int64(maxGB * (1 << 30)),
	}
	if hosts != "" {
		c.Hosts = strings.Split(hosts, ",")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	go func() {
		for {
			if n, err := c.Trim(); err != nil {
				slog.Warn("trimming the cache", "error", err)
			} else if n > 0 {
				slog.Info("trimmed the cache", "bytes", n)
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Hour):
			}
		}
	}()

	srv := &http.Server{Addr: addr, Handler: c, ReadHeaderTimeout: 30 * time.Second}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdown); err != nil {
			slog.Warn("shutting down", "error", err)
		}
	}()
	slog.Info("serving downloads", "addr", addr, "dir", dir, "max GB", maxGB)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
