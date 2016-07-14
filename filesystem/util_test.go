package filesystem

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
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
