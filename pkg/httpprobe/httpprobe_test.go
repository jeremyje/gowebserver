package httpprobe

import (
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestProbeError(t *testing.T) {
	perr := ProbeError{
		Code:    http.StatusServiceUnavailable,
		Message: "bad",
	}
	want := "HTTP Code: 503 - bad"
	got := perr.Error()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ProbeError.Error() mismatch (-want +got):\n%s", diff)
	}
}

func TestProbe(t *testing.T) {
	insecure := httptest.NewServer(&okHandler{})
	defer insecure.Close()
	secure := httptest.NewTLSServer(&okHandler{})
	defer secure.Close()
	pool := x509.NewCertPool()
	pool.AddCert(secure.Certificate())

	tests := []struct {
		name     string
		args     Args
		wantCode int
	}{
		{
			name:     "empty",
			args:     Args{},
			wantCode: http.StatusServiceUnavailable,
		},
		{
			name: "corrupt",
			args: Args{
				URL: string(rune(0x7f)),
			},
			wantCode: http.StatusServiceUnavailable,
		},
		{
			name: "HTTP Success",
			args: Args{
				URL: insecure.URL + "/ok",
			},
			wantCode: 0,
		},
		{
			name: "HTTP Success (Delayed)",
			args: Args{
				URL: insecure.URL + "/delay",
			},
			wantCode: 0,
		},
		{
			name: "HTTP Success (Timeout)",
			args: Args{
				Timeout: time.Millisecond,
				URL:     insecure.URL + "/delay",
			},
			wantCode: http.StatusServiceUnavailable,
		},
		{
			name: "HTTP Failure",
			args: Args{
				URL: insecure.URL + "/fail",
			},
			wantCode: http.StatusBadGateway,
		},
		{
			name: "HTTP Bad Request",
			args: Args{
				URL: insecure.URL + "/code/400",
			},
			wantCode: 400,
		},
		{
			name: "invalid port",
			args: Args{
				URL: insecure.URL + "00000",
			},
			wantCode: 503,
		},
		{
			name: "HTTPS Success",
			args: Args{
				URL:             secure.URL + "/ok",
				CertificatePool: pool,
			},
			wantCode: 0,
		},
		{
			name: "HTTPS Failure",
			args: Args{
				URL:             secure.URL + "/fail",
				CertificatePool: pool,
			},
			wantCode: http.StatusBadGateway,
		},
		{
			name: "HTTPS No Pool",
			args: Args{
				URL: secure.URL + "/ok",
			},
			wantCode: http.StatusServiceUnavailable,
		},
	}

	//time.Sleep(time.Second)
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			//t.Parallel()

			err := Probe(tc.args)
			if tc.wantCode == 0 {
				if err != nil {
					t.Errorf("got error %s", err)
				}
			} else {
				var perr ProbeError
				if errors.As(err, &perr) {
					if perr.Code != tc.wantCode {
						t.Errorf("want: %v, got %v", tc.wantCode, perr)
					}
				} else {
					t.Errorf("%v is not ProbeError", err)
				}
			}
		})
	}
}

type okHandler struct {
}

func (h *okHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/ok" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	} else if r.URL.Path == "/delay" {
		time.Sleep(time.Millisecond * 100)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	} else if r.URL.Path == "/fail" {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("failure"))
	} else if strings.HasPrefix(r.URL.Path, "/code/") {
		parts := strings.Split(r.URL.Path, "/")
		code, err := strconv.Atoi(parts[len(parts)-1])
		w.WriteHeader(code)
		w.Write([]byte(fmt.Sprintf("returning code '%d', '%s'", code, err)))
	} else {
		w.WriteHeader(http.StatusNotImplemented)
		w.Write([]byte(fmt.Sprintf("'%s' is not handled.", r.URL.Path)))
	}
}
