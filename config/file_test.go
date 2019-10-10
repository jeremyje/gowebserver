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

package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const noDefaultsConfigFile = `verbose: true
path: "/home/example"
serve-path: /serving
configurationfile: "/something.yaml"
http:
  port: 1
https:
  port: 2
  certificate:
    private-key: private.pem
    path: public.pem
    hosts: "hosts"
    duration: 1
    actascertificateauthority: false
    onlygeneratecertificate: false
    forceoverwrite: false
metrics:
  enabled: false
  path: /metrics
`

func TestNoDefaultConfig(t *testing.T) {
	assert := assert.New(t)

	fp, err := writeTempFile(noDefaultsConfigFile)
	defer os.Remove(fp.Name())
	assert.Nil(err)

	conf := &Config{}
	err = loadWithConfigFile(fp.Name(), conf)
	assert.Nil(err)

	assert.Equal(conf.Verbose, true)
	assert.Equal(conf.Path, "/home/example")
	assert.Equal(conf.ServePath, "/serving")
	assert.Equal(conf.HTTP.Port, 1)
	assert.Equal(conf.HTTPS.Port, 2)
	assert.Equal(conf.HTTPS.Certificate.PrivateKeyFilePath, "private.pem")
	assert.Equal(conf.HTTPS.Certificate.CertificateFilePath, "public.pem")
	assert.Equal(conf.HTTPS.Certificate.CertificateHosts, "hosts")
	assert.Equal(conf.HTTPS.Certificate.CertificateValidDuration, 1)
	assert.Equal(conf.Metrics.Enabled, false)
	assert.Equal(conf.Metrics.Path, "/metrics")
}

func TestPopulatedYamlConfig(t *testing.T) {
	assert := assert.New(t)

	fp, err := writeTempFile(populatedConfigYaml)
	defer os.Remove(fp.Name())
	assert.Nil(err)

	conf := &Config{}
	err = loadWithConfigFile(fp.Name(), conf)
	assert.Nil(err)

	assert.Equal(conf.Verbose, true)
	assert.Equal(conf.Path, "/home/folder")
	assert.Equal(conf.ServePath, "/serving")
	assert.Equal(conf.HTTP.Port, 1000)
	assert.Equal(conf.HTTPS.Port, 2000)
	assert.Equal(conf.HTTPS.Certificate.PrivateKeyFilePath, "private-key.pem")
	assert.Equal(conf.HTTPS.Certificate.CertificateFilePath, "public-certificate.pem")
	assert.Equal(conf.HTTPS.Certificate.CertificateHosts, "gowebserver.com")
	assert.Equal(conf.HTTPS.Certificate.CertificateValidDuration, 9000)
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
