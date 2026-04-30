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

package gowebserver

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	_ "embed"

	"github.com/google/go-cmp/cmp"
)

func TestCheckError(t *testing.T) {
	checkError(nil)
}

func mustTempDir(tb testing.TB) string {
	dir, err := os.MkdirTemp(os.TempDir(), "gowebserver")
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

func mustFile(tb testing.TB, path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		tb.Fatalf("cannot read file '%s', %s", path, err)
	}
	return data
}

var (
	//go:embed testdata/hi-template.html
	hiTemplateHTML []byte
	//go:embed testdata/hi-template-want.html
	hiTemplateWantHTML []byte
	//go:embed testdata/broken-template.html
	brokenTemplateHTML []byte
)

func TestExecuteTemplate(t *testing.T) {
	testCases := []struct {
		name    string
		input   []byte
		want    []byte
		wantErr bool
	}{
		{
			name:  "testdata/hi-template.html",
			input: hiTemplateHTML,
			want:  hiTemplateWantHTML,
		},
		{
			name:    "template-index.html",
			input:   templateIndexHTML,
			wantErr: true,
		},
		{
			name:    "testdata/broken-template.html",
			input:   brokenTemplateHTML,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			w := &bytes.Buffer{}
			var params = struct {
				TestString string
			}{"test-string"}
			if err := executeTemplate(tc.input, params, w); err != nil {
				if !tc.wantErr {
					t.Error(err)
				}
			} else {
				if tc.wantErr {
					t.Error("expected an error")
				}
				if diff := cmp.Diff(string(tc.want), w.String()); diff != "" {
					t.Errorf("executeTemplate() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestHumanizeDate(t *testing.T) {
	testCases := []struct {
		date          time.Time
		wantDate      string
		wantTimestamp string
	}{
		{
			date:          time.Date(2022, time.January, 1, 3, 4, 5, 0, time.UTC),
			wantDate:      "2022-01-01",
			wantTimestamp: "2022-01-01 03:04:05AM",
		},
		{
			date:          time.Date(2040, time.December, 31, 6, 7, 8, 0, time.UTC),
			wantDate:      "2040-12-31",
			wantTimestamp: "2040-12-31 06:07:08AM",
		},
		{
			date:          time.Date(2000, time.May, 25, 0, 0, 0, 0, time.UTC),
			wantDate:      "2000-05-25",
			wantTimestamp: "2000-05-25 12:00:00AM",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.date.String(), func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tc.wantDate, humanizeDate(tc.date)); diff != "" {
				t.Errorf("humanizeDate() mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantTimestamp, humanizeTimestamp(tc.date)); diff != "" {
				t.Errorf("humanizeTimestamp() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestIsEven(t *testing.T) {
	if isEven(1) {
		t.Error("1 is even")
	}
	if !isEven(2) {
		t.Error("2 is odd")
	}
}

func TestIsOdd(t *testing.T) {
	if !isOdd(1) {
		t.Error("1 is even")
	}
	if isOdd(2) {
		t.Error("2 is odd")
	}
}

func TestIsImage(t *testing.T) {
	testCases := []struct {
		name    string
		isImage bool
		isAudio bool
		isVideo bool
		isMedia bool
	}{
		{
			name:    "testdata/hi-template.html",
			isImage: false,
		},
		{
			name:    "testdata/image.jpg",
			isImage: true,
			isMedia: true,
		},
		{
			name:    "testdata/image.gif",
			isImage: true,
			isMedia: true,
		},
		{
			name:    "testdata/image.jpeg",
			isImage: true,
			isMedia: true,
		},
		{
			name:    "testdata/image.png",
			isImage: true,
			isMedia: true,
		},
		{
			name:    "testdata/image.mp4",
			isVideo: true,
			isMedia: true,
		},
		{
			name:    "testdata.png/image.mp4",
			isVideo: true,
			isMedia: true,
		},
		{
			name:    "testdata/sound.mp3",
			isAudio: true,
		},
		{
			name:    "testdata/sound.wav",
			isAudio: true,
		},
		{
			name: "testdata/doc.txt",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("isImage(%s)", tc.name), func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tc.isImage, isImage(tc.name)); diff != "" {
				t.Errorf("isImage() mismatch (-want +got):\n%s", diff)
			}
		})

		t.Run(fmt.Sprintf("isAudio(%s)", tc.name), func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tc.isAudio, isAudio(tc.name)); diff != "" {
				t.Errorf("isAudio() mismatch (-want +got):\n%s", diff)
			}
		})

		t.Run(fmt.Sprintf("isVideo(%s)", tc.name), func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tc.isVideo, isVideo(tc.name)); diff != "" {
				t.Errorf("isVideo() mismatch (-want +got):\n%s", diff)
			}
		})

		t.Run(fmt.Sprintf("isMedia(%s)", tc.name), func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tc.isMedia, isMedia(tc.name)); diff != "" {
				t.Errorf("isMedia() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestStepBeginEnd(t *testing.T) {
	testCases := []struct {
		size  int
		begin map[int]interface{}
		end   map[int]interface{}
		step  int
	}{
		{
			size:  0,
			begin: map[int]interface{}{},
			end:   map[int]interface{}{},
			step:  0,
		},
		{
			size:  4,
			begin: map[int]interface{}{0: nil, 1: nil, 2: nil, 3: nil},
			end:   map[int]interface{}{0: nil, 1: nil, 2: nil, 3: nil},
			step:  0,
		},
		{
			size:  4,
			begin: map[int]interface{}{0: nil, 1: nil, 2: nil, 3: nil},
			end:   map[int]interface{}{0: nil, 1: nil, 2: nil, 3: nil},
			step:  1,
		},
		{
			size:  4,
			begin: map[int]interface{}{0: nil, 2: nil},
			end:   map[int]interface{}{1: nil, 3: nil},
			step:  2,
		},
		{
			size:  4,
			begin: map[int]interface{}{0: nil, 3: nil},
			end:   map[int]interface{}{2: nil, 3: nil},
			step:  3,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%+v", tc), func(t *testing.T) {
			t.Parallel()

			for i := 0; i < tc.size; i++ {
				_, want := tc.begin[i]
				if diff := cmp.Diff(want, stepBegin(i, tc.step, tc.size)); diff != "" {
					t.Errorf("stepBegin(%d, %d, %d) mismatch (-want +got):\n%s", i, tc.step, tc.size, diff)
				}
				_, want = tc.end[i]
				if diff := cmp.Diff(want, stepEnd(i, tc.step, tc.size)); diff != "" {
					t.Errorf("stepEnd(%d, %d, %d) mismatch (-want +got):\n%s", i, tc.step, tc.size, diff)
				}
			}
		})
	}
}

func TestUrlEncode(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		{input: "", want: ""},
		{input: "'", want: "%27"},
		{input: "/a/b/c/d.txt", want: "%2Fa%2Fb%2Fc%2Fd.txt"},
		{input: "/a/b/c/d    .txt", want: "%2Fa%2Fb%2Fc%2Fd%20%20%20%20.txt"},
		{input: `weird %1.txt`, want: `weird%20%251.txt`},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tc.want, urlEncode(tc.input)); diff != "" {
				t.Errorf("urlEncode() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
