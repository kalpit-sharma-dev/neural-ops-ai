//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/neuralops/platform/internal/finops"
	"github.com/neuralops/platform/internal/finops/domain"
	"github.com/neuralops/platform/internal/observability"
)

// TestTenantIsolationFinOpsCosts ensures tenant-a ingest does not appear as tenant-b data.
func TestTenantIsolationFinOpsCosts(t *testing.T) {
	store := observability.NewStore()
	svc := finops.NewService(&observability.FinOpsCloudAdapter{Store: store})
	ctx := context.Background()

	if _, err := svc.IngestAll(ctx, "tenant-alpha"); err != nil {
		t.Fatalf("ingest alpha: %v", err)
	}

	to := time.Now().UTC()
	from := to.AddDate(0, 0, -30)
	q := domain.CostsQuery{Scope: "all", From: from, To: to}

	alpha, err := svc.GetCosts(ctx, "tenant-alpha", q)
	if err != nil {
		t.Fatalf("alpha costs: %v", err)
	}
	beta, err := svc.GetCosts(ctx, "tenant-beta", q)
	if err != nil {
		t.Fatalf("beta costs: %v", err)
	}

	if alpha.Total <= 0 {
		t.Fatal("expected tenant-alpha to have ingested costs")
	}
	if beta.Total > 0 && alpha.Total == beta.Total {
		t.Fatalf("tenant-beta should not mirror tenant-alpha spend (alpha=%v beta=%v)", alpha.Total, beta.Total)
	}
}
