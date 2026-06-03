package finops

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// DefaultIngestCadence is the daily billing connector schedule (REQ-FINOPS-001).
const DefaultIngestCadence = 24 * time.Hour

// StartScheduledIngest runs billing ingest on a configurable cadence until ctx is cancelled.
func StartScheduledIngest(ctx context.Context, svc *Service, tenantID string, cadence time.Duration, log *zap.Logger) {
	if svc == nil {
		return
	}
	if cadence <= 0 {
		cadence = DefaultIngestCadence
	}
	if tenantID == "" {
		tenantID = "default"
	}
	go func() {
		ticker := time.NewTicker(cadence)
		defer ticker.Stop()
		run := func() {
			if _, err := svc.IngestAll(ctx, tenantID); err != nil && log != nil {
				log.Warn("scheduled finops ingest failed",
					zap.String("tenantId", tenantID),
					zap.Error(err),
				)
			}
			svc.CheckStaleIngest(ctx, tenantID)
		}
		run()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				run()
			}
		}
	}()
}

// StartFreshnessMonitor periodically checks ingest lag and alerts when stale.
func StartFreshnessMonitor(ctx context.Context, svc *Service, tenantID string, interval time.Duration, log *zap.Logger) {
	if svc == nil {
		return
	}
	if interval <= 0 {
		interval = time.Hour
	}
	if tenantID == "" {
		tenantID = "default"
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				status := svc.CheckStaleIngest(ctx, tenantID)
				if status.Stale && log != nil {
					log.Warn("finops billing ingest stale",
						zap.String("tenantId", tenantID),
						zap.Float64("lagSeconds", status.LagSeconds),
					)
				}
			}
		}
	}()
}
