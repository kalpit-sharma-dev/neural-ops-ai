package finops

import (
	"context"
	"time"

	"github.com/neuralops/platform/internal/finops/budgets"
	"github.com/neuralops/platform/internal/finops/domain"
)

const defaultStaleIngestThreshold = 26 * time.Hour

// IngestFreshness summarizes billing ingest freshness (REQ §9 dogfooding).
type IngestFreshness struct {
	LastIngestAt time.Time `json:"lastIngestAt"`
	LagSeconds   float64   `json:"lagSeconds"`
	Stale        bool      `json:"stale"`
	Providers    []string  `json:"providers"`
	Snapshots    []domain.IngestSnapshot `json:"snapshots,omitempty"`
}

// StaleIngestAlerter routes stale billing data alerts.
type StaleIngestAlerter interface {
	RouteStaleIngest(ctx context.Context, tenantID string, lagSeconds float64)
}

// SetStaleIngestAlerter configures stale ingest alert routing.
func (s *Service) SetStaleIngestAlerter(a StaleIngestAlerter) {
	s.staleAlert = a
}

// IngestStatus returns freshness metrics for the tenant.
func (s *Service) IngestStatus(ctx context.Context, tenantID string) IngestFreshness {
	snaps, _ := s.repo.ListIngestSnapshots(ctx, tenantID)
	var last time.Time
	providers := make([]string, 0, len(snaps))
	for _, snap := range snaps {
		providers = append(providers, snap.Provider)
		if snap.IngestedAt.After(last) {
			last = snap.IngestedAt
		}
	}
	lag := budgets.StaleIngestLag(last)
	stale := last.IsZero() || time.Duration(lag)*time.Second > defaultStaleIngestThreshold
	recordIngestLag(tenantID, lag)
	return IngestFreshness{
		LastIngestAt: last,
		LagSeconds:   lag,
		Stale:        stale,
		Providers:    providers,
		Snapshots:    snaps,
	}
}

// CheckStaleIngest alerts when billing data is older than the daily cadence threshold.
func (s *Service) CheckStaleIngest(ctx context.Context, tenantID string) IngestFreshness {
	status := s.IngestStatus(ctx, tenantID)
	if status.Stale && s.staleAlert != nil {
		s.staleAlert.RouteStaleIngest(ctx, tenantID, status.LagSeconds)
	}
	return status
}
