package gowebserver

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	_ "embed"

	"github.com/google/go-cmp/cmp"
	gomainTesting "github.com/jeremyje/gomain/testing"
	gowsTesting "github.com/jeremyje/gowebserver/v2/internal/gowebserver/testing"
)

var (
	//go:embed testdata/multiindex.html
	wantMultiIndex []byte

	wantIndex      = []byte("index.html")
	wantSiteJs     = []byte("site.js")
	wantAssets1Txt = []byte("assets/1.txt")
)

func TestServeAndDie(t *testing.T) {
	baseURL, close := serveAsync(t, &Config{Debug: true})
	defer close()

	resp, err := http.Get(baseURL + "/diediedie")
	if err == nil || resp != nil {
		t.Errorf("got response when server should be dead., Response: %v, Err: %s", resp, err)
	}

	resp, err = http.Get(baseURL)
	if err == nil || resp != nil {
		t.Errorf("got response when server should be dead., Response: %v, Err: %s", resp, err)
	}
}

func TestDieDieDieDisabled(t *testing.T) {
	baseURL, close := serveAsync(t, &Config{})
	defer close()

	resp, err := http.Get(baseURL + "/diediedie")
	if resp.StatusCode != http.StatusOK || err != nil {
		t.Errorf("got error response, Response: %v, Err: %s", resp, err)
	}

	resp, err = http.Get(baseURL)
	if resp.StatusCode != http.StatusOK || err != nil {
		t.Errorf("server should still be alive, Response: %v, Err: %s", resp, err)
	}
}

func TestServe(t *testing.T) {
	ch := make(chan error)

	httpServer, err := New(&Config{})
	if err != nil {
		t.Fatal(err)
	}

	close := gomainTesting.Main(httpServer.Serve)
	go func() {
		time.Sleep(time.Second)
		ch <- close()
	}()

	if err := <-ch; err != nil {
		if !strings.Contains(err.Error(), "closed network connection") {
			t.Error(err)
		}
	}
}

func TestWebServer_Serve_Multi(t *testing.T) {
	zipPath := gowsTesting.MustZipFilePath(t)
	rarPath := gowsTesting.MustRarFilePath(t)
	tarXzPath := gowsTesting.MustTarXzFilePath(t)

	cfg := &Config{
		Serve: []Serve{
			{
				Source:   zipPath,
				Endpoint: "/zip",
			},
			{
				Source:   rarPath,
				Endpoint: "/rar",
			},
			{
				Source:   tarXzPath,
				Endpoint: "/tar.gz",
			},
		},
	}

	baseURL, close := serveAsync(t, cfg)
	defer close()

	testCases := []struct {
		url  string
		want []byte
	}{
		{url: baseURL, want: wantMultiIndex},
		{url: baseURL + "/zip", want: wantIndex},
		{url: baseURL + "/zip/", want: wantIndex},
		{url: baseURL + "/zip/site.js", want: wantSiteJs},
		{url: baseURL + "/zip/assets/1.txt", want: wantAssets1Txt},
		{url: baseURL + "/rar", want: wantIndex},
		{url: baseURL + "/rar/", want: wantIndex},
		{url: baseURL + "/rar/site.js", want: wantSiteJs},
		{url: baseURL + "/rar/assets/1.txt", want: wantAssets1Txt},
		{url: baseURL + "/tar.gz", want: wantIndex},
		{url: baseURL + "/tar.gz/", want: wantIndex},
		{url: baseURL + "/tar.gz/site.js", want: wantSiteJs},
		{url: baseURL + "/tar.gz/assets/1.txt", want: wantAssets1Txt},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.url, func(t *testing.T) {
			resp, err := http.Get(tc.url)
			if err != nil {
				t.Error(err)
			} else {
				got, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Error(err)
				}
				if diff := cmp.Diff(tc.want, got); diff != "" {
					t.Errorf("body mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestWebServer_Serve(t *testing.T) {
	archivePaths := []string{"/.", "/", "/index.html", "/site.js", "/assets/", "/assets/fivesix", "/assets/fivesix/", "/assets/fivesix/5.txt", "/assets/more/3.txt", "/assets/1.txt"}
	testCases := []struct {
		source string
		paths  []string
	}{
		{
			source: gowsTesting.MustZipFilePath(t),
			paths:  archivePaths,
		},
		{
			source: gowsTesting.MustRarFilePath(t),
			paths:  archivePaths,
		},
		{
			source: gowsTesting.MustSevenZipFilePath(t),
			paths:  archivePaths,
		},
		{
			source: gowsTesting.MustTarFilePath(t),
			paths:  archivePaths,
		},
		{
			source: gowsTesting.MustTarGzFilePath(t),
			paths:  archivePaths,
		},
		{
			source: gowsTesting.MustTarBzip2FilePath(t),
			paths:  archivePaths,
		},
		{
			source: gowsTesting.MustTarXzFilePath(t),
			paths:  archivePaths,
		},
		{
			source: gowsTesting.MustTarLz4FilePath(t),
			paths:  archivePaths,
		},
		{
			source: "http://example.com/",
			paths:  []string{"/"},
		},
		{
			source: "https://github.com/jeremyje/gowebserver.git",
			paths:  []string{"/", "/README.md"},
		},
		/*
			TODO: This breaks because of https://github.com/go-git/go-git/issues/143.
			{
				source: "git@github.com:jeremyje/gowebserver.git",
				paths:  []string{"/", "/README.md"},
			},
		*/
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.source, func(t *testing.T) {
			t.Parallel()

			cfg := &Config{
				Serve: []Serve{
					{
						Source:   tc.source,
						Endpoint: "/",
					},
				},
			}

			baseURL, close := serveAsync(t, cfg)
			defer close()

			for _, path := range tc.paths {
				resp, err := http.Get(baseURL + path)
				if err != nil {
					t.Error(err)
				} else {
					if resp.StatusCode != http.StatusOK {
						t.Errorf("status for '%s' got: %d, want 200", path, resp.StatusCode)
					}
				}
			}
		})
	}
}

func serveAsync(tb testing.TB, cfg *Config) (string, func()) {
	ws, err := New(cfg)
	if err != nil {
		tb.Fatal(err)
	}

	wsi, ok := ws.(*webServerImpl)
	if !ok {
		tb.Fatalf("WebServer is not of type *webServerImpl, %+v", ws)
	}

	close := gomainTesting.Main(wsi.Serve)

	var httpPort int
	for i := 0; i < 600; i++ {
		httpPort, _ = wsi.getPorts()
		if httpPort != 0 {
			break
		}
		if i%10 == 0 && i != 0 {
			tb.Logf("waited %d seconds", i*100)
		}
		time.Sleep(time.Millisecond * 100)
	}

	baseURL := fmt.Sprintf("http://localhost:%d", httpPort)
	if err := waitAvailable(baseURL); err != nil {
		tb.Error(err)
	}

	return baseURL, func() {
		close()
	}
}

func waitAvailable(url string) error {
	for i := 0; i < 10; i++ {
		if _, err := http.Get(url); err == nil {
			return nil
		}
		time.Sleep(time.Millisecond * 100)
	}
	return fmt.Errorf("exhausted retries while waiting for '%s'", url)
}

func TestNew(t *testing.T) {
	testCases := []struct {
		config *Config
		want   string
	}{
		{config: nil, want: "/"},
		{config: &Config{}, want: "/"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%+v", tc.config), func(t *testing.T) {
			//t.Parallel()
			got, err := New(tc.config)
			if err != nil {
				t.Fatal(err)
			}

			if got == nil {
				t.Error("WebServer is nil")
			}
		})
	}
}

func TestNormalizeHTTPPath(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		{input: "", want: "/"},
		{input: "/", want: "/"},
		{input: "//", want: "/"},
		{input: "///", want: "/"},
		{input: "gowebserver/", want: "/gowebserver/"},
		{input: "/gowebserver/", want: "/gowebserver/"},
		{input: "/gowebserver", want: "/gowebserver/"},
		{input: "/goweb/server", want: "/goweb/server/"},
		{input: "goweb/server", want: "/goweb/server/"},
		{input: "goweb/server/", want: "/goweb/server/"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			got := normalizeHTTPPath(tc.input)
			if tc.want != got {
				t.Errorf("expected: %v, got: %v", tc.want, got)
			}
		})
	}
}
