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

// Creates a new Certificate Builder.
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

func (this *certificateBuilderImpl) SetRsa1024() CertificateBuilder {
	this.rsaBits = 1024
	this.ecdsaCurve = nil
	this.isDirty = true
	return this
}

func (this *certificateBuilderImpl) SetRsa2048() CertificateBuilder {
	this.rsaBits = 2048
	this.ecdsaCurve = nil
	this.isDirty = true
	return this
}

func (this *certificateBuilderImpl) SetRsa4096() CertificateBuilder {
	this.rsaBits = 4096
	this.ecdsaCurve = nil
	this.isDirty = true
	return this
}

func (this *certificateBuilderImpl) SetEcdsaP224() CertificateBuilder {
	this.rsaBits = 0
	this.ecdsaCurve = elliptic.P224()
	this.isDirty = true
	return this
}

func (this *certificateBuilderImpl) SetEcdsaP256() CertificateBuilder {
	this.rsaBits = 0
	this.ecdsaCurve = elliptic.P256()
	this.isDirty = true
	return this
}

func (this *certificateBuilderImpl) SetEcdsaP384() CertificateBuilder {
	this.rsaBits = 0
	this.ecdsaCurve = elliptic.P384()
	this.isDirty = true
	return this
}

func (this *certificateBuilderImpl) SetEcdsaP521() CertificateBuilder {
	this.rsaBits = 0
	this.ecdsaCurve = elliptic.P521()
	this.isDirty = true
	return this
}

func (this *certificateBuilderImpl) SetValidDurationInDays(numDays int) CertificateBuilder {
	this.validDuration = time.Duration(time.Hour * 24 * 365)
	this.isDirty = true
	return this
}

func (this *certificateBuilderImpl) SetStartValidTime(startTime time.Time) CertificateBuilder {
	this.certValidStart = startTime
	this.isDirty = true
	return this
}

func (this *certificateBuilderImpl) SetHostName(hostName string) CertificateBuilder {
	this.hostName = hostName
	this.isDirty = true
	return this
}

func (this *certificateBuilderImpl) SetUseSelfAsCertificateAuthority(useSelf bool) CertificateBuilder {
	this.useSelfAsCertificateAuthority = useSelf
	this.isDirty = true
	return this
}

func (this *certificateBuilderImpl) SetOrganization(organization string, unit string) CertificateBuilder {
	this.organization = organization
	this.organizationUnit = unit
	this.isDirty = true
	return this
}

func (this *certificateBuilderImpl) SetCountry(country string) CertificateBuilder {
	this.country = country
	this.isDirty = true
	return this
}

func (this *certificateBuilderImpl) SetLocality(locality string) CertificateBuilder {
	this.locality = locality
	this.isDirty = true
	return this
}

func (this *certificateBuilderImpl) SetProvince(province string) CertificateBuilder {
	this.province = province
	this.isDirty = true
	return this
}

func (this *certificateBuilderImpl) GetCertificate() ([]byte, error) {
	this.buildCertificateIfNecessary()
	return this.x509PemBytes, this.buildError
}

func (this *certificateBuilderImpl) GetCertificateString() (string, error) {
	certBytes, err := this.GetCertificate()
	return string(certBytes), err
}

func (this *certificateBuilderImpl) GetPrivateKey() ([]byte, error) {
	this.buildCertificateIfNecessary()
	return this.privateKeyPemBytes, this.buildError
}

func (this *certificateBuilderImpl) GetPrivateKeyString() (string, error) {
	keyBytes, err := this.GetPrivateKey()
	return string(keyBytes), err
}

func (this *certificateBuilderImpl) WriteCertificate(path string) error {
	x509PemBytes, err := this.GetCertificate()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, x509PemBytes, 0660)
}

func (this *certificateBuilderImpl) WritePrivateKey(path string) error {
	privateKeyPemBytes, err := this.GetPrivateKey()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, privateKeyPemBytes, 0660)
}

func (this *certificateBuilderImpl) buildCertificateIfNecessary() error {
	if this.isDirty {
		return this.buildCertificate()
	}
	return this.buildError
}

func (this *certificateBuilderImpl) buildCertificate() error {
	var privateKey interface{}
	var err error
	if this.ecdsaCurve == nil {
		privateKey, err = rsa.GenerateKey(rand.Reader, this.rsaBits)
	} else {
		privateKey, err = ecdsa.GenerateKey(this.ecdsaCurve, rand.Reader)
	}
	if err != nil {
		this.buildError = err
		return err
	}

	certValidEnd := this.certValidStart.Add(this.validDuration)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		this.buildError = err
		return err
	}
	certName := pkix.Name{
		Country:            []string{this.country},
		Organization:       []string{this.organization},
		OrganizationalUnit: []string{this.organizationUnit},
		Locality:           []string{this.locality},
		Province:           []string{this.province},
		CommonName:         this.organization,
	}
	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               certName,
		Issuer:                certName,
		NotBefore:             this.certValidStart,
		NotAfter:              certValidEnd,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(this.hostName, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if this.useSelfAsCertificateAuthority {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(privateKey), privateKey)
	if err != nil {
		this.buildError = err
		return err
	}

	this.x509PemBytes = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	pemPriv, err := pemBlockForKey(privateKey)
	if err != nil {
		this.buildError = err
		return err
	}
	this.privateKeyPemBytes = pem.EncodeToMemory(pemPriv)
	this.isDirty = false
	return nil
}
