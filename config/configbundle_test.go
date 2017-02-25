package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const emptyConfigYaml = `verbose: false
directory: ""
serve-path: ""
upload-directory: ""
upload-serve-path: ""
http:
  port: 0
https:
  port: 0
  certificate:
    private-key: ""
    path: ""
    hosts: ""
    duration: 0
metrics:
  enabled: false
  path: ""
`

const populatedConfigYaml = `verbose: true
directory: /home/directory
serve-path: /serving
upload-directory: /home/upload
upload-serve-path: /postage
http:
  port: 1000
https:
  port: 2000
  certificate:
    private-key: private-key.pem
    path: public-certificate.pem
    hosts: gowebserver.com
    duration: 9000
metrics:
  enabled: true
  path: /prometheus
`

func TestEmptyConfig(t *testing.T) {
	assert := assert.New(t)

	conf := &Config{}
	assert.NotNil(conf)

	assert.Equal(conf.String(), emptyConfigYaml)
}

func TestPopulatedConfig(t *testing.T) {
	assert := assert.New(t)

	conf := &Config{}
	conf.Verbose = true
	conf.Directory = "/home/directory"
	conf.ServePath = "/serving"
	conf.UploadDirectory = "/home/upload"
	conf.UploadServePath = "/postage"
	conf.HTTP.Port = 1000
	conf.HTTPS.Port = 2000
	conf.HTTPS.Certificate.PrivateKeyFilePath = "private-key.pem"
	conf.HTTPS.Certificate.CertificateFilePath = "public-certificate.pem"
	conf.HTTPS.Certificate.CertificateHosts = "gowebserver.com"
	conf.HTTPS.Certificate.CertificateValidDuration = 9000
	conf.HTTPS.Certificate.ActAsCertificateAuthority = true
	conf.HTTPS.Certificate.OnlyGenerateCertificate = true
	conf.HTTPS.Certificate.ForceOverwrite = true
	conf.Metrics.Enabled = true
	conf.Metrics.Path = "/prometheus"

	assert.Equal(conf.String(), populatedConfigYaml)
}
