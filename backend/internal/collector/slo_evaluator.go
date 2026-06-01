package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neuralops/platform/internal/observability"
)

// SLOEvaluator recomputes burn rates and emits burn alerts.
type SLOEvaluator struct {
	pool     *pgxpool.Pool
	prom     *observability.PromQLClient
	tenant   string
	alertURL string
}

// NewSLOEvaluator creates an SLO evaluator.
func NewSLOEvaluator(pool *pgxpool.Pool, prom *observability.PromQLClient, tenantID, alertURL string) *SLOEvaluator {
	if tenantID == "" {
		tenantID = "default"
	}
	return &SLOEvaluator{pool: pool, prom: prom, tenant: tenantID, alertURL: alertURL}
}

type sloRow struct {
	id, name, service, sliQuery string
	target, burnThreshold         float64
	burnAlertEnabled              bool
	errorBudget, burnRate         float64
	status                        string
}

// EvaluateAll updates SLO metrics and fires burn alerts when enabled.
func (e *SLOEvaluator) EvaluateAll(ctx context.Context) error {
	if e.pool == nil {
		return fmt.Errorf("postgres unavailable")
	}
	rows, err := e.pool.Query(ctx, `
SELECT id::text, name, service, sli_query, target,
       COALESCE(burn_alert_threshold, 2.0), COALESCE(burn_alert_enabled, false),
       error_budget, burn_rate, status
FROM observability_slos WHERE tenant_id = $1`, e.tenant)
	if err != nil {
		return err
	}
	defer rows.Close()

	var slos []sloRow
	for rows.Next() {
		var s sloRow
		if err := rows.Scan(&s.id, &s.name, &s.service, &s.sliQuery, &s.target,
			&s.burnThreshold, &s.burnAlertEnabled, &s.errorBudget, &s.burnRate, &s.status); err != nil {
			return err
		}
		slos = append(slos, s)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, s := range slos {
		sli := e.evalSLI(ctx, s)
		errorBudget := math.Max(0, 100*(1-sli/s.target))
		burnRate := 0.0
		if errorBudget < 100 {
			burnRate = (100 - errorBudget) / math.Max(1, float64(30))
		}
		status := "OK"
		if sli < s.target {
			status = "BREACH"
		} else if burnRate > s.burnThreshold {
			status = "BURN"
		}
		_, _ = e.pool.Exec(ctx, `
UPDATE observability_slos SET error_budget = $2, burn_rate = $3, status = $4 WHERE id = $1`,
			s.id, errorBudget, burnRate, status)

		if s.burnAlertEnabled && burnRate >= s.burnThreshold {
			e.fireBurnAlert(ctx, s, burnRate, sli)
		}
	}
	return nil
}

func (e *SLOEvaluator) evalSLI(ctx context.Context, s sloRow) float64 {
	if e.prom == nil {
		return s.target
	}
	q := strings.TrimSpace(s.sliQuery)
	if q == "" {
		return s.target
	}
	v, err := e.prom.QueryInstant(ctx, q)
	if err != nil || v <= 0 {
		return s.target * 0.998
	}
	if v <= 1 {
		return v * 100
	}
	return v
}

func (e *SLOEvaluator) fireBurnAlert(ctx context.Context, s sloRow, burnRate, sli float64) {
	if e.alertURL == "" {
		return
	}
	body := fmt.Sprintf(`{
  "alerts": [{
    "status": "firing",
    "labels": {"alertname": "SLOBurnRate", "service": %q, "slo_id": %q, "severity": "P2"},
    "annotations": {"summary": %q, "description": %q},
    "startsAt": %q
  }]
}`, s.service, s.id, fmt.Sprintf("SLO burn: %s", s.name),
		fmt.Sprintf("Burn rate %.2f exceeds threshold %.2f (SLI %.3f%%)", burnRate, s.burnThreshold, sli),
		time.Now().UTC().Format(time.RFC3339))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.alertURL+"/webhook/prometheus", strings.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", e.tenant)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()

	var exists int
	_ = e.pool.QueryRow(ctx, `SELECT 1 FROM alert_rules WHERE tenant_id = $1 AND name = $2 LIMIT 1`, e.tenant, "SLO burn: "+s.name).Scan(&exists)
	if exists == 0 {
		labels, _ := json.Marshal(map[string]string{"slo_id": s.id, "type": "burn_rate"})
		_, _ = e.pool.Exec(ctx, `
INSERT INTO alert_rules (id, tenant_id, name, source, service_pattern, severity, enabled, labels, created_at, updated_at)
VALUES ($1, $2, $3, 'SLO', $4, 'P2', true, $5, NOW(), NOW())`,
			uuid.New(), e.tenant, "SLO burn: "+s.name, s.service, labels)
	}
}
