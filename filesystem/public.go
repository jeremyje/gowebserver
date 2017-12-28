package filesystem

import (
	"net/http"
)

// New creates a filesystem for the HTTP server from an archive file.
func New(path string) (http.FileSystem, error) {
	if isSupportedZip(path) {
		staged := newZipFs(path)
		return staged.handler, staged.err
	} else if isSupportedTar(path) {
		staged := newTarFs(path)
		return staged.handler, staged.err
	} else if isSupportedGit(path) {
		staged := newGitFs(path)
		return staged.handler, staged.err
	}
	return newNative(path)
}
