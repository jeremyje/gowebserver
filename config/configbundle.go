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
	"gopkg.in/yaml.v2"
)

// HTTP holds the configuration for HTTP serving.
type HTTP struct {
	Port int `yaml:"port"`
}

// HTTPS holds the configuration for HTTPS serving.
type HTTPS struct {
	Port        int         `yaml:"port"`
	Certificate Certificate `yaml:"certificate"`
}

// Certificate holds the certificate/private key configuration for HTTPS.
type Certificate struct {
	PrivateKeyFilePath        string `yaml:"private-key"`
	CertificateFilePath       string `yaml:"path"`
	CertificateHosts          string `yaml:"hosts"`
	CertificateValidDuration  int    `yaml:"duration"`
	ActAsCertificateAuthority bool   `yaml:"-"`
	OnlyGenerateCertificate   bool   `yaml:"-"`
	ForceOverwrite            bool   `yaml:"-"`
}

// Metrics holds the metrics configuration.
type Metrics struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

// Config is the root of the server configuration.
type Config struct {
	Verbose           bool   `yaml:"verbose"`
	Path              string `yaml:"path"`
	ServePath         string `yaml:"serve-path"`
	ConfigurationFile string `yaml:"-"`

	UploadPath      string `yaml:"upload-path"`
	UploadServePath string `yaml:"upload-serve-path"`

	HTTP    HTTP    `yaml:"http"`
	HTTPS   HTTPS   `yaml:"https"`
	Metrics Metrics `yaml:"metrics"`
}

// String returns a string representation of the config.
func (c *Config) String() string {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err.Error()
	}

	return string(data)
}
