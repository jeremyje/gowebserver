package gowebserver

import (
	_ "embed"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"facette.io/natsort"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var (
	//go:embed custom-index.html
	customIndexHTML []byte
)

type DirEntry struct {
	Name     string
	FullPath string
	Size     uint64
	ModTime  time.Time
	IsDir    bool
}

type CustomIndexReport struct {
	Root       string
	DirEntries []*DirEntry
	Images     []*DirEntry
}

type customIndexHandler struct {
	baseHandler  http.Handler
	baseFS       fs.FS
	enhancedList bool
	tp           trace.TracerProvider
}

func (c *customIndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rootTrace := c.tp.Tracer("customIndex")
	ctx, span := rootTrace.Start(r.Context(), r.URL.Path)
	defer span.End()
	if c.enhancedList {
		path := r.URL.Path
		path = cleanPath(strings.TrimPrefix(path, "/"))

		zap.S().With("url", r.URL, "path", path).Info("custom index")
		if strings.HasSuffix(r.URL.Path, "/") || path == "." {
			_, openSpan := rootTrace.Start(ctx, "Open")
			f, err := c.baseFS.Open(path)
			openSpan.End()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer f.Close()

			readDirCtx, readDirSpan := rootTrace.Start(ctx, "Read Directory")
			defer readDirSpan.End()
			rdf, ok := f.(fs.ReadDirFile)
			if ok {
				now := time.Now()
				entries, err := rdf.ReadDir(-1)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				params := &CustomIndexReport{
					Root:       path,
					DirEntries: []*DirEntry{},
				}

				files := map[string]*DirEntry{}
				sortedFiles := []string{}
				for _, entry := range entries {
					_, statFileSpan := rootTrace.Start(readDirCtx, entry.Name())
					defer statFileSpan.End()

					size := int64(0)
					t := now
					stat, err := entry.Info()
					if err == nil {
						t = stat.ModTime()
						size = stat.Size()
					}
					newEntry := &DirEntry{
						FullPath: filepath.Join(path, entry.Name()),
						Name:     entry.Name(),
						Size:     uint64(size),
						ModTime:  t,
						IsDir:    entry.IsDir(),
					}
					files[newEntry.Name] = newEntry
					sortedFiles = append(sortedFiles, newEntry.Name)
				}

				_, sortFileSpan := rootTrace.Start(readDirCtx, "sort")
				natsort.Sort(sortedFiles)
				sortFileSpan.End()

				_, generateSpan := rootTrace.Start(readDirCtx, "generate")
				defer generateSpan.End()
				for _, name := range sortedFiles {
					entry := files[name]
					params.DirEntries = append(params.DirEntries, entry)
				}

				if err := executeTemplate(customIndexHTML, params, w); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				return
			}
		}
	}
	c.baseHandler.ServeHTTP(w, r)
}

func newCustomIndex(baseHandler http.Handler, baseFS fs.FS, tp trace.TracerProvider, enhancedList bool) http.Handler {
	return &customIndexHandler{
		baseHandler:  baseHandler,
		baseFS:       baseFS,
		enhancedList: enhancedList,
		tp:           tp,
	}
}
