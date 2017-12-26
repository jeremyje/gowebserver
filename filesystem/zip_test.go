package filesystem

import (
	"fmt"
	test "github.com/jeremyje/gowebserver/testing"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"path/filepath"
	"testing"
)

func TestZipFs(t *testing.T) {
	assert := assert.New(t)

	path, err := test.GetZipFilePath()
	assert.Nil(err)

	handler, localArchivePath, dir, err := newZipFs(path)
	assert.Nil(err)
	assert.Equal(localArchivePath, path)
	assert.NotNil(handler)

	// zip does not support strip prefix so testing/testassets/ is required.
	verifyLocalFileFromDefaultAsset(filepath.Join(dir, "testing", "testassets"), assert)
}

func TestIsSupportedZip(t *testing.T) {
	assert := assert.New(t)

	assert.True(isSupportedZip("ok.zip"))
	assert.True(isSupportedZip("ok.ZIP"))

	assert.False(isSupportedZip("ok.tar"))
	assert.False(isSupportedZip("ok.tar.gz"))
	assert.False(isSupportedZip("ok.tar.bz2"))

	assert.False(isSupportedZip("ok.tar.lzma"))
}

func verifyLocalFileFromDefaultAsset(dir string, assert *assert.Assertions) {
	assert.Nil(verifyLocalFile(dir, "index.html"))
	assert.Nil(verifyLocalFile(dir, "site.js"))
	assert.Nil(verifyLocalFile(dir, "assets/1.txt"))
	assert.Nil(verifyLocalFile(dir, "assets/2.txt"))
	assert.Nil(verifyLocalFile(dir, "assets/more/3.txt"))
	assert.Nil(verifyLocalFile(dir, "assets/four/4.txt"))
	assert.Nil(verifyLocalFile(dir, "assets/fivesix/5.txt"))
	assert.Nil(verifyLocalFile(dir, "assets/fivesix/6.txt"))
}

func verifyLocalFile(dir string, assetPath string) error {
	fullPath := filepath.Join(dir, assetPath)
	if !exists(fullPath) {
		return fmt.Errorf("%s does not exist when it's expected to", fullPath)
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

func scanDir(dir string, assert *assert.Assertions) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}

	for _, file := range files {
		log.Printf("*** %s", filepath.Join(dir, file.Name()))
		if file.IsDir() {
			scanDir(filepath.Join(dir, file.Name()), assert)
		}
	}
}
