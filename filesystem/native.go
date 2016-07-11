package filesystem

import (
	"net/http"
)

func newNative(directory string) http.FileSystem {
	return http.Dir(directory + "/")
}
