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
