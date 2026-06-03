package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/gateway/auth"
	"github.com/neuralops/platform/internal/gateway/publicpaths"
	"github.com/neuralops/platform/internal/observability"
)

// GovernanceEnforcement applies ABAC and data residency policies after authentication.
func GovernanceEnforcement(gs *observability.GovernanceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if gs == nil || publicpaths.IsPublic(c.Request.URL.Path) {
			c.Next()
			return
		}
		principal, ok := auth.PrincipalFromGin(c)
		if !ok {
			c.Next()
			return
		}
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		method := c.Request.Method
		allowed, reason := gs.EvaluateABAC(c.Request.Context(), principal.TenantID, string(principal.Role), principal.Email, method, path)
		if !allowed {
			observability.RecordGovernanceDenied(reason)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status": "error", "errorCode": "GOV001", "message": reason,
			})
			return
		}
		if method == http.MethodPost && isIngestPath(path) {
			region := c.GetHeader("X-Data-Region")
			if region == "" {
				region = c.GetHeader("X-Region")
			}
			okRegion, reason := gs.ResidencyAllowsIngest(c.Request.Context(), principal.TenantID, region)
			if !okRegion {
				observability.RecordGovernanceDenied(reason)
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"status": "error", "errorCode": "GOV002", "message": reason,
				})
				return
			}
		}
		if method == http.MethodPost && strings.Contains(path, "/exports/") {
			allowed, reason = gs.EvaluateABAC(c.Request.Context(), principal.TenantID, string(principal.Role), principal.Email, "export", path)
			if !allowed {
				observability.RecordGovernanceDenied(reason)
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"status": "error", "errorCode": "GOV001", "message": reason,
				})
				return
			}
		}
		c.Next()
	}
}

func isIngestPath(path string) bool {
	return strings.HasSuffix(path, "/logs") ||
		strings.HasSuffix(path, "/metrics") ||
		strings.HasSuffix(path, "/events") ||
		strings.HasSuffix(path, "/traces")
}
