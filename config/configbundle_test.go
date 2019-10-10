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
	"testing"

	"github.com/stretchr/testify/assert"
)

const emptyConfigYaml = `verbose: false
path: ""
serve-path: ""
upload-path: ""
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
path: /home/folder
serve-path: /serving
upload-path: /home/upload
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
	conf.Path = "/home/folder"
	conf.ServePath = "/serving"
	conf.UploadPath = "/home/upload"
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
