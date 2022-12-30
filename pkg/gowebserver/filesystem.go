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
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/bodgit/sevenzip"
	archiver "github.com/mholt/archiver/v4"
	"go.opentelemetry.io/otel/trace"
)

const (
	defaultFileMode = fs.FileMode(0644)
)

var (
	archives = []string{".tar", ".tar.gz", ".tar.bz2", ".tar.xz", ".tar.lz4", ".tar.br", ".tar.zst", ".rar", ".zip"}
)

func splitNestedFSPath(path string) []string {
	parts := strings.Split(path, "/")
	segments := []string{}
	archiveDir := false

	cur := []string{}
	for _, part := range parts {
		cur = append(cur, part)
		if strings.HasSuffix(part, nestedDirSuffix) {
			undelimitedPart := strings.TrimSuffix(part, nestedDirSuffix)
			if isSupportedArchive(undelimitedPart) || isSupportedSevenZip(undelimitedPart) {
				segments = append(segments, strings.Join(cur, "/"))
				cur = []string{}
				archiveDir = true
			} else {
				archiveDir = false
			}
		}
	}
	if len(cur) > 0 {
		segments = append(segments, strings.Join(cur, "/"))
	} else if archiveDir {
		segments = append(segments, ".")
	}
	return segments
}

func joinNestedFSPath(paths []string) string {
	return strings.Join(paths, "/")
}

func newHandlerFromFS(path string, tp trace.TracerProvider, enhancedList bool) (http.Handler, func() error, error) {
	if !isSupportedGit(path) && isSupportedHTTP(path) {
		handler, err := newHTTPReverseProxy(path)
		return handler, nilFuncWithError, err
	}
	vFS, err := newRawFSFromURI(path)
	if err != nil {
		return nil, nilFuncWithError, err
	}
	nFS := newNestedFS(vFS)

	ci, err := newCustomIndex(http.FileServer(http.FS(nFS)), nFS, tp, enhancedList)
	if err != nil {
		return nil, nilFuncWithError, err
	}
	return ci, nFS.Close, nil
}

func newRawFSFromURI(path string) (FileSystem, error) {
	if isSupportedSevenZip(path) {
		return newSevenZipFS(path)
	} else if isSupportedArchive(path) {
		return newArchiveFSFromLocalPath(path)
	} else if isSupportedGit(path) {
		return newGitFS(path)
	}
	return newLocalFS(path, func() error { return nil })
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
		return fmt.Errorf("cannot open input file: %w", err)
	}
	defer zf.Close()

	return copyFile(zf, f.CreatedTime(), f.ModTime(), filePath)
}

type opener interface {
	Open() (io.ReadCloser, error)
	ModTime() time.Time
	CreatedTime() time.Time
}

func newSevenZipFS(filePath string) (*localFS, error) {
	tmpDir, localFilePath, cleanup, err := stageRemoteFile(filePath)
	if err != nil {
		return nil, err
	}

	// Extract archive
	r, err := sevenzip.OpenReader(localFilePath)
	if err != nil {
		logError(cleanup())
		return nil, err
	}
	defer r.Close()

	// Iterate through the files in the archive,
	// printing some of their contents.
	for _, f := range r.File {
		name := sanitizeFileName(f.Name)
		filePath := filepath.Join(tmpDir, name)
		if f.FileInfo().IsDir() {
			err = createDirectory(filePath)
			if err != nil {
				logError(cleanup())
				return nil, fmt.Errorf("cannot create directory: %s, %w", filePath, err)
			}
		} else {
			dirPath := filepath.Dir(filePath)
			err = createDirectory(dirPath)
			if err != nil {
				logError(cleanup())
				return nil, fmt.Errorf("cannot create directory: %s, %w", dirPath, err)
			}

			err := writeFileFromArchiveEntry(&sevenZipOpener{f: f}, filePath)
			if err != nil {
				logError(cleanup())
				return nil, fmt.Errorf("cannot write zip file entry: %s, %w", name, err)
			}
		}
	}

	return newLocalFS(tmpDir, cleanup)
}

type sevenZipOpener struct {
	f *sevenzip.File
}

func (s *sevenZipOpener) Open() (io.ReadCloser, error) {
	return s.f.Open()
}

func (s *sevenZipOpener) ModTime() time.Time {
	return s.f.Modified
}

func (s *sevenZipOpener) CreatedTime() time.Time {
	return s.f.Created
}

func coerceToReaderAt(file fs.File) (io.ReaderAt, error) {
	readerAt, ok := file.(io.ReaderAt)
	if ok {
		return readerAt, nil
	} else {
		// TODO: This is super inefficient because it's reading a nested zip file into memory.
		data, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(data), err
	}
}

func archiverFileSystemFromArchive(file fs.File) (fs.FS, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	readerAt, err := coerceToReaderAt(file)
	if err != nil {
		return nil, err
	}

	format, _, err := archiver.Identify(stat.Name(), file)
	if err != nil && !errors.Is(err, archiver.ErrNoMatch) {
		return nil, err
	}
	if format != nil {
		// TODO: we only really need Extractor and Decompressor here, not the combined interfaces...
		if af, ok := format.(archiver.Archival); ok {
			r := io.NewSectionReader(readerAt, 0, stat.Size())
			return archiver.ArchiveFS{Stream: r, Format: af}, nil
		}
	}
	return nil, fmt.Errorf("archive not recognized")
}
