//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/neuralops/platform/internal/seed"
	"github.com/stretchr/testify/require"
)

func TestSeedGeneratorVolumes(t *testing.T) {
	now := time.Now().UTC()
	logs := seed.GenerateAllLogs(now, seed.DefaultLogCount)
	require.Len(t, logs, seed.DefaultLogCount)

	txns := seed.GenerateTransactions(now)
	require.Len(t, txns, seed.DefaultTransactionCount)

	deployments := seed.GenerateDeployments(now)
	require.Len(t, deployments, seed.DefaultDeploymentCount)

	upiErrors := 0
	for _, log := range logs {
		if log.Service == "upi-service" && log.Severity == "ERROR" {
			upiErrors++
		}
	}
	require.Greater(t, upiErrors, 100, "UPI outage should generate a large ERROR volume")
}

func TestSeedPostgresRelationalData(t *testing.T) {
	ctx := context.Background()
	pool, dsn := setupAuthPostgres(t, ctx)

	require.NoError(t, seed.SeedPostgres(ctx, seed.PostgresOptions{DSN: dsn}))

	var incidentCount int
	require.NoError(t, pool.QueryRow(ctx, `
SELECT COUNT(*) FROM incidents WHERE tenant_id = $1`, seed.DemoTenantID).Scan(&incidentCount))
	require.Greater(t, incidentCount, 5)

	var deploymentCount int
	require.NoError(t, pool.QueryRow(ctx, `
SELECT COUNT(*) FROM deployments WHERE tenant_id = $1`, seed.DemoTenantID).Scan(&deploymentCount))
	require.Greater(t, deploymentCount, 10)
}
