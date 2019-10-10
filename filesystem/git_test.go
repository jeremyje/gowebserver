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
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestIsSupportedGit(t *testing.T) {
	assert := assert.New(t)

	assert.True(isSupportedGit("git@github.com:jeremyje/gowebserver.git"))
	assert.True(isSupportedGit("https://github.com/jeremyje/gowebserver.git"))

	assert.False(isSupportedGit("ok.tar.gz"))
	assert.False(isSupportedGit("ok.zip"))
}

func TestGitFsOverHttp(t *testing.T) {
	runGitFsTest(t, "https://github.com/jeremyje/gowebserver.git")
}

func runGitFsTest(t *testing.T, path string) {
	assert := assert.New(t)

	staged := newGitFs(path)
	log.Printf("Local: %s    Dir: %s, Error %s", staged.localFilePath, staged.tmpDir, staged.err)
	assert.Nil(staged.err)
	assert.Equal(staged.localFilePath, path)
	assert.NotNil(staged.handler)
	assert.Nil(verifyFileMissing(staged.tmpDir, ".gitignore"))
	assert.Nil(verifyFileMissing(staged.tmpDir, ".git"))
	assert.Nil(verifyFileMissing(staged.tmpDir, ".gitmodules"))
	assert.Nil(verifyFileExist(staged.tmpDir, "README.md"))
	assert.Nil(verifyFileExist(staged.tmpDir, "cert/cert.go"))
}
