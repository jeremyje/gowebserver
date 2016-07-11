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

	err := createCertificate(conf)
	if err != nil {
		log.Fatal(err)
	}
	if conf.Https.Certificate.OnlyGenerateCertificate {
		return
	}
	httpServer := server.NewWebServer().
		SetPorts(conf.Http.Port, conf.Https.Port).
		SetMetricsEnabled(conf.Metrics.Enabled).
		SetServePath(conf.ServePath, conf.Metrics.Path).
		SetCertificateFile(conf.Https.Certificate.CertificateFilePath).
		SetPrivateKey(conf.Https.Certificate.PrivateKeyFilePath).
		SetVerbose(conf.Verbose)
	err = httpServer.SetDirectory(conf.Directory)
	if err != nil {
		log.Fatal(err)
	}
	httpServer.Serve()
}

func createCertificate(conf *config.Config) error {
	_, certErr := os.Stat(conf.Https.Certificate.CertificateFilePath)
	_, privateKeyErr := os.Stat(conf.Https.Certificate.PrivateKeyFilePath)
	if conf.Https.Certificate.ForceOverwrite || (os.IsNotExist(certErr) && os.IsNotExist(privateKeyErr)) {
		certBuilder := cert.NewCertificateBuilder().
			SetRsa2048().
			SetValidDurationInDays(conf.Https.Certificate.CertificateValidDuration).
			SetUseSelfAsCertificateAuthority(conf.Https.Certificate.ActAsCertificateAuthority)
		err := certBuilder.WriteCertificate(conf.Https.Certificate.CertificateFilePath)
		if err != nil {
			return err
		}
		err = certBuilder.WritePrivateKey(conf.Https.Certificate.PrivateKeyFilePath)
		if err != nil {
			return err
		}
	}
	return nil
}
