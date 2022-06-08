package gowebserver

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

	fs := os.DirFS(cleanPath(dirPath(absPath)))
	return &localFS{
		concreteFS: newConcreteFS(fs, cleanup),
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
