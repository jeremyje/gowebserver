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
	err = httpServer.SetDirectory(conf.Directory)
	if err != nil {
		log.Fatal(err)
	}
	err = httpServer.SetUpload(conf.UploadDirectory, conf.UploadServePath)
	if err != nil {
		log.Fatal(err)
	}
	httpServer.Serve()
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
			return err
		}
		err = certBuilder.WritePrivateKey(conf.HTTPS.Certificate.PrivateKeyFilePath)
		if err != nil {
			return err
		}
	}
	return nil
}
