package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/neuralops/platform/internal/gateway/config"
)

// TokenPair contains issued access and refresh tokens.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

// JWTIssuer signs and validates internal platform JWTs.
type JWTIssuer struct {
	issuer     string
	audience   string
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	ttl        time.Duration
	refreshTTL time.Duration
}

// NewJWTIssuer creates a JWT issuer from configuration.
func NewJWTIssuer(cfg config.AuthConfig) (*JWTIssuer, error) {
	privateKey, publicKey, err := loadRSAKeys(cfg)
	if err != nil {
		return nil, err
	}
	ttl := cfg.AccessTokenTTL
	if ttl <= 0 {
		ttl = time.Hour
	}
	refreshTTL := cfg.RefreshTokenTTL
	if refreshTTL <= 0 {
		refreshTTL = 7 * 24 * time.Hour
	}
	return &JWTIssuer{
		issuer:     cfg.JWTIssuer,
		audience:   cfg.JWTAudience,
		privateKey: privateKey,
		publicKey:  publicKey,
		ttl:        ttl,
		refreshTTL: refreshTTL,
	}, nil
}

func loadRSAKeys(cfg config.AuthConfig) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	if strings.TrimSpace(cfg.JWTPrivateKeyPEM) != "" {
		block, _ := pem.Decode([]byte(cfg.JWTPrivateKeyPEM))
		if block == nil {
			return nil, nil, fmt.Errorf("invalid jwt private key pem")
		}
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			parsed, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err2 != nil {
				return nil, nil, fmt.Errorf("parse jwt private key: %w", err)
			}
			rsaKey, ok := parsed.(*rsa.PrivateKey)
			if !ok {
				return nil, nil, fmt.Errorf("jwt private key must be rsa")
			}
			return rsaKey, &rsaKey.PublicKey, nil
		}
		return key, &key.PublicKey, nil
	}

	if strings.TrimSpace(cfg.JWTPublicKeyPEM) != "" {
		block, _ := pem.Decode([]byte(cfg.JWTPublicKeyPEM))
		if block == nil {
			return nil, nil, fmt.Errorf("invalid jwt public key pem")
		}
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, nil, err
		}
		rsaPub, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, nil, fmt.Errorf("jwt public key must be rsa")
		}
		return nil, rsaPub, nil
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return key, &key.PublicKey, nil
}

// Issue creates access and refresh tokens for a principal.
func (j *JWTIssuer) Issue(principal Principal) (TokenPair, error) {
	if j.privateKey == nil {
		return TokenPair{}, fmt.Errorf("jwt signing key not configured")
	}
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"sub":      principal.UserID,
		"userId":   principal.UserID,
		"tenantId": principal.TenantID,
		"email":    principal.Email,
		"role":     string(principal.Role),
		"plan":     principal.Plan,
		"iss":      j.issuer,
		"aud":      j.audience,
		"iat":      now.Unix(),
		"exp":      now.Add(j.ttl).Unix(),
	}
	if len(principal.Services) > 0 {
		claims["services"] = principal.Services
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	accessToken, err := token.SignedString(j.privateKey)
	if err != nil {
		return TokenPair{}, err
	}

	refreshBytes := make([]byte, 32)
	if _, err := rand.Read(refreshBytes); err != nil {
		return TokenPair{}, err
	}
	refreshToken := fmt.Sprintf("%x", refreshBytes)

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(j.ttl.Seconds()),
	}, nil
}

// Validate parses an internal access token.
func (j *JWTIssuer) Validate(tokenString string) (Principal, error) {
	if j.publicKey == nil {
		return Principal{}, fmt.Errorf("jwt public key not configured")
	}
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return j.publicKey, nil
	})
	if err != nil || !token.Valid {
		return Principal{}, fmt.Errorf("invalid jwt: %w", err)
	}
	if j.issuer != "" {
		if iss, _ := claims["iss"].(string); iss != j.issuer {
			return Principal{}, fmt.Errorf("invalid issuer")
		}
	}
	if j.audience != "" && !claimContainsAudience(claims["aud"], j.audience) {
		return Principal{}, fmt.Errorf("invalid audience")
	}

	principal := Principal{
		UserID:   stringClaim(claims, "sub", "userId", "user_id"),
		TenantID: stringClaim(claims, "tenantId", "tenant_id", "tid"),
		Email:    stringClaim(claims, "email"),
		Role:     ParseRole(stringClaim(claims, "role")),
		Plan:     stringClaim(claims, "plan"),
		Services: stringSliceClaim(claims, "services"),
		AuthType: "jwt",
	}
	if principal.TenantID == "" {
		principal.TenantID = "default"
	}
	return principal, nil
}

// RefreshTokenTTL returns configured refresh token lifetime.
func (j *JWTIssuer) RefreshTokenTTL() time.Duration {
	return j.refreshTTL
}

// PublicKeyPEM exports the verification key for other services.
func (j *JWTIssuer) PublicKeyPEM() string {
	if j.publicKey == nil {
		return ""
	}
	encoded := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(j.publicKey),
	})
	return string(encoded)
}
