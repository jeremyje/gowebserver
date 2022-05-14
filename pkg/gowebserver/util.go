// Copyright 2022 Jeremy Edwards
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

package gowebserver

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"go.uber.org/zap"
)

func checkError(err error) {
	if err != nil {
		zap.S().Error(err)
		zap.S().Sync()
	}
}

const fsDirMode = os.FileMode(0777)

type createFsResult struct {
	handler       http.Handler
	cleanup       func()
	localFilePath string
	tmpDir        string
	err           error
}

func (r createFsResult) withError(err error) createFsResult {
	r.err = err
	return r
}

func (r createFsResult) withHTTPHandler(handler http.Handler, cleanup func(), err error) createFsResult {
	r.handler = handler
	r.cleanup = cleanup
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
	tmpDir, cleanup, err := createTempDirectory()
	if err != nil {
		return createFsResult{err: fmt.Errorf("cannot create temp directory, %s", err)}
	}

	return createFsResult{
		localFilePath: localFilePath,
		tmpDir:        tmpDir,
		cleanup:       cleanup,
		err:           nil,
	}
}

func createTempDirectory() (string, func(), error) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "gowebserver")

	if err != nil {
		return "", nilFunc, fmt.Errorf("cannot create temp directory, %s", err)
	}
	return tmpDir, func() {
		tryDeleteDirectory(tmpDir)
	}, nil
}

func tryDeleteDirectory(path string) {
	if !exists(path) {
		return
	}

	if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
		zap.S().With("error", err, "directory", path).Error("cannot delete directory")
	}
}

func tryDeleteFile(path string) {
	if !exists(path) {
		return
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		zap.S().With("error", err, "file", path).Error("cannot delete file")
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

		if _, err := io.Copy(f, resp.Body); err != nil {
			return "", err
		}
		return f.Name(), nil
	}
	return path, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirPath(dirPath string) string {
	return strings.TrimRight(dirPath, "/") + "/"
}

func copyFile(reader io.Reader, filePath string) error {
	fsf, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("cannot create target file %s, %s", filePath, err)
	}
	defer fsf.Close()

	_, err = io.Copy(fsf, reader)
	if err != nil {
		return fmt.Errorf("cannot copy to target file %s, %s", filePath, err)
	}
	return nil
}

func nilFunc() {
}
