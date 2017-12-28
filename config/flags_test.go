package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultConfiguration(t *testing.T) {
	assert := assert.New(t)

	conf := loadFromFlags()
	conf.HTTP.Port = 8080
	conf.HTTPS.Port = 8443

	assert.NotNil(conf)

	assert.Equal(conf.Verbose, false)
	assert.Equal(conf.Path, "")
	assert.Equal(conf.ServePath, "/")
	assert.Equal(conf.ConfigurationFile, "")
	assert.Equal(conf.HTTP.Port, 8080)
	assert.Equal(conf.HTTPS.Port, 8443)
	assert.Equal(conf.HTTPS.Certificate.PrivateKeyFilePath, "rsa.pem")
	assert.Equal(conf.HTTPS.Certificate.CertificateFilePath, "cert.pem")
	assert.Equal(conf.HTTPS.Certificate.CertificateHosts, "")
	assert.Equal(conf.HTTPS.Certificate.CertificateValidDuration, 5475)
	assert.Equal(conf.HTTPS.Certificate.ActAsCertificateAuthority, false)
	assert.Equal(conf.HTTPS.Certificate.OnlyGenerateCertificate, false)
	assert.Equal(conf.HTTPS.Certificate.ForceOverwrite, false)
	assert.Equal(conf.Metrics.Enabled, true)
	assert.Equal(conf.Metrics.Path, "/metrics")
	assert.Equal(conf.UploadPath, "uploaded-files")
	assert.Equal(conf.UploadServePath, "/upload.asp")
}
