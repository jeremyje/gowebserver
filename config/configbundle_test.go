package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const YAML_OUTPUT = `serve:
  directory: 
  serve-path: /
http:
  port: 8080
https:
  port: 8443
  privatekey: rsa.pem
  certificate:
    path: cert.pem
    hosts: 
    duration: 365
    authority: false
    onlygenerate: false
metrics:
  enabled: true
  path: /metrics
`

func TestDefaultConfiguration(t *testing.T) {
	assert := assert.New(t)

	conf := Get()
	conf.Http.Port = 8080
	conf.Https.Port = 8443

	assert.NotNil(conf)
    assert.Equal(conf.Metrics.Enabled, true)
}

func TestString(t *testing.T) {
	assert := assert.New(t)

	conf := Get()
	conf.Http.Port = 8080
	conf.Https.Port = 8443

	assert.NotNil(conf)
    assert.Equal(conf.String(), YAML_OUTPUT)
}
