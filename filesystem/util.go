package filesystem

import (
	"fmt"
	"github.com/jeremyje/gowebserver/termhook"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const fsDirMode = os.FileMode(0777)

type createFsResult struct {
	handler       http.FileSystem
	localFilePath string
	tmpDir        string
	err           error
}

func (r createFsResult) withError(err error) createFsResult {
	r.err = err
	return r
}

func (r createFsResult) withHandler(handler http.FileSystem, err error) createFsResult {
	r.handler = handler
	r.err = err
	return r
}

func createDirectory(path string) error {
	return os.MkdirAll(dirPath(path), fsDirMode)
}

func stageRemoteFile(maybeRemoteFilePath string) createFsResult {
	localFilePath, err := downloadFile(maybeRemoteFilePath)
	if err != nil {
		return createFsResult{err: fmt.Errorf("cannot download file %s, %s", maybeRemoteFilePath, err)}
	}
	tmpDir, err := createTempDirectory()
	if err != nil {
		return createFsResult{err: fmt.Errorf("cannot create temp directory, %s", err)}
	}

	return createFsResult{
		localFilePath: localFilePath,
		tmpDir:        tmpDir,
		err:           nil,
	}
}

func createTempDirectory() (string, error) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "gowebserver")

	if err != nil {
		return "", fmt.Errorf("cannot create temp directory, %s", err)
	}
	termhook.Add(func() {
		tryDeleteDirectory(tmpDir)
	})
	return tmpDir, nil
}

func tryDeleteDirectory(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		log.Fatalf("cannot delete directory: %s, Error= %v", path, err)
	}
}

func tryDeleteFile(path string) {
	err := os.Remove(path)
	if err != nil {
		log.Fatalf("cannot delete file: %s, Error= %v", path, err)
	}
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

func copyFile(reader io.Reader, filePath string) error {
	fsf, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("Cannot create target file %s, %s", filePath, err)
	}
	defer fsf.Close()

	_, err = io.Copy(fsf, reader)
	if err != nil {
		return fmt.Errorf("Cannot copy to target file %s, %s", filePath, err)
	}
	return nil
}
