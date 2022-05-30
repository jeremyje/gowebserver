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
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	_ "embed"

	"github.com/google/go-cmp/cmp"
)

func TestCheckError(t *testing.T) {
	checkError(nil)
}

func TestDirPath(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		{input: "/", want: "/"},
		{input: "/abc", want: "/abc/"},
		{input: "/abc/", want: "/abc/"},
		{input: "/abc//////////", want: "/abc/"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got := dirPath(tc.input)
			if tc.want != got {
				t.Errorf("expected: %v, got: %v", tc.want, got)
			}
		})
	}
}

func TestCreateTempDirectory(t *testing.T) {
	dir, cleanup, err := createTempDirectory()
	if err != nil {
		t.Error(err)
	}
	if !exists(dir) {
		t.Errorf("'%s' does not exist when it should", dir)
	}

	if !strings.Contains(dir, "gowebserver") {
		t.Errorf("'%s' does not contain 'gowebserver'", dir)
	}
	cleanup()
	if exists(dir) {
		t.Errorf("'%s' exists when it should not", dir)
	}
}

func TestDownloadFileOnLocalFile(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "gowebserver")
	if err != nil {
		t.Error(err)
	}
	path := f.Name()
	err = ioutil.WriteFile(path, []byte("ok"), os.FileMode(0644))
	if err != nil {
		t.Error(err)
	}

	localPath, err := downloadFile(path)
	if localPath != path {
		t.Errorf("want: %v, got: %v", path, localPath)
	}

	if err != nil {
		t.Error(err)
	}
}

func TestDownloadFileOnHttpsFile(t *testing.T) {
	remotePath := "https://raw.githubusercontent.com/jeremyje/gowebserver/master/Makefile"
	localPath, err := downloadFile(remotePath)
	if err != nil {
		t.Error(err)
	}
	if localPath == remotePath {
		t.Errorf("'%s' is the local and remote path, they should be different", localPath)
	}
	if !exists(localPath) {
		t.Errorf("'%s' does not exist locally", localPath)
	}
}

func TestSanitizeFileName(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		{input: "///////////////\\\\\\\\\\", want: "."},
		{input: "../ok", want: "ok"},
		{input: "/ok/", want: "ok"},
		{input: "../whatever.json", want: "whatever.json"},
		{input: "../what ever!@#$%^&*()+_=-.json", want: "what ever_-.json"},
		{input: "../abc/def..tar.gz", want: "abc/def.tar.gz"},
		{input: "./././././../.../..../abc.tar.gz/.....", want: "abc.tar.gz"},
		{input: ".file", want: ".file"},
		{input: "../ok/.file", want: "ok/.file"},
		{input: ".", want: "."},
		{input: "/", want: "."},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got := sanitizeFileName(tc.input)
			if tc.want != got {
				t.Errorf("want: %v, got: %v", tc.want, got)
			}
		})
	}
}

func mustTempDir(tb testing.TB) string {
	dir, cleanup, err := createTempDirectory()
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(cleanup)
	return dir
}

func mustFile(tb testing.TB, path string) []byte {
	data, err := ioutil.ReadFile(path)
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

type angryReader struct {
}

func (a *angryReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("failure")
}

func TestCopyFileErrors(t *testing.T) {
	testCases := []struct {
		filePath string
		r        io.Reader
		wantErr  string
	}{
		{
			filePath: "dir-does-not-exist/target-file.txt",
			r:        &angryReader{},
			wantErr:  "cannot create target file dir-does-not-exist/target-file.txt, open dir-does-not-exist/target-file.txt: no such file or directory",
		},
		{
			filePath: "target-file.txt",
			r:        &angryReader{},
			wantErr:  "cannot copy to target file target-file.txt, failure",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.filePath, func(t *testing.T) {
			t.Parallel()
			if err := copyFile(tc.r, tc.filePath); err != nil {
				if diff := cmp.Diff(tc.wantErr, err.Error()); diff != "" {
					t.Errorf("copyFile() mismatch (-want +got):\n%s", diff)
				}
			} else {
				t.Error("expected an error")
			}
		})
	}
}
