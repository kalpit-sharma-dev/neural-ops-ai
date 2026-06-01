package auth

import (
	"context"

	"github.com/gin-gonic/gin"
)

type contextKey string

const principalContextKey contextKey = "gateway_principal"

// Role represents RBAC roles.
type Role string

const (
	RoleAdmin        Role = "ADMIN"
	RoleSRE          Role = "SRE"
	RoleDeveloper    Role = "DEVELOPER"
	RoleReadOnly     Role = "READONLY"
	RoleAlertManager Role = "ALERT_MANAGER"
)

// Principal is the authenticated caller.
type Principal struct {
	UserID   string   `json:"userId"`
	TenantID string   `json:"tenantId"`
	Email    string   `json:"email"`
	Role     Role     `json:"role"`
	Plan     string   `json:"plan"`
	Services []string `json:"services,omitempty"`
	AuthType string   `json:"authType"`
}

// WithPrincipal stores principal on context.
func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey, principal)
}

// AttachPrincipal stores principal on gin and request context.
func AttachPrincipal(c *gin.Context, principal Principal) {
	c.Set("principal", principal)
	c.Request = c.Request.WithContext(WithPrincipal(c.Request.Context(), principal))
}

// FromContext returns principal from context.
func FromContext(ctx context.Context) (Principal, bool) {
	value, ok := ctx.Value(principalContextKey).(Principal)
	return value, ok
}

// FromGin returns principal stored by middleware.
func ParseRole(value string) Role {
	switch Role(value) {
	case RoleAdmin, RoleSRE, RoleDeveloper, RoleReadOnly, RoleAlertManager:
		return Role(value)
	default:
		return RoleReadOnly
	}
}
