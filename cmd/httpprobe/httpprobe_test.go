package main

import (
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func TestProbe(t *testing.T) {
	insecure := httptest.NewServer(&okHandler{})
	defer insecure.Close()
	secure := httptest.NewTLSServer(&okHandler{})
	defer secure.Close()

	certFile := filepath.Join(t.TempDir(), "test-probe.cert")

	if err := ioutil.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: secure.Certificate().Raw,
	}), 0644); err != nil {
		t.Fatal(err)
	}
	content, err := ioutil.ReadFile(certFile)
	t.Logf("%s %s", content, err)

	tests := []struct {
		url      string
		certFile string
		wantCode int
	}{
		{
			url:      insecure.URL,
			wantCode: errnoSuccess,
		},
		{
			url:      insecure.URL,
			certFile: "does-not-exist",
			wantCode: errnoFileNotFound,
		},
		{
			url:      "bad",
			wantCode: 503,
		},
		{
			url:      secure.URL,
			certFile: certFile,
			wantCode: errnoSuccess,
		},
		{
			url:      secure.URL,
			wantCode: 503,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("%s_%s", tc.url, tc.certFile), func(t *testing.T) {
			got := run(tc.url, tc.certFile, time.Second)
			if got != tc.wantCode {
				t.Errorf("want: %v, got %v", tc.wantCode, got)
			}
		})
	}
}

type okHandler struct {
}

func (h *okHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
