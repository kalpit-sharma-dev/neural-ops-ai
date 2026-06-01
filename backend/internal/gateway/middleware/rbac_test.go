package middleware_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/neuralops/platform/internal/gateway/auth"
	"github.com/stretchr/testify/require"
)

func TestRBACAllowedMatrix(t *testing.T) {
	require.True(t, rbacAllowed(auth.RoleAdmin, http.MethodPost, "/api/v1/incidents/abc/acknowledge"))
	require.True(t, rbacAllowed(auth.RoleSRE, http.MethodPost, "/api/v1/incidents/abc/acknowledge"))
	require.False(t, rbacAllowed(auth.RoleReadOnly, http.MethodPost, "/api/v1/incidents/abc/acknowledge"))
	require.True(t, rbacAllowed(auth.RoleReadOnly, http.MethodGet, "/api/v1/search/logs"))
	require.True(t, rbacAllowed(auth.RoleAlertManager, http.MethodGet, "/api/v1/alerts"))
	require.False(t, rbacAllowed(auth.RoleDeveloper, http.MethodPost, "/api/v1/settings"))
}

func TestDeveloperScopeHeader(t *testing.T) {
	require.True(t, serviceAllowed([]string{"upi-service"}, "upi-service"))
	require.False(t, serviceAllowed([]string{"upi-service"}, "ledger-service"))
}

func serviceAllowed(allowed []string, service string) bool {
	service = strings.ToLower(strings.TrimSpace(service))
	for _, candidate := range allowed {
		if strings.ToLower(strings.TrimSpace(candidate)) == service {
			return true
		}
	}
	return false
}

// rbacAllowed mirrors middleware.allowed for unit testing without gin.
func rbacAllowed(role auth.Role, method, path string) bool {
	switch role {
	case auth.RoleAdmin:
		return true
	case auth.RoleSRE:
		if method == http.MethodGet {
			return true
		}
		return contains(path, "/incidents/") && (endsWith(path, "/acknowledge") || endsWith(path, "/resolve"))
	case auth.RoleDeveloper:
		if method != http.MethodGet && !contains(path, "/logs") {
			return false
		}
		return contains(path, "/logs") || contains(path, "/search") || contains(path, "/incidents")
	case auth.RoleReadOnly:
		return method == http.MethodGet || method == http.MethodOptions
	case auth.RoleAlertManager:
		if contains(path, "/alerts") {
			return true
		}
		return method == http.MethodGet
	default:
		return method == http.MethodGet
	}
}

func contains(s, part string) bool { return len(s) >= len(part) && (s == part || len(part) == 0 || indexContains(s, part)) }
func indexContains(s, part string) bool {
	for i := 0; i+len(part) <= len(s); i++ {
		if s[i:i+len(part)] == part {
			return true
		}
	}
	return false
}
func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
