// Copyright 2022 Jeremy Edwards
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
	"crypto/x509"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jeremyje/gowebserver/v2/pkg/httpprobe"
	"errors"
)

const (
	errnoSuccess      = 0
	errnoFileNotFound = 2
)

var (
	publicCertificate = flag.String("public-certificate", "", "X.509 public certificate file to validate.")
	url               = flag.String("url", "http://localhost:8080", "URL of the HTTP/HTTPS endpoint to probe for availability.")
	timeout           = flag.Duration("timeout", time.Second*5, "Time to wait for a response before giving up.")
)

func main() {
	exitCode(run(*url, *publicCertificate, *timeout))
}

func exitCode(exitCode int) {
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

func run(url string, publicCertFile string, timeout time.Duration) int {
	certData, err := readPublicCertificate(publicCertFile)
	if err != nil {
		return errnoFileNotFound
	}

	err = httpprobe.Probe(httpprobe.Args{
		URL:             url,
		CertificatePool: certData,
		Timeout:         timeout,
	})

	var perr httpprobe.ProbeError
	if errors.As(err, &perr) {
		return perr.Code
	}

	return errnoSuccess
}

func readPublicCertificate(publicCertFile string) (*x509.CertPool, error) {
	if publicCertFile == "" {
		return nil, nil
	}

	data, err := os.ReadFile(publicCertFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %s, %w", publicCertFile, err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(data) {
		return nil, fmt.Errorf("cannot parse certificate file '%s', %w", publicCertFile, err)
	}
	return pool, nil
}
