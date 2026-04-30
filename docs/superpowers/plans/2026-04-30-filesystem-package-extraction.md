# Filesystem Package Extraction Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extract the `filesystem*.go` files from `pkg/gowebserver/` into an isolated `pkg/filesystem/` library, untangling any `net/http` couplings in the process.

**Architecture:** Create `pkg/filesystem` as a standalone `fs.FS`-based library with no `net/http` dependency. The only HTTP coupling (`newHandlerFromFS`) stays in `pkg/gowebserver/filesystem.go` as a thin adapter that calls into `pkg/filesystem`. Filesystem-specific utilities (temp dirs, archive extraction, sanitize) move with the code; template/web utilities stay in `pkg/gowebserver`.

**Tech Stack:** Go 1.24+, `github.com/mholt/archives`, `github.com/bodgit/sevenzip`, `github.com/go-git/go-git/v5`, `go.uber.org/zap`

---

## File Map

### New files (create)
| File | Responsibility |
|------|---------------|
| `pkg/filesystem/filesystem.go` | `FileSystem` interface + `concreteFS` (from gowebserver/filesystem_common.go) |
| `pkg/filesystem/filesystem_archive.go` | `archiveFS`, `newArchiveFSFromLocalPath`, `newArchiveFSFromFile` |
| `pkg/filesystem/filesystem_common.go` | `NewRawFSFromURI`, archive helpers, `splitNestedFSPath`, `joinNestedFSPath` |
| `pkg/filesystem/filesystem_git.go` | `newGitFS`, `IsSupportedGit`, `cloneOptions` (build tag: `!aix`) |
| `pkg/filesystem/filesystem_git_unsupported.go` | aix stub for `newGitFS`/`IsSupportedGit` |
| `pkg/filesystem/filesystem_local.go` | `localFS`, `cleanPath`, `newLocalFS`, sort helpers |
| `pkg/filesystem/filesystem_nested.go` | `nestedFS`, `NewNestedFS`, `augmentedFile`, virtual dir entries |
| `pkg/filesystem/util.go` | `createDirectory`, `stageRemoteFile`, `createTempDirectory`, `tryDelete*`, `downloadFile`, `exists`, `dirPath`, `copyFile`, `SanitizeFileName`, `nilFuncWithError`, `nilFunc`, `logError` |
| `pkg/filesystem/filesystem_test.go` | Tests for FileSystem implementations (ported from gowebserver) |
| `pkg/filesystem/util_test.go` | Tests for util functions (ported from gowebserver/util_test.go) |

### Modified files
| File | Change |
|------|--------|
| `pkg/gowebserver/filesystem.go` | Replace with thin adapter: only `newHandlerFromFS`, calling `pkg/filesystem` |
| `pkg/gowebserver/util.go` | Remove filesystem-specific functions; keep template/web/media helpers |
| `pkg/gowebserver/upload.go` | Use `filesystem.SanitizeFileName` instead of local `sanitizeFileName` |
| `pkg/gowebserver/util_test.go` | Remove tests for moved functions; update `mustTempDir` |

### Deleted files (from pkg/gowebserver/)
`filesystem_common.go`, `filesystem_archive.go`, `filesystem_git.go`, `filesystem_git_unsupported.go`, `filesystem_local.go`, `filesystem_nested.go`, `filesystem_test.go`

---

### Task 1: Create `pkg/filesystem/filesystem.go` — FileSystem interface and concreteFS

**Files:**
- Create: `pkg/filesystem/filesystem.go`

This is the core interface file. Content is lifted verbatim from `pkg/gowebserver/filesystem_common.go` with only the package declaration changed.

- [ ] **Step 1: Create the file**

```go
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
	"io"
	"io/fs"
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
	return io.ReadAll(f)
}

func (c *concreteFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if c.asReadDirFS != nil {
		return c.asReadDirFS.ReadDir(cleanPath(name))
	}

	f, err := c.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

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
```

- [ ] **Step 2: Verify it compiles (will fail due to missing dependencies, that's OK)**

```bash
cd /home/coder/project/gowebserver && go build ./pkg/filesystem/... 2>&1 | head -20
```

---

### Task 2: Create `pkg/filesystem/util.go` — filesystem utilities

**Files:**
- Create: `pkg/filesystem/util.go`

All filesystem-specific utilities from `pkg/gowebserver/util.go` move here. `SanitizeFileName` is exported (capital S) because `upload.go` in gowebserver also needs it. `logError` moves here too (from filesystem_nested.go).

- [ ] **Step 1: Create the file**

```go
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
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

const fsDirMode = os.FileMode(0777)

func createDirectory(path string) error {
	return os.MkdirAll(dirPath(path), fsDirMode)
}

func stageRemoteFile(maybeRemoteFilePath string) (string, string, func() error, error) {
	localFilePath, fileCleanup, err := downloadFile(maybeRemoteFilePath)
	if err != nil {
		fileCleanup()
		return "", "", nilFuncWithError, fmt.Errorf("cannot download file %s, %w", maybeRemoteFilePath, err)
	}

	tmpDir, cleanup, err := createTempDirectory()
	if err != nil {
		logError(fileCleanup())
		cleanup()
		return "", "", nilFuncWithError, fmt.Errorf("cannot create temp directory, %w", err)
	}

	return tmpDir, localFilePath, func() error {
		cleanup()
		return fileCleanup()
	}, nil
}

func createTempDirectory() (string, func(), error) {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "gowebserver")

	if err != nil {
		return "", nilFunc, fmt.Errorf("cannot create temp directory, %w", err)
	}
	return tmpDir, func() {
		tryDeleteDirectory(tmpDir)
	}, nil
}

func tryDeleteDirectory(path string) {
	if !exists(path) {
		return
	}

	if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
		zap.S().With("error", err, "directory", path).Error("cannot delete directory")
	}
}

func tryDeleteFile(path string) {
	if !exists(path) {
		return
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		zap.S().With("error", err, "file", path).Error("cannot delete file")
	}
}

func downloadFile(path string) (string, func() error, error) {
	if strings.HasPrefix(strings.ToLower(path), "http") {
		cleanup := nilFuncWithError
		f, err := os.CreateTemp(os.TempDir(), "gowebserverdl")
		if err != nil {
			return "", cleanup, err
		}
		defer f.Close()
		fileName := f.Name()
		cleanup = func() error {
			return os.Remove(fileName)
		}
		resp, err := http.Get(path)
		if err != nil {
			return "", cleanup, err
		}
		defer resp.Body.Close()

		if _, err := io.Copy(f, resp.Body); err != nil {
			return "", cleanup, err
		}
		return f.Name(), cleanup, nil
	}
	return path, nilFuncWithError, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirPath(dirPath string) string {
	return strings.TrimRight(dirPath, "/") + "/"
}

func copyFile(reader io.Reader, createdTime time.Time, modifiedTime time.Time, filePath string) error {
	fsf, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("cannot create target file %s, %w", filePath, err)
	}
	defer fsf.Close()

	_, err = io.Copy(fsf, reader)
	if err != nil {
		os.Remove(fsf.Name())
		return fmt.Errorf("cannot copy to target file %s, %w", filePath, err)
	}
	return os.Chtimes(filePath, createdTime, modifiedTime)
}

var (
	validChars = map[rune]interface{}{
		'.':  nil,
		'-':  nil,
		'_':  nil,
		' ':  nil,
		'\\': nil,
		'/':  nil,
		'$':  nil,
		'#':  nil,
	}
)

// SanitizeFileName removes path traversal and unsafe characters from a file name.
func SanitizeFileName(fileName string) string {
	name := strings.ReplaceAll(filepath.Clean(fileName), "..", ".")
	sanitized := ""
	for _, r := range name {
		if _, ok := validChars[r]; ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9') || ok {
			sanitized = sanitized + string(r)
		}
	}

	parts := strings.Split(strings.ReplaceAll(sanitized, "\\", "/"), "/")
	sanitizedParts := []string{}
	for _, part := range parts {
		if strings.ReplaceAll(part, ".", "") != "" {
			sanitizedParts = append(sanitizedParts, part)
		}
	}

	return filepath.Clean(strings.Join(sanitizedParts, string(filepath.Separator)))
}

func nilFunc() {
}

func nilFuncWithError() error {
	return nil
}

func logError(err error) {
	if err != nil {
		zap.S().Errorf("%s", err)
	}
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd /home/coder/project/gowebserver && go build ./pkg/filesystem/... 2>&1 | head -20
```

---

### Task 3: Create `pkg/filesystem/filesystem_local.go`

**Files:**
- Create: `pkg/filesystem/filesystem_local.go`

Content is identical to gowebserver's version except the package name and it uses the local `newArchiveFSFromLocalPath` (defined in the same package).

- [ ] **Step 1: Create the file**

```go
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
```

Note: `newLocalFS` used `fs` as a variable name which shadows the import. Changed to `baseFS`.

- [ ] **Step 2: Verify it compiles**

```bash
cd /home/coder/project/gowebserver && go build ./pkg/filesystem/... 2>&1 | head -20
```

---

### Task 4: Create `pkg/filesystem/filesystem_archive.go`

**Files:**
- Create: `pkg/filesystem/filesystem_archive.go`

Content matches gowebserver's version exactly, package name changed.

- [ ] **Step 1: Create the file**

```go
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
	baseFS, err := archives.FileSystem(ctx, name, nil)
	if err != nil {
		return nil, err
	}

	return &archiveFS{
		concreteFS: newConcreteFS(baseFS, func() error { return nil }),
	}, nil
}

func newArchiveFSFromFile(f fs.File) (*archiveFS, error) {
	baseFS, err := archiverFileSystemFromArchive(f)
	if err != nil {
		return nil, err
	}
	return &archiveFS{
		concreteFS: newConcreteFS(baseFS, f.Close),
	}, nil
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd /home/coder/project/gowebserver && go build ./pkg/filesystem/... 2>&1 | head -20
```

---

### Task 5: Create `pkg/filesystem/filesystem_common.go` — archive helpers and NewRawFSFromURI

**Files:**
- Create: `pkg/filesystem/filesystem_common.go`

This is the "entry point" of the library. `NewRawFSFromURI` (exported) replaces the old `newRawFSFromURI`. The `newHandlerFromFS` function does NOT move here — it stays in gowebserver.

- [ ] **Step 1: Create the file**

```go
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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/bodgit/sevenzip"
	"github.com/mholt/archives"
)

var (
	archiveExtList = []string{".tar", ".tar.gz", ".tar.bz2", ".tar.xz", ".tar.lz4", ".tar.br", ".tar.zst", ".rar", ".zip", ".7z"}
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

// NewRawFSFromURI creates a FileSystem for the given URI.
// Supports local directories, zip/tar/rar/.7z archives, and git repositories.
func NewRawFSFromURI(path string) (FileSystem, error) {
	if isSupportedSevenZip(path) {
		return newSevenZipFS(path)
	} else if isSupportedArchive(path) {
		return newArchiveFSFromLocalPath(path)
	} else if IsSupportedGit(path) {
		return newGitFS(path)
	}
	return newLocalFS(path, func() error { return nil })
}

func isSupportedArchive(filePath string) bool {
	for _, suffix := range archiveExtList {
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

	r, err := sevenzip.OpenReader(localFilePath)
	if err != nil {
		logError(cleanup())
		return nil, err
	}
	defer r.Close()

	for _, f := range r.File {
		name := SanitizeFileName(f.Name)
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
	}
	// This is very inefficient because it's reading a nested zip file into memory.
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), err
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
	ctx := context.Background()
	format, _, err := archives.Identify(ctx, stat.Name(), file)
	if err != nil && !errors.Is(err, archives.NoMatch) {
		return nil, err
	}
	if format != nil {
		if af, ok := format.(archives.Archival); ok {
			r := io.NewSectionReader(readerAt, 0, stat.Size())
			return archives.ArchiveFS{
				Stream: r,
				Format: af,
			}, nil
		}
	}
	return nil, fmt.Errorf("archive not recognized")
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd /home/coder/project/gowebserver && go build ./pkg/filesystem/... 2>&1 | head -20
```

---

### Task 6: Create `pkg/filesystem/filesystem_nested.go`

**Files:**
- Create: `pkg/filesystem/filesystem_nested.go`

`logError` moves to `util.go` (Task 2). `NewNestedFS` is the exported constructor (returns `FileSystem`). The `nestedFS` type stays unexported. `verifyFileSystem` in the test will be updated to take `FileSystem` instead of `*nestedFS`.

- [ ] **Step 1: Create the file**

```go
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
	"bytes"
	"io"
	"io/fs"
	"os"
	"sort"
	"strings"
	"time"
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
		nestedSubFS := newNestedFSInternal(subFS)
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

// NewNestedFS wraps a FileSystem so that archives inside it are accessible as virtual subdirectories.
// Archive files appear as both files and as "<name>-dir" virtual directories that can be navigated into.
func NewNestedFS(baseFS FileSystem) FileSystem {
	return newNestedFSInternal(baseFS)
}

func newNestedFSInternal(baseFS FileSystem) *nestedFS {
	return &nestedFS{
		mapped: map[string]*nestedFS{},
		baseFS: baseFS,
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
```

- [ ] **Step 2: Verify it compiles**

```bash
cd /home/coder/project/gowebserver && go build ./pkg/filesystem/... 2>&1 | head -20
```

---

### Task 7: Create `pkg/filesystem/filesystem_git.go` and `filesystem_git_unsupported.go`

**Files:**
- Create: `pkg/filesystem/filesystem_git.go`
- Create: `pkg/filesystem/filesystem_git_unsupported.go`

`isSupportedGit` is exported as `IsSupportedGit` (gowebserver needs it in `newHandlerFromFS`).

- [ ] **Step 1: Create `pkg/filesystem/filesystem_git.go`**

```go
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

//go:build !aix
// +build !aix

package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func newGitFS(filePath string) (*localFS, error) {
	if !IsSupportedGit(filePath) {
		return nil, fmt.Errorf("%s is not a valid git repository", filePath)
	}

	tmpDir, cleanup, err := createTempDirectory()
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("cannot create temp directory, %w", err)
	}

	for _, opts := range cloneOptions(filePath) {
		if _, err = git.PlainClone(tmpDir, false, opts); err == nil {
			break
		}
	}
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("could not clone %s, %w", filePath, err)
	}

	tryDeleteDirectory(filepath.Join(tmpDir, ".git"))
	tryDeleteFile(filepath.Join(tmpDir, ".gitignore"))
	tryDeleteFile(filepath.Join(tmpDir, ".gitmodules"))
	return newLocalFS(tmpDir, func() error {
		cleanup()
		return nil
	})
}

// IsSupportedGit returns true if filePath is a git repository URL (ends in .git).
func IsSupportedGit(filePath string) bool {
	return strings.HasSuffix(strings.ToLower(filePath), ".git")
}

func cloneOptions(filePath string) []*git.CloneOptions {
	return []*git.CloneOptions{
		{
			URL:          filePath,
			Progress:     os.Stdout,
			Depth:        1,
			SingleBranch: true,
		},
		{
			URL:           filePath,
			Progress:      os.Stdout,
			Depth:         1,
			SingleBranch:  true,
			ReferenceName: plumbing.NewBranchReferenceName("main"),
		},
	}
}
```

- [ ] **Step 2: Create `pkg/filesystem/filesystem_git_unsupported.go`**

```go
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

//go:build aix
// +build aix

package filesystem

import (
	"fmt"
)

func newGitFS(filePath string) (*localFS, error) {
	return nil, fmt.Errorf("%s is not a valid git repository", filePath)
}

// IsSupportedGit always returns false on aix.
func IsSupportedGit(filePath string) bool {
	return false
}
```

- [ ] **Step 3: Verify the full package compiles**

```bash
cd /home/coder/project/gowebserver && go build ./pkg/filesystem/... 2>&1
```

Expected: no errors.

---

### Task 8: Create `pkg/filesystem/filesystem_test.go`

**Files:**
- Create: `pkg/filesystem/filesystem_test.go`

Ported from `pkg/gowebserver/filesystem_test.go`. Key changes:
- Package name: `filesystem` (same package, can access unexported types)
- `isSupportedHTTP` test removed (that function lives in gowebserver)
- `verifyFileSystem` signature changes from `*nestedFS` to `FileSystem`
- `newRawFSFromURI` → `NewRawFSFromURI` (now exported)
- `newNestedFS` → `NewNestedFS` (returns `FileSystem`)

- [ ] **Step 1: Create `pkg/filesystem/filesystem_test.go`**

```go
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
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	gowsTesting "github.com/jeremyje/gowebserver/v2/internal/gowebserver/testing"
)

var (
	zeroTime               = time.Time{}
	commonFSRootDirList    = []string{"assets", "bytype", "index.html", "site.js", "weird #1.txt", "weird#.txt", "weird$.txt"}
	nestedZipFSRootDirList = []string{"single-testassets.zip", "single-testassets.zip-dir", "testassets", "testassets.7z", "testassets.7z-dir", "testassets.rar", "testassets.rar-dir", "testassets.tar", "testassets.tar-dir", "testassets.tar.bz2", "testassets.tar.bz2-dir", "testassets.tar.gz", "testassets.tar.gz-dir", "testassets.tar.lz4", "testassets.tar.lz4-dir", "testassets.tar.xz", "testassets.tar.xz-dir", "testassets.zip", "testassets.zip-dir", "testing.go", "testing_test.go"}
)

var (
	_ FileSystem    = (*localFS)(nil)
	_ FileSystem    = (*archiveFS)(nil)
	_ FileSystem    = (*nestedFS)(nil)
	_ io.ReaderAt   = (*augmentedFile)(nil)
	_ io.ReadSeeker = (*augmentedFile)(nil)
)

func TestDirlessArchive(t *testing.T) {
	nodirZipPath := gowsTesting.MustNoDirZipFilePath(t)

	vFS, err := NewRawFSFromURI(nodirZipPath)
	if err != nil {
		t.Error(err)
	}
	defer vFS.Close()
	nFS := NewNestedFS(vFS)
	defer nFS.Close()

	verifyReadDir(t, nFS, "assets", []string{"1.txt", "2.txt", "fivesix", "four", "more"})
}

func TestVirtualDirectory(t *testing.T) {
	nestedZipPath := gowsTesting.MustNestedZipFilePath(t)

	vFS, err := NewRawFSFromURI(nestedZipPath)
	if err != nil {
		t.Error(err)
	}
	defer vFS.Close()
	nFS := NewNestedFS(vFS)
	defer nFS.Close()

	dirs, err := nFS.ReadDir("")
	if err != nil {
		t.Fatal(err)
	}

	hasVirtual := false
	for _, dir := range dirs {
		if dir.IsDir() {
			if diff := cmp.Diff(fs.ModeDir, dir.Type()); diff != "" {
				t.Errorf("Type() mismatch (-want +got):\n%s", diff)
			}
			info, err := dir.Info()
			if err != nil {
				t.Errorf("dir.Info() error= %s", err)
			}
			if info == nil {
				t.Error("Info() is nil")
			}

			if strings.HasSuffix(dir.Name(), nestedDirSuffix) {
				hasVirtual = true
				vDir, ok := dir.(*virtualDirEntry)
				if ok {
					if diff := cmp.Diff(int64(0), vDir.Size()); diff != "" {
						t.Errorf("Size() mismatch (-want +got):\n%s", diff)
					}

					if diff := cmp.Diff(fs.ModeDir, vDir.Mode()); diff != "" {
						t.Errorf("Mode() mismatch (-want +got):\n%s", diff)
					}

					if vDir.Sys() == nil {
						t.Errorf("%s.Sys() is nil", vDir.Name())
					}

					if vDir.ModTime().Before(emptyTime) {
						t.Errorf("ModTime() is before empty time, %s", vDir.ModTime())
					}
				} else {
					t.Errorf("%s is not a virtual directory but has the suffix", dir.Name())
				}
			}
		}
	}

	if !hasVirtual {
		t.Error("no virtual directories were verified")
	}
}

func TestNestedFileSystem(t *testing.T) {
	zipPath := gowsTesting.MustZipFilePath(t)
	rarPath := gowsTesting.MustRarFilePath(t)
	sevenZipPath := gowsTesting.MustSevenZipFilePath(t)
	tarPath := gowsTesting.MustTarFilePath(t)
	tarGzPath := gowsTesting.MustTarGzFilePath(t)
	tarBz2Path := gowsTesting.MustTarBzip2FilePath(t)
	tarXzPath := gowsTesting.MustTarXzFilePath(t)
	tarLz4Path := gowsTesting.MustTarLz4FilePath(t)
	nodirZipPath := gowsTesting.MustNoDirZipFilePath(t)
	singleZipPath := gowsTesting.MustSingleZipFilePath(t)
	nestedZipPath := gowsTesting.MustNestedZipFilePath(t)

	testCases := []struct {
		uri          string
		baseDir      string
		isNested     bool
		wantRootDirs []string
	}{
		{uri: zipPath, baseDir: ""},
		{uri: rarPath, baseDir: ""},
		{uri: tarPath, baseDir: ""},
		{uri: tarGzPath, baseDir: ""},
		{uri: tarBz2Path, baseDir: ""},
		{uri: tarXzPath, baseDir: ""},
		{uri: tarLz4Path, baseDir: ""},

		{uri: sevenZipPath, baseDir: ""},

		{uri: nodirZipPath, baseDir: ""},

		{uri: singleZipPath, baseDir: "testassets", wantRootDirs: []string{"testassets"}},

		{uri: nestedZipPath, baseDir: "testassets", wantRootDirs: nestedZipFSRootDirList},
		{uri: nestedZipPath, baseDir: "single-testassets.zip-dir/testassets", isNested: true, wantRootDirs: nestedZipFSRootDirList},
		{uri: nestedZipPath, baseDir: "testassets.tar-dir", isNested: true, wantRootDirs: nestedZipFSRootDirList},
		{uri: nestedZipPath, baseDir: "testassets.tar.bz2-dir", isNested: true, wantRootDirs: nestedZipFSRootDirList},
		{uri: nestedZipPath, baseDir: "testassets.tar.lz4-dir", isNested: true, wantRootDirs: nestedZipFSRootDirList},
		{uri: nestedZipPath, baseDir: "testassets.zip-dir", isNested: true, wantRootDirs: nestedZipFSRootDirList},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%s %s", tc.uri, tc.baseDir), func(t *testing.T) {
			t.Parallel()

			vFS, err := NewRawFSFromURI(tc.uri)
			if err != nil {
				t.Error(err)
			}
			defer vFS.Close()
			nFS := NewNestedFS(vFS)
			defer nFS.Close()

			baseDir := tc.baseDir
			t.Run("verifyFileSystem", func(t *testing.T) {
				verifyFileSystem(t, nFS, baseDir)
			})

			if !tc.isNested {
				t.Run("verifyReadDir rawFS", func(t *testing.T) {
					verifyReadDir(t, vFS, baseDir, commonFSRootDirList)
				})
			}
			t.Run("verifyReadDir nestedFS", func(t *testing.T) {
				verifyReadDir(t, nFS, baseDir, commonFSRootDirList)
				if tc.wantRootDirs == nil {
					verifyReadDir(t, nFS, "", commonFSRootDirList)
				} else {
					verifyReadDir(t, nFS, "", tc.wantRootDirs)
				}
			})
		})
	}
}

func verifyFileSystem(tb testing.TB, nFS FileSystem, baseDir string) {
	indexFilePath := filepath.Join(baseDir, "index.html")
	fp, err := nFS.Open(indexFilePath)
	if err != nil {
		tb.Fatalf("cannot open '%s', err=%s", indexFilePath, err)
	}

	if stat, err := fp.Stat(); err != nil {
		tb.Errorf("'%s' stat error, %s", indexFilePath, err)
	} else {
		nStat, err := nFS.Stat(indexFilePath)
		if err != nil {
			tb.Errorf("cannot stat from nestedFS, %s", err)
		}
		if diff := cmp.Diff(stat.IsDir(), nStat.IsDir()); diff != "" {
			tb.Errorf("nestedFS.IsDir() mismatch (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff(stat.ModTime(), nStat.ModTime()); diff != "" {
			tb.Errorf("nestedFS.ModTime() mismatch (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff(stat.Mode(), nStat.Mode()); diff != "" {
			tb.Errorf("nestedFS.Mode() mismatch (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff(stat.Name(), nStat.Name()); diff != "" {
			tb.Errorf("nestedFS.Name() mismatch (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff(stat.Size(), nStat.Size()); diff != "" {
			tb.Errorf("nestedFS.Size() mismatch (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff(false, stat.IsDir()); diff != "" {
			tb.Errorf("[%s].Stat.IsDir() mismatch (-want +got):\n%s", stat.Name(), diff)
		}

		if zeroTime.After(stat.ModTime()) {
			tb.Errorf("[%s].Stat.ModTime() should be in the past, %v", stat.Name(), stat.ModTime())
		}

		if diff := cmp.Diff("index.html", stat.Name()); diff != "" {
			tb.Errorf("[%s].Stat.Name() mismatch (-want +got):\n%s", stat.Name(), diff)
		}

		if diff := cmp.Diff(int64(10), stat.Size()); diff != "" {
			tb.Errorf("[%s].Stat.Size() mismatch (-want +got):\n%s", stat.Name(), diff)
		}
	}
	data, err := io.ReadAll(fp)
	if err != nil {
		tb.Fatalf("cannot read '%s', err=%s", indexFilePath, err)
	}

	if diff := cmp.Diff("index.html", string(data)); diff != "" {
		tb.Errorf("index.html from FS mismatch (-want +got):\n%s", diff)
	}

	data, err = nFS.ReadFile(indexFilePath)
	if err != nil {
		tb.Errorf("cannot read %s via nestedFS, %s", indexFilePath, err)
	}
	if diff := cmp.Diff("index.html", string(data)); diff != "" {
		tb.Errorf("index.html from nestedFS mismatch (-want +got):\n%s", diff)
	}

	verifyLocalFileFromDefaultAsset(tb, nFS, baseDir)
}

func verifyReadDir(tb testing.TB, vFS FileSystem, baseDir string, want []string) {
	entries, err := vFS.ReadDir(baseDir)
	if err != nil {
		tb.Fatalf("cannot read directory '%s', %s", baseDir, err)
	}

	names := []string{}
	for _, dir := range entries {
		names = append(names, dir.Name())
	}
	if diff := cmp.Diff(want, names); diff != "" {
		tb.Errorf("ReadDir mismatch (-want +got):\n%s", diff)
	}
}

func TestIsSupported(t *testing.T) {
	testCases := []struct {
		input               string
		isSupportedArchive  bool
		isSupportedGit      bool
		isSupportedSevenZip bool
	}{
		{input: "ok.tar", isSupportedArchive: true},
		{input: "ok.tar.gz", isSupportedArchive: true},
		{input: "ok.tar.bz2", isSupportedArchive: true},
		{input: "ok.tar.xz", isSupportedArchive: true},
		{input: "ok.tar.lz4", isSupportedArchive: true},
		{input: "ok.tar.br", isSupportedArchive: true},
		{input: "ok.tar.zst", isSupportedArchive: true},

		{input: "ok.zip", isSupportedArchive: true},
		{input: "ok.tar.lzma"},
		{input: "ok.7z", isSupportedArchive: true, isSupportedSevenZip: true},
		{input: "ok.7Z", isSupportedArchive: true, isSupportedSevenZip: true},

		{input: "git@github.com:jeremyje/gowebserver.git", isSupportedGit: true},
		{input: "https://github.com/jeremyje/gowebserver.git", isSupportedGit: true},

		{input: "http://www.google.com/"},
		{input: "http://www.google.com.7z", isSupportedArchive: true, isSupportedSevenZip: true},

		{input: ""},
		{input: "/"},
	}

	tf := func(t *testing.T, name string, f func(string) bool, input string, want bool) {
		t.Run(fmt.Sprintf("%s %s", name, input), func(t *testing.T) {
			t.Parallel()
			got := f(input)
			if want != got {
				t.Errorf("want: %v, got: %v", want, got)
			}
		})
	}

	for _, tc := range testCases {
		tc := tc

		tf(t, "isSupportedArchive", isSupportedArchive, tc.input, tc.isSupportedArchive)
		tf(t, "IsSupportedGit", IsSupportedGit, tc.input, tc.isSupportedGit)
		tf(t, "isSupportedSevenZip", isSupportedSevenZip, tc.input, tc.isSupportedSevenZip)
	}
}

func verifyLocalFileFromDefaultAsset(tb testing.TB, vFS fs.FS, baseDir string) {
	for _, fileName := range []string{"index.html", "site.js", "weird$.txt", "weird #1.txt", "weird#.txt", "assets/1.txt", "assets/2.txt", "assets/more/3.txt", "assets/four/4.txt", "assets/fivesix/5.txt", "assets/fivesix/6.txt"} {
		name := fileName
		if baseDir != "" {
			name = filepath.Join(baseDir, name)
		}
		if err := verifyLocalFile(vFS, baseDir, name); err != nil {
			tb.Error(err)
		}
	}

	verifyFileMissing(tb, vFS, "does-not-exist/")
	verifyFileMissing(tb, vFS, "does-not-exist")
}

func verifyLocalFile(vFS fs.FS, baseDir string, assetPath string) error {
	f, err := vFS.Open(assetPath)
	if err != nil {
		return fmt.Errorf("%s does not exist when it's expected to, %s", assetPath, err)
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	wantContent := strings.TrimLeft(strings.TrimPrefix(assetPath, baseDir), "/")
	if string(data) != wantContent {
		return fmt.Errorf("The test asset file does not contain it's relative file path as the body, File= %s, WantBody= '%s', Body= '%s'", assetPath, wantContent, string(data))
	}
	return nil
}

func verifyFileExist(tb testing.TB, vFS fs.FS, assetPath string) {
	f, err := vFS.Open(assetPath)
	if err != nil {
		tb.Errorf("cannot find '%s', %s", assetPath, err)
	}

	if _, err := f.Stat(); err != nil {
		tb.Errorf("cannot stat '%s', %s", assetPath, err)
	}
}

func verifyFileMissing(tb testing.TB, vFS fs.FS, assetPath string) {
	f, err := vFS.Open(assetPath)
	if f != nil {
		tb.Errorf("'%s' file handle is not nil when it should be", assetPath)
	}
	if err == nil {
		tb.Errorf("wanted error for reading '%s' as it should not exist", assetPath)
	}
}

func TestGitFsOverHttp(t *testing.T) {
	runGitFsTest(t, "https://github.com/jeremyje/gowebserver.git")
}

func runGitFsTest(tb testing.TB, path string) {
	vFS, err := newGitFS(path)
	defer func() {
		if err := vFS.Close(); err != nil {
			tb.Error(err)
		}
	}()

	if err != nil {
		tb.Fatal(err)
	}

	verifyFileMissing(tb, vFS, ".gitignore")
	verifyFileMissing(tb, vFS, ".git")
	verifyFileMissing(tb, vFS, ".gitmodules")
	verifyFileExist(tb, vFS, "README.md")
	verifyFileExist(tb, vFS, ".github/dependabot.yml")
	verifyFileExist(tb, vFS, "cmd/gowebserver/gowebserver.go")
}

func TestNestedFSPath(t *testing.T) {
	testCases := []struct {
		input string
		want  []string
	}{
		{input: "", want: []string{""}},
		{input: ".", want: []string{"."}},
		{input: "a/b/c", want: []string{"a/b/c"}},
		{input: "a/b/c.ok/d/e/f.zip", want: []string{"a/b/c.ok/d/e/f.zip"}},
		{input: "./a/b/c.zip", want: []string{"./a/b/c.zip"}},
		{input: "./a/b/c.zip-dir/.d/e/f.tar.gz-dir/g/h/i.txt", want: []string{"./a/b/c.zip-dir", ".d/e/f.tar.gz-dir", "g/h/i.txt"}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			gotSplit := splitNestedFSPath(tc.input)
			if diff := cmp.Diff(tc.want, gotSplit); diff != "" {
				t.Errorf("splitNestedFSPath mismatch (-want +got):\n%s", diff)
			}
			gotJoin := joinNestedFSPath(gotSplit)

			if diff := cmp.Diff(tc.input, gotJoin); diff != "" {
				t.Errorf("joinNestedFSPath mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
```

- [ ] **Step 2: Run the filesystem package tests**

```bash
cd /home/coder/project/gowebserver && go test -race -count=1 -run 'TestNestedFSPath|TestIsSupported|TestVirtualDirectory|TestDirlessArchive' ./pkg/filesystem/... 2>&1
```

Expected: all listed tests PASS.

---

### Task 9: Create `pkg/filesystem/util_test.go`

**Files:**
- Create: `pkg/filesystem/util_test.go`

Ported from the filesystem-util-related portions of `pkg/gowebserver/util_test.go`. `sanitizeFileName` becomes `SanitizeFileName`.

- [ ] **Step 1: Create the file**

```go
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
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestDirPath(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		{input: "/", want: "/"},
		{input: "/abc", want: "/abc/"},
		{input: "/abc/", want: "/abc/"},
		{input: "/abc//////////", want: "/abc/"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got := dirPath(tc.input)
			if tc.want != got {
				t.Errorf("expected: %v, got: %v", tc.want, got)
			}
		})
	}
}

func TestCreateTempDirectory(t *testing.T) {
	dir, cleanup, err := createTempDirectory()
	if err != nil {
		t.Error(err)
	}
	if !exists(dir) {
		t.Errorf("'%s' does not exist when it should", dir)
	}

	if !strings.Contains(dir, "gowebserver") {
		t.Errorf("'%s' does not contain 'gowebserver'", dir)
	}
	cleanup()
	if exists(dir) {
		t.Errorf("'%s' exists when it should not", dir)
	}
}

func TestDownloadFileOnLocalFile(t *testing.T) {
	f, err := os.CreateTemp(os.TempDir(), "gowebserver")
	if err != nil {
		t.Error(err)
	}

	t.Cleanup(func() {
		if err := os.Remove(f.Name()); err != nil {
			t.Errorf("cannot cleanup temp file, %s, %s", f.Name(), err)
		}
	})

	path := f.Name()
	err = os.WriteFile(path, []byte("ok"), os.FileMode(0644))
	if err != nil {
		t.Error(err)
	}

	localPath, cleanup, err := downloadFile(path)
	if localPath != path {
		t.Errorf("want: %v, got: %v", path, localPath)
	}

	if err != nil {
		t.Error(err)
	}

	if !exists(localPath) {
		t.Errorf("local file %s should exist", localPath)
	}

	if err := cleanup(); err != nil {
		t.Error(err)
	}

	if !exists(localPath) {
		t.Errorf("local file %s should exist", localPath)
	}
}

func TestDownloadFileOnHttpsFile(t *testing.T) {
	remotePath := "https://raw.githubusercontent.com/jeremyje/gowebserver/main/Makefile"
	localPath, cleanup, err := downloadFile(remotePath)
	if err != nil {
		t.Error(err)
	}
	if localPath == remotePath {
		t.Errorf("'%s' is the local and remote path, they should be different", localPath)
	}
	if !exists(localPath) {
		t.Errorf("'%s' does not exist locally", localPath)
	}
	if err := cleanup(); err != nil {
		t.Errorf("cannot cleanup file, %s", err)
	}
	if exists(localPath) {
		t.Errorf("'%s' should have been cleaned up", localPath)
	}
}

func TestSanitizeFileName(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		{input: "///////////////\\\\\\\\\\", want: "."},
		{input: "../ok", want: "ok"},
		{input: "/ok/", want: "ok"},
		{input: "../whatever.json", want: "whatever.json"},
		{input: "../what ever!@#$%^&*()+_=-.json", want: "what ever#$_-.json"},
		{input: "../abc/def..tar.gz", want: "abc/def.tar.gz"},
		{input: "./././././../.../..../abc.tar.gz/.....", want: "abc.tar.gz"},
		{input: ".file", want: ".file"},
		{input: "../ok/.file", want: "ok/.file"},
		{input: ".", want: "."},
		{input: "/", want: "."},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got := SanitizeFileName(tc.input)
			if tc.want != got {
				t.Errorf("want: %v, got: %v", tc.want, got)
			}
		})
	}
}

type angryReader struct{}

func (a *angryReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("failure")
}

func TestCopyFileErrors(t *testing.T) {
	testCases := []struct {
		filePath string
		r        io.Reader
		wantErr  string
	}{
		{
			filePath: "dir-does-not-exist/target-file.txt",
			r:        &angryReader{},
			wantErr:  "cannot create target file dir-does-not-exist/target-file.txt, open dir-does-not-exist/target-file.txt: no such file or directory",
		},
		{
			filePath: "target-file.txt",
			r:        &angryReader{},
			wantErr:  "cannot copy to target file target-file.txt, failure",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.filePath, func(t *testing.T) {
			t.Parallel()
			if err := copyFile(tc.r, time.Now(), time.Now(), tc.filePath); err != nil {
				if diff := cmp.Diff(tc.wantErr, err.Error()); diff != "" {
					t.Errorf("copyFile() mismatch (-want +got):\n%s", diff)
				}
			} else {
				t.Error("expected an error")
			}
		})
	}
}
```

- [ ] **Step 2: Run util tests**

```bash
cd /home/coder/project/gowebserver && go test -race -count=1 -run 'TestDirPath|TestCreateTempDirectory|TestDownloadFileOnLocalFile|TestSanitizeFileName|TestCopyFileErrors' ./pkg/filesystem/... 2>&1
```

Expected: all listed tests PASS.

---

### Task 10: Replace `pkg/gowebserver/filesystem.go` with the thin adapter

**Files:**
- Modify: `pkg/gowebserver/filesystem.go`
- Delete: `pkg/gowebserver/filesystem_common.go`
- Delete: `pkg/gowebserver/filesystem_archive.go`
- Delete: `pkg/gowebserver/filesystem_git.go`
- Delete: `pkg/gowebserver/filesystem_git_unsupported.go`
- Delete: `pkg/gowebserver/filesystem_local.go`
- Delete: `pkg/gowebserver/filesystem_nested.go`
- Delete: `pkg/gowebserver/filesystem_test.go`

The existing `pkg/gowebserver/filesystem.go` has two responsibilities: the helper functions (moved) and `newHandlerFromFS` (stays). Replace the entire file with just the adapter.

- [ ] **Step 1: Replace `pkg/gowebserver/filesystem.go`**

Write this exact content (replacing all existing content):

```go
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
	"net/http"

	"github.com/jeremyje/gowebserver/v2/pkg/filesystem"
	"go.opentelemetry.io/otel/trace"
)

func newHandlerFromFS(path string, tp trace.TracerProvider, enhancedList bool) (http.Handler, func() error, error) {
	if !filesystem.IsSupportedGit(path) && isSupportedHTTP(path) {
		handler, err := newHTTPReverseProxy(path)
		return handler, nilFuncWithError, err
	}
	vFS, err := filesystem.NewRawFSFromURI(path)
	if err != nil {
		return nil, nilFuncWithError, err
	}
	nFS := filesystem.NewNestedFS(vFS)

	ci, err := newCustomIndex(http.FileServer(http.FS(nFS)), nFS, tp, enhancedList)
	if err != nil {
		return nil, nilFuncWithError, err
	}
	return ci, nFS.Close, nil
}
```

- [ ] **Step 2: Delete the old gowebserver filesystem files**

```bash
rm /home/coder/project/gowebserver/pkg/gowebserver/filesystem_common.go \
   /home/coder/project/gowebserver/pkg/gowebserver/filesystem_archive.go \
   /home/coder/project/gowebserver/pkg/gowebserver/filesystem_git.go \
   /home/coder/project/gowebserver/pkg/gowebserver/filesystem_git_unsupported.go \
   /home/coder/project/gowebserver/pkg/gowebserver/filesystem_local.go \
   /home/coder/project/gowebserver/pkg/gowebserver/filesystem_nested.go \
   /home/coder/project/gowebserver/pkg/gowebserver/filesystem_test.go
```

- [ ] **Step 3: Verify gowebserver compiles**

```bash
cd /home/coder/project/gowebserver && go build ./pkg/gowebserver/... 2>&1
```

Expected: no errors.

---

### Task 11: Update `pkg/gowebserver/util.go` and `upload.go`

**Files:**
- Modify: `pkg/gowebserver/util.go`
- Modify: `pkg/gowebserver/upload.go`
- Modify: `pkg/gowebserver/util_test.go`

Remove moved functions from `util.go`. Update `upload.go` to call `filesystem.SanitizeFileName`. Update `util_test.go` to remove tests for moved functions and fix `mustTempDir`.

- [ ] **Step 1: Edit `pkg/gowebserver/util.go`**

Remove these functions entirely (they now live in `pkg/filesystem/util.go`):
- `fsDirMode` const
- `createDirectory`
- `stageRemoteFile`
- `createTempDirectory`
- `tryDeleteDirectory`
- `tryDeleteFile`
- `downloadFile`
- `exists`
- `dirPath`
- `copyFile`
- `validChars` var
- `sanitizeFileName`
- `nilFunc`
- `nilFuncWithError`

Keep these functions (they are web/template utilities, not filesystem):
- `checkError`
- `ensureDirs`
- `createTemplate`
- `executeTemplate`
- `humanizeDate`, `humanizeTimestamp`
- `isImage`, `isAudio`, `isVideo`, `isMedia`
- `isOdd`, `isEven`
- `stepBegin`, `stepEnd`
- `urlEncode`

Add import for `filesystem` package and keep `nilFuncWithError` (still needed by `newHandlerFromFS` in the same package):

The final `pkg/gowebserver/util.go` should be:

```go
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
	"html/template"
	"io"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"go.uber.org/zap"
)

func checkError(err error) {
	if err != nil {
		zap.S().Error(err)
		zap.S().Sync()
	}
}

func nilFuncWithError() error {
	return nil
}

func ensureDirs(fileName string) error {
	absFullPath, err := filepath.Abs(filepath.Clean(fileName))
	if err != nil {
		return err
	}

	fullBaseDir := filepath.Dir(absFullPath)
	return os.MkdirAll(fullBaseDir, 0766)
}

func createTemplate(tmplText []byte) (*template.Template, error) {
	tmpl := template.New("").Funcs(template.FuncMap{
		"humanizeBytes":     humanize.Bytes,
		"isImage":           isImage,
		"isAudio":           isAudio,
		"isVideo":           isVideo,
		"isMedia":           isMedia,
		"isOdd":             isOdd,
		"isEven":            isEven,
		"humanizeDate":      humanizeDate,
		"humanizeTimestamp": humanizeTimestamp,
		"stepBegin":         stepBegin,
		"stepEnd":           stepEnd,
		"urlEncode":         urlEncode,
	})
	return tmpl.Parse(string(tmplText))
}

// Deprecated: Use createTemplate() instead.
func executeTemplate(tmplText []byte, params interface{}, w io.Writer) error {
	t, err := createTemplate(tmplText)
	if err != nil {
		return err
	}

	if err := t.Execute(w, params); err != nil {
		return err
	}
	return nil
}

func humanizeDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func humanizeTimestamp(t time.Time) string {
	return t.Format("2006-01-02 03:04:05PM")
}

func isImage(name string) bool {
	return strings.HasPrefix(mime.TypeByExtension(filepath.Ext(name)), "image")
}

func isAudio(name string) bool {
	return strings.HasPrefix(mime.TypeByExtension(filepath.Ext(name)), "audio")
}

func isVideo(name string) bool {
	return strings.HasPrefix(mime.TypeByExtension(filepath.Ext(name)), "video")
}

func isMedia(name string) bool {
	return isVideo(name) || isImage(name)
}

func isOdd(v int) bool {
	return v%2 == 1
}

func isEven(v int) bool {
	return v%2 == 0
}

func stepBegin(val int, step int, max int) bool {
	if step == 0 {
		return true
	}
	return val%step == 0
}

func stepEnd(val int, step int, max int) bool {
	if step == 0 {
		return true
	}
	return val%step == step-1 || val+1 == max
}

func urlEncode(u string) string {
	return url.PathEscape(u)
}
```

- [ ] **Step 2: Edit `pkg/gowebserver/upload.go`**

Change the `sanitizeFileName` call to `filesystem.SanitizeFileName`. Add the import for `pkg/filesystem`.

Find this line (around line 102 in upload.go):
```go
fileName := sanitizeFileName(files[i].Filename)
```

Replace with:
```go
fileName := filesystem.SanitizeFileName(files[i].Filename)
```

Also add to the import block:
```go
"github.com/jeremyje/gowebserver/v2/pkg/filesystem"
```

- [ ] **Step 3: Edit `pkg/gowebserver/util_test.go`**

Remove tests for moved functions (`TestDirPath`, `TestCreateTempDirectory`, `TestDownloadFileOnLocalFile`, `TestDownloadFileOnHttpsFile`, `TestSanitizeFileName`, `TestCopyFileErrors`).

Update `mustTempDir` to not call `createTempDirectory` (which no longer exists in this package):

```go
func mustTempDir(tb testing.TB) string {
	dir, err := os.MkdirTemp(os.TempDir(), "gowebserver")
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}
```

Also remove the `angryReader` type and `mustFile` function if they're only used by the deleted tests (keep `mustFile` since it's used by `gowebserver_test.go`).

Remove unused imports: `bytes`, `fmt`, `strings`, `time` (if no longer needed after removals).

- [ ] **Step 4: Verify the full module compiles**

```bash
cd /home/coder/project/gowebserver && go build ./... 2>&1
```

Expected: no errors.

- [ ] **Step 5: Run the gowebserver tests**

```bash
cd /home/coder/project/gowebserver && go test -race -count=1 ./pkg/gowebserver/... 2>&1 | grep -E 'FAIL|ok|error' | head -20
```

Expected: `ok` for the gowebserver package.

---

### Task 12: Final verification — run all tests and commit

- [ ] **Step 1: Run the full test suite**

```bash
cd /home/coder/project/gowebserver && go test -race -count=1 -run 'TestDirPath|TestCreateTempDirectory|TestDownloadFileOnLocalFile|TestSanitizeFileName|TestCopyFileErrors|TestNestedFSPath|TestIsSupported|TestVirtualDirectory|TestDirlessArchive|TestNestedFileSystem|TestCheckError|TestHumanizeDate|TestIsEven|TestIsOdd|TestStepBeginEnd|TestUrlEncode|TestExecuteTemplate' ./... 2>&1 | grep -E 'FAIL|ok|---' | head -40
```

Expected: all tests PASS.

- [ ] **Step 2: Run make lint**

```bash
cd /home/coder/project/gowebserver && make lint 2>&1
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
cd /home/coder/project/gowebserver && git add pkg/filesystem/ pkg/gowebserver/filesystem.go pkg/gowebserver/filesystem_common.go pkg/gowebserver/filesystem_archive.go pkg/gowebserver/filesystem_git.go pkg/gowebserver/filesystem_git_unsupported.go pkg/gowebserver/filesystem_local.go pkg/gowebserver/filesystem_nested.go pkg/gowebserver/filesystem_test.go pkg/gowebserver/util.go pkg/gowebserver/upload.go pkg/gowebserver/util_test.go && git commit -m "$(cat <<'EOF'
Extract filesystem implementations to pkg/filesystem package

Moves local, archive, git, and nested filesystem implementations
from pkg/gowebserver to the new standalone pkg/filesystem library.
The only net/http coupling (newHandlerFromFS) remains in gowebserver
as a thin adapter. SanitizeFileName is now exported from pkg/filesystem.

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
EOF
)"
```

---

## Self-Review

### Spec Coverage
- [x] Move filesystem*.go files → Tasks 1, 3, 4, 5, 6, 7
- [x] Untangle net/http → `newHandlerFromFS` stays in gowebserver (Task 10), `net/http` in filesystem only used for downloading remote archives in `downloadFile` (acceptable)
- [x] New `pkg/filesystem` package → Tasks 1-9
- [x] Tests preserved and working → Tasks 8, 9, 12
- [x] Gowebserver still compiles and tests pass → Tasks 10, 11, 12

### Type Consistency
- `FileSystem` interface defined in Task 1, used consistently throughout
- `NewRawFSFromURI` returns `(FileSystem, error)` — matches usage in Task 10
- `NewNestedFS` returns `FileSystem` — matches `nFS.Close` usage (via `io.Closer` in interface) and `http.FS(nFS)` (via `fs.FS` in interface)
- `IsSupportedGit` used in Task 10 — exported in Task 7
- `SanitizeFileName` exported in Task 2 — used in Task 11

### Placeholder Scan
No TBDs, TODOs, or "similar to Task N" patterns found. All code blocks are complete.
