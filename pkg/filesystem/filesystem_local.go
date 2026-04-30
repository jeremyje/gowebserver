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
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type localFS struct {
	*concreteFS
	absPath string
	relPath string
	cleanup func() error
}

func cleanPath(path string) string {
	return strings.ReplaceAll(filepath.Clean(path), "\\", "/")
}

func (l *localFS) getFS(name string) (FileSystem, error) {
	absPath := filepath.Join(l.absPath, name)
	return newArchiveFSFromLocalPath(absPath)
}

func (l *localFS) Open(name string) (fs.File, error) {
	return l.concreteFS.Open(name)
}

func (l *localFS) Stat(name string) (fs.FileInfo, error) {
	return l.concreteFS.Stat(name)
}

func (l *localFS) ReadFile(name string) ([]byte, error) {
	return l.concreteFS.ReadFile(name)
}

func (l *localFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return sortDirEntryListWithError(l.concreteFS.ReadDir(name))
}

func newLocalFS(relPath string, cleanup func() error) (*localFS, error) {
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		return nil, err
	}

	baseFS := os.DirFS(cleanPath(dirPath(absPath)))
	return &localFS{
		concreteFS: newConcreteFS(baseFS, cleanup),
		absPath:    absPath,
		relPath:    relPath,
		cleanup:    cleanup,
	}, nil
}

func sortDirEntryListWithError(original []fs.DirEntry, err error) ([]fs.DirEntry, error) {
	if err != nil {
		logError(err)
		return nil, err
	}

	return augmentDirEntryList(original, false), nil
}

func augmentDirEntryListWithError(original []fs.DirEntry, err error) ([]fs.DirEntry, error) {
	if err != nil {
		logError(err)
		return nil, err
	}

	return augmentDirEntryList(original, true), nil
}
