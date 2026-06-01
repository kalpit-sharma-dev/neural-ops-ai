package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/gateway/auth"
	"github.com/neuralops/platform/internal/gateway/publicpaths"
	"github.com/neuralops/platform/internal/security"
	"go.uber.org/zap"
)

// AuditLog records API access to the audit repository.
func AuditLog(repo *security.AuditRepository, log *zap.Logger) gin.HandlerFunc {
	if repo == nil {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		if publicpaths.IsPublic(c.Request.URL.Path) {
			c.Next()
			return
		}

		c.Next()

		result := "success"
		if c.Writer.Status() >= 400 {
			result = "failure"
		}

		userID := ""
		tenantID := "default"
		if principal, ok := auth.PrincipalFromGin(c); ok {
			userID = principal.UserID
			if principal.TenantID != "" {
				tenantID = principal.TenantID
			}
		}

		resourceType, resourceID := parseResource(c.Request.URL.Path)
		entry := security.AuditEntry{
			UserID:       userID,
			TenantID:     tenantID,
			Action:       c.Request.Method + " " + c.FullPath(),
			ResourceType: resourceType,
			ResourceID:   resourceID,
			IP:           c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			Result:       result,
		}

		if err := repo.Insert(c.Request.Context(), entry); err != nil && log != nil {
			log.Warn("audit log insert failed", zap.Error(err), zap.String("path", c.Request.URL.Path))
		}
	}
}

func parseResource(path string) (resourceType, resourceID string) {
	trimmed := strings.Trim(path, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) < 2 {
		return "", ""
	}

	// /api/v1/<resource>/...
	idx := 0
	if parts[0] == "api" {
		idx = 2
		if len(parts) <= idx {
			return "", ""
		}
	}

	resourceType = parts[idx]
	if len(parts) > idx+1 && parts[idx+1] != "" && !isActionSegment(parts[idx+1]) {
		resourceID = parts[idx+1]
	}
	return resourceType, resourceID
}

func isActionSegment(segment string) bool {
	switch strings.ToLower(segment) {
	case "search", "query", "overview", "health", "metrics", "events", "traces", "webhooks":
		return true
	default:
		return false
	}
}
