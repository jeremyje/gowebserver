package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultConfiguration(t *testing.T) {
	assert := assert.New(t)

	conf := loadFromFlags()
	conf.Http.Port = 8080
	conf.Https.Port = 8443

	assert.NotNil(conf)

	assert.Equal(conf.Verbose, false)
	assert.Equal(conf.Directory, "")
	assert.Equal(conf.ServePath, "/")
	assert.Equal(conf.ConfigurationFile, "")
	assert.Equal(conf.Http.Port, 8080)
	assert.Equal(conf.Https.Port, 8443)
	assert.Equal(conf.Https.Certificate.PrivateKeyFilePath, "rsa.pem")
	assert.Equal(conf.Https.Certificate.CertificateFilePath, "cert.pem")
	assert.Equal(conf.Https.Certificate.CertificateHosts, "")
	assert.Equal(conf.Https.Certificate.CertificateValidDuration, 5475)
	assert.Equal(conf.Https.Certificate.ActAsCertificateAuthority, false)
	assert.Equal(conf.Https.Certificate.OnlyGenerateCertificate, false)
	assert.Equal(conf.Https.Certificate.ForceOverwrite, false)
	assert.Equal(conf.Metrics.Enabled, true)
	assert.Equal(conf.Metrics.Path, "/metrics")
	assert.Equal(conf.UploadDirectory, "upload")
	assert.Equal(conf.UploadServePath, "/upload")
}
