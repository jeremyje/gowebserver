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
	"io/fs"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/cloudfra/gowebserver/pkg/ffmpeg"
	"github.com/google/go-cmp/cmp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

func TestFirstTiles(t *testing.T) {
	// Videos are known by the system's MIME types, which Windows may not
	// have for these.
	for ext, typ := range map[string]string{".mp4": "video/mp4", ".mov": "video/quicktime"} {
		if err := mime.AddExtensionType(ext, typ); err != nil {
			t.Fatal(err)
		}
	}
	fsys := fstest.MapFS{
		"album/10.jpg":      {},
		"album/2.JPG":       {},
		"album/1.png":       {},
		"album/b.mp4":       {},
		"album/A.mov":       {},
		"album/notes.txt":   {},
		"album/sub/3.jpg":   {},
		"album/sub.jpg/x":   {}, // a folder named like a photo
		"elsewhere/0.jpg":   {},
		"album/archive.zip": {},
	}
	names := func(tiles []tileFile) []string {
		out := make([]string, 0, len(tiles))
		for _, f := range tiles {
			out = append(out, f.name)
		}
		return out
	}
	// The grid's order: videos, then photos, each by name, numbers in
	// numeric order.
	if diff := cmp.Diff([]string{"album/A.mov", "album/b.mp4", "album/1.png", "album/2.JPG"}, names(firstTiles(fsys, "album", 4, true))); diff != "" {
		t.Errorf("first 4 (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff([]string{"album/1.png", "album/2.JPG", "album/10.jpg"}, names(firstTiles(fsys, "album", 10, false))); diff != "" {
		t.Errorf("without video thumbnails (-want +got):\n%s", diff)
	}
	if got := firstTiles(fsys, "missing", 4, true); got != nil {
		t.Errorf("a missing folder gave %v", got)
	}
}

// prefetchTest serves a folder of files (path to contents) with the file
// browser, making perDir thumbnails ahead in subfolders, with ff, which
// may be nil.
type prefetchTest struct {
	h      http.Handler
	th     *thumbnailHandler
	p      *prefetcher
	reader *sdkmetric.ManualReader
}

func newPrefetchTest(t *testing.T, perDir int, ff *ffmpeg.FFmpeg, files map[string][]byte) *prefetchTest {
	t.Helper()
	dir := t.TempDir()
	for name, data := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	mc, reader, _ := testMonitoring()
	thumbs, err := newThumbnailer(Thumbnails{Enabled: true, CachePath: thumbnailCacheMemory, Prefetch: perDir}, ff, mc)
	if err != nil {
		t.Fatal(err)
	}
	thumbs.prefetch.quiet = 10 * time.Millisecond
	h, closeFS, err := newHandlerFromFS(dir, mc, true, thumbs)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := thumbs.close(); err != nil {
			t.Error(err)
		}
		if err := closeFS(); err != nil {
			t.Error(err)
		}
	})
	return &prefetchTest{h: h, th: h.(*thumbnailHandler), p: thumbs.prefetch, reader: reader}
}

// list asks for the listing of the folder at target, and returns the
// queue it started.
func (pt *prefetchTest) list(t *testing.T, target string) chan struct{} {
	t.Helper()
	if rec := thumbGet(t, pt.h, target, nil); rec.Code != http.StatusOK {
		t.Fatalf("listing %s: status %d", target, rec.Code)
	}
	pt.p.mu.Lock()
	defer pt.p.mu.Unlock()
	return pt.p.done
}

func waitFor(t *testing.T, done chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("thumbnails were still being made ahead after 10s")
	}
}

// stored reports whether the photo name's thumbnail at size x size
// (fit=outside), in the format picked once the image is decoded, is in
// the store.
func (pt *prefetchTest) stored(t *testing.T, name string, size int) bool {
	t.Helper()
	return pt.storedAs(t, name, size, "")
}

// storedAs is stored for a thumbnail in format.
func (pt *prefetchTest) storedAs(t *testing.T, name string, size int, format string) bool {
	t.Helper()
	info, err := fs.Stat(pt.th.fsys, name)
	if err != nil {
		t.Fatal(err)
	}
	spec := &thumbnailSpec{width: size, height: size, fit: fitOutside, format: format, quality: thumbnailQuality}
	return pt.th.t.store.has(path.Join(pt.th.source, thumbnailPath(name, info, spec, false)))
}

func TestPrefetchSubfolders(t *testing.T) {
	photo := encodeJPEG(t, stripes(400, 300))
	pt := newPrefetchTest(t, 2, nil, map[string][]byte{
		"top.jpg":        photo,
		"a/10.jpg":       photo,
		"a/2.jpg":        photo,
		"a/1.jpg":        photo,
		"a/clip.mp4":     []byte("not really a video"), // no ffmpeg, so no thumbnail
		"a/deeper/d.jpg": photo,
		"b/x.png":        encodePNG(t, stripes(40, 30)),
		"b/notes.txt":    []byte("text"),
	})
	waitFor(t, pt.list(t, "/"))

	for name, want := range map[string]bool{
		"a/1.jpg": true, "a/2.jpg": true, "b/x.png": true,
		// Past the first two of its folder.
		"a/10.jpg": false,
		// In the folder listed, whose page asks for them itself.
		"top.jpg": false,
		// Two folders down.
		"a/deeper/d.jpg": false,
	} {
		if got := pt.stored(t, name, 256); got != want {
			t.Errorf("%s stored = %v, want %v", name, got, want)
		}
	}
	// Opening the folder finds its first tiles made.
	rec := thumbGet(t, pt.h, "/a/1.jpg?width=256&height=256&fit=outside", nil)
	decodeThumb(t, rec)
	if got := rec.Header().Get("Server-Timing"); got != "cache;desc=hit" {
		t.Errorf("Server-Timing = %q, want it from the store", got)
	}
	if diff := cmp.Diff(map[string]int64{"kind=photo,result=made": 3}, collected(t, pt.reader)["thumbnail_prefetches_total"]); diff != "" {
		t.Errorf("thumbnail_prefetches_total (-want +got):\n%s", diff)
	}

	// Listed again, they're found in the store and not made again.
	waitFor(t, pt.list(t, "/"))
	if diff := cmp.Diff(map[string]int64{"kind=photo,result=made": 3, "kind=photo,result=cached": 3}, collected(t, pt.reader)["thumbnail_prefetches_total"]); diff != "" {
		t.Errorf("thumbnail_prefetches_total after listing again (-want +got):\n%s", diff)
	}
}

func TestPrefetchTileSize(t *testing.T) {
	photo := encodeJPEG(t, stripes(400, 300))
	pt := newPrefetchTest(t, 1, nil, map[string][]byte{"top.jpg": photo, "a/1.jpg": photo})
	// The grid asks for 128-pixel tiles; the slideshow's previews aren't
	// tiles, and don't change that.
	decodeThumb(t, thumbGet(t, pt.h, "/top.jpg?width=128&height=128&fit=outside", nil))
	decodeThumb(t, thumbGet(t, pt.h, "/top.jpg?width=1024&height=1024&fit=inside", nil))
	waitFor(t, pt.list(t, "/"))
	if !pt.stored(t, "a/1.jpg", 128) {
		t.Error("a/1.jpg wasn't made at the size the grid asked for")
	}
	if pt.stored(t, "a/1.jpg", 256) {
		t.Error("a/1.jpg was made at the default size too")
	}
}

func TestPrefetchWebP(t *testing.T) {
	bin, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg is not installed")
	}
	ff, err := ffmpeg.Find(bin)
	if err != nil {
		t.Fatal(err)
	}
	photo := encodeJPEG(t, stripes(400, 300))
	pt := newPrefetchTest(t, 1, ff, map[string][]byte{"top.jpg": photo, "a/1.jpg": photo, "b/1.jpg": photo})
	if !waitForPhotoTools(t, pt.h).webp {
		t.Skip("ffmpeg has no libwebp encoder")
	}
	// Before the grid has asked for any, tiles are made at the default
	// size as WebP, which is what a browser gets.
	waitFor(t, pt.list(t, "/"))
	if !pt.storedAs(t, "a/1.jpg", 256, "webp") {
		t.Error("a/1.jpg wasn't made as WebP at the default size")
	}
	// A browser that shows WebP asks for 128-pixel tiles; one listing
	// later, they're made as it asked for them, and served from the store.
	decodeThumb(t, thumbGet(t, pt.h, "/top.jpg?width=128&height=128&fit=outside", browserAccept))
	waitFor(t, pt.list(t, "/"))
	rec := thumbGet(t, pt.h, "/b/1.jpg?width=128&height=128&fit=outside", browserAccept)
	decodeThumb(t, rec)
	if got := rec.Header().Get("Content-Type"); got != "image/webp" {
		t.Errorf("Content-Type = %q, want image/webp", got)
	}
	if got := rec.Header().Get("Server-Timing"); got != "cache;desc=hit" {
		t.Errorf("Server-Timing = %q, want it from the store", got)
	}
}

func TestPrefetchWaitsUntilIdle(t *testing.T) {
	photo := encodeJPEG(t, stripes(400, 300))
	pt := newPrefetchTest(t, 1, nil, map[string][]byte{"a/1.jpg": photo})
	// A thumbnail request still being served.
	served := pt.p.served()
	done := pt.list(t, "/")
	time.Sleep(20 * pt.p.quiet)
	if pt.stored(t, "a/1.jpg", 256) {
		t.Fatal("a thumbnail was made ahead while one was being asked for")
	}
	served()
	waitFor(t, done)
	if !pt.stored(t, "a/1.jpg", 256) {
		t.Error("a/1.jpg wasn't made once requests stopped")
	}
}

func TestPrefetchNewListingThrowsQueueAway(t *testing.T) {
	photo := encodeJPEG(t, stripes(400, 300))
	pt := newPrefetchTest(t, 1, nil, map[string][]byte{"a/1.jpg": photo, "b/1.jpg": photo})
	served := pt.p.served()
	first := pt.list(t, "/")
	// Opening a, which has no subfolders, replaces the queue for /.
	second := pt.list(t, "/a/")
	waitFor(t, first)
	served()
	waitFor(t, second)
	for _, name := range []string{"a/1.jpg", "b/1.jpg"} {
		if pt.stored(t, name, 256) {
			t.Errorf("%s was made ahead after another folder was listed", name)
		}
	}
}
