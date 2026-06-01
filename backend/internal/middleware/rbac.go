package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Permission names used by RBAC middleware.
const (
	PermLogsRead       = "logs.read"
	PermIncidentsRead  = "incidents.read"
	PermIncidentsWrite = "incidents.write"
	PermAlertsManage   = "alerts.manage"
	PermSettingsWrite  = "settings.write"
)

var rolePermissions = map[Role]map[string]bool{
	RoleAdmin: {
		PermLogsRead: true, PermIncidentsRead: true, PermIncidentsWrite: true,
		PermAlertsManage: true, PermSettingsWrite: true,
	},
	RoleSRE: {
		PermLogsRead: true, PermIncidentsRead: true, PermIncidentsWrite: true, PermAlertsManage: true,
	},
	RoleDeveloper: {
		PermLogsRead: true, PermIncidentsRead: true,
	},
	RoleReadOnly: {
		PermLogsRead: true, PermIncidentsRead: true,
	},
	RoleAlertManager: {
		PermAlertsManage: true,
	},
}

// HasPermission reports whether a role grants a permission.
func HasPermission(role Role, permission string) bool {
	perms, ok := rolePermissions[role]
	if !ok {
		return false
	}
	return perms[permission]
}

// CheckPermission returns middleware that enforces a single permission.
func CheckPermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, ok := PrincipalFromGin(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status": "error", "errorCode": "AUTH004", "message": "missing principal",
			})
			return
		}
		if !HasPermission(principal.Role, permission) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status": "error", "errorCode": "AUTH005", "message": "insufficient permissions",
			})
			return
		}
		c.Next()
	}
}
