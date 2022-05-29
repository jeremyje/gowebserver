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

// Package certtool is the public interface for integrating with auto generating certificates.
package certtool

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	oneYear                          = time.Hour * 24 * 365
	rsaPrivateKeyPEMType             = "RSA PRIVATE KEY"
	ecPrivateKeyPEMType              = "EC PRIVATE KEY"
	numDistinguishedNameSegmentParts = 2
	subjectDelimiter                 = "/"
	partDelimiter                    = "="
	subjectDelimiterReplacement      = "\u2318"
)

// Args of creating a self-signed X.509 public certificate/private key pair.
type Args struct {
	// CA indicates we need to create a CA certificate.
	CA bool

	// CommonName
	CommonName string
	// Country of the entity representing the certificate.
	Country string
	// Organization of the entity representing the certificate.
	Organization string
	// OrganizationalUnit of the entity representing the certificate.
	OrganizationalUnit string
	// Locality of the entity representing the certificate.
	Locality string
	// Province (or state) of the entity representing the certificate.
	Province string

	// Hostnames is a list of hostname (optional :port) of the endpoint used by the certificate.
	Hostnames []string
	// Validity is how long the certificate should be valid for.
	Validity time.Duration
	// ParentKeyPair is the root public certificate within the chain of trust.
	ParentKeyPair *KeyPair

	// KeyType is the type of key to generate.
	KeyType *KeyType
}

func (args *Args) GetKeyType() *KeyType {
	if args == nil || args.KeyType == nil {
		return defaultKeyType()
	}

	return args.KeyType
}

// KeyType is the key descriptor.
type KeyType struct {
	// Algorithm of the encryption
	Algorithm string
	// KeyLength is the length in bytes of the key.
	KeyLength int
}

// KeyPair is the X.509 public certificate/private key pair
type KeyPair struct {
	// PublicCertificate of the X.509 key pair.
	PublicCertificate []byte
	// PrivateKey of the X.509 key pair.
	PrivateKey []byte
}

func fillDefaults(args *Args) {
	if args.Country == "" {
		args.Country = "US"
	}
	if args.Organization == "" {
		args.Organization = "Certtool"
	}
	if args.CommonName == "" {
		args.CommonName = args.Organization
	}
	if args.OrganizationalUnit == "" {
		args.OrganizationalUnit = "None"
	}
	if args.Locality == "" {
		args.Locality = "Seattle"
	}
	if args.Province == "" {
		args.Province = "WA"
	}

	if args.Validity == 0 {
		args.Validity = oneYear
	}

	if args.KeyType == nil {
		args.KeyType = defaultKeyType()
	}
}

func defaultKeyType() *KeyType {
	return &KeyType{
		Algorithm: "RSA",
		KeyLength: 2048,
	}
}

func GenerateKeyPair(args *Args) (*KeyPair, error) {
	fillDefaults(args)

	return createCertificateAndPrivateKeyPEM(args)
}

func createCertificateAndPrivateKeyPEM(args *Args) (*KeyPair, error) {
	var sigAlg x509.SignatureAlgorithm

	startTimestamp := time.Now()
	expirationTimestamp := startTimestamp.Add(args.Validity)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create serial number for X.509 certificate")
	}

	pkixName := argsToPkixName(args, serialNumber.String())

	keyType := args.GetKeyType()
	switch strings.ToUpper(keyType.Algorithm) {
	case "RSA":
		sigAlg = x509.SHA512WithRSA
	case "ECDSA":
		sigAlg = x509.ECDSAWithSHA512
	default:
		return nil, fmt.Errorf("key algorithm, %s, is not valid", keyType.Algorithm)
	}

	certTemplate := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkixName,
		Issuer:                pkixName,
		NotBefore:             startTimestamp,
		NotAfter:              expirationTimestamp,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		SignatureAlgorithm:    sigAlg,
		IsCA:                  args.CA,
	}

	if args.CA {
		certTemplate.KeyUsage |= x509.KeyUsageCertSign
	}

	for _, hostname := range args.Hostnames {
		if ipAddress := net.ParseIP(hostname); ipAddress != nil {
			certTemplate.IPAddresses = append(certTemplate.IPAddresses, ipAddress)
		} else {
			certTemplate.DNSNames = append(certTemplate.DNSNames, hostname)
		}
	}

	certTemplate.DNSNames = append(certTemplate.DNSNames, "localhost")
	certTemplate.IPAddresses = append(certTemplate.IPAddresses, net.ParseIP("127.0.0.1"))

	privateKey, err := generatePrivateKeyFromType(*keyType)
	if err != nil {
		return nil, errors.Wrap(err, "cannot generate private key")
	}

	parentPrivateKey := privateKey
	parentTemplate := certTemplate

	if args.ParentKeyPair != nil {
		if len(args.ParentKeyPair.PublicCertificate) > 0 && len(args.ParentKeyPair.PrivateKey) == 0 {
			return nil, fmt.Errorf("root public certificate data was set but root private key data was not")
		} else if len(args.ParentKeyPair.PublicCertificate) == 0 && len(args.ParentKeyPair.PrivateKey) > 0 {
			return nil, fmt.Errorf("root private key data was set but root public certificate data was not")
		} else if len(args.ParentKeyPair.PublicCertificate) > 0 && len(args.ParentKeyPair.PrivateKey) > 0 {
			parentPublicCertificateT, parentPrivateKeyT, readErr := ReadKeyPair(args.ParentKeyPair.PublicCertificate, args.ParentKeyPair.PrivateKey)
			if readErr != nil {
				return nil, readErr
			}

			parentTemplate = *parentPublicCertificateT
			parentPrivateKey = parentPrivateKeyT
		}
	}

	cert, err := x509.CreateCertificate(rand.Reader, &certTemplate, &parentTemplate, publicKey(privateKey), parentPrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create X.509 public certificate")
	}

	publicCertificatePemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})

	pemPrivateKey, err := pemBlockForKey(privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create PEM block for private key")
	}

	privateKeyPemBytes := pem.EncodeToMemory(pemPrivateKey)

	return &KeyPair{
		PublicCertificate: publicCertificatePemBytes,
		PrivateKey:        privateKeyPemBytes,
	}, nil
}

func argsToPkixName(args *Args, serialNumber string) pkix.Name {
	return pkix.Name{
		Country:            []string{args.Country},
		Organization:       []string{args.Organization},
		OrganizationalUnit: []string{args.OrganizationalUnit},
		Locality:           []string{args.Locality},
		Province:           []string{args.Province},
		CommonName:         args.CommonName,
		SerialNumber:       serialNumber,
	}
}

func generatePrivateKeyFromType(keyType KeyType) (interface{}, error) {
	switch strings.ToUpper(keyType.Algorithm) {
	case "RSA":
		if keyType.KeyLength < 2048 {
			return nil, errors.Errorf("'%s-%d' key type has a key length below 2048", keyType.Algorithm, keyType.KeyLength)
		}
		return rsa.GenerateKey(rand.Reader, keyType.KeyLength)
	case "ECDSA":
		switch keyType.KeyLength {
		case 224:
			return ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
		case 256:
			return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		case 384:
			return ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		case 521:
			fallthrough
		case 0:
			return ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
		}
	}
	return nil, errors.Errorf("key type '%v' is not valid", keyType)
}

// ReadKeyPair takes PEM-encoded public certificate/private key pairs and returns the Go classes for them so they can be used for encryption or signing.
func ReadKeyPair(publicCertFileData []byte, privateKeyFileData []byte) (*x509.Certificate, interface{}, error) {
	// Verify that we can load the public/private key pair.
	publicCertPemBlock, remainder := pem.Decode(publicCertFileData)
	if len(remainder) > 0 {
		return nil, nil, errors.Errorf("public certificate has a PEM remainder of %d bytes", len(remainder))
	}

	publicCertificate, err := x509.ParseCertificate(publicCertPemBlock.Bytes)
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot parse X.509 certificate")
	}

	privateKeyPemBlock, remainder := pem.Decode(privateKeyFileData)
	if len(remainder) > 0 {
		return nil, nil, errors.Errorf("private key has a PEM remainder of %d bytes", len(remainder))
	}

	if privateKeyPemBlock.Type == rsaPrivateKeyPEMType {
		privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyPemBlock.Bytes)
		if err != nil {
			return nil, nil, errors.Wrap(err, "cannot parse PKCS1 encoding for private key")
		}

		return publicCertificate, privateKey, nil
	} else if privateKeyPemBlock.Type == ecPrivateKeyPEMType {
		privateKey, err := x509.ParseECPrivateKey(privateKeyPemBlock.Bytes)
		if err != nil {
			return nil, nil, errors.Wrap(err, "cannot parse elliptical curve private key")
		}

		return publicCertificate, privateKey, nil
	}

	return nil, nil, errors.Errorf("cannot parse private key PEM type, %s, is not supported", privateKeyPemBlock.Type)
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

func pemBlockForKey(privateKey interface{}) (*pem.Block, error) {
	switch k := privateKey.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: rsaPrivateKeyPEMType, Bytes: x509.MarshalPKCS1PrivateKey(k)}, nil
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, err
		}

		return &pem.Block{Type: ecPrivateKeyPEMType, Bytes: b}, nil
	default:
		return nil, errors.Errorf("private Key %v is not a valid private key", privateKey)
	}
}

// ParseName parses the RFC-2253 encoded Distinguished Names.
func ParseName(subject string) (pkix.Name, error) {
	name := pkix.Name{}
	subject = strings.ReplaceAll(subject, "\\"+subjectDelimiter, subjectDelimiterReplacement)
	parts := strings.Split(subject, subjectDelimiter)

	for _, part := range parts {
		vals := strings.Split(part, partDelimiter)
		if len(vals) < numDistinguishedNameSegmentParts {
			continue
		} else if len(vals) > numDistinguishedNameSegmentParts {
			return pkix.Name{}, errors.Errorf("AttributeType '%s' has too many parts, %v", vals[0], vals)
		}

		value := strings.ReplaceAll(vals[1], subjectDelimiterReplacement, ",")

		switch strings.ToUpper(vals[0]) {
		case "CN":
			name.CommonName = value
		case "OU":
			name.OrganizationalUnit = append(name.OrganizationalUnit, value)
		case "O":
			name.Organization = append(name.Organization, value)
		case "L":
			name.Locality = append(name.Locality, value)
		case "ST":
			name.Province = append(name.Province, value)
		case "C":
			name.Country = append(name.Country, value)
		case "STREET":
			name.StreetAddress = append(name.StreetAddress, value)
		case "POSTALCODE":
			name.PostalCode = append(name.PostalCode, value)
		default:
			return pkix.Name{}, errors.Errorf("'%s' is not a valid RFC-2253 AttributeType", vals[0])
		}
	}

	return name, nil
}

// ReadKeyPairFromFile is a convenience method for loading the key pair from a file.
func ReadKeyPairFromFile(publicCertificateFile string, privateKeyFile string) (*KeyPair, error) {
	if publicCertificateFile == "" && privateKeyFile == "" {
		return nil, fmt.Errorf("public certificate and private key were not provided")
	}
	if publicCertificateFile != "" && privateKeyFile == "" {
		return nil, fmt.Errorf("public certificate was provided without a private key")
	}
	if publicCertificateFile == "" && privateKeyFile != "" {
		return nil, fmt.Errorf("private key was provided without a public certificate")
	}

	pub, err := ioutil.ReadFile(publicCertificateFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read the public certificate file (%s), %s", publicCertificateFile, err)
	}
	priv, err := ioutil.ReadFile(privateKeyFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read the private key file (%s), %s", privateKeyFile, err)
	}

	return &KeyPair{
		PublicCertificate: pub,
		PrivateKey:        priv,
	}, nil
}

func GenerateAndWriteKeyPair(args *Args, publicCertificateFile string, privateKeyFile string) (*KeyPair, error) {
	if len(publicCertificateFile) == 0 {
		return nil, fmt.Errorf("public certificate file path must not be empty")
	}
	if len(privateKeyFile) == 0 {
		return nil, fmt.Errorf("private key file path must not be empty")
	}
	kp, err := GenerateKeyPair(args)
	if err != nil {
		return nil, err
	}
	if err := WriteKeyPair(kp, publicCertificateFile, privateKeyFile); err != nil {
		return nil, err
	}
	return kp, nil
}

func WriteKeyPair(kp *KeyPair, publicCertificateFile string, privateKeyFile string) error {
	if err := writeFile(publicCertificateFile, kp.PublicCertificate); err != nil {
		return err
	}
	if err := writeFile(privateKeyFile, kp.PrivateKey); err != nil {
		return err
	}
	return nil
}

func writeFile(fileName string, data []byte) error {
	return ioutil.WriteFile(fileName, data, 0644)
}
