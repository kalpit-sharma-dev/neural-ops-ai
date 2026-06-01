package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/gateway/auth"
	"github.com/neuralops/platform/internal/gateway/publicpaths"
)

// RBAC enforces role-based access control.
func RBAC() gin.HandlerFunc {
	return func(c *gin.Context) {
		if publicpaths.IsPublic(c.Request.URL.Path) {
			c.Next()
			return
		}

		principal, ok := auth.PrincipalFromGin(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status": "error", "errorCode": "AUTH004", "message": "missing principal",
			})
			return
		}

		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		if !allowed(principal.Role, method, path) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status": "error", "errorCode": "AUTH005", "message": "insufficient permissions",
			})
			return
		}

		if perm, ok := PermissionForRequest(method, path); ok && !HasPermission(principal.Role, perm) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status": "error", "errorCode": "AUTH005", "message": "insufficient permissions",
			})
			return
		}
		c.Next()
	}
}

func allowed(role auth.Role, method, path string) bool {
	switch role {
	case auth.RoleAdmin:
		return true
	case auth.RoleSRE:
		if method == http.MethodGet {
			return true
		}
		return strings.Contains(path, "/incidents/") &&
			(strings.HasSuffix(path, "/acknowledge") || strings.HasSuffix(path, "/resolve"))
	case auth.RoleDeveloper:
		if method != http.MethodGet && !isIngestionWrite(path) {
			return false
		}
		return isDeveloperPath(path)
	case auth.RoleReadOnly:
		return method == http.MethodGet || method == http.MethodOptions
	case auth.RoleAlertManager:
		if strings.Contains(path, "/alerts") || strings.Contains(path, "/notifications") {
			return true
		}
		return method == http.MethodGet
	default:
		return method == http.MethodGet
	}
}

func isIngestionWrite(path string) bool {
	return strings.Contains(path, "/logs") ||
		strings.Contains(path, "/metrics") ||
		strings.Contains(path, "/events") ||
		strings.Contains(path, "/traces")
}

func isDeveloperPath(path string) bool {
	return strings.Contains(path, "/logs") ||
		strings.Contains(path, "/search") ||
		strings.Contains(path, "/incidents") ||
		strings.Contains(path, "/dashboard") ||
		strings.Contains(path, "/chat")
}
