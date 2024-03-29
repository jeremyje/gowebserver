package gowebserver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	gomainTesting "github.com/jeremyje/gomain/testing"
	"github.com/jeremyje/gowebserver/v2/pkg/certtool"
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
	close := gomainTesting.Main(runApplication)

	ch := make(chan error)
	go func() {
		time.Sleep(time.Second)
		ch <- close()
	}()
	err := <-ch
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
		Monitoring: Monitoring{
			Metrics: Metrics{},
		},
		Upload: Serve{},
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

	close := gomainTesting.Main(httpServer.Serve)
	close()
	// Output:
}

func xTestWebServerFull(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	backback := filepath.Dir(filepath.Dir(cwd))
	conf := &Config{
		Verbose: true,
		Serve: []Serve{
			{Source: cwd, Endpoint: "/cwd"},
			{Source: backback, Endpoint: "/root"},
		},
		EnhancedList: true,
		Debug:        true,
		HTTP:         HTTP{Port: 8082},
		HTTPS:        HTTPS{Port: 0},
		Monitoring: Monitoring{
			DebugEndpoint: "/debug",
			Metrics: Metrics{
				Enabled: true,
				Path:    "/metrics",
			},
			Trace: Trace{
				Enabled: true,
				URI:     "http://jaeger:14268/api/traces",
			},
		},
		Upload: Serve{
			Source:   "/tmp/upload",
			Endpoint: "/upload",
		},
	}

	logger, syncFunc := configLogger(conf.Verbose)
	defer syncFunc()

	httpServer, err := New(conf)
	if err != nil {
		logger.Sugar().Fatal(err)
	}

	close := gomainTesting.Main(httpServer.Serve)
	time.Sleep(500 * time.Second)
	close()
	// Output:
}
