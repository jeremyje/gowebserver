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
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	gowsTesting "github.com/jeremyje/gowebserver/v2/internal/gowebserver/testing"
)

var (
	zeroTime               = time.Time{}
	commonFSRootDirList    = []string{"assets", "index.html", "site.js", "weird #1.txt", "weird#.txt", "weird$.txt"}
	nestedZipFSRootDirList = []string{"single-testassets.zip", "single-testassets.zip-dir", "testassets", "testassets.7z", "testassets.rar", "testassets.rar-dir", "testassets.tar", "testassets.tar-dir", "testassets.tar.bz2", "testassets.tar.bz2-dir", "testassets.tar.gz", "testassets.tar.gz-dir", "testassets.tar.lz4", "testassets.tar.lz4-dir", "testassets.tar.xz", "testassets.tar.xz-dir", "testassets.zip", "testassets.zip-dir", "testing.go", "testing_test.go"}
)

var (
	_ FileSystem    = (*localFS)(nil)
	_ FileSystem    = (*archiveFS)(nil)
	_ FileSystem    = (*nestedFS)(nil)
	_ io.ReaderAt   = (*augmentedFile)(nil)
	_ io.ReadSeeker = (*augmentedFile)(nil)
)

func TestDirlessArchive(t *testing.T) {
	nodirZipPath := gowsTesting.MustNoDirZipFilePath(t)

	vFS, err := newRawFSFromURI(nodirZipPath)
	if err != nil {
		t.Error(err)
	}
	defer vFS.Close()
	nFS := newNestedFS(vFS)
	defer nFS.Close()

	verifyReadDir(t, nFS, "assets", []string{"1.txt", "2.txt", "fivesix", "four", "more"})
}

func TestVirtualDirectory(t *testing.T) {
	nestedZipPath := gowsTesting.MustNestedZipFilePath(t)

	vFS, err := newRawFSFromURI(nestedZipPath)
	if err != nil {
		t.Error(err)
	}
	defer vFS.Close()
	nFS := newNestedFS(vFS)
	defer nFS.Close()

	dirs, err := nFS.ReadDir("")
	if err != nil {
		t.Fatal(err)
	}

	hasVirtual := false
	for _, dir := range dirs {
		if dir.IsDir() {
			if diff := cmp.Diff(fs.ModeDir, dir.Type()); diff != "" {
				t.Errorf("Type() mismatch (-want +got):\n%s", diff)
			}
			info, err := dir.Info()
			if err != nil {
				t.Errorf("dir.Info() error= %s", err)
			}
			if info == nil {
				t.Error("Info() is nil")
			}

			if strings.HasSuffix(dir.Name(), nestedDirSuffix) {
				hasVirtual = true
				vDir, ok := dir.(*virtualDirEntry)
				if ok {
					if diff := cmp.Diff(int64(0), vDir.Size()); diff != "" {
						t.Errorf("Size() mismatch (-want +got):\n%s", diff)
					}

					if diff := cmp.Diff(fs.ModeDir, vDir.Mode()); diff != "" {
						t.Errorf("Mode() mismatch (-want +got):\n%s", diff)
					}

					if vDir.Sys() == nil {
						t.Errorf("%s.Sys() is nil", vDir.Name())
					}

					if vDir.ModTime().Before(emptyTime) {
						t.Errorf("ModTime() is before empty time, %s", vDir.ModTime())
					}
				} else {
					t.Errorf("%s is not a virtual directory but has the suffix", dir.Name())
				}
			}
		}
	}

	if !hasVirtual {
		t.Error("no virtual directories were verified")
	}
}

func TestNestedFileSystem(t *testing.T) {
	zipPath := gowsTesting.MustZipFilePath(t)
	rarPath := gowsTesting.MustRarFilePath(t)
	sevenZipPath := gowsTesting.MustSevenZipFilePath(t)
	tarPath := gowsTesting.MustTarFilePath(t)
	tarGzPath := gowsTesting.MustTarGzFilePath(t)
	tarBz2Path := gowsTesting.MustTarBzip2FilePath(t)
	tarXzPath := gowsTesting.MustTarXzFilePath(t)
	tarLz4Path := gowsTesting.MustTarLz4FilePath(t)
	nodirZipPath := gowsTesting.MustNoDirZipFilePath(t)
	singleZipPath := gowsTesting.MustSingleZipFilePath(t)
	nestedZipPath := gowsTesting.MustNestedZipFilePath(t)

	testCases := []struct {
		uri          string
		baseDir      string
		isNested     bool
		wantRootDirs []string
	}{
		{uri: zipPath, baseDir: ""},
		{uri: rarPath, baseDir: ""},
		{uri: tarPath, baseDir: ""},
		{uri: tarGzPath, baseDir: ""},
		{uri: tarBz2Path, baseDir: ""},
		{uri: tarXzPath, baseDir: ""},
		{uri: tarLz4Path, baseDir: ""},

		{uri: sevenZipPath, baseDir: ""},

		{uri: nodirZipPath, baseDir: ""},

		{uri: singleZipPath, baseDir: "testassets", wantRootDirs: []string{"testassets"}},

		{uri: nestedZipPath, baseDir: "testassets", wantRootDirs: nestedZipFSRootDirList},
		{uri: nestedZipPath, baseDir: "single-testassets.zip-dir/testassets", isNested: true, wantRootDirs: nestedZipFSRootDirList},
		{uri: nestedZipPath, baseDir: "testassets.tar-dir", isNested: true, wantRootDirs: nestedZipFSRootDirList},
		{uri: nestedZipPath, baseDir: "testassets.tar.bz2-dir", isNested: true, wantRootDirs: nestedZipFSRootDirList},
		{uri: nestedZipPath, baseDir: "testassets.tar.lz4-dir", isNested: true, wantRootDirs: nestedZipFSRootDirList},
		{uri: nestedZipPath, baseDir: "testassets.zip-dir", isNested: true, wantRootDirs: nestedZipFSRootDirList},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%s %s", tc.uri, tc.baseDir), func(t *testing.T) {
			t.Parallel()

			vFS, err := newRawFSFromURI(tc.uri)
			if err != nil {
				t.Error(err)
			}
			defer vFS.Close()
			nFS := newNestedFS(vFS)
			defer nFS.Close()

			baseDir := tc.baseDir
			t.Run("verifyFileSystem", func(t *testing.T) {
				verifyFileSystem(t, nFS, baseDir)
			})

			if !tc.isNested {
				t.Run("verifyReadDir rawFS", func(t *testing.T) {
					verifyReadDir(t, vFS, baseDir, commonFSRootDirList)
				})
			}
			t.Run("verifyReadDir nestedFS", func(t *testing.T) {
				verifyReadDir(t, nFS, baseDir, commonFSRootDirList)
				if tc.wantRootDirs == nil {
					verifyReadDir(t, nFS, "", commonFSRootDirList)
				} else {
					verifyReadDir(t, nFS, "", tc.wantRootDirs)
				}
			})
		})
	}
}

func verifyFileSystem(tb testing.TB, nFS *nestedFS, baseDir string) {
	indexFilePath := filepath.Join(baseDir, "index.html")
	fp, err := nFS.Open(indexFilePath)
	if err != nil {
		tb.Fatalf("cannot open '%s', err=%s", indexFilePath, err)
	}

	if stat, err := fp.Stat(); err != nil {
		tb.Errorf("'%s' stat error, %s", indexFilePath, err)
	} else {
		nStat, err := nFS.Stat(indexFilePath)
		if err != nil {
			tb.Errorf("cannot stat from nestedFS, %s", err)
		}
		if diff := cmp.Diff(stat.IsDir(), nStat.IsDir()); diff != "" {
			tb.Errorf("nestedFS.IsDir() mismatch (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff(stat.ModTime(), nStat.ModTime()); diff != "" {
			tb.Errorf("nestedFS.ModTime() mismatch (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff(stat.Mode(), nStat.Mode()); diff != "" {
			tb.Errorf("nestedFS.Mode() mismatch (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff(stat.Name(), nStat.Name()); diff != "" {
			tb.Errorf("nestedFS.Name() mismatch (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff(stat.Size(), nStat.Size()); diff != "" {
			tb.Errorf("nestedFS.Size() mismatch (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff(false, stat.IsDir()); diff != "" {
			tb.Errorf("[%s].Stat.IsDir() mismatch (-want +got):\n%s", stat.Name(), diff)
		}

		if zeroTime.After(stat.ModTime()) {
			tb.Errorf("[%s].Stat.ModTime() should be in the past, %v", stat.Name(), stat.ModTime())
		}

		if diff := cmp.Diff("index.html", stat.Name()); diff != "" {
			tb.Errorf("[%s].Stat.Name() mismatch (-want +got):\n%s", stat.Name(), diff)
		}

		if diff := cmp.Diff(int64(10), stat.Size()); diff != "" {
			tb.Errorf("[%s].Stat.Size() mismatch (-want +got):\n%s", stat.Name(), diff)
		}
	}
	data, err := io.ReadAll(fp)
	if err != nil {
		tb.Fatalf("cannot read '%s', err=%s", indexFilePath, err)
	}

	if diff := cmp.Diff("index.html", string(data)); diff != "" {
		tb.Errorf("index.html from FS mismatch (-want +got):\n%s", diff)
	}

	data, err = nFS.ReadFile(indexFilePath)
	if err != nil {
		tb.Errorf("cannot read %s via nestedFS, %s", indexFilePath, err)
	}
	if diff := cmp.Diff("index.html", string(data)); diff != "" {
		tb.Errorf("index.html from nestedFS mismatch (-want +got):\n%s", diff)
	}

	// zip does not support strip prefix so testing/testassets/ is required.
	verifyLocalFileFromDefaultAsset(tb, nFS, baseDir)
}

func verifyReadDir(tb testing.TB, vFS FileSystem, baseDir string, want []string) {
	entries, err := vFS.ReadDir(baseDir)
	if err != nil {
		tb.Fatalf("cannot read directory '%s', %s", baseDir, err)
	}

	names := []string{}
	for _, dir := range entries {
		names = append(names, dir.Name())
	}
	if diff := cmp.Diff(want, names); diff != "" {
		tb.Errorf("ReadDir mismatch (-want +got):\n%s", diff)
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

func verifyLocalFileFromDefaultAsset(tb testing.TB, vFS fs.FS, baseDir string) {
	for _, fileName := range []string{"index.html", "site.js", "weird$.txt", "weird #1.txt", "weird#.txt", "assets/1.txt", "assets/2.txt", "assets/more/3.txt", "assets/four/4.txt", "assets/fivesix/5.txt", "assets/fivesix/6.txt"} {
		name := fileName
		if baseDir != "" {
			name = filepath.Join(baseDir, name)
		}
		if err := verifyLocalFile(vFS, baseDir, name); err != nil {
			tb.Error(err)
		}
	}

	verifyFileMissing(tb, vFS, "does-not-exist/")
	verifyFileMissing(tb, vFS, "does-not-exist")
}

func verifyLocalFile(vFS fs.FS, baseDir string, assetPath string) error {
	f, err := vFS.Open(assetPath)
	if err != nil {
		return fmt.Errorf("%s does not exist when it's expected to, %s", assetPath, err)
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	//parts := strings.Split(assetPath, nestedDirSuffix+"/")
	wantContent := strings.TrimLeft(strings.TrimPrefix(assetPath, baseDir), "/")
	if string(data) != wantContent {
		return fmt.Errorf("The test asset file does not contain it's relative file path as the body, File= %s, WantBody= '%s', Body= '%s'", assetPath, wantContent, string(data))
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
	vFS, err := newGitFS(path)
	defer func() {
		if err := vFS.Close(); err != nil {
			tb.Error(err)
		}
	}()

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
		{input: "./a/b/c.zip-dir/.d/e/f.tar.gz-dir/g/h/i.txt", want: []string{"./a/b/c.zip-dir", ".d/e/f.tar.gz-dir", "g/h/i.txt"}},
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
				t.Errorf("joinNestedFSPath mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
