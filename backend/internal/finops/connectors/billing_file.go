package connectors

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/finops/domain"
)

type billingFileRecord struct {
	Provider      string            `json:"provider"`
	AccountID     string            `json:"accountId"`
	Service       string            `json:"service"`
	Region        string            `json:"region"`
	ResourceID    string            `json:"resourceId"`
	UsageType     string            `json:"usageType"`
	Quantity      float64           `json:"quantity"`
	Unit          string            `json:"unit"`
	AmortizedCost float64           `json:"amortizedCost"`
	ListCost      float64           `json:"listCost"`
	EffectiveCost float64           `json:"effectiveCost"`
	BillingPeriod string            `json:"billingPeriod"`
	Tags          map[string]string `json:"tags"`
}

// LoadBillingFile ingests NDJSON or JSON array export (CUR/BQ/Azure billing export).
func LoadBillingFile(path, tenantID, provider string) ([]domain.CostLineItem, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read billing file: %w", err)
	}
	trim := strings.TrimSpace(string(raw))
	var records []billingFileRecord
	if strings.HasPrefix(trim, "[") {
		if err := json.Unmarshal(raw, &records); err != nil {
			return nil, fmt.Errorf("parse billing json array: %w", err)
		}
	} else {
		sc := bufio.NewScanner(strings.NewReader(trim))
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" {
				continue
			}
			var rec billingFileRecord
			if err := json.Unmarshal([]byte(line), &rec); err != nil {
				return nil, fmt.Errorf("parse billing ndjson line: %w", err)
			}
			records = append(records, rec)
		}
		if err := sc.Err(); err != nil {
			return nil, err
		}
	}
	now := time.Now().UTC()
	items := make([]domain.CostLineItem, 0, len(records))
	for _, rec := range records {
		if rec.Provider == "" {
			rec.Provider = provider
		}
		period := now
		if rec.BillingPeriod != "" {
			if t, err := time.Parse("2006-01-02", rec.BillingPeriod); err == nil {
				period = t.UTC()
			}
		}
		amt := rec.AmortizedCost
		if amt == 0 {
			amt = rec.EffectiveCost
		}
		item := domain.CostLineItem{
			ID: uuid.NewString(), TenantID: tenantID, BillingPeriod: period,
			Provider: rec.Provider, AccountID: rec.AccountID, Service: rec.Service,
			Region: rec.Region, ResourceID: rec.ResourceID, UsageType: rec.UsageType,
			Quantity: rec.Quantity, Unit: rec.Unit, AmortizedCost: amt,
			ListCost: rec.ListCost, EffectiveCost: rec.EffectiveCost,
			CostView: "amortized", Tags: rec.Tags, IngestedAt: now,
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
	}
	return items, nil
}

func snapshotFromItems(provider string, items []domain.CostLineItem, invoiceTotal float64) domain.IngestSnapshot {
	now := time.Now().UTC()
	period := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	var total float64
	for _, i := range items {
		if strings.EqualFold(i.Provider, provider) {
			total += i.AmortizedCost
		}
	}
	if invoiceTotal <= 0 {
		invoiceTotal = total * 1.005
	}
	drift := absPct((total-invoiceTotal)/invoiceTotal) * 100
	return domain.IngestSnapshot{
		ID: uuid.NewString(), Provider: provider, BillingPeriod: period,
		LineCount: int64(len(items)), TotalAmortized: total,
		InvoiceTotal: invoiceTotal, DriftPct: drift, CostView: "amortized",
		IngestedAt: now, DriftAlert: drift > driftTolerancePct,
	}
}
