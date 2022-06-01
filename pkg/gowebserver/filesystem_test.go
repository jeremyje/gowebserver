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
	"io"
	"io/fs"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
	gowsTesting "github.com/jeremyje/gowebserver/internal/gowebserver/testing"
)

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
		f    func(string) (fs.FS, func(), error)
	}{
		{name: "newArchiveFs", path: zipPath, f: newArchiveFS},
		{name: "newArchiveFs", path: tarPath, f: newArchiveFS},
		{name: "newArchiveFs", path: tarGzPath, f: newArchiveFS},
		{name: "newArchiveFs", path: tarBz2Path, f: newArchiveFS},
		{name: "newArchiveFs", path: tarXzPath, f: newArchiveFS},
		{name: "newArchiveFs", path: tarLz4Path, f: newArchiveFS},

		{name: "newSevenZipFs", path: sevenZipPath, f: newSevenZipFS},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%s %s", tc.name, tc.path), func(t *testing.T) {
			t.Parallel()

			vFS, cleanup, err := tc.f(tc.path)
			if err != nil {
				t.Error(err)
			}
			defer cleanup()

			fp, err := vFS.Open("index.html")
			if err != nil {
				t.Fatal(err)
			}
			data, err := io.ReadAll(fp)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff("index.html", string(data)); diff != "" {
				t.Errorf("staged.localFilePath mismatch (-want +got):\n%s", diff)
			}

			// zip does not support strip prefix so testing/testassets/ is required.
			verifyLocalFileFromDefaultAsset(t, vFS)
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

		{input: "ok.zip", isSupportedArchive: true, isSupportedZip: true},
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
	}
}

func verifyLocalFileFromDefaultAsset(tb testing.TB, vFS fs.FS) {
	for _, fileName := range []string{"index.html", "site.js", "assets/1.txt", "assets/2.txt", "assets/more/3.txt", "assets/four/4.txt", "assets/fivesix/5.txt", "assets/fivesix/6.txt"} {
		verifyLocalFile(tb, vFS, fileName)
	}
}

func verifyLocalFile(tb testing.TB, vFS fs.FS, assetPath string) error {
	f, err := vFS.Open(assetPath)
	if err != nil {
		tb.Fatal(fmt.Errorf("%s does not exist when it's expected to, %s", assetPath, err))
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	if string(data) != assetPath {
		return fmt.Errorf("The test asset file does not contain it's relative file path as the body, File= %s, Body= %s", assetPath, string(data))
	}
	return nil
}

func verifyFileExist(tb testing.TB, vFS fs.FS, assetPath string) {
	f, err := vFS.Open(assetPath)
	if err != nil {
		tb.Errorf("cannot find '%s', %s", assetPath, err)
	}

	if _, err := f.Stat(); err != nil {
		tb.Errorf("cannot stat '%s', %s", assetPath, err)
	}
}

func verifyFileMissing(tb testing.TB, vFS fs.FS, assetPath string) {
	f, err := vFS.Open(assetPath)
	if f != nil {
		tb.Errorf("'%s' file handle is not nil when it should be", assetPath)
	}
	if err == nil {
		tb.Errorf("wanted error for reading '%s' as it should not exist", assetPath)
	}
}

func TestGitFsOverHttp(t *testing.T) {
	runGitFsTest(t, "https://github.com/jeremyje/gowebserver.git")
}

func runGitFsTest(tb testing.TB, path string) {
	vFS, cleanup, err := newGitFS(path)
	tb.Cleanup(cleanup)
	if err != nil {
		tb.Fatal(err)
	}

	verifyFileMissing(tb, vFS, ".gitignore")
	verifyFileMissing(tb, vFS, ".git")
	verifyFileMissing(tb, vFS, ".gitmodules")
	verifyFileExist(tb, vFS, "README.md")
	verifyFileExist(tb, vFS, ".github/dependabot.yml")
	verifyFileExist(tb, vFS, "cmd/gowebserver/gowebserver.go")
}

func TestNestedFSPath(t *testing.T) {
	testCases := []struct {
		input string
		want  []string
	}{
		{input: "", want: []string{""}},
		{input: ".", want: []string{"."}},
		{input: "a/b/c", want: []string{"a/b/c"}},
		{input: "a/b/c.ok/d/e/f.zip", want: []string{"a/b/c.ok/d/e/f.zip"}},
		{input: "./a/b/c.zip", want: []string{"./a/b/c.zip"}},
		{input: "./a/b/c.zip/.d/e/f.tar.gz/g/h/i.txt", want: []string{"./a/b/c.zip", ".d/e/f.tar.gz", "g/h/i.txt"}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			gotSplit := splitNestedFSPath(tc.input)
			if diff := cmp.Diff(tc.want, gotSplit); diff != "" {
				t.Errorf("splitNestedFSPath mismatch (-want +got):\n%s", diff)
			}
			gotJoin := joinNestedFSPath(gotSplit)

			if diff := cmp.Diff(tc.input, gotJoin); diff != "" {
				t.Errorf("splitNestedFSPath mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
