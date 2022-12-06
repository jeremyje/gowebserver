package gowebserver

import (
	"context"
	"io/fs"

	archiver "github.com/mholt/archiver/v4"
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
	fs, err := archiver.FileSystem(ctx, name)
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
