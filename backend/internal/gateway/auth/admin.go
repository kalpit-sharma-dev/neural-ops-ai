package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/security"
)

// AdminUser is a user row for admin UI.
type AdminUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Active   bool   `json:"active"`
	TenantID string `json:"tenantId"`
}

// AdminAPIKey is an API key row for admin UI.
type AdminAPIKey struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Prefix    string    `json:"prefix"`
	Scopes    []string  `json:"scopes"`
	CreatedAt time.Time `json:"createdAt"`
}

// ListUsersByTenant lists users for admin UI.
func ListUsersByTenant(ctx context.Context, store *IdentityStore, tenantID string) ([]AdminUser, error) {
	if store == nil || store.pool == nil {
		return nil, nil
	}
	rows, err := store.pool.Query(ctx, `
SELECT id::text, email, role, tenant_id::text, active FROM users WHERE tenant_id = $1 ORDER BY email`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AdminUser, 0)
	for rows.Next() {
		var u AdminUser
		if err := rows.Scan(&u.ID, &u.Email, &u.Role, &u.TenantID, &u.Active); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// CreateUser provisions (or re-invites) a tenant user. Re-inviting an existing
// email updates the role and re-activates the account, keeping the operation
// idempotent for admin UIs.
func CreateUser(ctx context.Context, store *IdentityStore, tenantID, email, role string) (AdminUser, error) {
	if store == nil || store.pool == nil {
		return AdminUser{}, fmt.Errorf("identity store unavailable")
	}
	if tenantID == "" || email == "" {
		return AdminUser{}, fmt.Errorf("tenant id and email are required")
	}
	if role == "" {
		role = string(RoleReadOnly)
	}
	var u AdminUser
	err := store.pool.QueryRow(ctx, `
INSERT INTO users (tenant_id, email, role)
VALUES ($1, $2, $3)
ON CONFLICT (tenant_id, email) DO UPDATE SET role = EXCLUDED.role, active = TRUE
RETURNING id::text, email, role, tenant_id::text, active`,
		tenantID, email, role,
	).Scan(&u.ID, &u.Email, &u.Role, &u.TenantID, &u.Active)
	if err != nil {
		return AdminUser{}, err
	}
	return u, nil
}

// UpdateUserRole changes a user's RBAC role within a tenant.
func UpdateUserRole(ctx context.Context, store *IdentityStore, tenantID, userID, role string) error {
	if store == nil || store.pool == nil {
		return fmt.Errorf("identity store unavailable")
	}
	tag, err := store.pool.Exec(ctx, `
UPDATE users SET role = $1 WHERE id = $2 AND tenant_id = $3`, role, userID, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// SetUserActive activates or deactivates a user without deleting the record.
// Deactivating also revokes outstanding refresh tokens so existing sessions end.
func SetUserActive(ctx context.Context, store *IdentityStore, tenantID, userID string, active bool) error {
	if store == nil || store.pool == nil {
		return fmt.Errorf("identity store unavailable")
	}
	tag, err := store.pool.Exec(ctx, `
UPDATE users SET active = $1 WHERE id = $2 AND tenant_id = $3`, active, userID, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	if !active {
		_ = store.RevokeUserRefreshTokens(ctx, userID)
	}
	return nil
}

// ListAPIKeysByTenant lists API keys without secrets.
func ListAPIKeysByTenant(ctx context.Context, store *IdentityStore, tenantID string) ([]AdminAPIKey, error) {
	if store == nil || store.pool == nil {
		return nil, nil
	}
	rows, err := store.pool.Query(ctx, `
SELECT id::text, name, role, created_at FROM api_keys WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AdminAPIKey, 0)
	for rows.Next() {
		var k AdminAPIKey
		var role string
		if err := rows.Scan(&k.ID, &k.Name, &role, &k.CreatedAt); err != nil {
			return nil, err
		}
		k.Prefix = "no_"
		k.Scopes = []string{role}
		out = append(out, k)
	}
	return out, rows.Err()
}

// CreateAPIKey creates an API key and returns the raw secret once.
func CreateAPIKey(ctx context.Context, store *IdentityStore, tenantID, name, role string) (AdminAPIKey, string, error) {
	if store == nil || store.pool == nil {
		return AdminAPIKey{}, "", fmt.Errorf("identity store unavailable")
	}
	raw, err := generateAPIKeySecret()
	if err != nil {
		return AdminAPIKey{}, "", err
	}
	hash := security.HashAPIKey(raw)
	id := uuid.New()
	_, err = store.pool.Exec(ctx, `
INSERT INTO api_keys (id, tenant_id, key_hash, name, role) VALUES ($1, $2, $3, $4, $5)`,
		id, tenantID, hash, name, role,
	)
	if err != nil {
		return AdminAPIKey{}, "", err
	}
	return AdminAPIKey{
		ID: id.String(), Name: name, Prefix: raw[:8], Scopes: []string{role}, CreatedAt: time.Now().UTC(),
	}, raw, nil
}

func generateAPIKeySecret() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "no_" + hex.EncodeToString(buf), nil
}
