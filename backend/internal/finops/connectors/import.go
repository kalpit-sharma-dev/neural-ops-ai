package connectors

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/finops/domain"
)

// ImportCustomFeed normalizes CSV/JSON cost rows into FOCUS line items (REQ-FINOPS-004).
func ImportCustomFeed(_ context.Context, tenantID string, req domain.ImportRequest) (domain.ImportBatch, []domain.CostLineItem, error) {
	if req.Source == "" {
		return domain.ImportBatch{}, nil, fmt.Errorf("source required")
	}
	format := req.Format
	if format == "" {
		format = "json"
	}
	dedupeKey := req.DedupeKey
	if dedupeKey == "" {
		h := sha256.Sum256([]byte(tenantID + req.Source + fmt.Sprint(len(req.Items))))
		dedupeKey = hex.EncodeToString(h[:8])
	}
	now := time.Now().UTC()
	items := make([]domain.CostLineItem, 0, len(req.Items))
	var total float64
	for _, row := range req.Items {
		provider := row.Provider
		if provider == "" {
			provider = "custom"
		}
		item := domain.CostLineItem{
			ID: uuid.NewString(), TenantID: tenantID, BillingPeriod: now,
			Provider: provider, Service: row.Service, Region: row.Region,
			ResourceID: row.ResourceID, UsageType: row.UsageType,
			EffectiveCost: row.EffectiveCost, AmortizedCost: row.EffectiveCost,
			ListCost: row.EffectiveCost * 1.05, CostView: "effective",
			Tags: row.Tags, IngestedAt: now,
		}
		if item.Tags == nil {
			item.Tags = map[string]string{"source": req.Source}
		} else {
			item.Tags["source"] = req.Source
		}
		if t, ok := item.Tags["team"]; ok {
			item.Team = t
		}
		total += row.EffectiveCost
		items = append(items, item)
	}
	batch := domain.ImportBatch{
		ID: uuid.NewString(), Source: req.Source, Format: format,
		LineCount: int64(len(items)), TotalUSD: total, Version: 1,
		DedupeKey: dedupeKey, ImportedAt: now,
	}
	return batch, items, nil
}

// VersionSnapshot increments version for idempotent restatement (REQ-FINOPS-005).
func VersionSnapshot(existing []domain.IngestSnapshot, snap domain.IngestSnapshot) domain.IngestSnapshot {
	maxVer := 0
	for _, e := range existing {
		if e.Provider == snap.Provider && e.BillingPeriod.Equal(snap.BillingPeriod) {
			if e.Version > maxVer {
				maxVer = e.Version
			}
		}
	}
	snap.Version = maxVer + 1
	return snap
}

// DedupeLineItems replaces items with same natural key on re-ingest.
func DedupeLineItems(existing, incoming []domain.CostLineItem) []domain.CostLineItem {
	key := func(i domain.CostLineItem) string {
		return i.Provider + "|" + i.ResourceID + "|" + i.BillingPeriod.Format("2006-01-02") + "|" + i.UsageType
	}
	seen := map[string]domain.CostLineItem{}
	for _, i := range existing {
		seen[key(i)] = i
	}
	for _, i := range incoming {
		seen[key(i)] = i
	}
	out := make([]domain.CostLineItem, 0, len(seen))
	for _, v := range seen {
		out = append(out, v)
	}
	return out
}
