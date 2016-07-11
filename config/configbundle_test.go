package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const EMPTY_CONFIG_YAML = `verbose: false
directory: ""
serve-path: ""
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

const POPULATED_CONFIG_YAML = `verbose: true
directory: /home/directory
serve-path: /serving
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

	assert.Equal(conf.String(), EMPTY_CONFIG_YAML)
}

func TestPopulatedConfig(t *testing.T) {
	assert := assert.New(t)

	conf := &Config{}
	conf.Verbose = true
	conf.Directory = "/home/directory"
	conf.ServePath = "/serving"
	conf.Http.Port = 1000
	conf.Https.Port = 2000
	conf.Https.Certificate.PrivateKeyFilePath = "private-key.pem"
	conf.Https.Certificate.CertificateFilePath = "public-certificate.pem"
	conf.Https.Certificate.CertificateHosts = "gowebserver.com"
	conf.Https.Certificate.CertificateValidDuration = 9000
	conf.Https.Certificate.ActAsCertificateAuthority = true
	conf.Https.Certificate.OnlyGenerateCertificate = true
	conf.Https.Certificate.ForceOverwrite = true
	conf.Metrics.Enabled = true
	conf.Metrics.Path = "/prometheus"

	assert.Equal(conf.String(), POPULATED_CONFIG_YAML)
}
