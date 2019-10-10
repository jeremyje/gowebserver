// Copyright 2019 Jeremy Edwards
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

package filesystem

import (
	"log"
	"testing"

	test "github.com/jeremyje/gowebserver/testing"
	"github.com/stretchr/testify/assert"
)

func TestIsSupportedTar(t *testing.T) {
	assert := assert.New(t)

	assert.True(isSupportedTar("ok.tar"))
	assert.True(isSupportedTar("ok.tar.gz"))
	assert.True(isSupportedTar("ok.tar.bz2"))

	assert.False(isSupportedTar("ok.tar.lzma"))
	assert.False(isSupportedTar("ok.zip"))
}

func TestIsRegularTar(t *testing.T) {
	assert := assert.New(t)

	assert.True(isRegularTar("ok.tar"))
	assert.True(isRegularTar("ok.TAR"))
	assert.False(isRegularTar("ok.tar.gz"))
	assert.False(isRegularTar("ok.tar.bz2"))

	assert.False(isRegularTar("ok.tar.lzma"))
	assert.False(isRegularTar("ok.zip"))
}

func TestIsTarGzip(t *testing.T) {
	assert := assert.New(t)

	assert.False(isTarGzip("ok.tar"))
	assert.True(isTarGzip("ok.tar.gz"))
	assert.True(isTarGzip("ok.TAR.GZ"))
	assert.False(isTarGzip("ok.tar.bz2"))

	assert.False(isTarGzip("ok.tar.lzma"))
	assert.False(isTarGzip("ok.zip"))
}

func TestIsTarBzip2(t *testing.T) {
	assert := assert.New(t)

	assert.False(isTarBzip2("ok.tar"))
	assert.False(isTarBzip2("ok.tar.gz"))
	assert.True(isTarBzip2("ok.tar.bz2"))
	assert.True(isTarBzip2("ok.TAR.BZ2"))

	assert.False(isTarBzip2("ok.tar.lzma"))
	assert.False(isTarBzip2("ok.zip"))
}

func TestTarFsRegular(t *testing.T) {
	path, err := test.GetTarFilePath()
	runTarFsTest(t, path, err)
}

func TestTarFsGz(t *testing.T) {
	path, err := test.GetTarGzFilePath()
	runTarFsTest(t, path, err)
}

func TestTarFsBzip2(t *testing.T) {
	path, err := test.GetTarBzip2FilePath()
	runTarFsTest(t, path, err)
}

func runTarFsTest(t *testing.T, path string, err error) {
	assert := assert.New(t)
	assert.Nil(err)

	staged := newTarFs(path)
	log.Printf("Local: %s    Dir: %s, Error %s", staged.localFilePath, staged.tmpDir, staged.err)
	assert.Nil(staged.err)
	assert.Equal(staged.localFilePath, path)
	assert.NotNil(staged.handler)
	verifyLocalFileFromDefaultAsset(staged.tmpDir, assert)
}
