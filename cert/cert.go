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

package cert

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
)

// ReadCertificateFromFile reads a certificate from a file.
func ReadCertificateFromFile(certPath string) (*x509.Certificate, error) {
	certData, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, err
	}
	return ReadCertificateFromBytes(certData)
}

// ReadCertificateFromBytes reads a certificate from a byte string.
func ReadCertificateFromBytes(certData []byte) (*x509.Certificate, error) {
	pemData, extraBytes := pem.Decode(certData)
	if len(pemData.Bytes) == 0 {
		return nil, fmt.Errorf("certificate is not encoded in PEM format, %d bytes", len(certData))
	}
	if len(extraBytes) > 0 {
		return nil, fmt.Errorf("certificate had additional information after the PEM encoded data, %d bytes", len(extraBytes))
	}
	return x509.ParseCertificate(pemData.Bytes)
}

// WriteDefaultCertificate writes a X.509 Certificate and RSA private key using default configuration.
func WriteDefaultCertificate(certPath string, privateKeyPath string) error {
	certBuilder := NewCertificateBuilder()
	err := certBuilder.WriteCertificate(certPath)
	if err != nil {
		return err
	}
	return certBuilder.WritePrivateKey(privateKeyPath)
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) (*pem.Block, error) {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}, nil
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, err
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}, nil
	default:
		return nil, errors.New("invalid PEM format")
	}
}
