package config

import (
	"flag"
	"os"
	"os/user"
	"strconv"
)

// Serving Flags
var serveDirectoryFlag = flag.String("serve.directory", "", "The directory on the local filesystem to serve.")
var servePathFlag = flag.String("serve.path", "/", "The HTTP/HTTPS serving root path for the hosted filesystem directory.")
var configFileFlag = flag.String("config.file", "", "YAML formatted configuration file. (overrides flag values)")

// HTTP Flags
var httpPortFlag *int

// HTTPS Flags
var httpsPortFlag *int
var privateKeyFilePathFlag = flag.String("https.privatekey", "rsa.pem", "Certificate to host HTTPS with.")
var certificateFilePathFlag = flag.String("https.certificate.path", "cert.pem", "Certificate to host HTTPS with.")
var certHostsFlag = flag.String("https.certificate.hosts", "", "Comma-separated hostnames and IPs to generate a certificate for.")
var validDurationFlag = flag.Int("https.certificate.duration", 365, "Certificate valid duration.")
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

func Load() *ServerConfiguration {
    flag.Parse()
    return Get()
}