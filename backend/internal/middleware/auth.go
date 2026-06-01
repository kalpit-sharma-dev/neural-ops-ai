package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const principalContextKey contextKey = "neuralops_principal"

// Role represents RBAC roles aligned with the platform permission matrix.
type Role string

const (
	RoleAdmin        Role = "ADMIN"
	RoleSRE          Role = "SRE"
	RoleDeveloper    Role = "DEVELOPER"
	RoleReadOnly     Role = "READONLY"
	RoleAlertManager Role = "ALERT_MANAGER"
)

// Principal is the authenticated caller injected into request context.
type Principal struct {
	UserID   string   `json:"userId"`
	TenantID string   `json:"tenantId"`
	Email    string   `json:"email"`
	Role     Role     `json:"role"`
	Plan     string   `json:"plan"`
	Services []string `json:"services,omitempty"`
	AuthType string   `json:"authType"`
}

// JWTClaims are validated for RS256 platform tokens.
type JWTClaims struct {
	UserID   string `json:"userId"`
	TenantID string `json:"tenantId"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Plan     string `json:"plan"`
	jwt.RegisteredClaims
}

// WithPrincipal stores principal on context.
func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey, principal)
}

// PrincipalFromContext returns principal from context.
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	value, ok := ctx.Value(principalContextKey).(Principal)
	return value, ok
}

// PrincipalFromGin returns principal from gin context.
func PrincipalFromGin(c *gin.Context) (Principal, bool) {
	if value, ok := c.Get("principal"); ok {
		if principal, ok := value.(Principal); ok {
			return principal, true
		}
	}
	return PrincipalFromContext(c.Request.Context())
}

// AttachPrincipal stores principal on gin and request context.
func AttachPrincipal(c *gin.Context, principal Principal) {
	c.Set("principal", principal)
	c.Request = c.Request.WithContext(WithPrincipal(c.Request.Context(), principal))
}

// ParseRole normalizes role strings.
func ParseRole(value string) Role {
	switch role := Role(strings.ToUpper(strings.TrimSpace(value))); role {
	case RoleAdmin, RoleSRE, RoleDeveloper, RoleReadOnly, RoleAlertManager:
		return role
	default:
		return RoleReadOnly
	}
}

// ClaimsToPrincipal converts validated JWT claims to a principal.
func ClaimsToPrincipal(claims JWTClaims, authType string) Principal {
	return Principal{
		UserID:   claims.UserID,
		TenantID: claims.TenantID,
		Email:    claims.Email,
		Role:     ParseRole(claims.Role),
		Plan:     claims.Plan,
		AuthType: authType,
	}
}

// RequireAuth aborts with 401 when principal is missing.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := PrincipalFromGin(c); !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status": "error", "errorCode": "AUTH003", "message": "authentication required",
			})
			return
		}
		c.Next()
	}
}
