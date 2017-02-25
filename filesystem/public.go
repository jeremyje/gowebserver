package filesystem

import (
	"net/http"
)

// New creates a filesystem for the HTTP server from an archive file.
func New(path string) (http.FileSystem, error) {
	if isSupportedZip(path) {
		handler, _, _, err := newZipFs(path)
		return handler, err
	} else if isSupportedTar(path) {
		handler, _, _, err := newTarFs(path)
		return handler, err
	} else {
		return newNative(path)
	}
}
