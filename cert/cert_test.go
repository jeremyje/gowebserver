package cert

import (
	"io/ioutil"
	"github.com/stretchr/testify/assert"
	"testing"
	"os"
)

func TestWriteDefaultCertificate(t *testing.T) {
	assert := assert.New(t)

	certFp, err := createTempFile()
	assert.Nil(err)
	certPath := certFp.Name()
	keyFp, err := createTempFile()
	assert.Nil(err)
	keyPath := keyFp.Name()

	certBuilder := NewCertificateBuilder()
	err = certBuilder.WriteCertificate(certPath)
	assert.Nil(err)
	err = certBuilder.WritePrivateKey(keyPath)
	assert.Nil(err)
	cert, err := ReadCertificateFromFile(certPath)
	assert.Nil(err)
	assert.Equal("US", cert.Subject.Country[0])
	assert.Equal("Seattle", cert.Subject.Locality[0])
	assert.Equal("Washington", cert.Subject.Province[0])
}

func createTempFile() (*os.File, error) {
    return ioutil.TempFile(os.TempDir(), "tempfile")
}
