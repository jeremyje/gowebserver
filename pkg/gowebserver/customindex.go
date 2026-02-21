package gowebserver

import (
	_ "embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"mime"
	"net/http"
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
	baseFS       fs.FS
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

var (
	mimeIconMap = map[string]string{
		".":                               "folder",
		"image":                           "image",
		"application/pdf":                 "pdf",
		"audio":                           "audio",
		"text":                            "text",
		"video":                           "video",
		".txt":                            "text",
		".pdf":                            "pdf",
		".doc":                            "doc",
		".xls":                            "spreadsheet",
		".ppt":                            "presentation",
		".jpg":                            "image",
		".mp4":                            "video",
		".xvid":                           "video",
		".mp3":                            "audio",
		".zip":                            "archive",
		".cc":                             "code",
		".go":                             "code",
		".cs":                             "code",
		".java":                           "code",
		".cpp":                            "code",
		".sh":                             "terminal",
		".rar":                            "archive",
		".7z":                             "archive",
		".xz":                             "archive",
		".bz2":                            "archive",
		".tar":                            "archive",
		".gz":                             "archive",
		".ps1":                            "terminal",
		".psm1":                           "terminal",
		".cmd":                            "terminal",
		".bash":                           "terminal",
		".download":                       "download",
		"application/x-shellscript":       "terminal",
		"application/x-ms-dos-executable": "terminal",
		"application/x-msdownload":        "terminal",
		".db":                             "database",
		".epub":                           "ebook",
		".dwg":                            "cad",
		".svg":                            "vector",
		".psd":                            "photoshop",
		".html":                           "markup",
		".htm":                            "markup",
		".css":                            "stylesheet",
		".scss":                           "stylesheet",
		".js":                             "script",
		".ts":                             "script",
		".tsx":                            "script",
		".dat":                            "data",
		".crt":                            "certificate",
		".cert":                           "certificate",
		".pem":                            "key",
		".pkv":                            "key",
		".pk":                             "key",
		".key":                            "key",
		".log":                            "log",
		".bak":                            "backup",
		".bin":                            "binary",
		".pkg":                            "package",
		".rpm":                            "package",
		".msi":                            "package",
		".deb":                            "package",
		".snap":                           "package",
		".sqlite":                         "database",
		".pub":                            "certificate",
		"application/x-x509-ca-cert":      "certificate",
		"application/x-yaml":              "config",
		"application/illustrator":         "photoshop",
		".ds_store":                       "database",
		".ini":                            "config",
		"application/json":                "config",
		"font":                            "font",
		".config":                         "config",
		".cfg":                            "config",
		".yaml":                           "config",
		".yml":                            "config",
		"application/x-cd-image":          "disc",
		".iso":                            "disc",
		".docx":                           "doc",
		".xlsx":                           "spreadsheet",
		".pptx":                           "presentation",
		".md":                             "doc",
		".ttf":                            "font",
		".ai":                             "photoshop",
	}
)

func nameToIconClass(isDir bool, name string) string {
	ext := filepath.Ext(strings.ToLower(name))
	if isDir {
		return "folder"
	}

	if val, ok := mimeIconMap[ext]; ok {
		return val
	}

	mimeType := mime.TypeByExtension(ext)

	if mimeType != "" {
		if val, ok := mimeIconMap[mimeType]; ok {
			return val
		}

		if parts := strings.Split(mimeType, "/"); len(parts) > 1 {
			if val, ok := mimeIconMap[parts[0]]; ok {
				return val
			}
		}
	}

	log.Printf("%s > %s", ext, mimeType)
	return "unknown"
}

func (c *customIndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sortBy := canonicalizeSortBy(r.URL.Query().Get("sort"))
	rootTrace := c.tp.Tracer("customIndex")
	ctx, span := rootTrace.Start(r.Context(), r.URL.Path)
	defer span.End()
	span.SetAttributes(attribute.Bool("enhanced_list", c.enhancedList))
	if c.enhancedList {
		path := r.URL.Path
		path = cleanPath(strings.TrimPrefix(path, "/"))

		zap.S().With("url", r.URL, "path", path).Info("customIndexHandler")
		if strings.HasSuffix(r.URL.Path, "/") || path == "." {
			_, openSpan := rootTrace.Start(ctx, "Open")
			openSpan.SetAttributes(attribute.String("path", path))
			f, err := c.baseFS.Open(path)
			openSpan.End()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer func() {
				_, closeSpan := rootTrace.Start(ctx, "Close")
				closeSpan.SetAttributes(attribute.String("path", path))
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
					Root:               path,
					RootName:           strings.TrimSuffix(filepath.Base(path), nestedDirSuffix),
					DirEntries:         []*DirEntry{},
					SortBy:             sortBy,
					UseTimestamp:       strings.Contains(sortBy, "date"),
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

				files := newEntryList(sortBy)
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
						FullPath:  filepath.Join(path, entry.Name()),
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

func newCustomIndex(baseHandler http.Handler, baseFS fs.FS, tp trace.TracerProvider, enhancedList bool) (http.Handler, error) {
	tmpl, err := createTemplate(customIndexHTML)
	if err != nil {
		return nil, err
	}
	return &customIndexHandler{
		baseHandler:  baseHandler,
		baseFS:       baseFS,
		enhancedList: enhancedList,
		tp:           tp,
		tmpl:         tmpl,
	}, nil
}
