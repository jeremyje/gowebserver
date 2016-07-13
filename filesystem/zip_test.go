package filesystem

import (
	test "github.com/jeremyje/gowebserver/testing"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
	"fmt"
	"io/ioutil"
)

func TestZipFs(t *testing.T) {
	assert := assert.New(t)

	path, err := test.GetZipFilePath()
	assert.Nil(err)

	handler, localArchivePath, dir, err := newZipFs(path)
	assert.Nil(err)
	assert.Equal(localArchivePath, path)
	assert.NotNil(handler)
	assert.Nil(verifyLocalFile(dir, "index.html"))
	assert.True(exists(filepath.Join(dir, "index.html")))
	assert.True(exists(filepath.Join(dir, "site.js")))
	assert.True(exists(filepath.Join(dir, "assets/1.txt")))
	assert.True(exists(filepath.Join(dir, "assets/2.txt")))
	assert.True(exists(filepath.Join(dir, "assets/more/3.txt")))
	assert.True(exists(filepath.Join(dir, "assets/four/4.txt")))
	assert.True(exists(filepath.Join(dir, "assets/fivesix/5.txt")))
	assert.True(exists(filepath.Join(dir, "assets/fivesix/6.txt")))
}

func verifyLocalFile(dir string, assetPath string) error {
	fullPath := filepath.Join(dir, assetPath)
	if !exists(fullPath) {
		return fmt.Errorf("%s does not exist when it's expected to.", fullPath)
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
