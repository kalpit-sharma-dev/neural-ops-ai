package config

import "testing"

func TestValidateProduction_rejectsUnsafeAuth(t *testing.T) {
	cfg := &Config{
		Server:   ServerConfig{Port: 8080, Environment: "production"},
		Services: ServicesConfig{Ingestion: "http://ingestion:8081"},
		Auth: AuthConfig{
			Disabled:         false,
			AllowDevLogin:    true,
			OIDC:             OIDCConfig{Enabled: true},
			JWTPrivateKeyPEM: "test-private",
			JWTPublicKeyPEM:  "test-public",
		},
	}
	if err := cfg.ValidateProduction(); err == nil {
		t.Fatal("expected production validation error for dev login")
	}
}

func TestValidateProduction_acceptsSSO(t *testing.T) {
	cfg := &Config{
		Server:   ServerConfig{Environment: "staging"},
		Postgres: PostgresConfig{DSN: "postgres://user:pass@localhost:5432/neuralops"},
		Services: ServicesConfig{Ingestion: "http://ingestion:8081"},
		Auth: AuthConfig{
			AllowDevLogin:    false,
			OIDC:             OIDCConfig{Enabled: true},
			JWTPrivateKeyPEM: "test-private",
			JWTPublicKeyPEM:  "test-public",
		},
	}
	if err := cfg.ValidateProduction(); err != nil {
		t.Fatalf("expected ok: %v", err)
	}
}

func TestValidateProduction_requiresPostgres(t *testing.T) {
	cfg := &Config{
		Server:   ServerConfig{Environment: "production"},
		Services: ServicesConfig{Ingestion: "http://ingestion:8081"},
		Auth: AuthConfig{
			AllowDevLogin:    false,
			OIDC:             OIDCConfig{Enabled: true},
			JWTPrivateKeyPEM: "test-private",
			JWTPublicKeyPEM:  "test-public",
		},
	}
	if err := cfg.ValidateProduction(); err == nil {
		t.Fatal("expected postgres required in production")
	}
}
