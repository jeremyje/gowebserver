package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"net"
	"os/user"
	"strings"
	"time"
)

// CertificateBuilder creates X.509 certificates and private key pairs.
type CertificateBuilder interface {
	// SetRsa1024 sets to build certificate and private key using RSA-1024.
	SetRsa1024() CertificateBuilder
	// Set to build certificate and private key using RSA-2048.
	SetRsa2048() CertificateBuilder
	// Set to build certificate and private key using RSA-4096.
	SetRsa4096() CertificateBuilder
	// Set to build certificate and private key using ECDSA P-224 elliptical curve.
	SetEcdsaP224() CertificateBuilder
	// Set to build certificate and private key using ECDSA P-256 elliptical curve.
	SetEcdsaP256() CertificateBuilder
	// Set to build certificate and private key using ECDSA P-384 elliptical curve.
	SetEcdsaP384() CertificateBuilder
	// Set to build certificate and private key using ECDSA P-521 elliptical curve.
	SetEcdsaP521() CertificateBuilder
	// Set the duration (in days) for the generated certificate.
	SetValidDurationInDays(numDays int) CertificateBuilder
	// Set the start time that the certificate is valid. (default is now)
	SetStartValidTime(startTime time.Time) CertificateBuilder
	// Set the host names that the certificate is for.
	SetHostName(hostName string) CertificateBuilder
	// Set the certificate to be used as the certificate authority.
	SetUseSelfAsCertificateAuthority(useSelf bool) CertificateBuilder
	// Set the certificate organization.
	SetOrganization(organization string, unit string) CertificateBuilder
	// Sets the country of the certificate origin.
	SetCountry(country string) CertificateBuilder
	// Sets the locality (city) of the certificate origin.
	SetLocality(locality string) CertificateBuilder
	// Sets the province (state or providence) of the certificate origin.
	SetProvince(province string) CertificateBuilder
	// Gets the X.509 certificate in PEM format as a byte string.
	GetCertificate() ([]byte, error)
	// Gets the private key in PEM format as a byte string.
	GetPrivateKey() ([]byte, error)
	// Gets the X.509 certificate in PEM format as a string.
	GetCertificateString() (string, error)
	// Gets the private key in PEM format as a string.
	GetPrivateKeyString() (string, error)
	// Writes the X.509 certificate in PEM format to a file.
	WriteCertificate(path string) error
	// Gets the private key in PEM format to a file.
	WritePrivateKey(path string) error
}

type certificateBuilderImpl struct {
	rsaBits                       int
	ecdsaCurve                    elliptic.Curve
	isDirty                       bool
	buildError                    error
	x509PemBytes                  []byte
	privateKeyPemBytes            []byte
	validDuration                 time.Duration
	certValidStart                time.Time
	hostName                      string
	useSelfAsCertificateAuthority bool
	organization                  string
	organizationUnit              string
	country                       string
	locality                      string
	province                      string
}

// NewCertificateBuilder creates a new Certificate Builder.
func NewCertificateBuilder() CertificateBuilder {
	defaultCertOrg := "Some Company"
	defaultCertOrgUnit := "None"
	currentUser, err := user.Current()
	if err == nil {
		defaultCertOrg = currentUser.Name
		defaultCertOrgUnit = currentUser.Username
	}
	return &certificateBuilderImpl{
		rsaBits:                       2048,
		ecdsaCurve:                    nil,
		isDirty:                       true,
		buildError:                    nil,
		x509PemBytes:                  nil,
		privateKeyPemBytes:            nil,
		validDuration:                 time.Duration(time.Hour * 24 * 365),
		certValidStart:                time.Now(),
		hostName:                      "",
		useSelfAsCertificateAuthority: true,
		organization:                  defaultCertOrg,
		organizationUnit:              defaultCertOrgUnit,
		country:                       "US",
		locality:                      "Seattle",
		province:                      "Washington",
	}
}

func (cb *certificateBuilderImpl) SetRsa1024() CertificateBuilder {
	cb.rsaBits = 1024
	cb.ecdsaCurve = nil
	return cb.dirty()
}

func (cb *certificateBuilderImpl) SetRsa2048() CertificateBuilder {
	cb.rsaBits = 2048
	cb.ecdsaCurve = nil
	return cb.dirty()
}

func (cb *certificateBuilderImpl) SetRsa4096() CertificateBuilder {
	cb.rsaBits = 4096
	cb.ecdsaCurve = nil
	return cb.dirty()
}

func (cb *certificateBuilderImpl) SetEcdsaP224() CertificateBuilder {
	cb.rsaBits = 0
	cb.ecdsaCurve = elliptic.P224()
	return cb.dirty()
}

func (cb *certificateBuilderImpl) SetEcdsaP256() CertificateBuilder {
	cb.rsaBits = 0
	cb.ecdsaCurve = elliptic.P256()
	return cb.dirty()
}

func (cb *certificateBuilderImpl) SetEcdsaP384() CertificateBuilder {
	cb.rsaBits = 0
	cb.ecdsaCurve = elliptic.P384()
	return cb.dirty()
}

func (cb *certificateBuilderImpl) SetEcdsaP521() CertificateBuilder {
	cb.rsaBits = 0
	cb.ecdsaCurve = elliptic.P521()
	return cb.dirty()
}

func (cb *certificateBuilderImpl) SetValidDurationInDays(numDays int) CertificateBuilder {
	cb.validDuration = time.Duration(time.Hour * 24 * 365)
	return cb.dirty()
}

func (cb *certificateBuilderImpl) SetStartValidTime(startTime time.Time) CertificateBuilder {
	cb.certValidStart = startTime
	return cb.dirty()
}

func (cb *certificateBuilderImpl) SetHostName(hostName string) CertificateBuilder {
	cb.hostName = hostName
	return cb.dirty()
}

func (cb *certificateBuilderImpl) SetUseSelfAsCertificateAuthority(useSelf bool) CertificateBuilder {
	cb.useSelfAsCertificateAuthority = useSelf
	return cb.dirty()
}

func (cb *certificateBuilderImpl) SetOrganization(organization string, unit string) CertificateBuilder {
	cb.organization = organization
	cb.organizationUnit = unit
	return cb.dirty()
}

func (cb *certificateBuilderImpl) SetCountry(country string) CertificateBuilder {
	cb.country = country
	return cb.dirty()
}

func (cb *certificateBuilderImpl) SetLocality(locality string) CertificateBuilder {
	cb.locality = locality
	return cb.dirty()
}

func (cb *certificateBuilderImpl) SetProvince(province string) CertificateBuilder {
	cb.province = province
	return cb.dirty()
}

func (cb *certificateBuilderImpl) GetCertificate() ([]byte, error) {
	cb.buildCertificateIfNecessary()
	return cb.x509PemBytes, cb.buildError
}

func (cb *certificateBuilderImpl) GetCertificateString() (string, error) {
	certBytes, err := cb.GetCertificate()
	return string(certBytes), err
}

func (cb *certificateBuilderImpl) GetPrivateKey() ([]byte, error) {
	cb.buildCertificateIfNecessary()
	return cb.privateKeyPemBytes, cb.buildError
}

func (cb *certificateBuilderImpl) GetPrivateKeyString() (string, error) {
	keyBytes, err := cb.GetPrivateKey()
	return string(keyBytes), err
}

func (cb *certificateBuilderImpl) WriteCertificate(path string) error {
	x509PemBytes, err := cb.GetCertificate()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, x509PemBytes, 0660)
}

func (cb *certificateBuilderImpl) WritePrivateKey(path string) error {
	privateKeyPemBytes, err := cb.GetPrivateKey()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, privateKeyPemBytes, 0660)
}

func (cb *certificateBuilderImpl) buildCertificateIfNecessary() error {
	if cb.isDirty {
		return cb.buildCertificate()
	}
	return cb.buildError
}

func (cb *certificateBuilderImpl) buildCertificate() error {
	var privateKey interface{}
	var err error
	if cb.ecdsaCurve == nil {
		privateKey, err = rsa.GenerateKey(rand.Reader, cb.rsaBits)
	} else {
		privateKey, err = ecdsa.GenerateKey(cb.ecdsaCurve, rand.Reader)
	}
	if err != nil {
		return cb.saveError(err)
	}

	certValidEnd := cb.certValidStart.Add(cb.validDuration)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return cb.saveError(err)
	}
	certName := pkix.Name{
		Country:            []string{cb.country},
		Organization:       []string{cb.organization},
		OrganizationalUnit: []string{cb.organizationUnit},
		Locality:           []string{cb.locality},
		Province:           []string{cb.province},
		CommonName:         cb.organization,
	}
	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               certName,
		Issuer:                certName,
		NotBefore:             cb.certValidStart,
		NotAfter:              certValidEnd,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(cb.hostName, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if cb.useSelfAsCertificateAuthority {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(privateKey), privateKey)
	if err != nil {
		return cb.saveError(err)
	}

	cb.x509PemBytes = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	pemPriv, err := pemBlockForKey(privateKey)
	if err != nil {
		return cb.saveError(err)
	}
	cb.privateKeyPemBytes = pem.EncodeToMemory(pemPriv)
	cb.isDirty = false
	return nil
}

func (cb *certificateBuilderImpl) saveError(err error) error {
	if err != nil {
		cb.buildError = err
	}
	return err
}

func (cb *certificateBuilderImpl) dirty() *certificateBuilderImpl {
	cb.isDirty = true
	return cb
}
