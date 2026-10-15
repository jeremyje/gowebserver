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
	"image"
	"image/color"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

// browserAccept is what a browser sends for an <img>.
var browserAccept = http.Header{"Accept": {"image/avif,image/webp,image/apng,image/*,*/*;q=0.8"}}

// waitForPhotoTools asks h's thumbnailer what ffmpeg does for photos and
// waits until it knows.
func waitForPhotoTools(t *testing.T, h http.Handler) photoTools {
	t.Helper()
	th := h.(*thumbnailHandler).t
	th.photoTools()
	select {
	case <-th.probed:
	case <-time.After(30 * time.Second):
		t.Fatal("still finding out what ffmpeg does after 30s")
	}
	return th.photoTools()
}

// newFFmpegThumbnailHandler serves fsys with the ffmpeg installed on PATH,
// skipping the test when there's none that writes WebP.
func newFFmpegThumbnailHandler(t *testing.T, fsys fs.FS) http.Handler {
	t.Helper()
	bin, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg is not installed")
	}
	h, _ := newTestThumbnailHandler(t, fsys, bin)
	if !waitForPhotoTools(t, h).webp {
		t.Skip("ffmpeg has no libwebp encoder")
	}
	return h
}

func TestThumbnailFFmpegPhoto(t *testing.T) {
	// A 400x200 photo stored sideways, tagged to be turned 90 degrees
	// clockwise for display, as in TestThumbnailOrientation.
	jpg := withOrientation(encodeJPEG(t, stripes(400, 200)), 6)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "phone.jpg"), jpg, 0o600); err != nil {
		t.Fatal(err)
	}
	for name, fsys := range map[string]fs.FS{
		"on disk":       os.DirFS(dir),
		"in an archive": fstest.MapFS{"phone.jpg": {Data: jpg}},
	} {
		t.Run(name, func(t *testing.T) {
			h := newFFmpegThumbnailHandler(t, fsys)
			for _, tc := range []struct {
				name, query string
				header      http.Header
				format      string
				w, h        int
				// Whether ffmpeg made all of it, which is all decode.
				whole bool
				// The colors at the top and bottom: green and blue turned
				// the right way up, or red in the middle of a crop.
				top, bottom color.NRGBA
			}{
				{"webp", "width=64&height=64&fit=outside", browserAccept, "webp", 64, 128, true, thumbGreen, thumbBlue},
				{"webp fit inside", "width=64&height=64&fit=inside", browserAccept, "webp", 32, 64, true, thumbGreen, thumbBlue},
				{"webp by width", "width=50", browserAccept, "webp", 50, 100, true, thumbGreen, thumbBlue},
				{"webp asked for", "width=64&height=64&fit=outside&format=webp", nil, "webp", 64, 128, true, thumbGreen, thumbBlue},
				{"webp cropped", "width=50&height=50", browserAccept, "webp", 50, 50, false, thumbRed, thumbRed},
				{"jpeg", "width=64&height=64&fit=outside", nil, "jpeg", 64, 128, false, thumbGreen, thumbBlue},
				{"jpeg asked for", "width=64&height=64&fit=outside&format=jpeg", browserAccept, "jpeg", 64, 128, false, thumbGreen, thumbBlue},
			} {
				rec := thumbGet(t, h, "/phone.jpg?"+tc.query, tc.header)
				img := decodeThumb(t, rec)
				if got := rec.Header().Get("Content-Type"); got != "image/"+tc.format {
					t.Errorf("%s: Content-Type = %q, want image/%s", tc.name, got, tc.format)
				}
				if b := img.Bounds(); b.Dx() != tc.w || b.Dy() != tc.h {
					t.Errorf("%s: size = %dx%d, want %dx%d", tc.name, b.Dx(), b.Dy(), tc.w, tc.h)
					continue
				}
				b := img.Bounds()
				if c := img.At(b.Dx()/2, 4); !near(c, tc.top) {
					t.Errorf("%s: top = %v, want %v", tc.name, c, tc.top)
				}
				if c := img.At(b.Dx()/2, b.Dy()-5); !near(c, tc.bottom) {
					t.Errorf("%s: bottom = %v, want %v", tc.name, c, tc.bottom)
				}
				timing := rec.Header().Get("Server-Timing")
				if whole := !strings.Contains(timing, "scale"); whole != tc.whole {
					t.Errorf("%s: Server-Timing %q, want ffmpeg making all of it %v", tc.name, timing, tc.whole)
				}
			}
		})
	}
}

func TestThumbnailFFmpegVary(t *testing.T) {
	h := newFFmpegThumbnailHandler(t, fstest.MapFS{"a.jpg": {Data: encodeJPEG(t, stripes(40, 20))}})
	webp := thumbGet(t, h, "/a.jpg?width=20", browserAccept)
	jpg := thumbGet(t, h, "/a.jpg?width=20", nil)
	for _, rec := range []*http.Response{webp.Result(), jpg.Result()} {
		if got := rec.Header.Get("Vary"); !strings.Contains(got, "Accept") {
			t.Errorf("%s: Vary = %q, want Accept", rec.Header.Get("Content-Type"), got)
		}
	}
	if webp.Header().Get("ETag") == jpg.Header().Get("ETag") {
		t.Error("the WebP and JPEG thumbnails share an ETag")
	}
	// A format asked for doesn't depend on the browser.
	if got := thumbGet(t, h, "/a.jpg?width=20&format=png", browserAccept).Header().Get("Vary"); got != "" {
		t.Errorf("png asked for: Vary = %q, want none", got)
	}
}

func TestThumbnailFFmpegTransparent(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 40, 20))
	for x := range 20 { // the left half opaque red, the right clear
		for y := range 20 {
			src.SetNRGBA(x, y, thumbRed)
		}
	}
	h := newFFmpegThumbnailHandler(t, fstest.MapFS{"logo.png": {Data: encodePNG(t, src)}})
	rec := thumbGet(t, h, "/logo.png?width=20", browserAccept)
	img := decodeThumb(t, rec)
	if got := rec.Header().Get("Content-Type"); got != "image/webp" {
		t.Errorf("Content-Type = %q, want image/webp", got)
	}
	if _, _, _, a := img.At(2, 5).RGBA(); a != 0xffff {
		t.Errorf("left alpha = %#x, want opaque", a)
	}
	if _, _, _, a := img.At(17, 5).RGBA(); a > 0x1000 {
		t.Errorf("right alpha = %#x, want clear", a)
	}
}

func TestThumbnailFFmpegVideo(t *testing.T) {
	bin, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg is not installed")
	}
	dir := t.TempDir()
	video := filepath.Join(dir, "clip.mp4")
	if out, err := exec.Command(bin, "-hide_banner", "-loglevel", "error", "-nostdin",
		"-f", "lavfi", "-i", "color=c=green:size=320x180:rate=10:duration=12",
		"-c:v", "libx264", "-g", "30", "-pix_fmt", "yuv420p", video).CombinedOutput(); err != nil {
		t.Skipf("cannot make a test video: %v: %s", err, out)
	}
	h := newFFmpegThumbnailHandler(t, os.DirFS(dir))
	for _, tc := range []struct {
		header http.Header
		format string
	}{{browserAccept, "webp"}, {nil, "jpeg"}} {
		rec := thumbGet(t, h, "/clip.mp4?width=64&height=36&fit=outside", tc.header)
		img := decodeThumb(t, rec)
		if got := rec.Header().Get("Content-Type"); got != "image/"+tc.format {
			t.Errorf("Content-Type = %q, want image/%s", got, tc.format)
		}
		if b := img.Bounds(); b.Dx() != 64 || b.Dy() != 36 {
			t.Errorf("%s: size = %dx%d, want 64x36", tc.format, b.Dx(), b.Dy())
		}
		if c := img.At(32, 18); !near(c, color.NRGBA{G: 128, A: 255}) {
			t.Errorf("%s: middle = %v, want green", tc.format, c)
		}
	}
}

// TestThumbnailFFmpegFallsBack checks a photo ffmpeg can't read is made in
// Go.
func TestThumbnailFFmpegFallsBack(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script as ffmpeg")
	}
	// An ffmpeg that lists encoders, and fails at everything else.
	bin := filepath.Join(t.TempDir(), "ffmpeg")
	script := "#!/bin/sh\ncase \"$*\" in *-encoders*) printf ' ------\\n V....D mjpeg  MJPEG\\n'; exit 0;; esac\nexit 1\n"
	if err := os.WriteFile(bin, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	h, _ := newTestThumbnailHandler(t, fstest.MapFS{"a.jpg": {Data: encodeJPEG(t, stripes(40, 20))}}, bin)
	if tools := waitForPhotoTools(t, h); !tools.jpeg || tools.webp {
		t.Fatalf("photo tools = %+v, want JPEG decoding only", tools)
	}
	// Without libwebp, no WebP.
	rec := thumbGet(t, h, "/a.jpg?width=20", browserAccept)
	img := decodeThumb(t, rec)
	if got := rec.Header().Get("Content-Type"); got != "image/jpeg" {
		t.Errorf("Content-Type = %q, want image/jpeg", got)
	}
	if b := img.Bounds(); b.Dx() != 20 || b.Dy() != 10 {
		t.Errorf("size = %dx%d, want 20x10", b.Dx(), b.Dy())
	}
}

func TestChooseWebP(t *testing.T) {
	th := &thumbnailer{}
	accept := browserAccept.Get("Accept")
	if got := th.chooseWebP(&thumbnailSpec{}, accept).format; got != "" {
		t.Errorf("without ffmpeg: format %q, want it left to the image", got)
	}
	if got := th.chooseWebP(&thumbnailSpec{format: "webp"}, accept).format; got != "" {
		t.Errorf("webp asked for without ffmpeg: format %q, want it left to the image", got)
	}
	th.tools = &photoTools{jpeg: true, webp: true}
	th.ffmpeg = newFFmpeg(FFmpeg{Path: "/bin/true"})
	for _, tc := range []struct {
		format, accept, want string
	}{
		{"", accept, "webp"},
		{"", "image/png,*/*", ""},
		{"", "", ""},
		{"webp", "", "webp"},
		{"png", accept, "png"},
		{"jpeg", accept, "jpeg"},
	} {
		if got := th.chooseWebP(&thumbnailSpec{format: tc.format}, tc.accept).format; got != tc.want {
			t.Errorf("format %q, Accept %q: %q, want %q", tc.format, tc.accept, got, tc.want)
		}
	}
}

func TestJPEGReduce(t *testing.T) {
	for _, tc := range []struct {
		w, h, tw, th, want int
	}{
		// A 12 megapixel photo.
		{4032, 3024, 341, 256, 3},   // a small grid tile
		{4032, 3024, 683, 512, 2},   // a large one
		{4032, 3024, 1024, 768, 1},  // the slideshow's preview
		{4032, 3024, 2048, 1536, 0}, // its sharp version
		{4032, 3024, 4032, 3024, 0},
		// Never more than an eighth.
		{8000, 8000, 10, 10, 3},
		// Both sides have to stay large enough.
		{4000, 1000, 100, 600, 0},
	} {
		if got := jpegReduce(tc.w, tc.h, tc.tw, tc.th); got != tc.want {
			t.Errorf("jpegReduce(%d, %d, %d, %d) = %d, want %d", tc.w, tc.h, tc.tw, tc.th, got, tc.want)
		}
	}
}
