package config

import (
	"fmt"
)

// Bundle set of configuration values for the server.
type ServerConfiguration struct {
    // HTTP configuration
    HttpPort int
    ServePath string
    RootDirectory string
    ForceOverwrite bool
    
    // HTTPS configuration
    HttpsPort int
    CertificateFilePath string
    PrivateKeyFilePath string
    CertificateHosts string
    CertificateValidDuration int
    ActAsCertificateAuthority bool
    OnlyGenerateCertificate bool
    
    // Prometheus Metrics
    EnableMetrics bool
    MetricsHttpPath string
    
    ConfigurationFile string
}

func (this *ServerConfiguration) String() string {
    return "" +
        fmt.Sprintf("http:\n") +
        fmt.Sprintf("  port: %d\n", this.HttpPort) +
        fmt.Sprintf("  serve-path: %s\n", this.ServePath) +
        fmt.Sprintf("  root-directory: %s\n", this.RootDirectory) +
        fmt.Sprintf("https:\n") +
        fmt.Sprintf("  port: %d\n", this.HttpsPort) +
        fmt.Sprintf("")
}

func Get() *ServerConfiguration {
    return &ServerConfiguration{
        HttpPort: *httpPortFlag,
    }
}
