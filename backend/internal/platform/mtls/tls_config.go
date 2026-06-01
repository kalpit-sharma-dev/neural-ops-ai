package mtls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// ServerConfig builds TLS settings for services that require verified client certificates.
type ServerConfig struct {
	CertFile          string
	KeyFile           string
	CAFile            string
	RequireClientCert bool
}

// BuildServerTLS returns a tls.Config for mTLS-terminated HTTP servers.
func BuildServerTLS(cfg ServerConfig) (*tls.Config, error) {
	if cfg.CertFile == "" || cfg.KeyFile == "" {
		return nil, fmt.Errorf("server cert and key required")
	}
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("load server key pair: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		ClientAuth:   tls.NoClientCert,
	}
	if cfg.RequireClientCert {
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}
	if cfg.CAFile != "" {
		caPEM, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read client CA: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("parse client CA pem")
		}
		tlsConfig.ClientCAs = pool
	}
	return tlsConfig, nil
}

// VerifyPeerRequiresClientCert reports whether client certificate verification is enabled.
func VerifyPeerRequiresClientCert(cfg ServerConfig) bool {
	return cfg.RequireClientCert && cfg.CAFile != ""
}
