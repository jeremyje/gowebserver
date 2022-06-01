package gowebserver

import "testing"

func xTestMain(t *testing.T) {
	conf := &Config{
		Verbose:           true,
		Serve:             []Serve{{Source: "/home/coder/project/gowebserver/cmd/", Endpoint: "/mains/"}, {Source: "/home/coder/project/gowebserver/pkg/", Endpoint: "/code/"}, {Source: "/home/coder/project/gowebserver/", Endpoint: "/root/"}},
		ConfigurationFile: "",
		HTTP:              HTTP{Port: 8181},
		HTTPS:             HTTPS{Port: 8443},
		Monitoring: Monitoring{
			DebugEndpoint: "",
			Metrics: Metrics{
				Enabled: true,
				Path:    "/metrics",
			},
			Trace: Trace{
				Enabled: true,
				URI:     "http://coder:14268/api/traces",
			},
		},
		Upload:       Serve{},
		EnhancedList: true,
	}

	logger, syncFunc := configLogger(conf.Verbose)
	defer syncFunc()

	httpServer, err := New(conf)
	if err != nil {
		logger.Sugar().Fatal(err)
	}

	termCh := make(chan error)
	if err := httpServer.Serve(termCh); err != nil {
		logger.Sugar().With("error", err).Error("web server broke")
	}
	// Output:
}
