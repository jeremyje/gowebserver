package main

import (
	"flag"
	"github.com/jeremyje/gowebserver/cert"
	"github.com/jeremyje/gowebserver/server"
	"log"
	"os"
	"os/user"
	"strconv"
)

var httpsPortFlag *int
var httpPortFlag *int
var certificateFilePathFlag = flag.String("certificate", "cert.pem", "Certificate to host HTTPS with.")
var privateKeyFilePathFlag = flag.String("private_key", "rsa.pem", "Certificate to host HTTPS with.")
var rootDirectoryFlag = flag.String("directory", "", "The directory to serve.")
var servePathFlag = flag.String("http_root", "/", "HTTP serve root path for the filesystem.")

var metricsFlag = flag.Bool("metrics", true, "Enables server metrics for monitoring.")
var metricsPathFlag = flag.String("metrics_path", "/metrics", "The URL path for exporting server metrics for Prometheus montitoring.")

var certHostsFlag = flag.String("host", "", "Comma-separated hostnames and IPs to generate a certificate for.")
var validDurationFlag = flag.Int("certificate_duration", 365, "Certificate valid duration.")
var certificateAuthorityFlag = flag.Bool("ca", false, "(Experimental) Generate a root cert as a Certificate Authority")
var generateCertFlag = flag.Bool("generate_cert", false, "Only generate a self-signed certificate for the server.")

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

	httpPortFlag = flag.Int("port", defaultPortInt, "Port to run HTTP server.")
	httpsPortFlag = flag.Int("secure_port", defaultSecurePortInt, "Port to run HTTPS server.")
}

func main() {
	flag.Parse()
	err := createCertificate(*certificateFilePathFlag, *privateKeyFilePathFlag, *certHostsFlag, *validDurationFlag, *certificateAuthorityFlag, *generateCertFlag)
	if err != nil {
		log.Fatal(err)
	}
	if *generateCertFlag {
		return
	}
	httpServer := server.NewWebServer()
	httpServer.SetPorts(*httpPortFlag, *httpsPortFlag)
	err = httpServer.SetDirectory(*rootDirectoryFlag)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	httpServer.SetMetricsEnabled(*metricsFlag)
	httpServer.SetServePath(*servePathFlag, *metricsPathFlag)

	httpServer.SetCertificateFile(*certificateFilePathFlag)
	httpServer.SetPrivateKey(*privateKeyFilePathFlag)
	httpServer.Serve()
}

func createCertificate(certPath string, privateKeyPath string, hosts string, durationInDays int, isCa bool, overwrite bool) error {
	_, certErr := os.Stat(certPath)
	_, privateKeyErr := os.Stat(privateKeyPath)
	if overwrite || (os.IsNotExist(certErr) && os.IsNotExist(privateKeyErr)) {
		certBuilder := cert.NewCertificateBuilder()
		certBuilder.SetRsa2048().SetValidDurationInDays(durationInDays).SetUseSelfAsCertificateAuthority(isCa)
		err := certBuilder.WriteCertificate(certPath)
		if err != nil {
			return err
		}
		err = certBuilder.WritePrivateKey(privateKeyPath)
		if err != nil {
			return err
		}
	}
	return nil
}
