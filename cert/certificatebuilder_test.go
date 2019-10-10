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
	"crypto/x509"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
