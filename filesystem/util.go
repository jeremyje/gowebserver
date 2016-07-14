package filesystem

import (
	"github.com/jeremyje/gowebserver/termhook"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const DIR_MODE = os.FileMode(0777)

func createDirectory(path string) error {
	return os.MkdirAll(dirPath(path), DIR_MODE)
}

func stageRemoteFile(maybeRemoteFilePath string) (string, string, error) {
	localFilePath, err := downloadFile(maybeRemoteFilePath)
	if err != nil {
		return "", "", err
	}
	tmpDir, err := createTempDirectory()
	if err != nil {
		return "", "", err
	}

	termhook.Add(func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			log.Fatalf("Cannot delete directory: %s, Error= %v", tmpDir, err)
		}
	})

	return localFilePath, tmpDir, nil
}

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
