package config

import (
	"gopkg.in/yaml.v2"
)

type Http struct {
	Port int `yaml:"port"`
}

type Https struct {
	Port        int         `yaml:"port"`
	Certificate Certificate `yaml:"certificate"`
}

type Certificate struct {
	PrivateKeyFilePath        string `yaml:"private-key"`
	CertificateFilePath       string `yaml:"path"`
	CertificateHosts          string `yaml:"hosts"`
	CertificateValidDuration  int    `yaml:"duration"`
	ActAsCertificateAuthority bool   `yaml:"-"`
	OnlyGenerateCertificate   bool   `yaml:"-"`
	ForceOverwrite            bool   `yaml:"-"`
}

type Metrics struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

type Config struct {
	Verbose           bool   `yaml:"verbose"`
	Directory         string `yaml:"directory"`
	ServePath         string `yaml:"serve-path"`
	ConfigurationFile string `yaml:"-"`

	Http    Http    `yaml:"http"`
	Https   Https   `yaml:"https"`
	Metrics Metrics `yaml:"metrics"`
}

func (this *Config) String() string {
	data, err := yaml.Marshal(this)
	if err != nil {
		return err.Error()
	}

	return string(data)
}
