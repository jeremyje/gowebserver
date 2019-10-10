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
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirPath(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("/", dirPath("/"))
	assert.Equal("/abc/", dirPath("/abc"))
	assert.Equal("/abc/", dirPath("/abc/"))
	assert.Equal("/abc/", dirPath("/abc//////////"))
}

func TestCreateTempDirectory(t *testing.T) {
	assert := assert.New(t)

	dir, err := createTempDirectory()
	assert.Nil(err)
	assert.True(exists(dir))
	assert.Contains(dir, "gowebserver")
}

func TestDownloadFileOnLocalFile(t *testing.T) {
	assert := assert.New(t)

	f, err := ioutil.TempFile(os.TempDir(), "gowebserver")
	assert.Nil(err)
	path := f.Name()
	err = ioutil.WriteFile(path, []byte("ok"), os.FileMode(0644))
	assert.Nil(err)

	localPath, err := downloadFile(path)
	assert.Equal(localPath, path)
	assert.Nil(err)
}

func TestDownloadFileOnHttpsFile(t *testing.T) {
	assert := assert.New(t)

	remotePath := "https://raw.githubusercontent.com/jeremyje/gowebserver/master/Makefile"
	localPath, err := downloadFile(remotePath)
	assert.Nil(err)
	assert.NotEqual(localPath, remotePath)
	assert.True(exists(localPath))
}
