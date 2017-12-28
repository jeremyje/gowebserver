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

	handler, localArchivePath, dir, err := newGitFs(path)
	log.Printf("Local: %s    Dir: %s, Error %s", localArchivePath, dir, err)
	assert.Nil(err)
	assert.Equal(localArchivePath, path)
	assert.NotNil(handler)
	assert.Nil(verifyFileMissing(dir, ".gitignore"))
	assert.Nil(verifyFileMissing(dir, ".git"))
	assert.Nil(verifyFileMissing(dir, ".gitmodules"))
	assert.Nil(verifyFileExist(dir, "README.md"))
	assert.Nil(verifyFileExist(dir, "cert/cert.go"))
}
