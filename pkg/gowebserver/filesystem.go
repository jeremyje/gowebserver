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
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bodgit/sevenzip"
	git "github.com/go-git/go-git/v5"
	archiver "github.com/mholt/archiver/v4"
)

var (
	archives = []string{".tar", ".tar.gz", ".tar.bz2", ".tar.xz", ".tar.lz4", ".tar.br", ".tar.zst", ".rar", ".zip"}
)

func newFS(path string) (http.Handler, func(), error) {
	if isSupportedSevenZip(path) {
		staged := newSevenZipFs(path)
		return staged.handler, staged.cleanup, staged.err
	} else if isSupportedArchive(path) {
		staged := newArchiveFs(path)
		return staged.handler, staged.cleanup, staged.err
	} else if isSupportedGit(path) {
		staged := newGitFs(path)
		return staged.handler, staged.cleanup, staged.err
	} else if isSupportedHTTP(path) {
		staged := newHTTPReverseProxy(path)
		return staged.handler, staged.cleanup, staged.err
	}
	return newNative(path)
}

func newArchiveFs(filePath string) createFsResult {
	staged := stageRemoteFile(filePath)
	if staged.err != nil {
		return staged
	}
	fs, err := archiver.FileSystem(filePath)
	if err != nil {
		return staged.withError(err)
	}

	return staged.withHTTPHandler(http.FileServer(http.FS(fs)), nilFunc, nil)
}

func isSupportedArchive(filePath string) bool {
	for _, suffix := range archives {
		if strings.HasSuffix(strings.ToLower(filePath), suffix) {
			return true
		}
	}
	return false
}

func newSevenZipFs(filePath string) createFsResult {
	staged := stageRemoteFile(filePath)
	if staged.err != nil {
		return staged
	}
	// Extract archive
	r, err := sevenzip.OpenReader(staged.localFilePath)
	if err != nil {
		return staged.withError(err)
	}
	defer r.Close()

	// Iterate through the files in the archive,
	// printing some of their contents.
	for _, f := range r.File {
		name := sanitizeFileName(f.Name)
		filePath := filepath.Join(staged.tmpDir, name)
		if f.FileInfo().IsDir() {
			err = createDirectory(filePath)
			if err != nil {
				return staged.withError(fmt.Errorf("cannot create directory: %s, %s", filePath, err))
			}
		} else {
			dirPath := filepath.Dir(filePath)
			err = createDirectory(dirPath)
			if err != nil {
				return staged.withError(fmt.Errorf("cannot create directory: %s, %s", dirPath, err))
			}

			err := writeFileFromArchiveEntry(f, filePath)
			if err != nil {
				return staged.withError(fmt.Errorf("cannot write zip file entry: %s, %s", name, err))
			}
		}
	}
	return staged.withHTTPHandler(newNative(staged.tmpDir))
}

func isSupportedSevenZip(filePath string) bool {
	return strings.HasSuffix(strings.ToLower(filePath), ".7z")
}

func writeFileFromArchiveEntry(f opener, filePath string) error {
	zf, err := f.Open()
	if err != nil {
		return fmt.Errorf("cannot open input file: %s", err)
	}
	defer zf.Close()
	return copyFile(zf, filePath)
}

type opener interface {
	Open() (io.ReadCloser, error)
}

func newNative(directory string) (http.Handler, func(), error) {
	dir, err := filepath.Abs(directory)
	if err != nil {
		return nil, nilFunc, err
	}
	return http.FileServer(http.Dir(dirPath(dir))), nilFunc, nil
}

func newGitFs(filePath string) createFsResult {
	staged := createFsResult{
		localFilePath: filePath,
	}
	if !isSupportedGit(filePath) {
		return staged.withError(fmt.Errorf("%s is not a valid git repository", filePath))
	}

	tmpDir, cleanup, err := createTempDirectory()
	if err != nil {
		return staged.withError(fmt.Errorf("cannot create temp directory, %s", err))
	}
	staged.tmpDir = tmpDir
	_, err = git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL:          filePath,
		Progress:     os.Stdout,
		Depth:        1,
		SingleBranch: true,
	})
	if err != nil {
		return staged.withError(fmt.Errorf("could not clone %s, %s", filePath, err))
	}
	tryDeleteDirectory(filepath.Join(tmpDir, ".git"))
	tryDeleteFile(filepath.Join(tmpDir, ".gitignore"))
	tryDeleteFile(filepath.Join(tmpDir, ".gitmodules"))
	h, _, err := newNative(tmpDir)
	return staged.withHTTPHandler(h, cleanup, err)
}

func isSupportedGit(filePath string) bool {
	return strings.HasSuffix(strings.ToLower(filePath), ".git")
}
