package gowebserver

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	nestedDirSuffix = "-dir"
)

type subFSGetter interface {
	getFS(name string) (FileSystem, error)
}

type nestedFS struct {
	mapped map[string]*nestedFS
	baseFS FileSystem
}

func (n *nestedFS) Close() error {
	var lastErr error
	for _, v := range n.mapped {
		if err := v.Close(); err != nil {
			lastErr = err
		}
	}

	n.mapped = map[string]*nestedFS{}

	return lastErr
}

func (n *nestedFS) Open(name string) (fs.File, error) {
	paths := splitNestedFSPath(name)
	switch len(paths) {
	case 0:
		return nil, os.ErrNotExist
	case 1:
		return augmentFSFileDir(n.baseFS.Open(name))
	}
	subFS, err := n.getFS(paths[0])
	if err != nil {
		return nil, err
	}

	return augmentFSFile(subFS.Open(joinNestedFSPath(paths[1:])))
}

type augmentedFile struct {
	f                fs.File
	requiresSeekRead bool
	reader           *bytes.Reader
	data             []byte
	readErr          error
}

func (a *augmentedFile) Stat() (fs.FileInfo, error) {
	return a.f.Stat()
}

func (a *augmentedFile) Read(d []byte) (int, error) {
	if a.requiresSeekRead {
		r, err := a.getReader()
		if err != nil {
			return 0, err
		}
		return r.Read(d)
	}
	return a.f.Read(d)
}

func (a *augmentedFile) Close() error {
	a.data = nil
	a.readErr = nil
	return a.f.Close()
}

func (a *augmentedFile) getReader() (*bytes.Reader, error) {
	if a.reader != nil || a.readErr != nil {
		return a.reader, a.readErr
	}
	data, err := io.ReadAll(a.f)
	a.data = data
	a.readErr = err
	a.reader = bytes.NewReader(data)
	return a.reader, err
}

func (a *augmentedFile) Seek(offset int64, whence int) (int64, error) {
	seeker, ok := a.f.(io.Seeker)
	if ok {
		return seeker.Seek(offset, whence)
	}

	r, err := a.getReader()
	if err != nil {
		return 0, err
	}

	return r.Seek(offset, whence)
}

func (a *augmentedFile) ReadAt(p []byte, off int64) (n int, err error) {
	readerAt, ok := a.f.(io.ReaderAt)
	if ok {
		return readerAt.ReadAt(p, off)
	}

	r, err := a.getReader()
	if err != nil {
		return 0, err
	}

	return r.ReadAt(p, off)
}

func augmentFSFile(f fs.File, err error) (fs.File, error) {
	if err != nil {
		return f, err
	}

	if _, ok := f.(fs.ReadDirFile); ok {
		return f, err
	}

	if _, ok := f.(*augmentedFile); ok {
		return f, nil
	}

	if _, ok := f.(io.ReadSeeker); ok {
		if _, ok := f.(io.ReaderAt); ok {
			return &augmentedFile{
				f: f,
			}, nil
		}
	}
	return &augmentedFile{
		f:                f,
		requiresSeekRead: true,
	}, nil
}

func augmentFSFileDir(f fs.File, err error) (fs.File, error) {
	if err != nil {
		logError(err)
		return f, err
	}
	fa, err := f.Stat()
	if err == nil {
		if fa.IsDir() {
			dir, ok := f.(fs.ReadDirFile)
			if ok {
				return &fsReadDirFile{dir: dir}, nil
			}
		}
	}

	return f, nil
}

type fsReadDirFile struct {
	dir fs.ReadDirFile
}

func (f *fsReadDirFile) Stat() (fs.FileInfo, error) {
	return f.dir.Stat()
}

func (f *fsReadDirFile) Read(data []byte) (int, error) {
	return f.dir.Read(data)
}

func (f *fsReadDirFile) Close() error {
	return f.dir.Close()
}

func (f *fsReadDirFile) ReadDir(n int) ([]fs.DirEntry, error) {
	return augmentDirEntryListWithError(f.dir.ReadDir(n))
}

func (n *nestedFS) getFS(name string) (FileSystem, error) {
	name = strings.TrimSuffix(name, nestedDirSuffix)
	subFS, ok := n.mapped[name]
	if ok {
		return subFS, nil
	}
	if subFS, err := n.baseFS.getFS(name); err != nil {
		logError(err)
		return nil, err
	} else {
		nestedSubFS := newNestedFS(subFS)
		n.mapped[name] = nestedSubFS
		return nestedSubFS, nil
	}
}

func (n *nestedFS) Stat(name string) (fs.FileInfo, error) {
	paths := splitNestedFSPath(name)
	switch len(paths) {
	case 0:
		return nil, os.ErrNotExist
	case 1:
		return n.baseFS.Stat(name)
	}
	subFS, err := n.getFS(paths[0])
	if err != nil {
		return nil, err
	}
	r, err := subFS.Stat(joinNestedFSPath(paths[1:]))
	logError(err)
	return r, err
}

func (n *nestedFS) ReadFile(name string) ([]byte, error) {
	paths := splitNestedFSPath(name)
	switch len(paths) {
	case 0:
		return nil, os.ErrNotExist
	case 1:
		return n.baseFS.ReadFile(name)
	}
	subFS, err := n.getFS(paths[0])
	if err != nil {
		logError(err)
		return nil, err
	}
	r, err := subFS.ReadFile(joinNestedFSPath(paths[1:]))
	logError(err)
	return r, err
}

func (n *nestedFS) ReadDir(name string) ([]fs.DirEntry, error) {
	paths := splitNestedFSPath(name)
	switch len(paths) {
	case 0:
		return nil, os.ErrNotExist
	case 1:
		return augmentDirEntryListWithError(n.baseFS.ReadDir(name))
	}
	subFS, err := n.getFS(paths[0])
	if err != nil {
		logError(err)
		return nil, err
	}
	r, err := subFS.ReadDir(joinNestedFSPath(paths[1:]))
	logError(err)
	return r, err
}

func newNestedFS(baseFS FileSystem) *nestedFS {
	return &nestedFS{
		mapped: map[string]*nestedFS{},
		baseFS: baseFS,
	}
}

func logError(err error) {
	if err != nil {
		zap.S().Errorf("%s", err)
	}
}

func augmentDirEntryList(original []fs.DirEntry, addVirtualDirs bool) []fs.DirEntry {
	sortedNames := []string{}
	entryMap := map[string]fs.DirEntry{}
	for _, entry := range original {
		sortedNames = append(sortedNames, entry.Name())
		entryMap[entry.Name()] = entry
		if addVirtualDirs && isSupportedArchive(entry.Name()) {
			dirEntry := newVirtualDirEntry(entry)
			entryMap[dirEntry.Name()] = dirEntry
			sortedNames = append(sortedNames, dirEntry.Name())
		}
	}

	sort.Strings(sortedNames)

	result := []fs.DirEntry{}
	for _, name := range sortedNames {
		result = append(result, entryMap[name])
	}
	return result
}

func newVirtualDirEntry(entry fs.DirEntry) fs.DirEntry {
	name := entry.Name() + nestedDirSuffix
	return &virtualDirEntry{
		name:            name,
		backingDirEntry: entry,
	}
}

type virtualDirEntry struct {
	name            string
	backingDirEntry fs.DirEntry
	cachedModTime   time.Time
}

func (v *virtualDirEntry) Name() string {
	return v.name
}

func (v *virtualDirEntry) IsDir() bool {
	return true
}

func (v *virtualDirEntry) Type() fs.FileMode {
	return fs.ModeDir
}

func (v *virtualDirEntry) Info() (fs.FileInfo, error) {
	return v, nil
}

func (v *virtualDirEntry) Size() int64 {
	return 0
}

func (v *virtualDirEntry) Mode() fs.FileMode {
	return v.Type()
}

func (v *virtualDirEntry) Sys() any {
	return v.backingDirEntry
}

var (
	emptyTime time.Time
)

func (v *virtualDirEntry) ModTime() time.Time {
	if v.cachedModTime == emptyTime {
		info, err := v.backingDirEntry.Info()
		if err == nil && info != nil {
			v.cachedModTime = info.ModTime()
		} else {
			v.cachedModTime = time.Now()
		}
	}

	return v.cachedModTime
}
