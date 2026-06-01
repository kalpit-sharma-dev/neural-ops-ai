package auth

import (
	"fmt"
	"strings"

	"github.com/neuralops/platform/internal/gateway/config"
)

// APIKeyValidator validates ingestion API keys.
type APIKeyValidator struct {
	keys map[string]Principal
}

// NewAPIKeyValidator creates an API key validator.
func NewAPIKeyValidator(cfg config.AuthConfig) *APIKeyValidator {
	keys := make(map[string]Principal, len(cfg.APIKeys))
	for key, entry := range cfg.APIKeys {
		keys[key] = Principal{
			UserID:   entry.UserID,
			TenantID: entry.TenantID,
			Role:     ParseRole(entry.Role),
			AuthType: "api_key",
		}
	}
	return &APIKeyValidator{keys: keys}
}

// Validate returns principal for a valid API key.
func (v *APIKeyValidator) Validate(apiKey string) (Principal, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return Principal{}, fmt.Errorf("missing api key")
	}
	principal, ok := v.keys[apiKey]
	if !ok {
		return Principal{}, fmt.Errorf("invalid api key")
	}
	return principal, nil
}
