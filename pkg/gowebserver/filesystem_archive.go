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
	"context"
	"io/fs"

	"github.com/mholt/archives"
)

type archiveFS struct {
	*concreteFS
}

func (a *archiveFS) getFS(name string) (FileSystem, error) {
	f, err := a.Open(name)
	if err != nil {
		return nil, err
	}
	return newArchiveFSFromFile(f)
}

func (a *archiveFS) Open(name string) (fs.File, error) {
	return a.concreteFS.Open(name)
}

func (a *archiveFS) Stat(name string) (fs.FileInfo, error) {
	return a.concreteFS.Stat(name)
}

func (a *archiveFS) ReadFile(name string) ([]byte, error) {
	return a.concreteFS.ReadFile(name)
}

func (a *archiveFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return sortDirEntryListWithError(a.concreteFS.ReadDir(name))
}

func newArchiveFSFromLocalPath(name string) (*archiveFS, error) {
	ctx := context.Background()
	fs, err := archives.FileSystem(ctx, name, nil)
	if err != nil {
		return nil, err
	}

	return &archiveFS{
		concreteFS: newConcreteFS(fs, func() error { return nil }),
	}, nil
}

func newArchiveFSFromFile(f fs.File) (*archiveFS, error) {
	fs, err := archiverFileSystemFromArchive(f)
	if err != nil {
		return nil, err
	}
	return &archiveFS{
		concreteFS: newConcreteFS(fs, f.Close),
	}, nil
}
