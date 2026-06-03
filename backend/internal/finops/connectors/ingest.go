package connectors

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/finops/domain"
)

const driftTolerancePct = 1.0

// Ingestor normalizes cloud billing from live exports (CUR/BQ/Azure files) or inventory simulation.
type Ingestor struct {
	cloud  domain.CloudAssets
	config BillingConfig
}

// NewIngestor creates a billing ingestor.
func NewIngestor(cloud domain.CloudAssets) *Ingestor {
	return &Ingestor{cloud: cloud, config: LoadBillingConfig()}
}

// IngestAll ingests AWS, GCP, and Azure billing for the current period.
func (i *Ingestor) IngestAll(ctx context.Context, tenantID string) ([]domain.IngestSnapshot, []domain.CostLineItem, error) {
	_ = ctx
	now := time.Now().UTC()
	period := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	var snapshots []domain.IngestSnapshot
	var items []domain.CostLineItem

	for _, provider := range []string{"aws", "gcp", "azure"} {
		if liveItems, liveSnap, ok := i.ingestLiveFile(tenantID, provider); ok {
			items = append(items, liveItems...)
			snapshots = append(snapshots, liveSnap)
			continue
		}
		if !i.config.UseSimulation(provider) {
			continue
		}
		if i.cloud == nil {
			continue
		}
		assets := i.cloud.ListCloudAssets(provider)
		if len(assets) == 0 {
			continue
		}
		var total float64
		for _, a := range assets {
			daily := a.MonthlyUSD / 30.0
			for d := 0; d < 14; d++ {
				day := now.AddDate(0, 0, -d)
				amt := daily * (0.95 + float64(d%3)*0.02)
				item := domain.CostLineItem{
					ID: uuid.NewString(), TenantID: tenantID,
					BillingPeriod: day, Provider: a.Provider,
					AccountID: a.AccountID, Service: a.Type, Region: a.Region,
					ResourceID: a.ID, UsageType: "ComputeHours", Quantity: 24,
					Unit: "Hrs", AmortizedCost: amt, ListCost: amt * 1.08,
					EffectiveCost: amt, CostView: "amortized",
					Tags: a.Tags, IngestedAt: now,
				}
				if item.Tags == nil {
					item.Tags = map[string]string{}
				}
				if t, ok := item.Tags["team"]; ok {
					item.Team = t
				}
				if e, ok := item.Tags["env"]; ok {
					item.Environment = e
				}
				items = append(items, item)
				total += amt
			}
		}
		invoiceTotal := total * 1.005
		if provider == "aws" && i.config.AWSInvoiceUSD > 0 {
			invoiceTotal = i.config.AWSInvoiceUSD
		}
		drift := absPct((total-invoiceTotal)/invoiceTotal) * 100
		snapshots = append(snapshots, domain.IngestSnapshot{
			ID: uuid.NewString(), Provider: provider, BillingPeriod: period,
			LineCount: int64(len(assets) * 14), TotalAmortized: total,
			InvoiceTotal: invoiceTotal, DriftPct: drift, CostView: "amortized",
			IngestedAt: now, DriftAlert: drift > driftTolerancePct,
		})
	}
	if len(snapshots) == 0 && len(items) == 0 {
		return nil, nil, fmt.Errorf("no billing data ingested: configure FINOPS_*_FILE or cloud inventory")
	}
	return snapshots, items, nil
}

func (i *Ingestor) ingestLiveFile(tenantID, provider string) ([]domain.CostLineItem, domain.IngestSnapshot, bool) {
	var path string
	var invoice float64
	switch provider {
	case "aws":
		path = i.config.AWS_CURFile
		invoice = i.config.AWSInvoiceUSD
	case "gcp":
		path = i.config.GCPBillingFile
	case "azure":
		path = i.config.AzureBillingFile
	default:
		return nil, domain.IngestSnapshot{}, false
	}
	if path == "" {
		return nil, domain.IngestSnapshot{}, false
	}
	items, err := LoadBillingFile(path, tenantID, provider)
	if err != nil {
		return nil, domain.IngestSnapshot{}, false
	}
	if len(items) == 0 {
		return nil, domain.IngestSnapshot{}, false
	}
	return items, snapshotFromItems(provider, items, invoice), true
}

func absPct(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
