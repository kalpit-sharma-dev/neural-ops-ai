package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/platform/tenant"
)

// ExtractTenant reads X-Tenant-ID (or falls back) and stores it on the request context.
func ExtractTenant(defaultTenant string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := tenant.NormalizeID(c.GetHeader(tenant.HeaderName), defaultTenant)
		c.Set("tenant_id", tenantID)
		c.Request = c.Request.WithContext(tenant.WithContext(c.Request.Context(), tenantID))
		c.Next()
	}
}

// RequireTenant aborts when tenant ID is missing from context.
func RequireTenant(defaultTenant string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, ok := tenant.FromContext(c.Request.Context())
		if !ok {
			tenantID = tenant.NormalizeID(c.GetHeader(tenant.HeaderName), defaultTenant)
			if tenantID == "" {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"status": "error", "errorCode": "TEN002", "message": "tenant id required",
				})
				return
			}
			c.Request = c.Request.WithContext(tenant.WithContext(c.Request.Context(), tenantID))
		}
		c.Next()
	}
}

// TenantFromGin returns tenant ID from gin context or header.
func TenantFromGin(c *gin.Context, defaultTenant string) string {
	if value, ok := c.Get("tenant_id"); ok {
		if id, ok := value.(string); ok && id != "" {
			return id
		}
	}
	if id, ok := tenant.FromContext(c.Request.Context()); ok {
		return id
	}
	return tenant.NormalizeID(c.GetHeader(tenant.HeaderName), defaultTenant)
}
