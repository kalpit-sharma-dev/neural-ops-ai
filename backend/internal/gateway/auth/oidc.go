package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/neuralops/platform/internal/gateway/config"
)

type jwkSet struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// OIDCValidator validates OIDC bearer tokens using JWKS.
type OIDCValidator struct {
	issuer   string
	audience string
	jwksURL  string
	client   *http.Client
	keys     map[string]*rsa.PublicKey
	mu       sync.RWMutex
}

// NewOIDCValidator creates an OIDC validator.
func NewOIDCValidator(cfg config.OIDCConfig) *OIDCValidator {
	return &OIDCValidator{
		issuer:   cfg.Issuer,
		audience: cfg.Audience,
		jwksURL:  cfg.JWKSURL,
		client:   &http.Client{Timeout: 10 * time.Second},
		keys:     make(map[string]*rsa.PublicKey),
	}
}

// Validate validates an OIDC access token.
func (v *OIDCValidator) Validate(ctx context.Context, tokenString string) (Principal, error) {
	if v.jwksURL == "" {
		return Principal{}, fmt.Errorf("oidc jwks url not configured")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		kid, _ := token.Header["kid"].(string)
		return v.lookupKey(ctx, kid)
	}, jwt.WithValidMethods([]string{"RS256", "RS384", "RS512"}))
	if err != nil || !token.Valid {
		return Principal{}, fmt.Errorf("invalid oidc token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return Principal{}, fmt.Errorf("invalid oidc claims")
	}
	if v.issuer != "" {
		if iss, _ := claims["iss"].(string); iss != v.issuer {
			return Principal{}, fmt.Errorf("invalid oidc issuer")
		}
	}
	if v.audience != "" && !claimContainsAudience(claims["aud"], v.audience) {
		return Principal{}, fmt.Errorf("invalid oidc audience")
	}

	principal := Principal{
		UserID:   stringClaim(claims, "sub", "email"),
		TenantID: stringClaim(claims, "tenant_id", "tenantId", "tid"),
		Email:    stringClaim(claims, "email"),
		Role:     ParseRole(stringClaim(claims, "role")),
		Plan:     stringClaim(claims, "plan"),
		AuthType: "oidc",
	}
	if principal.TenantID == "" {
		principal.TenantID = "default"
	}
	return principal, nil
}

func (v *OIDCValidator) lookupKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.mu.RLock()
	if key, ok := v.keys[kid]; ok {
		v.mu.RUnlock()
		return key, nil
	}
	if len(v.keys) == 1 {
		for _, key := range v.keys {
			v.mu.RUnlock()
			return key, nil
		}
	}
	v.mu.RUnlock()

	if err := v.refreshJWKS(ctx); err != nil {
		return nil, err
	}

	v.mu.RLock()
	defer v.mu.RUnlock()
	if key, ok := v.keys[kid]; ok {
		return key, nil
	}
	for _, key := range v.keys {
		return key, nil
	}
	return nil, fmt.Errorf("jwks key not found")
}

func (v *OIDCValidator) refreshJWKS(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	if err != nil {
		return err
	}
	res, err := v.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return fmt.Errorf("fetch jwks: status=%d", res.StatusCode)
	}

	var set jwkSet
	if err := json.NewDecoder(res.Body).Decode(&set); err != nil {
		return err
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	for _, key := range set.Keys {
		if strings.ToUpper(key.Kty) != "RSA" {
			continue
		}
		rsaKey, err := rsaFromJWK(key.N, key.E)
		if err != nil {
			continue
		}
		kid := key.Kid
		if kid == "" {
			kid = "default"
		}
		v.keys[kid] = rsaKey
	}
	if len(v.keys) == 0 {
		return fmt.Errorf("no rsa keys in jwks")
	}
	return nil
}

func rsaFromJWK(n, e string) (*rsa.PublicKey, error) {
	nb, err := base64.RawURLEncoding.DecodeString(n)
	if err != nil {
		return nil, err
	}
	eb, err := base64.RawURLEncoding.DecodeString(e)
	if err != nil {
		return nil, err
	}
	exponent := 0
	for _, b := range eb {
		exponent = exponent<<8 + int(b)
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nb),
		E: exponent,
	}, nil
}
