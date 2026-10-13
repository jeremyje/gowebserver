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

package gowebserver

import (
	"net/http"
	"sort"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// collected is what the thumbnail metrics hold: for a counter or
// histogram named name, its count for each set of attributes written
// "k=v,k=v".
func collected(t *testing.T, reader *sdkmetric.ManualReader) map[string]map[string]int64 {
	t.Helper()
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(t.Context(), &rm); err != nil {
		t.Fatal(err)
	}
	key := func(set attribute.Set) string {
		kvs := set.ToSlice()
		parts := make([]string, 0, len(kvs))
		for _, kv := range kvs {
			parts = append(parts, string(kv.Key)+"="+kv.Value.String())
		}
		sort.Strings(parts)
		return strings.Join(parts, ",")
	}
	got := map[string]map[string]int64{}
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			points := map[string]int64{}
			switch data := m.Data.(type) {
			case metricdata.Sum[int64]:
				for _, p := range data.DataPoints {
					points[key(p.Attributes)] = p.Value
				}
			case metricdata.Histogram[float64]:
				for _, p := range data.DataPoints {
					points[key(p.Attributes)] = int64(p.Count)
				}
			}
			got[m.Name] = points
		}
	}
	return got
}

func TestThumbnailMetrics(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	spans := tracetest.NewSpanRecorder()
	mc := &monitoringContext{
		promProvider: sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)),
		tp:           sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spans)),
	}
	th, err := newThumbnailer(Thumbnails{Enabled: true, CachePath: thumbnailCacheMemory}, nil, mc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := th.close(); err != nil {
			t.Error(err)
		}
	})
	fsys := fstest.MapFS{
		"photo.jpg":  {Data: encodeJPEG(t, stripes(400, 200))},
		"broken.jpg": {Data: []byte("not a picture")},
		"clip.mp4":   {Data: []byte("not really a video")},
	}
	base := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	h := th.handler(base, fsys, "src")

	// Made, with where the time went for the browser's network panel.
	rec := thumbGet(t, h, "/photo.jpg?width=64", nil)
	decodeThumb(t, rec)
	timing := rec.Header().Get("Server-Timing")
	timed := strings.Split(timing, ", ")
	stages := make([]string, 0, len(timed))
	for _, part := range timed {
		name, dur, ok := strings.Cut(part, ";dur=")
		if !ok || dur == "" {
			t.Errorf("Server-Timing %q: %q has no duration", timing, part)
		}
		stages = append(stages, name)
	}
	if diff := cmp.Diff([]string{"wait", "decode", "scale", "encode", "store"}, stages); diff != "" {
		t.Errorf("Server-Timing %q stages (-want +got):\n%s", timing, diff)
	}
	// Then from the store, which is all there is to say about it.
	rec = thumbGet(t, h, "/photo.jpg?width=64", nil)
	decodeThumb(t, rec)
	if got := rec.Header().Get("Server-Timing"); got != "cache;desc=hit" {
		t.Errorf("Server-Timing of a stored thumbnail = %q", got)
	}
	thumbGet(t, h, "/broken.jpg?width=64", nil)
	// There is no ffmpeg, so no video thumbnails.
	thumbGet(t, h, "/clip.mp4?width=64", nil)

	got := collected(t, reader)
	want := map[string]map[string]int64{
		"thumbnail_requests_total": {
			"kind=photo,result=made":        1,
			"kind=photo,result=cached":      1,
			"kind=photo,result=unsupported": 1,
			"kind=video,result=unsupported": 1,
		},
		"thumbnail_request_duration_seconds": {
			"kind=photo,result=made":        1,
			"kind=photo,result=cached":      1,
			"kind=photo,result=unsupported": 1,
			"kind=video,result=unsupported": 1,
		},
		"thumbnail_stage_duration_seconds": {
			// Both photos waited and were decoded, or tried; one got further.
			"kind=photo,stage=wait":   2,
			"kind=photo,stage=decode": 2,
			"kind=photo,stage=scale":  1,
			"kind=photo,stage=encode": 1,
			"kind=photo,stage=store":  1,
			"kind=video,stage=wait":   1,
			"kind=video,stage=decode": 1,
		},
		"thumbnail_source_megapixels": {"kind=photo": 1},
		// Nothing is being made any more.
		"thumbnail_renders": {
			"kind=photo,state=waiting": 0,
			"kind=photo,state=running": 0,
			"kind=video,state=waiting": 0,
			"kind=video,state=running": 0,
		},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("thumbnail metrics (-want +got):\n%s", diff)
	}

	// The first request's trace: the stages under its span.
	var request sdktrace.ReadOnlySpan
	children := map[string]bool{}
	for _, s := range spans.Ended() {
		if s.Name() == "thumbnail" && request == nil {
			request = s
		}
	}
	if request == nil {
		t.Fatal("no thumbnail span")
	}
	for _, s := range spans.Ended() {
		if s.SpanContext().TraceID() == request.SpanContext().TraceID() && s != request {
			children[s.Name()] = true
		}
	}
	wantSpans := map[string]bool{"thumbnail.wait": true, "thumbnail.decode": true, "thumbnail.scale": true, "thumbnail.encode": true, "thumbnail.store": true}
	if diff := cmp.Diff(wantSpans, children); diff != "" {
		t.Errorf("spans of a thumbnail request (-want +got):\n%s", diff)
	}
	attrs := map[string]string{}
	for _, kv := range request.Attributes() {
		attrs[string(kv.Key)] = kv.Value.String()
	}
	if attrs["kind"] != "photo" || attrs["result"] != "made" {
		t.Errorf("thumbnail span attributes = %v, want kind photo and result made", attrs)
	}
}
