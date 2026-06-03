package finops_test

import (
	"context"
	"testing"

	"github.com/neuralops/platform/internal/finops"
	"github.com/neuralops/platform/internal/finops/domain"
)

type stubCloud struct{}

func (stubCloud) ListCloudAssets(string) []domain.CloudAsset {
	return []domain.CloudAsset{
		{ID: "a1", Provider: "aws", Type: "ec2", MonthlyUSD: 5000, Tags: map[string]string{"team": "payments"}},
	}
}

// TestTenantIsolation verifies ingest snapshots are tenant-scoped (BANK-006).
func TestTenantIsolation(t *testing.T) {
	svc := finops.NewService(stubCloud{})
	ctx := context.Background()
	snaps, err := svc.IngestAll(ctx, "tenant-a")
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}
	if len(snaps) == 0 {
		t.Fatal("tenant-a expected ingest snapshots")
	}
	statusB := svc.IngestStatus(ctx, "tenant-b")
	if len(statusB.Snapshots) > 0 {
		t.Fatalf("tenant-b should have no ingest snapshots without explicit ingest")
	}
	statusA := svc.IngestStatus(ctx, "tenant-a")
	if len(statusA.Snapshots) == 0 {
		t.Fatal("tenant-a ingest status should list snapshots")
	}
}
