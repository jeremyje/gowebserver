package main

import (
	"fmt"
	"github.com/jeremyje/gowebserver/cert"
	"github.com/jeremyje/gowebserver/config"
	"github.com/jeremyje/gowebserver/server"
	"log"
	"os"
)

func main() {
	conf := config.Load()
	if conf.Verbose {
		fmt.Printf("%v", conf)
	}

	checkError(createCertificate(conf))
	if conf.HTTPS.Certificate.OnlyGenerateCertificate {
		return
	}
	httpServer := server.NewWebServer().
		SetPorts(conf.HTTP.Port, conf.HTTPS.Port).
		SetMetricsEnabled(conf.Metrics.Enabled).
		SetServePath(conf.ServePath, conf.Metrics.Path).
		SetCertificateFile(conf.HTTPS.Certificate.CertificateFilePath).
		SetPrivateKey(conf.HTTPS.Certificate.PrivateKeyFilePath).
		SetVerbose(conf.Verbose)
	checkError(httpServer.SetPath(conf.Path))
	checkError(httpServer.SetUpload(conf.UploadPath, conf.UploadServePath))
	httpServer.Serve()
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func createCertificate(conf *config.Config) error {
	_, certErr := os.Stat(conf.HTTPS.Certificate.CertificateFilePath)
	_, privateKeyErr := os.Stat(conf.HTTPS.Certificate.PrivateKeyFilePath)
	if conf.HTTPS.Certificate.ForceOverwrite || (os.IsNotExist(certErr) && os.IsNotExist(privateKeyErr)) {
		certBuilder := cert.NewCertificateBuilder().
			SetRsa2048().
			SetValidDurationInDays(conf.HTTPS.Certificate.CertificateValidDuration).
			SetUseSelfAsCertificateAuthority(conf.HTTPS.Certificate.ActAsCertificateAuthority)
		err := certBuilder.WriteCertificate(conf.HTTPS.Certificate.CertificateFilePath)
		if err != nil {
			return fmt.Errorf("cannot write public certificate, %s", err)
		}
		err = certBuilder.WritePrivateKey(conf.HTTPS.Certificate.PrivateKeyFilePath)
		if err != nil {
			return fmt.Errorf("cannot write private key, %s", err)
		}
	}
	return nil
}
