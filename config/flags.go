package config

import (
	"flag"
	"os"
	"os/user"
	"strconv"
)

var httpsPortFlag *int
var httpPortFlag *int
var certificateFilePathFlag = flag.String("certificate.public", "cert.pem", "Certificate to host HTTPS with.")
var privateKeyFilePathFlag = flag.String("certificate.private", "rsa.pem", "Certificate to host HTTPS with.")
var rootDirectoryFlag = flag.String("directory", "", "The directory to serve.")
var servePathFlag = flag.String("http.path", "/", "HTTP serve root path for the filesystem.")

var metricsFlag = flag.Bool("metrics.enabled", true, "Enables server metrics for monitoring.")
var metricsPathFlag = flag.String("metrics.path", "/metrics", "The URL path for exporting server metrics for Prometheus montitoring.")

var certHostsFlag = flag.String("host", "", "Comma-separated hostnames and IPs to generate a certificate for.")
var validDurationFlag = flag.Int("certificate.duration", 365, "Certificate valid duration.")
var certificateAuthorityFlag = flag.Bool("certificate.authority", false, "(Experimental) Generate a root cert as a Certificate Authority")
var generateCertFlag = flag.Bool("certificate.onlygenerate", false, "Only generate a self-signed certificate for the server.")
var configFileFlag = flag.String("config.file", "", "Configuration file in YAML format.")

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
	httpsPortFlag = flag.Int("http.secureport", defaultSecurePortInt, "Port to run HTTPS server.")
}

func Load() *ServerConfiguration {
    flag.Parse()
    return Get()
}