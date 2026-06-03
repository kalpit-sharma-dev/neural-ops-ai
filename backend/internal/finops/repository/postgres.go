package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neuralops/platform/internal/finops/domain"
)

// PostgresStore delegates to memory until full PG queries are wired; persists budgets/rules when pool available.
type PostgresStore struct {
	pool     *pgxpool.Pool
	fallback Store
	strict   bool
}

// NewPostgresStore creates a hybrid store. When strict is true (FINOPS_REQUIRE_POSTGRES),
// line-item reads/writes never fall back to in-memory simulated data.
func NewPostgresStore(pool *pgxpool.Pool, fallback Store, strict bool) *PostgresStore {
	return &PostgresStore{pool: pool, fallback: fallback, strict: strict}
}

func (p *PostgresStore) available() bool { return p.pool != nil }

func (p *PostgresStore) requirePool() error {
	if p.strict && !p.available() {
		return fmt.Errorf("finops postgres store required but database pool is unavailable")
	}
	return nil
}

func (p *PostgresStore) allowFallback() bool {
	return !p.strict && p.fallback != nil
}

func (p *PostgresStore) SaveLineItems(ctx context.Context, tenantID string, items []domain.CostLineItem) error {
	if err := p.requirePool(); err != nil {
		return err
	}
	if p.allowFallback() {
		if err := p.fallback.SaveLineItems(ctx, tenantID, items); err != nil {
			return err
		}
	}
	if !p.available() || len(items) == 0 {
		return nil
	}
	for _, item := range items {
		tags, _ := json.Marshal(item.Tags)
		if tags == nil {
			tags = []byte("{}")
		}
		_, err := p.pool.Exec(ctx, `
INSERT INTO finops_cost_line_items (
  id, tenant_id, billing_period, provider, account_id, service, region, resource_id,
  usage_type, quantity, unit, amortized_cost, list_cost, effective_cost, cost_view,
  tags, team, environment, cost_center, ingested_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)
ON CONFLICT (id) DO UPDATE SET
  amortized_cost=EXCLUDED.amortized_cost, effective_cost=EXCLUDED.effective_cost, ingested_at=EXCLUDED.ingested_at`,
			item.ID, tenantID, item.BillingPeriod, item.Provider, item.AccountID, item.Service,
			item.Region, item.ResourceID, item.UsageType, item.Quantity, item.Unit,
			item.AmortizedCost, item.ListCost, item.EffectiveCost, item.CostView,
			tags, item.Team, item.Environment, item.CostCenter, item.IngestedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PostgresStore) ListLineItems(ctx context.Context, tenantID string, from, to time.Time, provider, scope string) ([]domain.CostLineItem, error) {
	if err := p.requirePool(); err != nil {
		return nil, err
	}
	if !p.available() {
		if p.allowFallback() {
			return p.fallback.ListLineItems(ctx, tenantID, from, to, provider, scope)
		}
		return nil, fmt.Errorf("finops postgres store required but database pool is unavailable")
	}
	q := `
SELECT id, billing_period, provider, account_id, service, region, resource_id, usage_type,
       quantity, unit, amortized_cost, list_cost, effective_cost, cost_view, tags,
       team, environment, cost_center, ingested_at
FROM finops_cost_line_items
WHERE tenant_id = $1 AND billing_period >= $2 AND billing_period <= $3`
	args := []any{tenantID, from, to}
	argN := 4
	if provider != "" {
		q += ` AND provider = $` + strconv.Itoa(argN)
		args = append(args, provider)
		argN++
	}
	if scope != "" && scope != "all" {
		q += ` AND team = $` + strconv.Itoa(argN)
		args = append(args, scope)
	}
	rows, err := p.pool.Query(ctx, q, args...)
	if err != nil {
		if p.allowFallback() {
			return p.fallback.ListLineItems(ctx, tenantID, from, to, provider, scope)
		}
		return nil, err
	}
	defer rows.Close()
	var out []domain.CostLineItem
	for rows.Next() {
		var item domain.CostLineItem
		var tags []byte
		if err := rows.Scan(
			&item.ID, &item.BillingPeriod, &item.Provider, &item.AccountID, &item.Service,
			&item.Region, &item.ResourceID, &item.UsageType, &item.Quantity, &item.Unit,
			&item.AmortizedCost, &item.ListCost, &item.EffectiveCost, &item.CostView, &tags,
			&item.Team, &item.Environment, &item.CostCenter, &item.IngestedAt,
		); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(tags, &item.Tags)
		item.TenantID = tenantID
		out = append(out, item)
	}
	return out, rows.Err()
}

func (p *PostgresStore) SaveIngestSnapshot(ctx context.Context, tenantID string, s domain.IngestSnapshot) error {
	if err := p.requirePool(); err != nil {
		return err
	}
	if !p.available() {
		if p.allowFallback() {
			return p.fallback.SaveIngestSnapshot(ctx, tenantID, s)
		}
		return fmt.Errorf("finops postgres store required but database pool is unavailable")
	}
	_, err := p.pool.Exec(ctx, `
INSERT INTO finops_ingest_snapshots (id, tenant_id, provider, billing_period, line_count, total_amortized, invoice_total, drift_pct, cost_view, ingested_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
ON CONFLICT (tenant_id, provider, billing_period, cost_view) DO UPDATE SET
  line_count=EXCLUDED.line_count, total_amortized=EXCLUDED.total_amortized, invoice_total=EXCLUDED.invoice_total,
  drift_pct=EXCLUDED.drift_pct, ingested_at=EXCLUDED.ingested_at`,
		s.ID, tenantID, s.Provider, s.BillingPeriod, s.LineCount, s.TotalAmortized, s.InvoiceTotal, s.DriftPct, s.CostView, s.IngestedAt)
	if err != nil {
		return err
	}
	if p.allowFallback() {
		return p.fallback.SaveIngestSnapshot(ctx, tenantID, s)
	}
	return nil
}

func (p *PostgresStore) ListAllocationRules(ctx context.Context, tenantID string) ([]domain.AllocationRule, error) {
	return p.fallback.ListAllocationRules(ctx, tenantID)
}

func (p *PostgresStore) CreateAllocationRule(ctx context.Context, tenantID string, r domain.AllocationRule) (domain.AllocationRule, error) {
	created, err := p.fallback.CreateAllocationRule(ctx, tenantID, r)
	if err != nil || !p.available() {
		return created, err
	}
	_, _ = p.pool.Exec(ctx, `
INSERT INTO finops_allocation_rules (id, tenant_id, name, dimension, tag_key, tag_value, priority, enabled, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		created.ID, tenantID, created.Name, created.Dimension, created.TagKey, created.TagValue, created.Priority, created.Enabled, created.CreatedAt, created.UpdatedAt)
	return created, nil
}

func (p *PostgresStore) UpdateAllocationRule(ctx context.Context, tenantID string, r domain.AllocationRule) (domain.AllocationRule, error) {
	return p.fallback.UpdateAllocationRule(ctx, tenantID, r)
}

func (p *PostgresStore) DeleteAllocationRule(ctx context.Context, tenantID, id string) error {
	if p.available() {
		_, _ = p.pool.Exec(ctx, `DELETE FROM finops_allocation_rules WHERE tenant_id=$1 AND id=$2`, tenantID, id)
	}
	return p.fallback.DeleteAllocationRule(ctx, tenantID, id)
}

func (p *PostgresStore) ListBudgets(ctx context.Context, tenantID string) ([]domain.Budget, error) {
	return p.fallback.ListBudgets(ctx, tenantID)
}

func (p *PostgresStore) CreateBudget(ctx context.Context, tenantID string, b domain.Budget) (domain.Budget, error) {
	created, err := p.fallback.CreateBudget(ctx, tenantID, b)
	if err != nil || !p.available() {
		return created, err
	}
	_, _ = p.pool.Exec(ctx, `
INSERT INTO finops_budgets (id, tenant_id, name, scope_type, scope_value, period, amount_usd, thresholds, notify_policy_id, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		created.ID, tenantID, created.Name, created.ScopeType, created.ScopeValue, created.Period, created.AmountUSD, `[50,80,100]`, created.NotifyPolicyID, created.CreatedAt, created.UpdatedAt)
	return created, nil
}

func (p *PostgresStore) UpdateBudget(ctx context.Context, tenantID string, b domain.Budget) (domain.Budget, error) {
	return p.fallback.UpdateBudget(ctx, tenantID, b)
}

func (p *PostgresStore) DeleteBudget(ctx context.Context, tenantID, id string) error {
	if p.available() {
		_, _ = p.pool.Exec(ctx, `DELETE FROM finops_budgets WHERE tenant_id=$1 AND id=$2`, tenantID, id)
	}
	return p.fallback.DeleteBudget(ctx, tenantID, id)
}

func (p *PostgresStore) ListAnomalies(ctx context.Context, tenantID, scope, severity, status string) ([]domain.Anomaly, error) {
	return p.fallback.ListAnomalies(ctx, tenantID, scope, severity, status)
}

func (p *PostgresStore) GetAnomaly(ctx context.Context, tenantID, id string) (domain.Anomaly, error) {
	return p.fallback.GetAnomaly(ctx, tenantID, id)
}

func (p *PostgresStore) SaveAnomaly(ctx context.Context, tenantID string, a domain.Anomaly) error {
	return p.fallback.SaveAnomaly(ctx, tenantID, a)
}

func (p *PostgresStore) UpdateAnomalyFeedback(ctx context.Context, tenantID, id, feedback string) (domain.Anomaly, error) {
	return p.fallback.UpdateAnomalyFeedback(ctx, tenantID, id, feedback)
}

func (p *PostgresStore) ListRecommendations(ctx context.Context, tenantID, recType, scope, status string) ([]domain.Recommendation, error) {
	return p.fallback.ListRecommendations(ctx, tenantID, recType, scope, status)
}

func (p *PostgresStore) GetRecommendation(ctx context.Context, tenantID, id string) (domain.Recommendation, error) {
	return p.fallback.GetRecommendation(ctx, tenantID, id)
}

func (p *PostgresStore) SaveRecommendation(ctx context.Context, tenantID string, r domain.Recommendation) error {
	return p.fallback.SaveRecommendation(ctx, tenantID, r)
}

func (p *PostgresStore) UpdateRecommendationStatus(ctx context.Context, tenantID, id, action string) (domain.Recommendation, error) {
	return p.fallback.UpdateRecommendationStatus(ctx, tenantID, id, action)
}

func (p *PostgresStore) SaveImportBatch(ctx context.Context, tenantID string, b domain.ImportBatch) error {
	return p.fallback.SaveImportBatch(ctx, tenantID, b)
}

func (p *PostgresStore) GetImportVersion(ctx context.Context, tenantID, dedupeKey string) (int, error) {
	return p.fallback.GetImportVersion(ctx, tenantID, dedupeKey)
}

func (p *PostgresStore) ReplaceLineItems(ctx context.Context, tenantID string, items []domain.CostLineItem) error {
	if err := p.requirePool(); err != nil {
		return err
	}
	if p.allowFallback() {
		if err := p.fallback.ReplaceLineItems(ctx, tenantID, items); err != nil {
			return err
		}
	}
	if !p.available() {
		return fmt.Errorf("finops postgres store required but database pool is unavailable")
	}
	if _, err := p.pool.Exec(ctx, `DELETE FROM finops_cost_line_items WHERE tenant_id = $1`, tenantID); err != nil {
		return err
	}
	return p.SaveLineItems(ctx, tenantID, items)
}

func (p *PostgresStore) ListSharedSplits(ctx context.Context, tenantID string) ([]domain.SharedSplitRule, error) {
	return p.fallback.ListSharedSplits(ctx, tenantID)
}

func (p *PostgresStore) SaveSharedSplits(ctx context.Context, tenantID string, rules []domain.SharedSplitRule) error {
	return p.fallback.SaveSharedSplits(ctx, tenantID, rules)
}

func (p *PostgresStore) ListTagSuggestions(ctx context.Context, tenantID string) ([]domain.TagSuggestion, error) {
	return p.fallback.ListTagSuggestions(ctx, tenantID)
}

func (p *PostgresStore) SaveTagSuggestions(ctx context.Context, tenantID string, s []domain.TagSuggestion) error {
	return p.fallback.SaveTagSuggestions(ctx, tenantID, s)
}

func (p *PostgresStore) GetAnomalySensitivity(ctx context.Context, tenantID, scope string) (domain.AnomalySensitivity, error) {
	return p.fallback.GetAnomalySensitivity(ctx, tenantID, scope)
}

func (p *PostgresStore) SaveAnomalySensitivity(ctx context.Context, tenantID string, s domain.AnomalySensitivity) error {
	return p.fallback.SaveAnomalySensitivity(ctx, tenantID, s)
}

func (p *PostgresStore) ListIngestSnapshots(ctx context.Context, tenantID string) ([]domain.IngestSnapshot, error) {
	if err := p.requirePool(); err != nil {
		return nil, err
	}
	if !p.available() {
		if p.allowFallback() {
			return p.fallback.ListIngestSnapshots(ctx, tenantID)
		}
		return nil, fmt.Errorf("finops postgres store required but database pool is unavailable")
	}
	rows, err := p.pool.Query(ctx, `
SELECT id, provider, billing_period, line_count, total_amortized, invoice_total, drift_pct, cost_view, ingested_at, COALESCE(version,1)
FROM finops_ingest_snapshots WHERE tenant_id = $1 ORDER BY ingested_at DESC`, tenantID)
	if err != nil {
		if p.allowFallback() {
			return p.fallback.ListIngestSnapshots(ctx, tenantID)
		}
		return nil, err
	}
	defer rows.Close()
	var out []domain.IngestSnapshot
	for rows.Next() {
		var s domain.IngestSnapshot
		if err := rows.Scan(&s.ID, &s.Provider, &s.BillingPeriod, &s.LineCount, &s.TotalAmortized,
			&s.InvoiceTotal, &s.DriftPct, &s.CostView, &s.IngestedAt, &s.Version); err != nil {
			return nil, err
		}
		s.DriftAlert = s.DriftPct > 1.0
		out = append(out, s)
	}
	return out, rows.Err()
}

func (p *PostgresStore) SaveChargebackStatement(ctx context.Context, tenantID string, st domain.ChargebackStatement) error {
	return p.fallback.SaveChargebackStatement(ctx, tenantID, st)
}

func (p *PostgresStore) GetChargebackStatement(ctx context.Context, tenantID, id string) (domain.ChargebackStatement, error) {
	return p.fallback.GetChargebackStatement(ctx, tenantID, id)
}

func (p *PostgresStore) ListChargebackStatements(ctx context.Context, tenantID, costCenter string) ([]domain.ChargebackStatement, error) {
	return p.fallback.ListChargebackStatements(ctx, tenantID, costCenter)
}

func (p *PostgresStore) SaveScenarioRun(ctx context.Context, tenantID string, r domain.ScenarioResult) error {
	return p.fallback.SaveScenarioRun(ctx, tenantID, r)
}

func (p *PostgresStore) ListScenarioRuns(ctx context.Context, tenantID string, limit int) ([]domain.ScenarioResult, error) {
	return p.fallback.ListScenarioRuns(ctx, tenantID, limit)
}
