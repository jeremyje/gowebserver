package gowebserver

import (
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"facette.io/natsort"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var (
	//go:embed custom-index.html
	customIndexHTML []byte
)

type EntryList struct {
	Entries    map[string]*DirEntry
	EntryOrder []string
	sortBy     string
}

func (l *EntryList) Len() int {
	return len(l.Entries)
}

func (l *EntryList) Less(i, j int) bool {
	iElem := l.Entries[l.EntryOrder[i]]
	jElem := l.Entries[l.EntryOrder[j]]
	if iElem.IsDir != jElem.IsDir {
		return iElem.IsDir
	}
	switch l.sortBy {
	case "size":
		return iElem.Size < jElem.Size
	case "size-desc":
		return jElem.Size < iElem.Size
	case "date":
		return iElem.ModTime.Before(jElem.ModTime)
	case "date-desc":
		return jElem.ModTime.Before(iElem.ModTime)
	case "name":
		return natsort.Compare(iElem.Name, jElem.Name)
	case "name-desc":
		return natsort.Compare(jElem.Name, iElem.Name)
	default:
		return natsort.Compare(iElem.Name, jElem.Name)
	}
}

func (l *EntryList) Swap(i, j int) {
	l.EntryOrder[i], l.EntryOrder[j] = l.EntryOrder[j], l.EntryOrder[i]
}

func (l *EntryList) add(entry *DirEntry) {
	l.Entries[entry.Name] = entry
	l.EntryOrder = append(l.EntryOrder, entry.Name)
}

func newEntryList(sortBy string) *EntryList {
	return &EntryList{
		sortBy:     sortBy,
		Entries:    map[string]*DirEntry{},
		EntryOrder: []string{},
	}
}

type DirEntry struct {
	Name      string
	FullPath  string
	Size      uint64
	ModTime   time.Time
	IsDir     bool
	IsArchive bool
	IconClass string
}

func (d *DirEntry) String() string {
	if d == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s: dir= %t, archive= %t iconClass= %q", d.Name, d.IsDir, d.IsArchive, d.IconClass)
}

type CustomIndexReport struct {
	Root               string
	RootName           string
	DirEntries         []*DirEntry
	SortBy             string
	UseTimestamp       bool
	HasNonMediaEntry   bool
	HasImage           bool
	HasVideo           bool
	ApplicationVersion string
}

type customIndexHandler struct {
	baseHandler  http.Handler
	searchFS     searchFS
	enhancedList bool
	tp           trace.TracerProvider
	tmpl         *template.Template
}

func canonicalizeSortBy(v string) string {
	parts := strings.Split(strings.ToLower(v), "=")
	key := parts[0]
	switch key {
	case "name", "name-desc", "size", "size-desc", "date", "date-desc":
		return key
	}
	return "name"
}

type indexArgs struct {
	sortBy string
	query  string
	path   string
}

func getIndexArgs(u *url.URL) *indexArgs {
	sortBy := canonicalizeSortBy(u.Query().Get("sort"))
	query := u.Query().Get("q")
	return &indexArgs{
		path:   cleanPath(strings.TrimPrefix(u.Path, "/")),
		sortBy: sortBy,
		query:  query,
	}
}

func (c *customIndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	args := getIndexArgs(r.URL)
	rootTrace := c.tp.Tracer("customIndex")
	ctx, span := rootTrace.Start(r.Context(), r.URL.Path)
	defer span.End()
	span.SetAttributes(attribute.Bool("enhanced_list", c.enhancedList))
	if c.enhancedList {
		zap.S().With("url", r.URL, "path", args.path).Info("customIndexHandler")
		if strings.HasSuffix(r.URL.Path, "/") || args.path == "." {
			_, openSpan := rootTrace.Start(ctx, "Open")
			openSpan.SetAttributes(attribute.String("path", args.path))
			f, err := c.searchFS.Open(args.path)
			openSpan.End()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer func() {
				_, closeSpan := rootTrace.Start(ctx, "Close")
				closeSpan.SetAttributes(attribute.String("path", args.path))
				f.Close()
				closeSpan.End()
			}()

			readDirCtx, readDirSpan := rootTrace.Start(ctx, "readDirectory")
			defer readDirSpan.End()
			rdf, ok := f.(fs.ReadDirFile)
			if ok {
				span.SetAttributes(attribute.Bool("custom_directory_list", true))
				readDirSpan.AddEvent("")
				now := time.Now()
				entries, err := rdf.ReadDir(-1)
				span.SetAttributes(attribute.Int("num_entries", len(entries)))
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				params := &CustomIndexReport{
					Root:               args.path,
					RootName:           strings.TrimSuffix(filepath.Base(args.path), nestedDirSuffix),
					DirEntries:         []*DirEntry{},
					SortBy:             args.sortBy,
					UseTimestamp:       strings.Contains(args.sortBy, "date"),
					ApplicationVersion: version,
				}

				allFiles := map[string]any{}
				archiveDirs := map[string]fs.DirEntry{}
				for _, entry := range entries {
					if strings.HasSuffix(entry.Name(), nestedDirSuffix) {
						archiveDirs[entry.Name()] = entry
					}
					allFiles[entry.Name()] = entry
				}
				actualArchiveDir := map[string]fs.DirEntry{}
				for name := range allFiles {
					if archiveDir, ok := archiveDirs[name+nestedDirSuffix]; ok {
						actualArchiveDir[name] = archiveDir
					}
				}

				files := newEntryList(args.sortBy)
				for _, entry := range entries {
					_, statFileSpan := rootTrace.Start(readDirCtx, entry.Name())

					if strings.HasSuffix(entry.Name(), nestedDirSuffix) {
						if _, ok := actualArchiveDir[strings.TrimSuffix(entry.Name(), nestedDirSuffix)]; ok {
							continue
						}
					}

					size := int64(0)
					t := now
					stat, err := entry.Info()
					if err == nil {
						t = stat.ModTime()
						size = stat.Size()
					}

					_, isArchive := actualArchiveDir[entry.Name()]
					isDir := entry.IsDir() || isArchive
					newEntry := &DirEntry{
						FullPath:  filepath.Join(args.path, entry.Name()),
						Name:      entry.Name(),
						Size:      uint64(size),
						ModTime:   t,
						IsDir:     isDir,
						IsArchive: isArchive,
						IconClass: nameToIconClass(isDir, entry.Name()),
					}
					files.add(newEntry)
					statFileSpan.End()
				}

				_, sortFileSpan := rootTrace.Start(readDirCtx, "sort")
				sortFileSpan.SetAttributes(attribute.Int("num_files", files.Len()))
				sort.Sort(files)
				sortFileSpan.End()

				_, generateSpan := rootTrace.Start(readDirCtx, "applyTemplate")
				generateSpan.SetAttributes(attribute.Int("num_files", files.Len()))
				defer generateSpan.End()
				hasNonMediaEntry := false
				hasImage := false
				hasVideo := false
				for _, name := range files.EntryOrder {
					entry := files.Entries[name]
					params.DirEntries = append(params.DirEntries, entry)
					if !isMedia(entry.Name) {
						hasNonMediaEntry = true
					}
					if isImage(entry.Name) {
						hasImage = true
					}
					if isVideo(entry.Name) {
						hasVideo = true
					}
				}
				params.HasNonMediaEntry = hasNonMediaEntry
				params.HasImage = hasImage
				params.HasVideo = hasVideo

				if err := c.tmpl.Execute(w, params); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				return
			}
		}
	}
	span.SetAttributes(attribute.Bool("custom_directory_list", false))
	c.baseHandler.ServeHTTP(w, r)
}

func newCustomIndex(baseHandler http.Handler, baseFS fs.FS, tp trace.TracerProvider, enhancedList bool, searchParams *searchParams) (http.Handler, func(), error) {
	tmpl, err := createTemplate(customIndexHTML)
	if err != nil {
		return nil, func() {}, err
	}
	ctx := context.Background()
	searchFS := newSearchFS(baseFS, searchParams)
	searchFS.Start(ctx)
	return &customIndexHandler{
			baseHandler:  baseHandler,
			searchFS:     searchFS,
			enhancedList: enhancedList,
			tp:           tp,
			tmpl:         tmpl,
		}, func() {
			searchFS.Stop()
		}, nil
}
