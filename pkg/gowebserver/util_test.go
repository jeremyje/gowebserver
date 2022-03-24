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
	"io/ioutil"
	"os"
	"strings"
	"testing"
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
		t.Errorf("expected: %v, got: %v", path, localPath)
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
