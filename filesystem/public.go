package filesystem

import (
	"net/http"
)

func New(path string) (http.FileSystem, error) {
	return newNative(path), nil
}
