package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neuralops/platform/internal/security"
)

// UserRecord represents a platform user.
type UserRecord struct {
	ID       string
	TenantID string
	Email    string
	Role     Role
	SSOSub   string
}

// TenantRecord represents a tenant subscription.
type TenantRecord struct {
	ID                 string
	Name               string
	PlanTier           string
	SubscriptionStatus string
}

// IdentityStore persists users, tenants, API keys, and auth sessions.
type IdentityStore struct {
	pool *pgxpool.Pool
}

// NewIdentityStore creates an identity store.
func NewIdentityStore(pool *pgxpool.Pool) *IdentityStore {
	return &IdentityStore{pool: pool}
}

// ValidateAPIKey looks up a hashed API key and returns the associated principal.
func (s *IdentityStore) ValidateAPIKey(ctx context.Context, rawKey string) (Principal, error) {
	if s == nil || s.pool == nil {
		return Principal{}, fmt.Errorf("identity store unavailable")
	}
	hash := security.HashAPIKey(strings.TrimSpace(rawKey))
	row := s.pool.QueryRow(ctx, `
SELECT k.tenant_id::text, k.role, k.name, t.plan_tier, t.subscription_status
FROM api_keys k
JOIN tenants t ON t.id = k.tenant_id
WHERE k.key_hash = $1`, hash)

	var tenantID, role, keyName, plan, subscriptionStatus string
	if err := row.Scan(&tenantID, &role, &keyName, &plan, &subscriptionStatus); err != nil {
		if err == pgx.ErrNoRows {
			return Principal{}, fmt.Errorf("invalid api key")
		}
		return Principal{}, err
	}
	if err := ValidateSubscriptionStatus(subscriptionStatus); err != nil {
		return Principal{}, err
	}
	return Principal{
		UserID:   "api-key-" + hash[:8],
		TenantID: tenantID,
		Email:    keyName + "@api-key.local",
		Role:     ParseRole(role),
		Plan:     plan,
		AuthType: "api_key",
	}, nil
}

// GetTenant returns tenant metadata and validates subscription status.
func (s *IdentityStore) GetTenant(ctx context.Context, tenantID string) (*TenantRecord, error) {
	if s == nil || s.pool == nil {
		return nil, fmt.Errorf("identity store unavailable")
	}
	row := s.pool.QueryRow(ctx, `
SELECT id::text, name, plan_tier, subscription_status
FROM tenants WHERE id = $1`, tenantID)
	var tenant TenantRecord
	if err := row.Scan(&tenant.ID, &tenant.Name, &tenant.PlanTier, &tenant.SubscriptionStatus); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, err
	}
	return &tenant, nil
}

// UpsertOIDCUser creates or updates a user from OIDC claims.
func (s *IdentityStore) UpsertOIDCUser(ctx context.Context, tenantID, email, role, ssoSub string) (*UserRecord, error) {
	if s == nil || s.pool == nil {
		return nil, fmt.Errorf("identity store unavailable")
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant id required")
	}
	tenant, err := s.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if err := ValidateSubscriptionStatus(tenant.SubscriptionStatus); err != nil {
		return nil, err
	}
	if email == "" {
		email = ssoSub
	}
	if role == "" {
		role = string(RoleReadOnly)
	}

	var user UserRecord
	err = s.pool.QueryRow(ctx, `
INSERT INTO users (tenant_id, email, role, sso_sub)
VALUES ($1, $2, $3, $4)
ON CONFLICT (tenant_id, email) DO UPDATE SET
  role = EXCLUDED.role,
  sso_sub = COALESCE(EXCLUDED.sso_sub, users.sso_sub)
RETURNING id::text, tenant_id::text, email, role, COALESCE(sso_sub, '')`,
		tenantID, email, role, ssoSub,
	).Scan(&user.ID, &user.TenantID, &user.Email, &user.Role, &user.SSOSub)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID returns a user by primary key.
func (s *IdentityStore) GetUserByID(ctx context.Context, userID string) (*UserRecord, error) {
	row := s.pool.QueryRow(ctx, `
SELECT u.id::text, u.tenant_id::text, u.email, u.role, COALESCE(u.sso_sub, ''), t.plan_tier, t.name
FROM users u
JOIN tenants t ON t.id = u.tenant_id
WHERE u.id = $1`, userID)
	var user UserRecord
	var plan, tenantName string
	if err := row.Scan(&user.ID, &user.TenantID, &user.Email, &user.Role, &user.SSOSub, &plan, &tenantName); err != nil {
		return nil, err
	}
	return &user, nil
}

// SaveOIDCState stores PKCE verifier for callback validation.
func (s *IdentityStore) SaveOIDCState(ctx context.Context, state, codeVerifier, redirectURI string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
INSERT INTO oidc_auth_states (state, code_verifier, redirect_uri, expires_at)
VALUES ($1, $2, $3, $4)`, state, codeVerifier, redirectURI, expiresAt)
	return err
}

// ConsumeOIDCState validates and deletes a PKCE state record.
func (s *IdentityStore) ConsumeOIDCState(ctx context.Context, state string) (codeVerifier, redirectURI string, err error) {
	row := s.pool.QueryRow(ctx, `
DELETE FROM oidc_auth_states
WHERE state = $1 AND expires_at > NOW()
RETURNING code_verifier, redirect_uri`, state)
	if err := row.Scan(&codeVerifier, &redirectURI); err != nil {
		return "", "", fmt.Errorf("invalid or expired oidc state")
	}
	return codeVerifier, redirectURI, nil
}

// SaveRefreshToken stores a hashed refresh token.
func (s *IdentityStore) SaveRefreshToken(ctx context.Context, userID, rawToken string, expiresAt time.Time) error {
	hash := hashToken(rawToken)
	_, err := s.pool.Exec(ctx, `
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4)`, uuid.New(), userID, hash, expiresAt)
	return err
}

// ConsumeRefreshToken validates a refresh token and returns the user ID.
func (s *IdentityStore) ConsumeRefreshToken(ctx context.Context, rawToken string) (string, error) {
	hash := hashToken(rawToken)
	var userID string
	err := s.pool.QueryRow(ctx, `
DELETE FROM refresh_tokens
WHERE token_hash = $1 AND expires_at > NOW()
RETURNING user_id::text`, hash).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token")
	}
	return userID, nil
}

// DevLogin finds a user by email within a tenant for local development.
func (s *IdentityStore) DevLogin(ctx context.Context, email, tenantID string) (*UserRecord, error) {
	row := s.pool.QueryRow(ctx, `
SELECT u.id::text, u.tenant_id::text, u.email, u.role, COALESCE(u.sso_sub, '')
FROM users u
WHERE u.email = $1 AND u.tenant_id = $2`, email, tenantID)
	var user UserRecord
	if err := row.Scan(&user.ID, &user.TenantID, &user.Email, &user.Role, &user.SSOSub); err != nil {
		return nil, fmt.Errorf("user not found")
	}
	return &user, nil
}

// RevokeRefreshToken deletes a specific refresh token hash.
func (s *IdentityStore) RevokeRefreshToken(ctx context.Context, rawToken string) error {
	if s == nil || s.pool == nil || strings.TrimSpace(rawToken) == "" {
		return nil
	}
	hash := hashToken(rawToken)
	_, err := s.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE token_hash = $1`, hash)
	return err
}

// RevokeUserRefreshTokens deletes all refresh tokens for a user.
func (s *IdentityStore) RevokeUserRefreshTokens(ctx context.Context, userID string) error {
	if s == nil || s.pool == nil || userID == "" {
		return nil
	}
	_, err := s.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
	return err
}

// SaveExchangeCode stores a one-time SPA exchange code for OIDC callback.
func (s *IdentityStore) SaveExchangeCode(ctx context.Context, rawCode, userID string, expiresAt time.Time) error {
	if s == nil || s.pool == nil {
		return fmt.Errorf("identity store unavailable")
	}
	hash := hashToken(rawCode)
	_, err := s.pool.Exec(ctx, `
INSERT INTO auth_exchange_codes (code_hash, user_id, expires_at)
VALUES ($1, $2, $3)`, hash, userID, expiresAt)
	return err
}

// ConsumeExchangeCode validates and deletes a one-time exchange code.
func (s *IdentityStore) ConsumeExchangeCode(ctx context.Context, rawCode string) (string, error) {
	if s == nil || s.pool == nil {
		return "", fmt.Errorf("identity store unavailable")
	}
	hash := hashToken(strings.TrimSpace(rawCode))
	var userID string
	err := s.pool.QueryRow(ctx, `
DELETE FROM auth_exchange_codes
WHERE code_hash = $1 AND expires_at > NOW()
RETURNING user_id::text`, hash).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("invalid or expired exchange code")
	}
	return userID, nil
}

// ValidateSubscriptionStatus ensures a tenant subscription is active.
func ValidateSubscriptionStatus(status string) error {
	normalized := strings.ToLower(strings.TrimSpace(status))
	if normalized == "" || normalized == "active" {
		return nil
	}
	return fmt.Errorf("tenant subscription inactive")
}

// GenerateExchangeCode creates a cryptographically secure one-time code.
func GenerateExchangeCode() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// GeneratePKCE creates PKCE verifier/challenge pair.
func GeneratePKCE() (verifier, challenge string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	verifier = base64.RawURLEncoding.EncodeToString(buf)
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(sum[:])
	return verifier, challenge, nil
}

// GenerateState creates a random OIDC state value.
func GenerateState() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
