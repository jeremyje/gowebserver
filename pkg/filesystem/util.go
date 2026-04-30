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

package filesystem

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

const fsDirMode = os.FileMode(0777)

func createDirectory(path string) error {
	return os.MkdirAll(dirPath(path), fsDirMode)
}

func stageRemoteFile(maybeRemoteFilePath string) (string, string, func() error, error) {
	localFilePath, fileCleanup, err := downloadFile(maybeRemoteFilePath)
	if err != nil {
		fileCleanup()
		return "", "", nilFuncWithError, fmt.Errorf("cannot download file %s, %w", maybeRemoteFilePath, err)
	}

	tmpDir, cleanup, err := createTempDirectory()
	if err != nil {
		logError(fileCleanup())
		cleanup()
		return "", "", nilFuncWithError, fmt.Errorf("cannot create temp directory, %w", err)
	}

	return tmpDir, localFilePath, func() error {
		cleanup()
		return fileCleanup()
	}, nil
}

func createTempDirectory() (string, func(), error) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "gowebserver")

	if err != nil {
		return "", nilFunc, fmt.Errorf("cannot create temp directory, %w", err)
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

func downloadFile(path string) (string, func() error, error) {
	if strings.HasPrefix(strings.ToLower(path), "http") {
		cleanup := nilFuncWithError
		f, err := os.CreateTemp(os.TempDir(), "gowebserverdl")
		if err != nil {
			return "", cleanup, err
		}
		defer f.Close()
		fileName := f.Name()
		cleanup = func() error {
			return os.Remove(fileName)
		}
		resp, err := http.Get(path)
		if err != nil {
			return "", cleanup, err
		}
		defer resp.Body.Close()

		if _, err := io.Copy(f, resp.Body); err != nil {
			return "", cleanup, err
		}
		return f.Name(), cleanup, nil
	}
	return path, nilFuncWithError, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirPath(dirPath string) string {
	return strings.TrimRight(dirPath, "/") + "/"
}

func copyFile(reader io.Reader, createdTime time.Time, modifiedTime time.Time, filePath string) error {
	fsf, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("cannot create target file %s, %w", filePath, err)
	}
	defer fsf.Close()

	_, err = io.Copy(fsf, reader)
	if err != nil {
		os.Remove(fsf.Name())
		return fmt.Errorf("cannot copy to target file %s, %w", filePath, err)
	}
	return os.Chtimes(filePath, createdTime, modifiedTime)
}

var (
	validChars = map[rune]interface{}{
		'.':  nil,
		'-':  nil,
		'_':  nil,
		' ':  nil,
		'\\': nil,
		'/':  nil,
		'$':  nil,
		'#':  nil,
	}
)

// SanitizeFileName removes path traversal and unsafe characters from a file name.
func SanitizeFileName(fileName string) string {
	name := strings.ReplaceAll(filepath.Clean(fileName), "..", ".")
	sanitized := ""
	for _, r := range name {
		if _, ok := validChars[r]; ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9') || ok {
			sanitized = sanitized + string(r)
		}
	}

	parts := strings.Split(strings.ReplaceAll(sanitized, "\\", "/"), "/")
	sanitizedParts := []string{}
	for _, part := range parts {
		if strings.ReplaceAll(part, ".", "") != "" {
			sanitizedParts = append(sanitizedParts, part)
		}
	}

	return filepath.Clean(strings.Join(sanitizedParts, string(filepath.Separator)))
}

func nilFunc() {
}

func nilFuncWithError() error {
	return nil
}

func logError(err error) {
	if err != nil {
		zap.S().Errorf("%s", err)
	}
}
