// Copyright 2019 Jeremy Edwards
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jeremyje/gowebserver/cert"
	"github.com/jeremyje/gowebserver/config"
	"github.com/jeremyje/gowebserver/server"
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
