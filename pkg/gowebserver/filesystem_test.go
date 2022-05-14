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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	gowsTesting "github.com/jeremyje/gowebserver/internal/gowebserver/testing"
)

func TestIsSupportedTar(t *testing.T) {
	testCases := []struct {
		input       string
		want        bool
		wantRegular bool
		wantGzip    bool
		wantBzip2   bool
	}{
		{input: "ok.tar", want: true, wantRegular: true, wantGzip: false, wantBzip2: false},
		{input: "ok.TAR", want: true, wantRegular: true, wantGzip: false, wantBzip2: false},
		{input: "ok.tar.gz", want: true, wantRegular: false, wantGzip: true, wantBzip2: false},
		{input: "ok.tar.GZ", want: true, wantRegular: false, wantGzip: true, wantBzip2: false},
		{input: "ok.tar.bz2", want: true, wantRegular: false, wantGzip: false, wantBzip2: true},
		{input: "ok.tar.BZ2", want: true, wantRegular: false, wantGzip: false, wantBzip2: true},

		{input: "ok.tar.lzma", want: false, wantRegular: false, wantGzip: false, wantBzip2: false},
		{input: "ok.zip", want: false, wantRegular: false, wantGzip: false, wantBzip2: false},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("isSupportedTar(%s)", tc.input), func(t *testing.T) {
			t.Parallel()
			got := isSupportedTar(tc.input)
			if tc.want != got {
				t.Errorf("expected: %v, got: %v", tc.want, got)
			}
		})

		t.Run(fmt.Sprintf("isRegularTar(%s)", tc.input), func(t *testing.T) {
			t.Parallel()
			got := isRegularTar(tc.input)
			if tc.wantRegular != got {
				t.Errorf("expected: %v, got: %v", tc.wantRegular, got)
			}
		})

		t.Run(fmt.Sprintf("isTarGzip(%s)", tc.input), func(t *testing.T) {
			t.Parallel()
			got := isTarGzip(tc.input)
			if tc.wantGzip != got {
				t.Errorf("expected: %v, got: %v", tc.wantGzip, got)
			}
		})

		t.Run(fmt.Sprintf("isTarBzip2(%s)", tc.input), func(t *testing.T) {
			t.Parallel()
			got := isTarBzip2(tc.input)
			if tc.wantBzip2 != got {
				t.Errorf("expected: %v, got: %v", tc.wantBzip2, got)
			}
		})
	}
}

func TestFileSystem(t *testing.T) {
	zipPath := gowsTesting.MustZipFilePath(t)
	sevenZipPath := gowsTesting.MustSevenZipFilePath(t)
	tarPath := gowsTesting.MustTarFilePath(t)
	tarGzPath := gowsTesting.MustTarGzFilePath(t)
	tarBz2Path := gowsTesting.MustTarBzip2FilePath(t)
	tarXzPath := gowsTesting.MustTarXzFilePath(t)
	tarLz4Path := gowsTesting.MustTarLz4FilePath(t)

	testCases := []struct {
		name string
		path string
		f    func(string) createFsResult
	}{
		{name: "newArchiveFs", path: zipPath, f: newArchiveFs},
		{name: "newArchiveFs", path: sevenZipPath, f: newArchiveFs},
		{name: "newArchiveFs", path: tarPath, f: newArchiveFs},
		{name: "newArchiveFs", path: tarGzPath, f: newArchiveFs},
		{name: "newArchiveFs", path: tarBz2Path, f: newArchiveFs},
		{name: "newArchiveFs", path: tarXzPath, f: newArchiveFs},
		{name: "newArchiveFs", path: tarLz4Path, f: newArchiveFs},

		{name: "newZipFs", path: zipPath, f: newZipFs},
		{name: "newSevenZipFs", path: sevenZipPath, f: newSevenZipFs},
		{name: "newTarFs", path: tarPath, f: newTarFs},
		{name: "newTarFs", path: tarGzPath, f: newTarFs},
		{name: "newTarFs", path: tarBz2Path, f: newTarFs},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%s %s", tc.name, tc.path), func(t *testing.T) {
			t.Parallel()

			staged := tc.f(tc.path)

			if diff := cmp.Diff(staged.localFilePath, tc.path); diff != "" {
				t.Errorf("staged.localFilePath mismatch (-want +got):\n%s", diff)
			}

			if staged.handler == nil {
				t.Error("staged.handler is nil")
			}

			// zip does not support strip prefix so testing/testassets/ is required.
			verifyLocalFileFromDefaultAsset(t, staged.tmpDir)
		})
	}
}

func TestIsSupported(t *testing.T) {
	testCases := []struct {
		input               string
		isSupportedArchive  bool
		isSupportedGit      bool
		isSupportedHTTP     bool
		isSupportedSevenZip bool
		isSupportedTar      bool
		isSupportedZip      bool
	}{
		{input: "ok.tar", isSupportedArchive: true, isSupportedTar: true},
		{input: "ok.tar.gz", isSupportedArchive: true, isSupportedTar: true},
		{input: "ok.tar.bz2", isSupportedArchive: true, isSupportedTar: true},
		{input: "ok.tar.xz", isSupportedArchive: true},
		{input: "ok.tar.lz4", isSupportedArchive: true},
		{input: "ok.tar.br", isSupportedArchive: true},
		{input: "ok.tar.zst", isSupportedArchive: true},

		{input: "ok.zip", isSupportedZip: true},
		{input: "ok.tar.lzma"},
		{input: "ok.7z", isSupportedSevenZip: true},
		{input: "ok.7Z", isSupportedSevenZip: true},

		{input: "git@github.com:jeremyje/gowebserver.git", isSupportedGit: true},
		{input: "https://github.com/jeremyje/gowebserver.git", isSupportedHTTP: true, isSupportedGit: true},

		{input: "http://www.google.com/", isSupportedHTTP: true},
		{input: "http://www.google.com.7z", isSupportedHTTP: true, isSupportedSevenZip: true},

		{input: ""},
		{input: "/"},
	}

	tf := func(t *testing.T, name string, f func(string) bool, input string, want bool) {
		t.Run(fmt.Sprintf("%s %s", name, input), func(t *testing.T) {
			t.Parallel()
			got := f(input)
			if want != got {
				t.Errorf("want: %v, got: %v", want, got)
			}
		})
	}

	for _, tc := range testCases {
		tc := tc

		tf(t, "isSupportedArchive", isSupportedArchive, tc.input, tc.isSupportedArchive)
		tf(t, "isSupportedGit", isSupportedGit, tc.input, tc.isSupportedGit)
		tf(t, "isSupportedHTTP", isSupportedHTTP, tc.input, tc.isSupportedHTTP)
		tf(t, "isSupportedSevenZip", isSupportedSevenZip, tc.input, tc.isSupportedSevenZip)
		tf(t, "isSupportedTar", isSupportedTar, tc.input, tc.isSupportedTar)
		tf(t, "isSupportedZip", isSupportedZip, tc.input, tc.isSupportedZip)
	}
}

func verifyLocalFileFromDefaultAsset(tb testing.TB, dir string) {
	verifyLocalFile(tb, dir, "index.html")
	verifyLocalFile(tb, dir, "site.js")
	verifyLocalFile(tb, dir, "assets/1.txt")
	verifyLocalFile(tb, dir, "assets/2.txt")
	verifyLocalFile(tb, dir, "assets/more/3.txt")
	verifyLocalFile(tb, dir, "assets/four/4.txt")
	verifyLocalFile(tb, dir, "assets/fivesix/5.txt")
	verifyLocalFile(tb, dir, "assets/fivesix/6.txt")
}

func verifyLocalFile(tb testing.TB, dir string, assetPath string) error {
	fullPath := filepath.Join(dir, assetPath)
	if !exists(fullPath) {
		return fmt.Errorf("%s does not exist when it's expected to", fullPath)
	}
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return err
	}
	if string(data) != assetPath {
		return fmt.Errorf("The test asset file does not contain it's relative file path as the body, File= %s, Body= %s", fullPath, string(data))
	}
	return nil
}

func verifyFileExist(tb testing.TB, dir string, assetPath string) {
	fullPath := filepath.Join(dir, assetPath)
	if !exists(fullPath) {
		tb.Errorf("%s does not exist when it's expected to", fullPath)
	}
}

func verifyFileMissing(tb testing.TB, dir string, assetPath string) {
	fullPath := filepath.Join(dir, assetPath)
	if exists(fullPath) {
		tb.Errorf("%s exists when it's expected to be deleted", fullPath)
	}
}

func TestGitFsOverHttp(t *testing.T) {
	runGitFsTest(t, "https://github.com/jeremyje/gowebserver.git")
}

func runGitFsTest(tb testing.TB, path string) {
	staged := newGitFs(path)
	tb.Logf("Local: %s    Dir: %s, Error %s", staged.localFilePath, staged.tmpDir, staged.err)
	if staged.err != nil {
		tb.Error(staged.err)
	}

	if diff := cmp.Diff(staged.localFilePath, path); diff != "" {
		tb.Errorf("staged.localFilePath mismatch (-want +got):\n%s", diff)
	}

	if staged.handler == nil {
		tb.Error("staged.handler is nil")
	}

	verifyFileMissing(tb, staged.tmpDir, ".gitignore")
	verifyFileMissing(tb, staged.tmpDir, ".git")
	verifyFileMissing(tb, staged.tmpDir, ".gitmodules")
	verifyFileExist(tb, staged.tmpDir, "README.md")
	verifyFileExist(tb, staged.tmpDir, ".github/workflows/dependabot.yml")
	verifyFileExist(tb, staged.tmpDir, "cmd/gowebserver/gowebserver.go")
}
