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
