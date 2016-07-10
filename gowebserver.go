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
	serverConf := config.Load()
	fmt.Printf("%v", serverConf)

	err := createCertificate(serverConf)
	if err != nil {
		log.Fatal(err)
	}
	if serverConf.OnlyGenerateCertificate {
		return
	}
	httpServer := server.NewWebServer().
		SetPorts(serverConf.HttpPort, serverConf.HttpPort).
		SetMetricsEnabled(serverConf.EnableMetrics).
		SetServePath(serverConf.ServePath, serverConf.MetricsHttpPath).
		SetCertificateFile(serverConf.CertificateFilePath).
		SetPrivateKey(serverConf.PrivateKeyFilePath)
	err = httpServer.SetDirectory(serverConf.ServeDirectory)
	if err != nil {
		log.Fatal(err)
	}
	httpServer.Serve()
}

func createCertificate(serverConf *config.ServerConfiguration) error {
	_, certErr := os.Stat(serverConf.CertificateFilePath)
	_, privateKeyErr := os.Stat(serverConf.PrivateKeyFilePath)
	if serverConf.ForceOverwrite || (os.IsNotExist(certErr) && os.IsNotExist(privateKeyErr)) {
		certBuilder := cert.NewCertificateBuilder().
			SetRsa2048().
			SetValidDurationInDays(serverConf.CertificateValidDuration).
			SetUseSelfAsCertificateAuthority(serverConf.ActAsCertificateAuthority)
		err := certBuilder.WriteCertificate(serverConf.CertificateFilePath)
		if err != nil {
			return err
		}
		err = certBuilder.WritePrivateKey(serverConf.PrivateKeyFilePath)
		if err != nil {
			return err
		}
	}
	return nil
}
