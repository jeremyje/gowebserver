// Copyright 2022 Cloudfra
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
	"io/fs"
	"path"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
)

func TestLoadTemplateInclude(t *testing.T) {
	fsys := fstest.MapFS{
		"template/page/page.html": {Data: []byte("<style>\n  {{template \"shared\"}}\n  {{include \"page.css\"}}\n</style>\n<script>\n  (function () {\n    {{include \"a.js\"}}\n\n    {{include \"b.js\"}}\n  })();\n</script>\n{{.Name}} {{include \"inline.css\"}}\n")},
		"template/page/page.css":  {Data: []byte("a {\n  color: red;\n}\n\nb {\n  color: blue;\n}\n")},
		"template/page/a.js":      {Data: []byte("var a = 1;\n")},
		"template/page/b.js":      {Data: []byte("function b() {\n\treturn a;\n}\n\n\n")},
	}
	// A file goes in at the indent of the line that includes it, with no
	// trailing space on its empty lines; only a whole line includes.
	want := "<style>\n  {{template \"shared\"}}\n  a {\n    color: red;\n  }\n\n  b {\n    color: blue;\n  }\n</style>\n<script>\n  (function () {\n    var a = 1;\n\n    function b() {\n    \treturn a;\n    }\n  })();\n</script>\n{{.Name}} {{include \"inline.css\"}}\n"
	got, err := loadTemplate(fsys, "page")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("loadTemplate() mismatch (-want +got):\n%s", diff)
	}
}

func TestLoadTemplateErrors(t *testing.T) {
	testCases := []struct {
		name    string
		fsys    fstest.MapFS
		wantErr string
	}{
		{
			name:    "no such template",
			fsys:    fstest.MapFS{},
			wantErr: "template/page/page.html",
		},
		{
			name:    "no such file",
			fsys:    fstest.MapFS{"template/page/page.html": {Data: []byte("{{include \"gone.js\"}}\n")}},
			wantErr: "template/page/gone.js",
		},
		{
			name: "template action in an included file",
			fsys: fstest.MapFS{
				"template/page/page.html": {Data: []byte("{{include \"page.js\"}}\n")},
				"template/page/page.js":   {Data: []byte("var name = {{.Name}};\n")},
			},
			wantErr: "template/page/page.js holds a template action",
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := loadTemplate(tc.fsys, "page")
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("loadTemplate() error = %v, want one with %q", err, tc.wantErr)
			}
		})
	}
}

// TestEmbeddedTemplates checks what ships in the binary: every template
// loads into a page that parses, every CSS and JavaScript file is part of
// one, and the tests stay out.
func TestEmbeddedTemplates(t *testing.T) {
	dirs, err := fs.ReadDir(templateFS, "template")
	if err != nil {
		t.Fatal(err)
	}
	if len(dirs) == 0 {
		t.Fatal("no templates are embedded")
	}
	for _, dir := range dirs {
		name := dir.Name()
		t.Run(name, func(t *testing.T) {
			page, err := loadTemplate(templateFS, name)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(page), "{{include") {
				t.Errorf("an include is left in the page; it must be a line of its own")
			}
			// page-theme defines the templates the pages use, so it
			// parses with them as a page does.
			if _, err := createTemplate(page); err != nil {
				t.Errorf("createTemplate() = %v", err)
			}

			html, err := fs.ReadFile(templateFS, path.Join("template", name, name+".html"))
			if err != nil {
				t.Fatal(err)
			}
			files, err := fs.ReadDir(templateFS, path.Join("template", name))
			if err != nil {
				t.Fatal(err)
			}
			for _, f := range files {
				switch {
				case f.IsDir():
					t.Errorf("%s/%s is embedded; tests and benchmarks must stay out of the binary", name, f.Name())
				case f.Name() == name+".html":
				case !strings.Contains(string(html), `{{include "`+f.Name()+`"}}`):
					t.Errorf("%s is embedded but %s.html doesn't include it", f.Name(), name)
				}
			}
		})
	}
}
