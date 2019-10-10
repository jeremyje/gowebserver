// Copyright 2019 Jeremy Edwards
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package filesystem

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/jeremyje/gowebserver/termhook"
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
	if err != nil && err != os.ErrNotExist {
		log.Fatalf("cannot delete directory: %s, Error= %v", path, err)
	}
}

func tryDeleteFile(path string) {
	err := os.Remove(path)
	if err != nil && err != os.ErrNotExist {
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
