package auth_test

import (
	"context"
	"testing"

	"github.com/neuralops/platform/internal/gateway/auth"
	"github.com/stretchr/testify/require"
)

func TestParseRole(t *testing.T) {
	require.Equal(t, auth.RoleAdmin, auth.ParseRole("ADMIN"))
	require.Equal(t, auth.RoleSRE, auth.ParseRole("SRE"))
	require.Equal(t, auth.RoleDeveloper, auth.ParseRole("DEVELOPER"))
	require.Equal(t, auth.RoleReadOnly, auth.ParseRole("unknown"))
}

func TestPrincipalContextRoundTrip(t *testing.T) {
	principal := auth.Principal{
		UserID:   "user-1",
		TenantID: "tenant-1",
		Email:    "demo@neuralops.ai",
		Role:     auth.RoleAdmin,
	}
	ctx := auth.WithPrincipal(context.Background(), principal)
	got, ok := auth.FromContext(ctx)
	require.True(t, ok)
	require.Equal(t, principal.UserID, got.UserID)
	require.Equal(t, principal.Role, got.Role)
}

func TestGenerateState(t *testing.T) {
	state, err := auth.GenerateState()
	require.NoError(t, err)
	require.NotEmpty(t, state)
}
