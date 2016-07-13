package filesystem

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
	"net/http"
)

func createTempDirectory() (string, error) {
	return ioutil.TempDir(os.TempDir(), "gowebserver")
}

func downloadFile(path string) (string, error) {
	if strings.HasPrefix(strings.ToLower(path), "http") {
		f, err := ioutil.TempFile(os.TempDir(), "gowebserverdl")
		if err != nil {
			return "", err
		}
		defer f.Close()
		resp, err := http.Get(path)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		io.Copy(f, resp.Body)
		return f.Name(), nil
	}
	return path, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func dirPath(dirPath string) string {
	return strings.TrimRight(dirPath, "/") + "/"
}
