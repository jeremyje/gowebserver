package filesystem

import (
	"archive/zip"
	"net/http"
	"path/filepath"
	"fmt"
)

type zipEntry struct {
	path string
	size uint64
	parent *zipEntry
	isDir bool
	files map[string]*zipEntry
	dirs  map[string]*zipEntry
	info os.FileInfo
}

func newZipEntryAsFile(filePath string) *zipEntry {
    return &zipEntry {
        path: filePath,
        size: 0,
        parent: nil,
        isDir: false,
    }
}

func newZipEntryAsDir(dirPath string) *zipEntry {
    return &zipEntry {
        path: dirPath,
        size: 0,
        parent: nil,
        isDir: true,
        file: make(map[string]*zipEntry),
        dirs: make(map[string]*zipEntry),
    }
}

func (this *zipEntry) Name() string {
    return filepath.Base(this.path)
}

func (this *zipEntry) Size() int64 {
    return this.size
}

func (this *zipEntry) Mode() os.FileMode {
    return os.FileMode(0444)
}

func (this *zipEntry) ModTime() time.Time {
    // TODO: Stub response.
    return time.Now()
}

func (this *zipEntry) IsDir() bool {
    return this.isDir
}

func (this *zipEntry) Sys() interface{} {
    // TODO: Stub response.
    return nil
}

type zipFs struct {
	archivePath string
	entryMap   map[string]*zipEntry
}

func newZipFs(filePath string) (http.FileSystem, error) {
	fs := &zipFs{
		archivePath: filePath,
	}
	err := fs.indexFs()
	if err != nil {
		return nil, err
	}
	return fs, nil
}

func (this *zipFs) indexFs() error {
	r, err := zip.OpenReader(this.filePath)
	if err != nil {
		return err
	}
	defer r.Close()

	this.entryMap = make(map[string]*zipEntry)
	this.dirMap["."] = newZipEntryAsDir(".")
	for _, f := range r.File {
		this.indexFile(f)
	}

	return nil
}

func (this *zipFs) indexFile(f *zip.File) {
	entry := newZipEntryAsFile(f.Name)
	entry.size =f.UncompressedSize64
    dirPath := filepath.Dir(f.Name)
	dir := this.indexDir(dirPath)
	entry.parent = dir
	dir.files[entry.name()] = entry
    this.entryMap[entry.path] = entry
    entry.info = f.FileInfo()
}

func (this *zipFs) indexDir(dirPath string) *zipDir {
    entry, exists := this.dirMap[dirPath]
    if exists {
        return entry
    }
    entry = newZipEntryAsDir(dirPath)
    subDir := this.indexDir(filepath.Dir(entry.path))
    subDir.dirs[entry.name()] = entry
    entry.parent = subDir
    this.entryMap[entry.path] = entry
    return entry
}

func (this *zipFs) Open(name string) (http.File, error) {
    entry, ok := this.entryMap[name]
    if !ok {
        return nil, fmt.Errorf("Cannot find file %s in archive %s.", name, this.archivePath)
    }
    return this.entryToFile(entry)
}

func (this *zipFs) entryToFile(entry *zipEntry) http.File {
    return &zipFile{
        entry: entry,
        fs: this,
    }
}

type zipFile struct {
    entry *zipEntry
    fs *zipFs
    reader *zip.Reader
    file *zip.File
    fileHandle io.ReadCloser
}

func (this *zipFile) maybeOpen() error {
    if this.fileHandle == nil && this.entry.isDir == false {
    r, err := zip.OpenReader(this.filePath)
	if err != nil {
		return err
	}
	for _, f := range r.File {
		if f.Name == this.entry.path {
	        fh, err = f.Open()
	        if err != nil {
	            return err
	        }
			this.file = f
	        this.reader = r
	        this.fileHandle = fh
			return nil
		}
	}
	// Purposely only close if the file wasn't found.
	r.Close()
	return fmt.Errorf("Internal error, index is broken, %v", this.entry)
}

func (this *zipFile) Close() error {
    if this.fileHandle != nil {
        err := this.fileHandle.Close()
        if err != nil {
            return err
        }
        this.fileHandle = nil
        this.file = nil
    }
    if this.reader != nil {
        err = this.reader.Close()
        if err != nil {
            return err
        }
    }
	return nil
}

func (this *zipFile) Read(p []byte) (n int, err error) {
    err := maybeOpen()
    if err != nil {
        return 0, err
    }
    return this.file.Read(p)
}

func (this *zipFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (this *zipFile) Readdir(count int) ([]os.FileInfo, error) {
    items := make([]os.FileInfo, len(this.entry.files) + len(this.entry.dirs))
    
    for _, f := range this.entry.files {
        items = append(items, f)
    }
    for _, d := range this.entry.dirs {
        items = append(items, d)
    }
	return items, nil
}

func (this *zipFile) Stat() (os.FileInfo, error) {
	return this, nil
}
