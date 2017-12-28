package config

import (
	"flag"
	"os"
	"os/user"
	"strconv"
)

// Serving Flags
var pathFlag = flag.String("path", "", "Path to serve (local filesystem, git, zip, tarball files).")
var servePathFlag = flag.String("servepath", "/", "The HTTP/HTTPS serving root path for the hosted path.")
var configFileFlag = flag.String("configfile", "", "YAML formatted configuration file. (overrides flag values)")
var verboseFlag = flag.Bool("verbose", false, "Print out extra information.")

// Upload Flags
var uploadPathFlag = flag.String("upload.path", "uploaded-files", "Local filesystem path where uploaded files are placed.")
var uploadServePathFlag = flag.String("upload.servepath", "/upload.asp", "The URL path for uploading files.")

// HTTP Flags
var httpPortFlag *int

// HTTPS Flags
var httpsPortFlag *int
var privateKeyFilePathFlag = flag.String("https.privatekey", "rsa.pem", "Certificate to host HTTPS with.")

// HTTPS Certificate Flags
var certificateFilePathFlag = flag.String("https.certificate.path", "cert.pem", "Certificate to host HTTPS with.")
var certHostsFlag = flag.String("https.certificate.hosts", "", "Comma-separated hostnames and IPs to generate a certificate for.")
var validDurationFlag = flag.Int("https.certificate.duration", 5475, "Certificate valid duration.")
var actAsCertificateAuthorityFlag = flag.Bool("https.certificate.authority", false, "(Experimental) Generate a root cert as a Certificate Authority")
var onlyGenerateCertFlag = flag.Bool("https.certificate.onlygenerate", false, "Only generate a self-signed certificate for the server.")

// Monitoring Flags
var metricsFlag = flag.Bool("metrics.enabled", true, "Enables server metrics for monitoring.")
var metricsPathFlag = flag.String("metrics.path", "/metrics", "The URL path for exporting server metrics for Prometheus monitoring.")

func init() {
	defaultPortInt := 8080
	defaultSecurePortInt := 8443
	currentUser, err := user.Current()
	if err == nil {
		if currentUser.Uid == "0" {
			defaultPortInt = 80
			defaultSecurePortInt = 443
		}
	}

	defaultPort := os.Getenv("PORT")
	if defaultPort != "" {
		port, err := strconv.Atoi(defaultPort)
		if err == nil {
			defaultPortInt = port
		}
	}

	httpPortFlag = flag.Int("http.port", defaultPortInt, "Port to run HTTP server.")
	httpsPortFlag = flag.Int("https.port", defaultSecurePortInt, "Port to run HTTPS server.")
}

func loadFromFlags() *Config {
	return &Config{
		Verbose:           *verboseFlag,
		Path:         *pathFlag,
		ServePath:         *servePathFlag,
		ConfigurationFile: *configFileFlag,
		UploadPath:   *uploadPathFlag,
		UploadServePath:   *uploadServePathFlag,
		HTTP: HTTP{
			Port: *httpPortFlag,
		},
		HTTPS: HTTPS{
			Port: *httpsPortFlag,
			Certificate: Certificate{
				PrivateKeyFilePath:        *privateKeyFilePathFlag,
				CertificateFilePath:       *certificateFilePathFlag,
				CertificateHosts:          *certHostsFlag,
				CertificateValidDuration:  *validDurationFlag,
				ActAsCertificateAuthority: *actAsCertificateAuthorityFlag,
				OnlyGenerateCertificate:   *onlyGenerateCertFlag,
				ForceOverwrite:            *onlyGenerateCertFlag, // No flag available yet.
			},
		},
		Metrics: Metrics{
			Enabled: *metricsFlag,
			Path:    *metricsPathFlag,
		},
	}
}
