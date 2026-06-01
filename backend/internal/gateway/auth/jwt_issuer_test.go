package auth_test

import (
	"testing"
	"time"

	"github.com/neuralops/platform/internal/gateway/auth"
	"github.com/neuralops/platform/internal/gateway/config"
)

func TestJWTIssuerIssueAndValidate(t *testing.T) {
	issuer, err := auth.NewJWTIssuer(config.AuthConfig{
		JWTIssuer:       "neuralops-test",
		JWTAudience:     "neuralops-api",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("NewJWTIssuer: %v", err)
	}

	principal := auth.Principal{
		UserID:   "user-123",
		TenantID: "00000000-0000-0000-0000-000000000002",
		Email:    "demo@neuralops.ai",
		Role:     auth.RoleAdmin,
		Plan:     "enterprise",
		AuthType: "jwt",
	}

	pair, err := issuer.Issue(principal)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Fatal("expected non-empty tokens")
	}

	validated, err := issuer.Validate(pair.AccessToken)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if validated.UserID != principal.UserID {
		t.Fatalf("user id mismatch: got %q want %q", validated.UserID, principal.UserID)
	}
	if validated.TenantID != principal.TenantID {
		t.Fatalf("tenant id mismatch: got %q want %q", validated.TenantID, principal.TenantID)
	}
}

func TestGenerateExchangeCode(t *testing.T) {
	code, err := auth.GenerateExchangeCode()
	if err != nil {
		t.Fatalf("GenerateExchangeCode: %v", err)
	}
	if len(code) != 64 {
		t.Fatalf("expected 64-char hex code, got len %d", len(code))
	}

	code2, err := auth.GenerateExchangeCode()
	if err != nil {
		t.Fatalf("GenerateExchangeCode second: %v", err)
	}
	if code == code2 {
		t.Fatal("exchange codes should be unique")
	}
}

func TestGeneratePKCE(t *testing.T) {
	verifier, challenge, err := auth.GeneratePKCE()
	if err != nil {
		t.Fatalf("GeneratePKCE: %v", err)
	}
	if verifier == "" || challenge == "" {
		t.Fatal("expected non-empty pkce values")
	}
	if verifier == challenge {
		t.Fatal("verifier and challenge must differ")
	}
}
