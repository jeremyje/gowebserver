package gowebserver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gowsTesting "github.com/jeremyje/gowebserver/v2/internal/gowebserver/testing"
)

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
	ci := newCustomIndex(http.FileServer(http.FS(nFS)), nFS, mc.getTraceProvider(), true)

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
