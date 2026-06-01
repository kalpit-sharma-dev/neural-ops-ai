package auth

import (
	"net/http"

	"github.com/neuralops/platform/internal/gateway/config"
)

// SAMLProvider handles SP-initiated SAML flows (crewjam/saml or dev fallback).
type SAMLProvider interface {
	WriteMetadata(w http.ResponseWriter)
	LoginURL(relayState string) (string, error)
	ParseACS(r *http.Request) (SAMLIdentity, error)
}

// NewSAMLProvider creates a production crewjam SP when metadata URL is configured, else dev parser.
func NewSAMLProvider(cfg config.SAMLConfig) (SAMLProvider, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	if cfg.MetadataURL != "" {
		return newCrewjamSAMLProvider(cfg)
	}
	return NewSAMLFlow(cfg)
}
