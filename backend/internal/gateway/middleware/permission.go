package middleware

import (
	"strings"

	"github.com/neuralops/platform/internal/gateway/auth"
)

// Permission represents a fine-grained RBAC action from the spec matrix.
type Permission string

const (
	PermissionLogsRead       Permission = "logs.read"
	PermissionIncidentsRead  Permission = "incidents.read"
	PermissionIncidentsWrite Permission = "incidents.write"
	PermissionAlertsManage   Permission = "alerts.manage"
	PermissionSettingsWrite  Permission = "settings.write"
)

// RolePermissions maps roles to allowed permissions.
var RolePermissions = map[auth.Role]map[Permission]bool{
	auth.RoleAdmin: {
		PermissionLogsRead: true, PermissionIncidentsRead: true, PermissionIncidentsWrite: true,
		PermissionAlertsManage: true, PermissionSettingsWrite: true,
	},
	auth.RoleSRE: {
		PermissionLogsRead: true, PermissionIncidentsRead: true, PermissionIncidentsWrite: true,
		PermissionAlertsManage: true,
	},
	auth.RoleDeveloper: {
		PermissionLogsRead: true, PermissionIncidentsRead: true,
	},
	auth.RoleReadOnly: {
		PermissionLogsRead: true, PermissionIncidentsRead: true,
	},
	auth.RoleAlertManager: {
		PermissionAlertsManage: true,
	},
}

// HasPermission reports whether a role may perform an action.
func HasPermission(role auth.Role, permission Permission) bool {
	perms, ok := RolePermissions[role]
	if !ok {
		return false
	}
	return perms[permission]
}

// PermissionForRequest maps HTTP method + path to a permission check.
func PermissionForRequest(method, path string) (Permission, bool) {
	switch {
	case containsAny(path, "/logs", "/search", "/traces"):
		return PermissionLogsRead, true
	case containsAny(path, "/incidents"):
		if method != "GET" && method != "HEAD" && method != "OPTIONS" {
			return PermissionIncidentsWrite, true
		}
		return PermissionIncidentsRead, true
	case containsAny(path, "/alerts", "/notifications"):
		return PermissionAlertsManage, true
	case containsAny(path, "/admin", "/dashboards", "/slos", "/synthetic", "/workflows"):
		return PermissionSettingsWrite, true
	case containsAny(path, "/apm", "/metrics/catalog", "/metrics/query", "/metrics/promql", "/topology", "/infra", "/databases", "/middleware", "/rum", "/security", "/integrations", "/zones", "/logs/metric-rules", "/logs/parsing-rules", "/anomalies/entities", "/notebooks"):
		return PermissionLogsRead, true
	case containsAny(path, "/settings"):
		return PermissionSettingsWrite, true
	default:
		return "", false
	}
}

func containsAny(path string, parts ...string) bool {
	for _, part := range parts {
		if strings.Contains(path, part) {
			return true
		}
	}
	return false
}
