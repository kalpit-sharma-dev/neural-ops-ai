package finops_test

import (
	"context"
	"testing"

	"github.com/neuralops/platform/internal/finops"
	"github.com/neuralops/platform/internal/finops/domain"
)

type stubCloudAssets struct{}

func (stubCloudAssets) ListCloudAssets(provider string) []domain.CloudAsset {
	return []domain.CloudAsset{
		{ID: "i-1", Provider: "aws", Type: "ec2", Name: "web", Region: "us-east-1",
			AccountID: "111122223333", MonthlyUSD: 3000, Tags: map[string]string{"team": "platform"}},
		{ID: "i-2", Provider: "gcp", Type: "gce", Name: "api", Region: "us-central1",
			AccountID: "proj-1", MonthlyUSD: 1500, Tags: map[string]string{"team": "payments"}},
	}
}

func TestIngestReconciliationDrift(t *testing.T) {
	svc := finops.NewService(stubCloudAssets{})
	snaps, err := svc.IngestAll(context.Background(), "reconcile-tenant")
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}
	if len(snaps) == 0 {
		t.Fatal("expected ingest snapshots")
	}
	for _, s := range snaps {
		if s.DriftPct > 1.0 {
			t.Fatalf("provider %s drift %.2f%% exceeds 1%% tolerance", s.Provider, s.DriftPct)
		}
	}
	status := svc.IngestStatus(context.Background(), "reconcile-tenant")
	if status.Stale {
		t.Fatalf("expected fresh ingest, lag=%v", status.LagSeconds)
	}
}
