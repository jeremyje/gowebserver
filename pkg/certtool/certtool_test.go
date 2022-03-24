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

package certtool

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	secretMessage = "this is a secret message"
)

func TestPublicKey(t *testing.T) {
	if testing.Short() {
		t.Skip("certificate generation takes a long time")
	}

	assert := assert.New(t)

	assert.Nil(publicKey(&Args{}))
	assert.NotNil(publicKey(&rsa.PrivateKey{}))
	assert.NotNil(publicKey(&ecdsa.PrivateKey{}))
}

func TestReadKeyPair_BadPublicCert(t *testing.T) {
	assert := assert.New(t)

	pubCert, pk, err := ReadKeyPair([]byte("bad"), []byte("bad"))
	assert.Nil(pubCert)
	assert.Nil(pk)
	assert.Contains(err.Error(), "public certificate has a PEM remainder")
}

func TestReadKeyPair_MalformedRootCerts(t *testing.T) {
	if testing.Short() {
		t.Skip("certificate generation takes a long time")
	}

	assert := assert.New(t)

	pair, err := createCertificateAndPrivateKeyPEM(&Args{
		KeyType: defaultKeyType(),
		ParentKeyPair: &KeyPair{
			PublicCertificate: []byte("lol"),
			PrivateKey:        []byte("lol"),
		},
	})
	assert.Contains(err.Error(), "public certificate has a PEM remainder")
	assert.Nil(pair)
}

func TestReadKeyPair_KeyTypeMismatch(t *testing.T) {
	if testing.Short() {
		t.Skip("certificate generation takes a long time")
	}

	assert := assert.New(t)

	rootPair, err := createCertificateAndPrivateKeyPEM(&Args{
		Validity:  time.Hour * 1,
		Hostnames: []string{"a.com", "b.com", "127.0.0.1"},
		KeyType:   defaultKeyType(),
		CA:        true,
	})
	if err != nil {
		t.Fatal(err)
	}

	derivedPair, err := createCertificateAndPrivateKeyPEM(&Args{
		KeyType: &KeyType{
			Algorithm: "ECDSA",
			KeyLength: 224,
		},
		ParentKeyPair: rootPair,
	})
	assert.Contains(err.Error(), "cannot create X.509 public certificate")
	assert.Nil(derivedPair)
}

func TestReadKeyPair_MalformedPublicCertificate(t *testing.T) {
	if testing.Short() {
		t.Skip("certificate generation takes a long time")
	}

	assert := assert.New(t)

	pair, err := createCertificateAndPrivateKeyPEM(&Args{
		KeyType: defaultKeyType(),
	})
	assert.Nil(err)
	assert.NotNil(pair.PrivateKey)
	assert.NotNil(pair.PublicCertificate)

	malformedPublicKey := []byte(strings.ReplaceAll(string(pair.PublicCertificate), "MII", "MIE"))

	pubCert, pk, err := ReadKeyPair(malformedPublicKey, pair.PrivateKey)
	assert.Nil(pubCert)
	assert.Nil(pk)
	assert.Contains(err.Error(), "malformed")
}

func TestReadKeyPair_MalformedPrivateKey(t *testing.T) {
	if testing.Short() {
		t.Skip("certificate generation takes a long time")
	}

	assert := assert.New(t)

	pair, err := createCertificateAndPrivateKeyPEM(&Args{
		KeyType: defaultKeyType(),
	})
	assert.Nil(err)
	assert.NotNil(pair.PrivateKey)
	assert.NotNil(pair.PublicCertificate)
	malformedPriv := []byte(strings.ReplaceAll(string(pair.PrivateKey), rsaPrivateKeyPEMType, ecPrivateKeyPEMType))

	pubCert, pk, err := ReadKeyPair(pair.PublicCertificate, malformedPriv)
	assert.Nil(pubCert)
	assert.Nil(pk)
	assert.Contains(err.Error(), "x509: failed to parse")
	assert.Contains(err.Error(), "private key")

	malformedPriv = pair.PrivateKey
	// Increment some bit in the middle of the payload.
	malformedPriv[500]++

	pubCert, pk, err = ReadKeyPair(pair.PublicCertificate, malformedPriv)
	assert.Nil(pubCert)
	assert.Nil(pk)
	assert.NotNil(err)

	malformedPriv = []byte(strings.ReplaceAll(string(pair.PrivateKey), rsaPrivateKeyPEMType, "IDK"))

	pubCert, pk, err = ReadKeyPair(pair.PublicCertificate, malformedPriv)
	assert.Nil(pubCert)
	assert.Nil(pk)
	assert.NotEmpty(err.Error())
}

func TestReadKeyPair_BadPrivateKey(t *testing.T) {
	if testing.Short() {
		t.Skip("certificate generation takes a long time")
	}

	assert := assert.New(t)

	pair, err := createCertificateAndPrivateKeyPEM(&Args{})
	assert.Nil(err)
	assert.NotNil(pair.PrivateKey)
	assert.NotNil(pair.PublicCertificate)

	pubCert, pk, err := ReadKeyPair(pair.PublicCertificate, []byte("bad"))
	assert.Nil(pubCert)
	assert.Nil(pk)
	assert.Contains(err.Error(), "private key has a PEM remainder")
}

func TestArgsToPkixName(t *testing.T) {
	testCases := []struct {
		input Args
		want  pkix.Name
	}{
		{
			Args{},
			pkix.Name{
				Country:            []string{""},
				Organization:       []string{""},
				OrganizationalUnit: []string{""},
				Locality:           []string{""},
				Province:           []string{""},
				CommonName:         "",
				SerialNumber:       "1",
			},
		},
		{
			Args{
				Country:            "country",
				Organization:       "organization",
				OrganizationalUnit: "organizationUnit",
				Locality:           "locality",
				Province:           "province",
				CommonName:         "organization",
			},
			pkix.Name{
				Country:            []string{"country"},
				Organization:       []string{"organization"},
				OrganizationalUnit: []string{"organizationUnit"},
				Locality:           []string{"locality"},
				Province:           []string{"province"},
				CommonName:         "organization",
				SerialNumber:       "1",
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%v", tc.input), func(t *testing.T) {
			t.Parallel()
			actual := argsToPkixName(&tc.input, "1")
			if !reflect.DeepEqual(actual, tc.want) {
				t.Errorf("pkix.Name are different\ngot %v\nwant: %v", actual, tc.want)
			}
		})
	}
}

func TestCreateCACertificateWithECDSA(t *testing.T) {
	if testing.Short() {
		t.Skip("certificate generation takes a long time")
	}

	testCases := []struct {
		keyType KeyType
	}{
		{keyType: KeyType{Algorithm: "ECDSA", KeyLength: 224}},
		{keyType: KeyType{Algorithm: "ECDSA", KeyLength: 256}},
		{keyType: KeyType{Algorithm: "ECDSA", KeyLength: 384}},
		{keyType: KeyType{Algorithm: "ECDSA", KeyLength: 521}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%v", tc.keyType), func(t *testing.T) {
			t.Parallel()
			assert := assert.New(t)

			rootPair, err := createCertificateAndPrivateKeyPEM(&Args{
				Validity:  time.Hour * 1,
				Hostnames: []string{"a.com", "b.com", "127.0.0.1"},
				KeyType:   &tc.keyType,
				CA:        true,
			})
			assert.Nil(err)

			derivedPair, err := createCertificateAndPrivateKeyPEM(&Args{
				Validity:      time.Hour * 1,
				Hostnames:     []string{"a.com", "b.com", "127.0.0.1"},
				KeyType:       &tc.keyType,
				CA:            false,
				ParentKeyPair: rootPair,
			})
			assert.Nil(err)

			// Verify that we can load the public/private key pair.
			rootPub, _, err := ReadKeyPair(rootPair.PublicCertificate, rootPair.PrivateKey)
			assert.Nil(err)
			assert.NotNil(rootPub)

			// Verify that we can load the public/private key pair.
			pub, pk, err := ReadKeyPair(derivedPair.PublicCertificate, derivedPair.PrivateKey)
			assert.Nil(err)
			assert.NotNil(pub)
			assert.NotNil(pk)
			pkEcdsa, ok := pk.(*ecdsa.PrivateKey)
			assert.True(ok)
			pubEcdsa, ok := pub.PublicKey.(*ecdsa.PublicKey)
			assert.True(ok)

			hash := sha256.Sum256([]byte(secretMessage))
			r, s, err := ecdsa.Sign(rand.Reader, pkEcdsa, hash[:])
			assert.Nil(err)
			verified := ecdsa.Verify(pubEcdsa, hash[:], r, s)
			assert.True(verified)

			// Validate certificate rootness.
			pool := x509.NewCertPool()
			ok = pool.AppendCertsFromPEM(rootPair.PublicCertificate)
			assert.True(ok)

			assert.Nil(pub.CheckSignatureFrom(rootPub))
		})
	}
}

func TestGenerateKeyPair(t *testing.T) {
	if testing.Short() {
		t.Skip("certificate generation takes a long time")
	}

	assert := assert.New(t)

	rootPair, err := GenerateKeyPair(&Args{
		Hostnames: []string{"a.com", "b.com", "127.0.0.1"},
	})
	assert.Nil(err)

	// Verify that we can load the public/private key pair.
	publicCert, privateKey, err := ReadKeyPair(rootPair.PublicCertificate, rootPair.PrivateKey)
	assert.Nil(err)
	assert.NotNil(publicCert)
	assert.NotNil(privateKey)

	// Verify that we can load the public/private key pair.
	pkRSA, ok := privateKey.(*rsa.PrivateKey)
	assert.True(ok)
	pubRSA, ok := publicCert.PublicKey.(*rsa.PublicKey)
	assert.True(ok)

	hash := sha256.Sum256([]byte(secretMessage))

	sig, err := rsa.SignPKCS1v15(rand.Reader, pkRSA, crypto.SHA256, hash[:])
	assert.Nil(err)
	err = rsa.VerifyPKCS1v15(pubRSA, crypto.SHA256, hash[:], sig)
	assert.Nil(err)
}

func TestFillDefaults(t *testing.T) {
	args := &Args{}
	fillDefaults(args)

	testCases := []struct {
		fieldName string
		got       string
		want      string
	}{
		{fieldName: "args.Country", got: args.Country, want: "US"},
		{fieldName: "args.Organization", got: args.Organization, want: "Certtool"},
		{fieldName: "args.CommonName", got: args.CommonName, want: "Certtool"},
		{fieldName: "args.OrganizationalUnit", got: args.OrganizationalUnit, want: "None"},
		{fieldName: "args.Locality", got: args.Locality, want: "Seattle"},
		{fieldName: "args.Province", got: args.Province, want: "WA"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%v", tc.fieldName), func(t *testing.T) {
			t.Parallel()
			if tc.got != tc.want {
				t.Errorf("%s = %s; want: %v", tc.fieldName, tc.got, tc.want)
			}
		})
	}
}

func TestCreateCertificateToBadPath(t *testing.T) {
	assert := assert.New(t)

	tmpDir, err := ioutil.TempDir("", "certtest")

	defer func() {
		assert.Nil(os.RemoveAll(tmpDir))
	}()

	assert.Nil(err)

	publicCertPath := filepath.Join(tmpDir, "public.cert")

	assert.Contains(GenerateAndWriteKeyPair(
		&Args{
			Validity:  time.Hour * 1,
			Hostnames: []string{"a.com", "b.com", "127.0.0.1"},
			KeyType:   defaultKeyType(),
		},
		"does-not-exist/pub.cert",
		"does-not-exist/private.key",
	).Error(), "does-not-exist/pub.cert")
	assert.Contains(GenerateAndWriteKeyPair(
		&Args{
			Validity:  time.Hour * 1,
			Hostnames: []string{"a.com", "b.com", "127.0.0.1"},
			KeyType:   defaultKeyType(),
		},
		publicCertPath,
		"does-not-exist/private.key",
	).Error(), "does-not-exist/private.key")
}

func TestCreateCertificate(t *testing.T) {
	if testing.Short() {
		t.Skip("certificate generation takes a long time")
	}

	assert := assert.New(t)

	tmpDir, err := ioutil.TempDir("", "certtest")

	defer func() {
		assert.Nil(os.RemoveAll(tmpDir))
	}()

	assert.Nil(err)

	publicCertPath := filepath.Join(tmpDir, "public.cert")
	privateKeyPath := filepath.Join(tmpDir, "private.key")

	err = GenerateAndWriteKeyPair(&Args{
		Validity:  time.Hour * 1,
		Hostnames: []string{"a.com", "b.com", "127.0.0.1"},
		KeyType:   defaultKeyType(),
	},
		publicCertPath, privateKeyPath)
	if err != nil {
		t.Fatal(err)
	}

	assert.FileExists(publicCertPath)
	assert.FileExists(privateKeyPath)

	publicCertFileData, err := ioutil.ReadFile(publicCertPath)
	assert.Nil(err)

	privateKeyFileData, err := ioutil.ReadFile(privateKeyPath)
	assert.Nil(err)

	// Verify that we can load the public/private key pair.
	pub, pk, err := ReadKeyPair(publicCertFileData, privateKeyFileData)
	assert.Nil(err)
	assert.NotNil(pub)
	assert.NotNil(pk)
	pkRSA, ok := pk.(*rsa.PrivateKey)
	assert.True(ok)

	// Verify that the public/private key pair can RSA encrypt/decrypt.
	pubKey, ok := pub.PublicKey.(*rsa.PublicKey)
	assert.True(ok, "pub.PublicKey is not of type, *rsa.PublicKey, %v", pub.PublicKey)

	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, pubKey, []byte(secretMessage), []byte{})
	assert.Nil(err)
	assert.NotEqual(string(ciphertext), secretMessage)

	cleartext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, pkRSA, ciphertext, []byte{})
	assert.Nil(err)
	assert.Equal(string(cleartext), string(secretMessage))
}

func TestBadValues(t *testing.T) {
	testCases := []struct {
		errorString string
		pub         string
		priv        string
		args        *Args
	}{
		{"root public certificate data was set but root private key data was not", "pub.cert", "priv.key",
			&Args{ParentKeyPair: &KeyPair{PublicCertificate: []byte("A")}, Validity: time.Second, Hostnames: []string{"127.0.0.1"}, KeyType: defaultKeyType()}},
		{"root private key data was set but root public certificate data was not", "pub.cert", "priv.key",
			&Args{ParentKeyPair: &KeyPair{PrivateKey: []byte("A")}, Validity: time.Second, Hostnames: []string{"127.0.0.1"}, KeyType: defaultKeyType()}},
		{"public certificate file path must not be empty", "", "priv.key", &Args{Validity: time.Second, Hostnames: []string{"127.0.0.1"}, KeyType: defaultKeyType()}},
		{"private key file path must not be empty", "pub.cert", "", &Args{Validity: time.Second, Hostnames: []string{"127.0.0.1"}, KeyType: defaultKeyType()}},
		//{"hostname list was empty. At least 1 hostname is required for generating a certificate-key pair", "pub.cert", "priv.key", &Args{}},
		{"cannot generate private key: key type '{ECDSA 2047}' is not valid", "pub.cert", "priv.key", &Args{Validity: time.Second, Hostnames: []string{"127.0.0.1"}, KeyType: &KeyType{
			Algorithm: "ECDSA",
			KeyLength: 2047,
		}}},
		//{"hostname list was empty. At least 1 hostname is required for generating a certificate-key pair", "pub.cert", "priv.key", &Args{Validity: time.Second, KeyType: defaultKeyType()}},
		//{"validity duration is required, otherwise the certificate would immediately expire", "pub.cert", "priv.key", &Args{Hostnames: []string{"127.0.0.1"}, KeyType: defaultKeyType()}},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.errorString, func(t *testing.T) {
			t.Parallel()
			err := GenerateAndWriteKeyPair(tc.args, tc.pub, tc.priv)
			if err == nil {
				t.Errorf("Expected an error with text, '%s'", tc.errorString)
			} else if err.Error() != tc.errorString {
				t.Errorf("Expected an error with text, '%s', got '%s'", tc.errorString, err)
			}
		})
	}
}

func TestParseName(t *testing.T) {
	testCases := []struct {
		subject  string
		expected pkix.Name
	}{
		{"/C=GB/ST=London/L=London/O=Global Security/OU=IT Department/CN=example.com", pkix.Name{
			Country:            []string{"GB"},
			Province:           []string{"London"},
			Locality:           []string{"London"},
			Organization:       []string{"Global Security"},
			OrganizationalUnit: []string{"IT Department"},
			CommonName:         "example.com",
		}},
		{"////////////CN=example.com", pkix.Name{
			CommonName: "example.com",
		}},
		{"STREET=123 Main Street/POSTALCODE=12345", pkix.Name{
			StreetAddress: []string{"123 Main Street"},
			PostalCode:    []string{"12345"},
		}},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.subject, func(t *testing.T) {
			t.Parallel()
			actual, err := ParseName(tc.subject)
			if err != nil {
				t.Errorf("Unexepected error for input '%s', %v", tc.subject, err)
			}
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("pkix.Name are different\ngot %v\nexpected: %v", actual, tc.expected)
			}
		})
	}
}
func TestBadParseName(t *testing.T) {
	testCases := []struct {
		subject  string
		expected string
	}{
		{"/LOL=CODE", "'LOL' is not a valid RFC-2253 AttributeType"},
		{"/ST=CODE=OK", "AttributeType 'ST' has too many parts, [ST CODE OK]"},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.subject, func(t *testing.T) {
			t.Parallel()
			_, err := ParseName(tc.subject)
			if err == nil {
				t.Errorf("Expected error '%s' for input %v", tc.expected, tc.subject)
			} else if err.Error() != tc.expected {
				t.Errorf("Expected error '%s' for input %v, got '%s'", tc.expected, tc.subject, err.Error())
			}
		})
	}
}

func TestPemBlockForKey_errors(t *testing.T) {
	assert := assert.New(t)

	block, err := pemBlockForKey("ok")
	assert.Nil(block)
	assert.Contains(err.Error(), "not a valid private key")

	block, err = pemBlockForKey(&ecdsa.PrivateKey{})
	assert.Nil(block)
	assert.Contains(err.Error(), "unknown elliptic curve")
}
