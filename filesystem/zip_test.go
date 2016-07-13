package filesystem

import (
	test "github.com/jeremyje/gowebserver/testing"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestZipFs(t *testing.T) {
	assert := assert.New(t)

	path, err := test.GetZipFilePath()
	assert.Nil(err)

	fs := newZipFs(path)
	assert.NotNil(fs)
}
