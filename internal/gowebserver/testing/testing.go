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

package testing

import (
	_ "embed"
	"io/ioutil"
	"os"
	"testing"
)

const testFileMode = os.FileMode(0644)

var (
	//go:embed testassets.zip
	zipAssets []byte
	//go:embed testassets.7z
	sevenZipAssets []byte
	//go:embed testassets.tar
	tarAssets []byte
	//go:embed testassets.tar.gz
	tarGzAssets []byte
	//go:embed testassets.tar.bz2
	tarBzip2Assets []byte
	//go:embed testassets.tar.xz
	tarXzAssets []byte
	//go:embed testassets.tar.lz4
	tarLz4Assets []byte
)

// MustCreateTempFile creates a temp file for testing.
func MustCreateTempFile(tb testing.TB) *os.File {
	f, err := ioutil.TempFile(os.TempDir(), "tempfile")
	fatalOnFail(tb, err)

	cleanupFile(tb, f.Name())
	return f
}

// MustWriteTempFile writes the contents to a file for testing.
func MustWriteTempFile(tb testing.TB, content string) *os.File {
	fp := MustCreateTempFile(tb)

	fatalOnFail(tb, ioutil.WriteFile(fp.Name(), []byte(content), testFileMode))
	return fp
}

// MustZipFilePath gets the .zip test asset file.
func MustZipFilePath(tb testing.TB) string {
	return mustWriteData(tb, mustCreateTempArchive(tb, ".zip"), zipAssets)
}

// MustSevenZipFilePath gets the .7z test asset file.
func MustSevenZipFilePath(tb testing.TB) string {
	return mustWriteData(tb, mustCreateTempArchive(tb, ".7z"), sevenZipAssets)
}

// MustTarFilePath gets the .tar test asset file.
func MustTarFilePath(tb testing.TB) string {
	return mustWriteData(tb, mustCreateTempArchive(tb, ".tar"), tarAssets)
}

// MustTarGzFilePath gets the .tar.gz test asset file.
func MustTarGzFilePath(tb testing.TB) string {
	return mustWriteData(tb, mustCreateTempArchive(tb, ".tar.gz"), tarGzAssets)
}

// MustTarBzip2FilePath gets .tar.bz2 test asset file.
func MustTarBzip2FilePath(tb testing.TB) string {
	return mustWriteData(tb, mustCreateTempArchive(tb, ".tar.bz2"), tarBzip2Assets)
}

// MustTarXzFilePath gets .tar.xz test asset file.
func MustTarXzFilePath(tb testing.TB) string {
	return mustWriteData(tb, mustCreateTempArchive(tb, ".tar.xz"), tarXzAssets)
}

// MustTarLz4FilePath gets .tar.lz4 test asset file.
func MustTarLz4FilePath(tb testing.TB) string {
	return mustWriteData(tb, mustCreateTempArchive(tb, ".tar.lz4"), tarLz4Assets)
}

func mustCreateTempArchive(tb testing.TB, suffix string) string {
	tf := MustCreateTempFile(tb)
	return tf.Name() + suffix
}

func mustWriteData(tb testing.TB, path string, data []byte) string {
	fatalOnFail(tb, ioutil.WriteFile(path, data, testFileMode))
	return path
}

func cleanupFile(tb testing.TB, name string) {
	tb.Cleanup(func() {
		if name != "" {
			err := os.Remove(name)
			if err != nil {
				tb.Errorf("cannot delete '%s', %s", name, err)
			}
		}
	})
}

func fatalOnFail(tb testing.TB, err error) {
	if err != nil {
		tb.Fatal(err)
	}
}
