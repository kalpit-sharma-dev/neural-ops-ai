package connectors

import (
	"context"
	"testing"
	"time"

	"github.com/neuralops/platform/internal/finops/domain"
)

func TestImportCustomFeed(t *testing.T) {
	batch, items, err := ImportCustomFeed(context.Background(), "t1", domain.ImportRequest{
		Source: "datadog", Items: []domain.ImportLineItem{
			{ResourceID: "dd-1", Service: "apm", EffectiveCost: 120, Tags: map[string]string{"team": "platform"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if batch.LineCount != 1 || len(items) != 1 {
		t.Fatalf("expected 1 item, got batch=%d items=%d", batch.LineCount, len(items))
	}
}

func TestVersionSnapshot(t *testing.T) {
	period := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	snap := VersionSnapshot([]domain.IngestSnapshot{
		{Provider: "aws", BillingPeriod: period, Version: 2},
	}, domain.IngestSnapshot{Provider: "aws", BillingPeriod: period})
	if snap.Version != 3 {
		t.Fatalf("expected version 3, got %d", snap.Version)
	}
}

func TestDedupeLineItems(t *testing.T) {
	day := time.Now().UTC()
	a := domain.CostLineItem{Provider: "aws", ResourceID: "r1", BillingPeriod: day, UsageType: "compute", EffectiveCost: 10}
	b := domain.CostLineItem{Provider: "aws", ResourceID: "r1", BillingPeriod: day, UsageType: "compute", EffectiveCost: 15}
	out := DedupeLineItems([]domain.CostLineItem{a}, []domain.CostLineItem{b})
	if len(out) != 1 || out[0].EffectiveCost != 15 {
		t.Fatalf("dedupe failed: %+v", out)
	}
}
