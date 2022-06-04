package gowebserver

import (
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

	r, err := subFS.Open(joinNestedFSPath(paths[1:]))
	logError(err)
	return r, err
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
