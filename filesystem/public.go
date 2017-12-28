package filesystem

import (
	"fmt"
	"net/http"
)

// New creates a filesystem for the HTTP server from an archive file.
func New(path string) (http.FileSystem, error) {
	if isSupportedZip(path) {
		handler, _, _, err := newZipFs(path)
		return handler, fmt.Errorf("cannot create hosted zip file, %s", err)
	} else if isSupportedTar(path) {
		handler, _, _, err := newTarFs(path)
		return handler, fmt.Errorf("cannot create hosted tarball, %s", err)
	} else if isSupportedGit(path) {
		handler, _, _, err := newGitFs(path)
		return handler, fmt.Errorf("cannot create hosted git repository, %s", err)
	}
	return newNative(path)
}
