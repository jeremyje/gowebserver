package filesystem

import (
	"net/http"
	"strings"
)

func New(path string) (http.FileSystem, error) {
	lcPath := path
	if strings.HasSuffix(lcPath, ".zip") {
		handler, _, _, err := newZipFs(path)
		return handler, err
	}
	return newNative(path)
}
