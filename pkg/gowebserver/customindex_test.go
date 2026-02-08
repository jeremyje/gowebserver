package gowebserver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gowsTesting "github.com/jeremyje/gowebserver/v2/internal/gowebserver/testing"
)

func TestNameToIconClass(t *testing.T) {
	testCases := []struct {
		input string
		isDir bool
		want  string
	}{
		{input: "abc", isDir: false, want: "unknown"},
		{input: "abc", isDir: true, want: "folder"},
		{input: "abc.txt", isDir: false, want: "text"},
		{input: "abc.pdf", isDir: false, want: "pdf"},
		{input: "abc.doc", isDir: false, want: "doc"},
		{input: "abc.xls", isDir: false, want: "spreadsheet"},
		{input: "abc.ppt", isDir: false, want: "presentation"},
		{input: "abc.jpg", isDir: false, want: "image"},
		{input: "abc.mp4", isDir: false, want: "video"},
		{input: "abc.m4v", isDir: false, want: "video"},
		{input: "abc.m4a", isDir: false, want: "audio"},
		{input: "abc.avi", isDir: false, want: "video"},
		{input: "abc.wmv", isDir: false, want: "video"},
		{input: "abc.flv", isDir: false, want: "video"},
		{input: "abc.mp3", isDir: false, want: "audio"},
		{input: "abc.ogg", isDir: false, want: "audio"},
		{input: "abc.m4a", isDir: false, want: "audio"},
		{input: "abc.flac", isDir: false, want: "audio"},
		{input: "abc.wav", isDir: false, want: "audio"},
		{input: "abc.zip", isDir: false, want: "archive"},
		{input: "abc.tar.gz", isDir: false, want: "archive"},
		{input: "abc.tar", isDir: false, want: "archive"},
		{input: "abc.tar.bz2", isDir: false, want: "archive"},
		{input: "abc.tar.xz", isDir: false, want: "archive"},
		{input: "abc.7z", isDir: false, want: "archive"},
		{input: "abc.rar", isDir: false, want: "archive"},
		{input: "abc.cc", isDir: false, want: "code"},
		{input: "abc.sh", isDir: false, want: "terminal"},
		{input: "abc.bash", isDir: false, want: "terminal"},
		{input: "abc.cmd", isDir: false, want: "terminal"},
		{input: "abc.ps1", isDir: false, want: "terminal"},
		{input: "abc.download", isDir: false, want: "download"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got := nameToIconClass(tc.isDir, tc.input)
			if got != tc.want {
				t.Errorf("got: %q, want: %q", got, tc.want)
			}
		})
	}
}

func TestCustomIndex(t *testing.T) {
	nestedZipPath := gowsTesting.MustNestedZipFilePath(t)

	vFS, err := newRawFSFromURI(nestedZipPath)
	if err != nil {
		t.Error(err)
	}
	defer vFS.Close()
	nFS := newNestedFS(vFS)
	defer nFS.Close()

	mc := &monitoringContext{}
	ci, err := newCustomIndex(http.FileServer(http.FS(nFS)), nFS, mc.getTraceProvider(), true)
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewUnstartedServer(ci)
	ts.EnableHTTP2 = true
	ts.StartTLS()
	defer ts.Close()

	verifyCustomIndex(t, ts.Client(), ts.URL, []string{"single-testassets.zip", "single-testassets.zip-dir/", "testassets/", "testassets.zip-dir/"})
	verifyCustomIndex(t, ts.Client(), ts.URL+"/testassets", []string{"index.html", "site.js", "assets/"})
	verifyCustomIndex(t, ts.Client(), ts.URL+"/testassets/", []string{"index.html", "site.js", "assets/"})
	verifyCustomIndex(t, ts.Client(), ts.URL+"/testassets.zip-dir/", []string{"index.html", "site.js", "assets/"})
	verifyCustomIndex(t, ts.Client(), ts.URL+"/testassets/assets/images", []string{"ocean.jpg", "nature.jpg"})
	verifyCustomIndex(t, ts.Client(), ts.URL+"/testassets/assets/images/", []string{"ocean.jpg", "nature.jpg"})
}

func verifyCustomIndex(t *testing.T, hc *http.Client, u string, substrs []string) {
	if len(substrs) == 0 {
		t.Error("must have at least 1 substring to check")
		return
	}
	res, err := hc.Get(u)
	if err != nil {
		t.Errorf("cannot GET %s, %s", u, err)
		return
	}

	bodyBytes, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Errorf("cannot read response from %s, %s", u, err)
		return
	}
	body := string(bodyBytes)
	missingSubstr := true
	for _, substr := range substrs {
		if !strings.Contains(body, substr) {
			t.Errorf("%s does not contain string '%s'", u, substr)
			missingSubstr = true
		}
	}
	if missingSubstr {
		t.Logf("--- GET: %s\n\n%s", u, body)
	}
}
