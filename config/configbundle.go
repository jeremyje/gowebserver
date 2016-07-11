package config

import (
	//"fmt"
)

type Http struct {
    Port int `yaml:"port"`
}

type Https struct {
    Port int `yaml:"port"`
    Certificate Certificate `yaml:"certificate"`
}

type Certificate struct {
    PrivateKeyFilePath string `yaml:"private-key"`
    CertificateFilePath string `yaml:"path"`
    CertificateHosts string `yaml:"hosts"`
    CertificateValidDuration int `yaml:"duration"`
    ActAsCertificateAuthority bool
    OnlyGenerateCertificate bool
    ForceOverwrite bool
}

type Metrics struct {
    Enabled bool `yaml:"enabled"`
    Path string `yaml:"path"`
}

type Config struct {
    Verbose bool `yaml:"verbose"`
    ServeDirectory string `yaml:"directory"`
    ServePath string `yaml:"serve-path"`
    ConfigurationFile string
    
    Http Http `yaml:"http"`
    Https Https `yaml:"https"`
    Metrics Metrics `yaml:"metrics"`
}
/*
// Bundle set of configuration values for the server.
type ServerConfiguration struct {
    
    // Serving configuration
    ServeDirectory string
    ServePath string
    ConfigurationFile string
    Verbose bool
    
    
    // HTTP configuration
    HttpPort int
    ForceOverwrite bool
    
    // HTTPS configuration
    HttpsPort int
    PrivateKeyFilePath string
    CertificateFilePath string
    CertificateHosts string
    CertificateValidDuration int
    ActAsCertificateAuthority bool
    OnlyGenerateCertificate bool
    
    // Prometheus Metrics
    EnableMetrics bool
    MetricsHttpPath string
}

func (this *ServerConfiguration) String() string {
    return "" +
        fmt.Sprintf("verbose: %t\n", this.Verbose) +
        fmt.Sprintf("serve:\n") +
        fmt.Sprintf("  directory: %s\n", this.ServeDirectory) +
        fmt.Sprintf("  serve-path: %s\n", this.ServePath) +
        fmt.Sprintf("http:\n") +
        fmt.Sprintf("  port: %d\n", this.HttpPort) +
        fmt.Sprintf("https:\n") +
        fmt.Sprintf("  port: %d\n", this.HttpsPort) +
        fmt.Sprintf("  privatekey: %s\n", this.PrivateKeyFilePath) +
        fmt.Sprintf("  certificate:\n") +
        fmt.Sprintf("    path: %s\n", this.CertificateFilePath) +
        fmt.Sprintf("    hosts: %s\n", this.CertificateHosts) +
        fmt.Sprintf("    duration: %d\n", this.CertificateValidDuration) +
        fmt.Sprintf("    authority: %t\n", this.ActAsCertificateAuthority) +
        fmt.Sprintf("    onlygenerate: %t\n", this.OnlyGenerateCertificate) +
        fmt.Sprintf("metrics:\n") +
        fmt.Sprintf("  enabled: %t\n", this.EnableMetrics) +
        fmt.Sprintf("  path: %s\n", this.MetricsHttpPath) +
        ""
}

func Get() *ServerConfiguration {
    return &ServerConfiguration{
        Verbose: *verboseFlag,
        ServeDirectory: *serveDirectoryFlag,
        ServePath: *servePathFlag,
        ConfigurationFile: *configFileFlag,
        HttpPort: *httpPortFlag,
        HttpsPort: *httpsPortFlag,
        PrivateKeyFilePath: *privateKeyFilePathFlag,
        CertificateFilePath: *certificateFilePathFlag,
        CertificateHosts: *certHostsFlag,
        CertificateValidDuration: *validDurationFlag,
        ActAsCertificateAuthority: *actAsCertificateAuthorityFlag,
        OnlyGenerateCertificate: *onlyGenerateCertFlag,
        EnableMetrics: *metricsFlag,
        MetricsHttpPath: *metricsPathFlag,
    }
}
*/
func Get() *Config {
    return &Config {
        Verbose: *verboseFlag,
        ServeDirectory: *serveDirectoryFlag,
        ServePath: *servePathFlag,
        ConfigurationFile: *configFileFlag,
        Http: Http {
          Port:   *httpPortFlag,
        },
        Https: Https {
            Port:   *httpsPortFlag,
            Certificate: Certificate {
                PrivateKeyFilePath: *privateKeyFilePathFlag,
                CertificateFilePath: *certificateFilePathFlag,
                CertificateHosts: *certHostsFlag,
                CertificateValidDuration: *validDurationFlag,
                ActAsCertificateAuthority: *actAsCertificateAuthorityFlag,
                OnlyGenerateCertificate: *onlyGenerateCertFlag,
                ForceOverwrite: *onlyGenerateCertFlag, // No flag available yet.
            },
        },
        Metrics: Metrics {
            Enabled: *metricsFlag,
            Path: *metricsPathFlag,
        },
    }
}
