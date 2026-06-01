package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/gateway/auth"
	"github.com/neuralops/platform/internal/gateway/config"
)

// Tenant validates tenant isolation, subscription status, and injects downstream headers.
func Tenant(cfg config.TenantConfig, store *auth.IdentityStore) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(cfg.AllowedTenants))
	for _, tenantID := range cfg.AllowedTenants {
		allowed[tenantID] = struct{}{}
	}

	return func(c *gin.Context) {
		principal, ok := auth.PrincipalFromGin(c)
		if !ok {
			c.Next()
			return
		}

		tenantID := principal.TenantID
		if tenantID == "" {
			tenantID = cfg.DefaultTenant
		}
		if len(allowed) > 0 {
			if _, ok := allowed[tenantID]; !ok {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"status": "error", "errorCode": "TEN001", "message": "tenant not allowed",
				})
				return
			}
		}

		if store != nil {
			tenant, err := store.GetTenant(c.Request.Context(), tenantID)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"status": "error", "errorCode": "TEN003", "message": "unknown tenant",
				})
				return
			}
			if status := strings.ToLower(tenant.SubscriptionStatus); status != "" && status != "active" {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"status": "error", "errorCode": "TEN004", "message": "tenant subscription inactive",
				})
				return
			}
			if principal.Plan == "" {
				principal.Plan = tenant.PlanTier
				auth.AttachPrincipal(c, principal)
			}
		}

		c.Request.Header.Set("X-Tenant-ID", tenantID)
		c.Request.Header.Set("X-User-ID", principal.UserID)
		c.Request.Header.Set("X-User-Role", string(principal.Role))
		if len(principal.Services) > 0 {
			c.Request.Header.Set("X-Allowed-Services", joinCSV(principal.Services))
		}
		c.Next()
	}
}

func joinCSV(values []string) string {
	if len(values) == 0 {
		return ""
	}
	out := values[0]
	for i := 1; i < len(values); i++ {
		out += "," + values[i]
	}
	return out
}
