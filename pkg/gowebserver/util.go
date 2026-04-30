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
