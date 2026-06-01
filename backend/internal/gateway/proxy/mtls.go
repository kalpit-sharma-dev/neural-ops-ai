package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// MTLSConfig configures optional client TLS for upstream service calls.
type MTLSConfig struct {
	Enabled    bool
	CertFile   string
	KeyFile    string
	CAFile     string
	ServerName string
}

// BuildTransport returns an http.RoundTripper with optional mTLS client credentials.
func BuildTransport(cfg MTLSConfig) (http.RoundTripper, error) {
	if !cfg.Enabled {
		return http.DefaultTransport, nil
	}
	if cfg.CertFile == "" || cfg.KeyFile == "" {
		return nil, fmt.Errorf("mTLS enabled but cert/key paths missing")
	}
	if _, err := os.Stat(cfg.CertFile); err != nil {
		return http.DefaultTransport, nil
	}
	if _, err := os.Stat(cfg.KeyFile); err != nil {
		return http.DefaultTransport, nil
	}

	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("load mTLS client cert: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}
	if cfg.ServerName != "" {
		tlsConfig.ServerName = cfg.ServerName
	}
	if cfg.CAFile != "" {
		caPEM, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read mTLS CA: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("parse mTLS CA pem")
		}
		tlsConfig.RootCAs = pool
	}

	return &http.Transport{
		TLSClientConfig: tlsConfig,
		Proxy:           http.ProxyFromEnvironment,
	}, nil
}

// MTLSConfigFromEnv reads mTLS settings from standard env vars.
func MTLSConfigFromEnv() MTLSConfig {
	enabled := strings.EqualFold(os.Getenv("INTERNAL_MTLS_ENABLED"), "true")
	return MTLSConfig{
		Enabled:    enabled,
		CertFile:   os.Getenv("INTERNAL_MTLS_CERT_FILE"),
		KeyFile:    os.Getenv("INTERNAL_MTLS_KEY_FILE"),
		CAFile:     os.Getenv("INTERNAL_MTLS_CA_FILE"),
		ServerName: os.Getenv("INTERNAL_MTLS_SERVER_NAME"),
	}
}
