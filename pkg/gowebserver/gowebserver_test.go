package gowebserver

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jeremyje/gowebserver/pkg/certtool"
)

func TestConfigLogger(t *testing.T) {
	for _, verbose := range []bool{false, true} {
		logger, close := configLogger(verbose)
		if logger == nil {
			t.Error("logger is nil")
		}
		close()
	}
}

func TestRunApplication(t *testing.T) {
	httpPortFlag = new(int)
	httpsPortFlag = new(int)

	ch := make(chan error)
	go func() {
		time.Sleep(time.Second)
		ch <- nil
	}()
	err := runApplication(ch)
	if err != nil {
		if !strings.Contains(err.Error(), "closed network connection") {
			t.Error(err)
		}
	}
}

func TestCreateCertificate(t *testing.T) {
	dir := mustTempDir(t)
	cfg := &Config{
		Verbose:           false,
		Serve:             []Serve{},
		ConfigurationFile: "",
		HTTP:              HTTP{},
		HTTPS: HTTPS{
			Port: 0,
			Certificate: Certificate{
				RootPrivateKeyFilePath:   filepath.Join(dir, "root-private.key"),
				RootCertificateFilePath:  filepath.Join(dir, "root-public.cert"),
				PrivateKeyFilePath:       filepath.Join(dir, "private.key"),
				CertificateFilePath:      filepath.Join(dir, "public.cert"),
				CertificateHosts:         "localhost",
				CertificateValidDuration: time.Hour,
				ForceOverwrite:           true,
			},
		},
		Metrics: Metrics{},
		Upload:  Serve{},
	}
	if err := createCertificate(cfg); err != nil {
		t.Error(err)
	}

	rootPub, _, err := certtool.ReadKeyPair(mustFile(t, filepath.Join(dir, "root-public.cert")), mustFile(t, filepath.Join(dir, "root-private.key")))
	if err != nil {
		t.Error(err)
	}

	pub, _, err := certtool.ReadKeyPair(mustFile(t, filepath.Join(dir, "public.cert")), mustFile(t, filepath.Join(dir, "private.key")))
	if err != nil {
		t.Error(err)
	}
	if err := pub.CheckSignatureFrom(rootPub); err != nil {
		t.Error(err)
	}
}

func ExampleWebServer_Serve() {
	conf := &Config{
		Verbose: false,
		HTTP: HTTP{
			Port: 0,
		},
		HTTPS: HTTPS{
			Port: 0,
		},
	}

	logger, syncFunc := configLogger(conf.Verbose)
	defer syncFunc()

	httpServer, err := New(conf)
	if err != nil {
		logger.Sugar().Fatal(err)
	}

	termCh := make(chan error)
	go func() {
		time.Sleep(time.Second)
		termCh <- nil
	}()

	httpServer.Serve(termCh)
	// Output:
}
