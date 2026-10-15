// Copyright 2022 Cloudfra
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
	"context"
	"errors"
	"io/fs"
	"path"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"facette.io/natsort"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Prefetching makes thumbnails before they're asked for. Each folder
// listed sends the thumbnailer through its immediate subfolders, making
// the first few tiles of each one's grid, so opening one shows pictures at
// once rather than waiting for them to be made.
//
// It's low priority: it only starts a thumbnail when no request for one
// has been seen for thumbnailPrefetchQuiet, and uses at most half the
// render slots. Listing another folder throws the rest away, and stops
// the thumbnails being made for it unless a request is waiting for them
// too. One made ahead is shared with a request that comes for it while
// it's being made, like any other.
const (
	// thumbnailPrefetchDirs is how many of a folder's subfolders are
	// visited, the first in the order the list shows them.
	thumbnailPrefetchDirs = 32
	// thumbnailPrefetchQuiet is how long thumbnail requests have to have
	// stopped before one is made ahead. The page asks for its tiles in
	// batches, with short gaps between them that aren't idle.
	thumbnailPrefetchQuiet = 500 * time.Millisecond
)

// The results of making a thumbnail ahead, the "result" of
// thumbnail_prefetches_total.
const (
	prefetchMade        = "made"
	prefetchCached      = "cached" // it was in the store already
	prefetchUnsupported = "unsupported"
	prefetchFailed      = "failed"
	prefetchCanceled    = "canceled" // another folder was listed first
)

// prefetcher makes thumbnails ahead (see above). One is shared by every
// served filesystem, so there's one queue, which the latest listing owns.
type prefetcher struct {
	t       *thumbnailer
	perDir  int // thumbnails made in each subfolder
	workers int // made at once
	// How long thumbnail requests have to have stopped:
	// thumbnailPrefetchQuiet, or less in tests.
	quiet time.Duration

	// Requests for thumbnails being served now, and when the last one
	// started or ended (unix nanoseconds), to tell when it's idle.
	active atomic.Int64
	last   atomic.Int64

	mu sync.Mutex
	// The current queue: cancel throws it away, and done closes when it
	// has stopped.
	cancel context.CancelFunc
	done   chan struct{}
	// The sizes the grid's tiles last asked for, by kind of file, which
	// are the ones to make.
	tiles map[string]thumbnailSpec
}

// The tile sizes made until the page has asked for any: what it asks for
// on a desktop screen. Their format is left to chooseWebP.
var defaultPrefetchTiles = map[string]thumbnailSpec{
	thumbnailKind(false): {width: 256, height: 256, fit: fitOutside, quality: thumbnailQuality},
	thumbnailKind(true):  {width: 512, height: 288, fit: fitOutside, quality: thumbnailQuality},
}

// newPrefetcher makes perDir thumbnails ahead in each subfolder, or
// returns nil when perDir is 0.
func newPrefetcher(t *thumbnailer, perDir, slots int) *prefetcher {
	if perDir <= 0 {
		return nil
	}
	return &prefetcher{t: t, perDir: perDir, workers: max(1, slots/2), quiet: thumbnailPrefetchQuiet, tiles: map[string]thumbnailSpec{}}
}

// served marks a thumbnail request as started; the function it returns
// marks it ended. It's safe on a nil prefetcher.
func (p *prefetcher) served() func() {
	if p == nil {
		return func() {}
	}
	p.active.Add(1)
	p.last.Store(time.Now().UnixNano())
	return func() {
		p.last.Store(time.Now().UnixNano())
		p.active.Add(-1)
	}
}

// learn remembers made as the size of a grid tile for kind of file when
// asked is one: covering a box at the file's shape, with nothing else
// asked for. made is what asked is made as, in the format the browser's
// Accept header picked. The slideshow's sizes aren't, and aren't made
// ahead.
func (p *prefetcher) learn(kind string, asked, made *thumbnailSpec) {
	if p == nil || asked.fit != fitOutside || asked.width == 0 || asked.height == 0 ||
		asked.format != "" || asked.background != nil || asked.quality != thumbnailQuality {
		return
	}
	p.mu.Lock()
	p.tiles[kind] = *made
	p.mu.Unlock()
}

// tile returns the size and format to make grid tiles of kind of file in:
// the last remembered, or until there's one, the default size in the
// format a browser that shows WebP gets.
func (p *prefetcher) tile(kind string) *thumbnailSpec {
	p.mu.Lock()
	s, ok := p.tiles[kind]
	p.mu.Unlock()
	if !ok {
		s = defaultPrefetchTiles[kind]
		return p.t.chooseWebP(&s, "image/webp")
	}
	return &s
}

// idle waits until no thumbnail has been asked for in p.quiet. It
// returns ctx's error if ctx ends first.
func (p *prefetcher) idle(ctx context.Context) error {
	for {
		wait := p.quiet
		if p.active.Load() == 0 {
			wait -= time.Since(time.Unix(0, p.last.Load()))
			if wait <= 0 {
				return nil
			}
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

// listed replaces the queue with the subfolders (names in dir, in the
// order the list shows them) of the folder dir that h's filesystem just
// listed. It's safe on a nil prefetcher.
func (p *prefetcher) listed(h *thumbnailHandler, dir string, subdirs []string) {
	if p == nil {
		return
	}
	// The page that was just listed asks for its own tiles next, so wait
	// for those as if it already had.
	p.last.Store(time.Now().UnixNano())
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	p.mu.Lock()
	if p.cancel != nil {
		p.cancel()
	}
	p.cancel, p.done = cancel, done
	p.mu.Unlock()
	if len(subdirs) > thumbnailPrefetchDirs {
		subdirs = subdirs[:thumbnailPrefetchDirs]
	}
	go func() {
		defer close(done)
		p.run(ctx, h, dir, subdirs)
	}()
}

// stop throws the queue away and waits for it to stop. It's safe on a
// nil prefetcher.
func (p *prefetcher) stop() {
	if p == nil {
		return
	}
	p.mu.Lock()
	cancel, done := p.cancel, p.done
	p.mu.Unlock()
	if cancel != nil {
		cancel()
		<-done
	}
}

// run makes the first thumbnails of each subfolder in turn, up to
// p.workers at once, until they're done or ctx ends.
func (p *prefetcher) run(ctx context.Context, h *thumbnailHandler, dir string, subdirs []string) {
	slots := make(chan struct{}, p.workers)
	var wg sync.WaitGroup
	defer wg.Wait()
	for _, sub := range subdirs {
		if p.idle(ctx) != nil {
			return
		}
		for _, f := range firstTiles(h.fsys, path.Join(dir, sub), p.perDir, p.t.videos()) {
			select {
			case slots <- struct{}{}:
			case <-ctx.Done():
				return
			}
			if p.idle(ctx) != nil {
				<-slots
				return
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() { <-slots }()
				h.prefetch(ctx, f.name, f.video, p.tile(thumbnailKind(f.video)))
			}()
		}
	}
}

type tileFile struct {
	name  string // in the filesystem
	video bool
}

// firstTiles returns the first n files of the grid of the folder dir,
// opened at its default sort: its videos, then its photos, by name.
// Videos are left out when they get no thumbnails.
func firstTiles(fsys fs.FS, dir string, n int, videos bool) []tileFile {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil
	}
	var vids, photos []string
	for _, e := range entries {
		switch name := e.Name(); {
		case e.IsDir():
		case isVideo(name):
			if videos {
				vids = append(vids, name)
			}
		case isImage(name):
			photos = append(photos, name)
		}
	}
	byName := func(a, b string) int {
		a, b = strings.ToLower(a), strings.ToLower(b)
		switch {
		case natsort.Compare(a, b):
			return -1
		case natsort.Compare(b, a):
			return 1
		}
		return 0
	}
	slices.SortFunc(vids, byName)
	slices.SortFunc(photos, byName)
	var tiles []tileFile
	for _, name := range vids {
		tiles = append(tiles, tileFile{name: path.Join(dir, name), video: true})
	}
	for _, name := range photos {
		tiles = append(tiles, tileFile{name: path.Join(dir, name)})
	}
	return tiles[:min(n, len(tiles))]
}

// prefetch makes and stores the thumbnail of the file name for spec,
// unless it's stored already. It's a span, thumbnail.prefetch, and is
// counted by result.
func (h *thumbnailHandler) prefetch(ctx context.Context, name string, video bool, spec *thumbnailSpec) {
	kind := thumbnailKind(video)
	ctx, span := h.t.metrics.tracer.Start(ctx, "thumbnail.prefetch")
	defer span.End()
	span.SetAttributes(attribute.String("path", name), attribute.String("spec", spec.key()), attribute.String("kind", kind))
	result := prefetchCached
	defer func() {
		span.SetAttributes(attribute.String("result", result))
		h.t.metrics.prefetches.Add(context.WithoutCancel(ctx), 1, metric.WithAttributes(kindAttr(kind), attribute.String("result", result)))
	}()

	info, err := fs.Stat(h.fsys, name)
	if err != nil {
		result = prefetchFailed
		return
	}
	key := path.Join(h.source, thumbnailPath(name, info, spec, video))
	if h.t.store.has(key) {
		return
	}
	_, err = h.t.shared(ctx, key, func(ctx context.Context) (*thumbnail, error) {
		return h.make(ctx, name, video, spec)
	})
	switch {
	case err == nil:
		result = prefetchMade
	case ctx.Err() != nil:
		result = prefetchCanceled
	case errors.Is(err, errThumbnailUnsupported):
		result = prefetchUnsupported
	default:
		result = prefetchFailed
	}
}
