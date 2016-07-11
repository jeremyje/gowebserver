package config

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

const NO_DEFAULTS_CONFIG_FILE = `verbose: true
directory: "/home/example"
serve-path: /serving
configurationfile: "/something.yaml"
http:
  port: 1
https:
  port: 2
  certificate:
    private-key: private-key.pem
    path: path.pem
    hosts: "hosts"
    duration: 1
    actascertificateauthority: false
    onlygeneratecertificate: false
    forceoverwrite: false
metrics:
  enabled: false
  path: /path
`

func TestNoDefaultConfig(t *testing.T) {
	assert := assert.New(t)

	fp, err := writeTempFile(POPULATED_CONFIG_YAML)
	defer os.Remove(fp.Name())
	assert.Nil(err)

	conf := &Config{}
	err = loadWithConfigFile(fp.Name(), conf)
	assert.Nil(err)

	assert.Equal(conf.Verbose, true)
	assert.Equal(conf.Directory, "/home/directory")
	assert.Equal(conf.ServePath, "/serving")
	assert.Equal(conf.Http.Port, 1000)
	assert.Equal(conf.Https.Port, 2000)
	assert.Equal(conf.Https.Certificate.PrivateKeyFilePath, "private-key.pem")
	assert.Equal(conf.Https.Certificate.CertificateFilePath, "public-certificate.pem")
	assert.Equal(conf.Https.Certificate.CertificateHosts, "gowebserver.com")
	assert.Equal(conf.Https.Certificate.CertificateValidDuration, 9000)
	assert.Equal(conf.Metrics.Enabled, true)
	assert.Equal(conf.Metrics.Path, "/prometheus")
}

func createTempFile() (*os.File, error) {
	return ioutil.TempFile(os.TempDir(), "tempfile")
}

func writeTempFile(content string) (*os.File, error) {
	fp, err := createTempFile()
	if err != nil {
		return fp, err
	}
	err = ioutil.WriteFile(fp.Name(), []byte(content), os.FileMode(0644))
	return fp, err
}
