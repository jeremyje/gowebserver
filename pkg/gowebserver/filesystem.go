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
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bodgit/sevenzip"
	archiver "github.com/mholt/archiver/v4"
	"go.uber.org/zap"
	git "gopkg.in/src-d/go-git.v4"
)

var (
	archives = []string{".tar", ".tar.gz", ".tar.bz2", ".tar.xz", ".tar.lz4", ".tar.br", ".tar.zst", ".rar"}
)

func newFS(path string) (http.Handler, func(), error) {
	if isSupportedZip(path) {
		staged := newZipFs(path)
		return staged.handler, staged.cleanup, staged.err
	} else if isSupportedSevenZip(path) {
		staged := newSevenZipFs(path)
		return staged.handler, staged.cleanup, staged.err
	} else if isSupportedTar(path) {
		staged := newTarFs(path)
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
		filePath := filepath.Join(staged.tmpDir, f.Name)
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
				return staged.withError(fmt.Errorf("cannot write zip file entry: %s, %s", f.Name, err))
			}
		}
	}
	return staged.withHTTPHandler(newNative(staged.tmpDir))
}

func isSupportedSevenZip(filePath string) bool {
	return strings.HasSuffix(strings.ToLower(filePath), ".7z")
}

func newZipFs(filePath string) createFsResult {
	staged := stageRemoteFile(filePath)
	if staged.err != nil {
		return staged
	}
	// Extract archive
	r, err := zip.OpenReader(staged.localFilePath)
	if err != nil {
		return staged.withError(err)
	}
	defer r.Close()

	// Iterate through the files in the archive,
	// printing some of their contents.
	for _, f := range r.File {
		filePath := filepath.Join(staged.tmpDir, f.Name)
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
				return staged.withError(fmt.Errorf("cannot write zip file entry: %s, %s", f.Name, err))
			}
		}
	}
	return staged.withHTTPHandler(newNative(staged.tmpDir))
}

func writeFileFromArchiveEntry(f opener, filePath string) error {
	zf, err := f.Open()
	if err != nil {
		return fmt.Errorf("cannot open input file: %s", err)
	}
	defer zf.Close()
	return copyFile(zf, filePath)
}

func isSupportedZip(filePath string) bool {
	return strings.HasSuffix(strings.ToLower(filePath), ".zip")
}

type opener interface {
	Open() (io.ReadCloser, error)
}

func newTarFs(filePath string) createFsResult {
	if !isSupportedTar(filePath) {
		return createFsResult{
			err: fmt.Errorf("%s is not a valid tarball", filePath),
		}
	}
	staged := stageRemoteFile(filePath)
	if staged.err != nil {
		return staged
	}

	var r io.Reader
	f, err := os.Open(staged.localFilePath)
	if err != nil {
		return staged.withError(err)
	}
	defer f.Close()
	r = f

	if isTarGzip(filePath) {
		gzf, err := gzip.NewReader(f)
		if err != nil {
			return staged.withError(err)
		}
		r = gzf
	} else if isTarBzip2(filePath) {
		bzf := bzip2.NewReader(f)
		r = bzf
	}
	tr := tar.NewReader(r)

	err = processTarEntries(tr, staged.tmpDir)
	if err != nil {
		return staged.withError(err)
	}
	return staged.withHTTPHandler(newNative(staged.tmpDir))
}

func processTarEntries(tr *tar.Reader, tmpDir string) error {
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("cannot get next tar entry, %s", err)
		}

		localPath := filepath.Join(tmpDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			err = createDirectory(localPath)
		case tar.TypeReg:
			err = writeFileFromTarEntry(localPath, tmpDir, tr)
		default:
			zap.S().With("headerName", header.Name, "header", header).Warn("Tar entry type not supported.")
		}
		if err != nil {
			return fmt.Errorf("error processing tar entry %v, %s", header.Typeflag, err)
		}
	}
	return nil
}

func writeFileFromTarEntry(localPath string, tmpDir string, tr *tar.Reader) error {
	err := createDirectory(filepath.Dir(localPath))
	if err != nil {
		return err
	}
	return copyFile(tr, localPath)
}

func isSupportedTar(filePath string) bool {
	return isRegularTar(filePath) || isTarGzip(filePath) || isTarBzip2(filePath)
}

func isRegularTar(filePath string) bool {
	return strings.HasSuffix(strings.ToLower(filePath), ".tar")
}

func isTarGzip(filePath string) bool {
	return strings.HasSuffix(strings.ToLower(filePath), ".tar.gz")
}

func isTarBzip2(filePath string) bool {
	return strings.HasSuffix(strings.ToLower(filePath), ".tar.bz2")
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
