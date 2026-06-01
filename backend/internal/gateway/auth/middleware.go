package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/gateway/config"
	"github.com/neuralops/platform/internal/gateway/publicpaths"
)

// Authenticator resolves API key, internal JWT, or legacy OIDC bearer credentials.
type Authenticator struct {
	cfg       config.AuthConfig
	tenantCfg config.TenantConfig
	jwt       *JWTValidator
	issuer    *JWTIssuer
	apiKeys   *APIKeyValidator
	store     *IdentityStore
	oidc      *OIDCValidator
}

// NewAuthenticator creates gateway authentication dependencies.
func NewAuthenticator(cfg config.Config, store *IdentityStore) (*Authenticator, error) {
	auth := &Authenticator{
		cfg:       cfg.Auth,
		tenantCfg: cfg.Tenant,
		apiKeys:   NewAPIKeyValidator(cfg.Auth),
		store:     store,
	}
	if cfg.Auth.OIDC.Enabled && cfg.Auth.OIDC.JWKSURL != "" {
		auth.oidc = NewOIDCValidator(cfg.Auth.OIDC)
	}
	if strings.TrimSpace(cfg.Auth.JWTPublicKeyPEM) != "" {
		validator, err := NewJWTValidator(cfg.Auth)
		if err != nil {
			return nil, err
		}
		auth.jwt = validator
	}
	if issuer, err := NewJWTIssuer(cfg.Auth); err == nil {
		auth.issuer = issuer
		if auth.jwt == nil && issuer.PublicKeyPEM() != "" {
			cfg.Auth.JWTPublicKeyPEM = issuer.PublicKeyPEM()
			validator, err := NewJWTValidator(cfg.Auth)
			if err == nil {
				auth.jwt = validator
			}
		}
	}
	return auth, nil
}

// Issuer returns the internal JWT issuer when configured.
func (a *Authenticator) Issuer() *JWTIssuer {
	return a.issuer
}

// Middleware authenticates requests and stores principal on context.
func (a *Authenticator) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if publicpaths.IsPublic(c.Request.URL.Path) {
			c.Next()
			return
		}

		if a.cfg.Disabled {
			principal := Principal{
				UserID:   "dev-user",
				TenantID: a.tenantCfg.DefaultTenant,
				Email:    "dev@neuralops.local",
				Role:     RoleAdmin,
				Plan:     "enterprise",
				AuthType: "disabled",
			}
			a.attachPrincipal(c, principal)
			c.Next()
			return
		}

		if apiKey := strings.TrimSpace(c.GetHeader("X-API-Key")); apiKey != "" {
			principal, err := a.validateAPIKey(c.Request.Context(), apiKey)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"status": "error", "errorCode": "AUTH001", "message": err.Error(),
				})
				return
			}
			a.attachPrincipal(c, principal)
			c.Next()
			return
		}

		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			token := strings.TrimSpace(authHeader[7:])
			principal, err := a.validateBearer(c.Request.Context(), token)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"status": "error", "errorCode": "AUTH002", "message": "invalid bearer token",
				})
				return
			}
			a.attachPrincipal(c, principal)
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"status": "error", "errorCode": "AUTH003", "message": "authentication required",
		})
	}
}

func (a *Authenticator) validateAPIKey(ctx context.Context, apiKey string) (Principal, error) {
	if a.store != nil {
		if principal, err := a.store.ValidateAPIKey(ctx, apiKey); err == nil {
			return principal, nil
		}
	}
	return a.apiKeys.Validate(apiKey)
}

func (a *Authenticator) validateBearer(ctx context.Context, token string) (Principal, error) {
	if a.issuer != nil {
		if principal, err := a.issuer.Validate(token); err == nil {
			return principal, nil
		}
	}
	if a.jwt != nil {
		if principal, err := a.jwt.Validate(token); err == nil {
			return principal, nil
		}
	}
	if a.oidc != nil && a.cfg.OIDC.Enabled {
		return a.oidc.Validate(ctx, token)
	}
	return Principal{}, fmt.Errorf("invalid bearer token")
}

func (a *Authenticator) attachPrincipal(c *gin.Context, principal Principal) {
	if principal.TenantID == "" {
		principal.TenantID = a.tenantCfg.DefaultTenant
	}
	c.Set("principal", principal)
	c.Request = c.Request.WithContext(WithPrincipal(c.Request.Context(), principal))
}

// PrincipalFromGin returns principal from gin context.
func PrincipalFromGin(c *gin.Context) (Principal, bool) {
	value, ok := c.Get("principal")
	if !ok {
		return Principal{}, false
	}
	principal, ok := value.(Principal)
	return principal, ok
}
