package middleware_test

import (
	"testing"

	"github.com/neuralops/platform/internal/gateway/auth"
	gwmiddleware "github.com/neuralops/platform/internal/gateway/middleware"
	"github.com/stretchr/testify/require"
)

func TestHasPermissionMatrix(t *testing.T) {
	require.True(t, gwmiddleware.HasPermission(auth.RoleAdmin, gwmiddleware.PermissionSettingsWrite))
	require.False(t, gwmiddleware.HasPermission(auth.RoleDeveloper, gwmiddleware.PermissionIncidentsWrite))
	require.True(t, gwmiddleware.HasPermission(auth.RoleAlertManager, gwmiddleware.PermissionAlertsManage))
}

func TestPermissionForRequest(t *testing.T) {
	perm, ok := gwmiddleware.PermissionForRequest("GET", "/api/v1/search/logs")
	require.True(t, ok)
	require.Equal(t, gwmiddleware.PermissionLogsRead, perm)

	perm, ok = gwmiddleware.PermissionForRequest("POST", "/api/v1/incidents/abc/acknowledge")
	require.True(t, ok)
	require.Equal(t, gwmiddleware.PermissionIncidentsWrite, perm)
}
