package gowebserver

import (
	"io"
	"io/fs"
	"io/ioutil"
)

type FileSystem interface {
	fs.FS
	fs.StatFS
	fs.ReadFileFS
	fs.ReadDirFS
	io.Closer
	subFSGetter
}

type concreteFS struct {
	baseFS       fs.FS
	asReadFileFS fs.ReadFileFS
	asStatFS     fs.StatFS
	asReadDirFS  fs.ReadDirFS
	asGlobFS     fs.GlobFS
	cleanup      func() error
}

func (c *concreteFS) Close() error {
	return c.cleanup()
}

func (c *concreteFS) Open(name string) (fs.File, error) {
	return c.baseFS.Open(cleanPath(name))
}

func (c *concreteFS) Stat(name string) (fs.FileInfo, error) {
	if c.asStatFS != nil {
		return c.asStatFS.Stat(cleanPath(name))
	}
	f, err := c.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.Stat()
}

func (c *concreteFS) ReadFile(name string) ([]byte, error) {
	if c.asReadFileFS != nil {
		return c.asReadFileFS.ReadFile(cleanPath(name))
	}

	f, err := c.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func (c *concreteFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if c.asReadDirFS != nil {
		return c.asReadDirFS.ReadDir(cleanPath(name))
	}

	f, err := c.Open(name)
	if err != nil {
		return nil, err
	}

	readDirFile, ok := f.(fs.ReadDirFile)
	if !ok {
		return nil, fs.ErrInvalid
	}
	return readDirFile.ReadDir(-1)
}

func newConcreteFS(baseFS fs.FS, close func() error) *concreteFS {
	asReadFileFS, ok := baseFS.(fs.ReadFileFS)
	if !ok {
		asReadFileFS = nil
	}
	asStatFS, ok := baseFS.(fs.StatFS)
	if !ok {
		asStatFS = nil
	}
	asReadDirFS, ok := baseFS.(fs.ReadDirFS)
	if !ok {
		asReadDirFS = nil
	}
	asGlobFS, ok := baseFS.(fs.GlobFS)
	if !ok {
		asGlobFS = nil
	}
	return &concreteFS{
		baseFS:       baseFS,
		asReadFileFS: asReadFileFS,
		asStatFS:     asStatFS,
		asReadDirFS:  asReadDirFS,
		asGlobFS:     asGlobFS,
		cleanup:      close,
	}
}
