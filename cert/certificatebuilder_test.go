package cert

import (
	"crypto/x509"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

var sentinel interface{}

func TestBuildDefaultCertificate(t *testing.T) {
	assert := assert.New(t)

	certBuilder := NewCertificateBuilder()

	certData, err := certBuilder.GetCertificate()
	assert.Nil(err)
	cert, err := ReadCertificateFromBytes(certData)
	assert.Nil(err)
	assert.Nil(err)
	assert.Equal("US", cert.Subject.Country[0])
	assert.Equal("Seattle", cert.Subject.Locality[0])
	assert.Equal("Washington", cert.Subject.Province[0])
	privateKey, err := certBuilder.GetPrivateKey()
	assert.Nil(err)
	assert.NotNil(privateKey)
}

func TestBuildElipticalCertificate(t *testing.T) {
	assert := assert.New(t)

	certBuilder := NewCertificateBuilder()
	certBuilder.SetEcdsaP521()
	certBuilder.SetUseSelfAsCertificateAuthority(true)
	certBuilder.SetOrganization("test-runner", "test")
	certBuilder.SetCountry("CA")
	certBuilder.SetLocality("Vancouver")
	certBuilder.SetProvince("British Columbia")

	certData, err := certBuilder.GetCertificate()
	assert.Nil(err)
	cert, err := ReadCertificateFromBytes(certData)
	assert.Nil(err)
	assert.Equal(x509.ECDSAWithSHA512, cert.SignatureAlgorithm)
	assert.Equal("test-runner", cert.Subject.Organization[0])
	assert.Equal("test", cert.Subject.OrganizationalUnit[0])
	assert.Equal("CA", cert.Subject.Country[0])
	assert.Equal("Vancouver", cert.Subject.Locality[0])
	assert.Equal("British Columbia", cert.Subject.Province[0])
	privateKey, err := certBuilder.GetPrivateKey()
	assert.Nil(err)
	assert.NotNil(privateKey)
}

func ExampleNewCertificateBuilder() {
	certBuilder := NewCertificateBuilder().
		SetCountry("US").
		SetProvince("WA").
		SetLocality("Seattle").
		SetOrganization("Golang Test Runner", "development")
	certData, _ := certBuilder.GetCertificate()
	cert, _ := ReadCertificateFromBytes(certData)

	fmt.Printf("%s", cert.Subject.Organization[0])
	// Output: Golang Test Runner
}

func BenchmarkNewCertificateBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		certBuilder := NewCertificateBuilder()
		sentinel, _ = certBuilder.GetCertificate()
	}
}
