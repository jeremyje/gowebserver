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
	"io/fs"
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
	return newHandlerFromFS(path)
}

func newHandlerFromFS(path string) (http.Handler, func(), error) {
	if !isSupportedGit(path) && isSupportedHTTP(path) {
		staged := newHTTPReverseProxy(path)
		return staged.handler, staged.cleanup, staged.err
	}
	vFS, cleanup, err := newRawFSFromURI(path)
	if err != nil {
		cleanup()
		return nil, nilFunc, err
	}

	return http.FileServer(http.FS(vFS)), cleanup, nil
}

func newRawFSFromURI(path string) (fs.FS, func(), error) {
	if isSupportedSevenZip(path) {
		return newSevenZipFS(path)
	} else if isSupportedArchive(path) {
		return newArchiveFS(path)
	} else if isSupportedGit(path) {
		return newGitFS(path)
	}
	return newLocalFS(path)
}

func isSupportedArchive(filePath string) bool {
	for _, suffix := range archives {
		if strings.HasSuffix(strings.ToLower(filePath), suffix) {
			return true
		}
	}
	return false
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

func newLocalFS(directory string) (fs.FS, func(), error) {
	dir, err := filepath.Abs(directory)
	if err != nil {
		return nil, nilFunc, err
	}
	return os.DirFS(filepath.Clean(dirPath(dir))), nilFunc, nil
}

func newArchiveFS(filePath string) (fs.FS, func(), error) {
	staged := stageRemoteFile(filePath)
	if staged.err != nil {
		return nil, staged.cleanup, staged.err
	}
	fs, err := archiver.FileSystem(filePath)
	return fs, nilFunc, err
}

func newSevenZipFS(filePath string) (fs.FS, func(), error) {
	staged := stageRemoteFile(filePath)
	if staged.err != nil {
		return nil, staged.cleanup, staged.err
	}

	// Extract archive
	r, err := sevenzip.OpenReader(staged.localFilePath)
	if err != nil {
		return nil, staged.cleanup, err
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
				return nil, staged.cleanup, fmt.Errorf("cannot create directory: %s, %s", filePath, err)
			}
		} else {
			dirPath := filepath.Dir(filePath)
			err = createDirectory(dirPath)
			if err != nil {
				return nil, staged.cleanup, fmt.Errorf("cannot create directory: %s, %s", dirPath, err)
			}

			err := writeFileFromArchiveEntry(f, filePath)
			if err != nil {
				return nil, staged.cleanup, fmt.Errorf("cannot write zip file entry: %s, %s", name, err)
			}
		}
	}

	localFS, cleanup, err := newLocalFS(staged.tmpDir)
	return localFS, func() {
		os.Remove(staged.localFilePath)
		staged.cleanup()
		cleanup()
	}, err
}

func newGitFS(filePath string) (fs.FS, func(), error) {
	if !isSupportedGit(filePath) {
		return nil, nilFunc, fmt.Errorf("%s is not a valid git repository", filePath)
	}

	tmpDir, cleanup, err := createTempDirectory()
	if err != nil {
		return nil, nilFunc, fmt.Errorf("cannot create temp directory, %s", err)
	}

	if _, err := git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL:          filePath,
		Progress:     os.Stdout,
		Depth:        1,
		SingleBranch: true,
	}); err != nil {
		return nil, nilFunc, fmt.Errorf("could not clone %s, %s", filePath, err)
	}

	tryDeleteDirectory(filepath.Join(tmpDir, ".git"))
	tryDeleteFile(filepath.Join(tmpDir, ".gitignore"))
	tryDeleteFile(filepath.Join(tmpDir, ".gitmodules"))
	lFS, localCleanup, err := newLocalFS(tmpDir)
	return lFS, func() {
		cleanup()
		localCleanup()
	}, err
}

func isSupportedGit(filePath string) bool {
	return strings.HasSuffix(strings.ToLower(filePath), ".git")
}
