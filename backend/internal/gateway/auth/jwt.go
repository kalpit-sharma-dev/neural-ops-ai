package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/neuralops/platform/internal/gateway/config"
)

// JWTValidator validates RS256 JWT access tokens.
type JWTValidator struct {
	issuer   string
	audience string
	key      *rsa.PublicKey
}

// NewJWTValidator creates a JWT validator from PEM public key.
func NewJWTValidator(cfg config.AuthConfig) (*JWTValidator, error) {
	if strings.TrimSpace(cfg.JWTPublicKeyPEM) == "" {
		return nil, fmt.Errorf("jwt public key not configured")
	}
	block, _ := pem.Decode([]byte(cfg.JWTPublicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("invalid jwt public key pem")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse jwt public key: %w", err)
	}
	rsaKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("jwt public key must be rsa")
	}
	return &JWTValidator{issuer: cfg.JWTIssuer, audience: cfg.JWTAudience, key: rsaKey}, nil
}

// Validate parses and validates a bearer token.
func (v *JWTValidator) Validate(tokenString string) (Principal, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Header["alg"])
		}
		return v.key, nil
	})
	if err != nil || !token.Valid {
		return Principal{}, fmt.Errorf("invalid jwt: %w", err)
	}

	if v.issuer != "" {
		if iss, _ := claims["iss"].(string); iss != v.issuer {
			return Principal{}, fmt.Errorf("invalid issuer")
		}
	}
	if v.audience != "" {
		if !claimContainsAudience(claims["aud"], v.audience) {
			return Principal{}, fmt.Errorf("invalid audience")
		}
	}

	principal := Principal{
		UserID:   stringClaim(claims, "sub", "user_id", "userId"),
		TenantID: stringClaim(claims, "tenant_id", "tenantId", "tid"),
		Email:    stringClaim(claims, "email"),
		Role:     ParseRole(stringClaim(claims, "role")),
		Plan:     stringClaim(claims, "plan"),
		AuthType: "jwt",
	}
	if principal.UserID == "" {
		return Principal{}, fmt.Errorf("jwt missing subject")
	}
	if principal.TenantID == "" {
		principal.TenantID = "default"
	}
	principal.Services = stringSliceClaim(claims, "services")
	if exp, ok := claims["exp"].(float64); ok && time.Now().Unix() > int64(exp) {
		return Principal{}, fmt.Errorf("jwt expired")
	}
	return principal, nil
}

func stringClaim(claims jwt.MapClaims, keys ...string) string {
	for _, key := range keys {
		if value, ok := claims[key].(string); ok && value != "" {
			return value
		}
	}
	return ""
}

func stringSliceClaim(claims jwt.MapClaims, key string) []string {
	raw, ok := claims[key]
	if !ok {
		return nil
	}
	switch values := raw.(type) {
	case []string:
		return values
	case []any:
		out := make([]string, 0, len(values))
		for _, item := range values {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func claimContainsAudience(raw any, expected string) bool {
	switch aud := raw.(type) {
	case string:
		return aud == expected
	case []any:
		for _, item := range aud {
			if value, ok := item.(string); ok && value == expected {
				return true
			}
		}
	}
	return false
}
