package config

import (
	"fmt"
	"strings"
)

// ValidateProduction rejects unsafe authentication settings in regulated environments.
func (c *Config) ValidateProduction() error {
	env := strings.ToLower(strings.TrimSpace(c.Server.Environment))
	if env != "production" && env != "staging" {
		return nil
	}
	if c.Auth.Disabled {
		return fmt.Errorf("%s: auth.disabled must be false", env)
	}
	if c.Auth.AllowDevLogin {
		return fmt.Errorf("%s: auth.allow_dev_login must be false", env)
	}
	if c.DemoMode {
		return fmt.Errorf("%s: demo_mode must be false", env)
	}
	if !c.Auth.OIDC.Enabled && !c.Auth.SAML.Enabled {
		return fmt.Errorf("%s: enable auth.oidc or auth.saml (SSO required)", env)
	}
	if c.Auth.JWTPrivateKeyPEM == "" && c.Auth.JWTPublicKeyPEM == "" {
		return fmt.Errorf("%s: JWT key pair must be configured (not dev defaults)", env)
	}
	return nil
}
