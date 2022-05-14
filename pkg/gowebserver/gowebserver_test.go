package gowebserver

import (
	"strings"
	"testing"
	"time"
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
