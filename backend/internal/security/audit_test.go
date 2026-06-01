package security_test

import (
	"context"
	"testing"

	"github.com/neuralops/platform/internal/security"
	"github.com/stretchr/testify/require"
)

func TestAuditInsertNilRepository(t *testing.T) {
	var repo *security.AuditRepository
	require.NoError(t, repo.Insert(context.Background(), security.AuditEntry{Action: "login"}))
}

func TestAuditInsertNilPoolSkipsValidation(t *testing.T) {
	repo := security.NewAuditRepository(nil)
	require.NoError(t, repo.Insert(context.Background(), security.AuditEntry{}))
}

func TestAuditWithClickHouseNilSafe(t *testing.T) {
	var repo *security.AuditRepository
	require.Nil(t, repo.WithClickHouse(nil))
}

func TestAuditActionsConstants(t *testing.T) {
	require.NotEmpty(t, security.AuditActionLoginSuccess)
	require.NotEmpty(t, security.AuditActionLogout)
}
