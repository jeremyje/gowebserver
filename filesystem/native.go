package filesystem

import (
	"net/http"
	"path/filepath"
)

func newNative(directory string) (http.FileSystem, error) {
	dir, err := filepath.Abs(directory)
	if err != nil {
		return nil, err
	}
	return http.Dir(dirPath(dir)), nil
}
