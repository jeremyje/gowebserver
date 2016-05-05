package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var httpsPortFlag *int
var httpPortFlag *int
var certificateFilePathFlag = flag.String("certificate", "cert.pem", "Certificate to host HTTPS with.")
var privateKeyFilePathFlag = flag.String("private_key", "rsa.pem", "Certificate to host HTTPS with.")
var rootDirectoryFlag = flag.String("directory", "", "The directory to serve.")

var certHostsFlag = flag.String("host", "", "Comma-separated hostnames and IPs to generate a certificate for.")
var validDurationFlag = flag.Int("certificate_duration", 365, "Certificate valid duration.")
var certificateAuthorityFlag = flag.Bool("ca", false, "(Experimental) Generate a root cert as a Certificate Authority")
var generateCertFlag = flag.Bool("generate_cert", false, "Only generate a self-signed certificate for the server.")

func init() {
	defaultPortInt := 8080
	defaultSecurePortInt := 8443
	currentUser, err := user.Current()
	if err == nil {
		if currentUser.Uid == "0" {
			defaultPortInt = 80
			defaultSecurePortInt = 443
		}
	}

	defaultPort := os.Getenv("PORT")
	if defaultPort != "" {
		port, err := strconv.Atoi(defaultPort)
		if err == nil {
			defaultPortInt = port
		}
	}

	httpPortFlag = flag.Int("port", defaultPortInt, "Port to run HTTP server.")
	httpsPortFlag = flag.Int("secure_port", defaultSecurePortInt, "Port to run HTTPS server.")
}

func main() {
	flag.Parse()
	err := createCertificate(*certificateFilePathFlag, *privateKeyFilePathFlag, *certHostsFlag, *validDurationFlag, *certificateAuthorityFlag, *generateCertFlag)
	if err != nil {
		log.Fatal(err)
	}
	if *generateCertFlag {
		return
	}
	server := NewWebServer()
	server.SetPorts(*httpPortFlag, *httpsPortFlag)
	err = server.SetDirectory(*rootDirectoryFlag)
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	server.SetCertificateFile(*certificateFilePathFlag)
	server.SetPrivateKey(*privateKeyFilePathFlag)
	server.Serve()
}

func createCertificate(certPath string, privateKeyPath string, hosts string, durationInDays int, isCa bool, overwrite bool) error {
	_, certErr := os.Stat(certPath)
	_, privateKeyErr := os.Stat(privateKeyPath)
	if overwrite || (os.IsNotExist(certErr) && os.IsNotExist(privateKeyErr)) {
		certBuilder := NewCertificateBuilder()
		certBuilder.SetRsa2048().SetValidDurationInDays(durationInDays).SetUseSelfAsCertificateAuthority(isCa)
		err := certBuilder.WriteCertificate(certPath)
		if err != nil {
			return err
		}
		err = certBuilder.WritePrivateKey(privateKeyPath)
		if err != nil {
			return err
		}
	}
	return nil
}

type WebServer interface {
	SetPorts(httpPort, httpsPort int)
	SetDirectory(dir string) error
	SetCertificateFile(certificateFilePath string)
	SetPrivateKey(privateKeyFilePath string)
	Serve()
}

type WebServerImpl struct {
	httpPort            string
	httpsPort           string
	certificateFilePath string
	privateKeyFilePath  string
	servingDirectory    string
}

func (this *WebServerImpl) SetPorts(httpPort, httpsPort int) {
	this.httpPort = ":" + strconv.Itoa(httpPort)
	this.httpsPort = ":" + strconv.Itoa(httpsPort)
}

func (this *WebServerImpl) SetDirectory(dir string) error {
	if len(dir) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		dir = cwd
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	this.servingDirectory = dir
	return nil
}

func (this *WebServerImpl) SetCertificateFile(certificateFilePath string) {
	this.certificateFilePath = certificateFilePath
}

func (this *WebServerImpl) SetPrivateKey(privateKeyFilePath string) {
	this.privateKeyFilePath = privateKeyFilePath
}

func (this *WebServerImpl) Serve() {
	log.Printf("Serving %s on %s and %s", this.servingDirectory, this.httpPort, this.httpsPort)
	fsHandler := http.FileServer(http.Dir(this.servingDirectory + "/"))
	go func() {
		err := http.ListenAndServeTLS(this.httpsPort, this.certificateFilePath, this.privateKeyFilePath, fsHandler)
		if err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		err := http.ListenAndServe(this.httpPort, fsHandler)
		if err != nil {
			log.Fatal(err)
		}
	}()
	ch := make(chan bool)
	<-ch
}

func NewWebServer() WebServer {
	return &WebServerImpl{
		httpPort:            "80",
		httpsPort:           "443",
		certificateFilePath: "",
		privateKeyFilePath:  "",
		servingDirectory:    "",
	}
}

// Creates X.509 certificates and private key pairs.
type CertificateBuilder interface {
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
	// Gets the X.509 certificate in PEM format as a byte string.
	GetCertificate() ([]byte, error)
	// Gets the private key in PEM format as a byte string.
	GetPrivateKey() ([]byte, error)
	// Writes the X.509 certificate in PEM format to a file.
	WriteCertificate(path string) error
	// Gets the private key in PEM format to a file.
	WritePrivateKey(path string) error
}

type CertificateBuilderImpl struct {
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
	certOrganization              string
	certOrganizationUnit          string
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
	return &CertificateBuilderImpl{
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
		certOrganization:              defaultCertOrg,
		certOrganizationUnit:          defaultCertOrgUnit,
	}
}

func (this *CertificateBuilderImpl) SetRsa2048() CertificateBuilder {
	this.rsaBits = 2048
	this.ecdsaCurve = nil
	this.isDirty = true
	return this
}

func (this *CertificateBuilderImpl) SetRsa4096() CertificateBuilder {
	this.rsaBits = 4096
	this.ecdsaCurve = nil
	this.isDirty = true
	return this
}

func (this *CertificateBuilderImpl) SetEcdsaP224() CertificateBuilder {
	this.rsaBits = 0
	this.ecdsaCurve = elliptic.P224()
	this.isDirty = true
	return this
}

func (this *CertificateBuilderImpl) SetEcdsaP256() CertificateBuilder {
	this.rsaBits = 0
	this.ecdsaCurve = elliptic.P256()
	this.isDirty = true
	return this
}

func (this *CertificateBuilderImpl) SetEcdsaP384() CertificateBuilder {
	this.rsaBits = 0
	this.ecdsaCurve = elliptic.P384()
	this.isDirty = true
	return this
}

func (this *CertificateBuilderImpl) SetEcdsaP521() CertificateBuilder {
	this.rsaBits = 0
	this.ecdsaCurve = elliptic.P521()
	this.isDirty = true
	return this
}

func (this *CertificateBuilderImpl) SetValidDurationInDays(numDays int) CertificateBuilder {
	this.validDuration = time.Duration(time.Hour * 24 * 365)
	this.isDirty = true
	return this
}

func (this *CertificateBuilderImpl) SetStartValidTime(startTime time.Time) CertificateBuilder {
	this.certValidStart = startTime
	this.isDirty = true
	return this
}

func (this *CertificateBuilderImpl) SetHostName(hostName string) CertificateBuilder {
	this.hostName = hostName
	this.isDirty = true
	return this
}

func (this *CertificateBuilderImpl) SetUseSelfAsCertificateAuthority(useSelf bool) CertificateBuilder {
	this.useSelfAsCertificateAuthority = useSelf
	this.isDirty = true
	return this
}

func (this *CertificateBuilderImpl) SetOrganization(organization string, unit string) CertificateBuilder {
	this.certOrganization = organization
	this.certOrganizationUnit = unit
	this.isDirty = true
	return this
}

func (this *CertificateBuilderImpl) GetCertificate() ([]byte, error) {
	this.buildCertificateIfNecessary()
	return this.x509PemBytes, this.buildError
}

func (this *CertificateBuilderImpl) GetPrivateKey() ([]byte, error) {
	this.buildCertificateIfNecessary()
	return this.privateKeyPemBytes, this.buildError
}

func (this *CertificateBuilderImpl) WriteCertificate(path string) error {
	x509PemBytes, err := this.GetCertificate()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, x509PemBytes, 0660)
}

func (this *CertificateBuilderImpl) WritePrivateKey(path string) error {
	privateKeyPemBytes, err := this.GetPrivateKey()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, privateKeyPemBytes, 0660)
}

func (this *CertificateBuilderImpl) buildCertificateIfNecessary() error {
	if this.isDirty {
		return this.buildCertificate()
	}
	return this.buildError
}

func (this *CertificateBuilderImpl) buildCertificate() error {
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
		Country:            []string{"US"},
		Organization:       []string{this.certOrganization},
		OrganizationalUnit: []string{this.certOrganizationUnit},
		Locality:           []string{"Seattle"},
		Province:           []string{"Washington"},
		CommonName:         this.certOrganization,
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
		return nil, errors.New("Invalid PEM format.")
	}
}
