package observability

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SRSRepo persists SRS parity entities in Postgres.
type SRSRepo struct {
	pool *pgxpool.Pool
}

// NewSRSRepo creates an SRS repository.
func NewSRSRepo(pool *pgxpool.Pool) *SRSRepo {
	return &SRSRepo{pool: pool}
}

func (r *SRSRepo) available() bool {
	return r != nil && r.pool != nil
}

// Available reports whether Postgres persistence is configured.
func (r *SRSRepo) Available() bool {
	return r.available()
}

func loadABACFromPostgres(ctx context.Context, pool *pgxpool.Pool, tenantID string) (ABACPolicy, error) {
	var raw []byte
	err := pool.QueryRow(ctx, `SELECT abac FROM tenant_governance WHERE tenant_id = $1`, tenantID).Scan(&raw)
	if err != nil {
		return ABACPolicy{}, err
	}
	var p ABACPolicy
	if err := json.Unmarshal(raw, &p); err != nil {
		return ABACPolicy{}, err
	}
	return p, nil
}

func loadResidencyFromPostgres(ctx context.Context, pool *pgxpool.Pool, tenantID string) (DataResidencyPolicy, error) {
	var raw []byte
	err := pool.QueryRow(ctx, `SELECT residency FROM tenant_governance WHERE tenant_id = $1`, tenantID).Scan(&raw)
	if err != nil {
		return DataResidencyPolicy{}, err
	}
	var p DataResidencyPolicy
	if err := json.Unmarshal(raw, &p); err != nil {
		return DataResidencyPolicy{}, err
	}
	return p, nil
}

func (r *SRSRepo) SaveABAC(ctx context.Context, tenantID string, p ABACPolicy) error {
	if !r.available() {
		return errors.New("postgres unavailable")
	}
	p.UpdatedAt = time.Now().UTC()
	b, _ := json.Marshal(p)
	_, err := r.pool.Exec(ctx, `
INSERT INTO tenant_governance (tenant_id, abac, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (tenant_id) DO UPDATE SET abac = EXCLUDED.abac, updated_at = NOW()`,
		tenantID, b)
	return err
}

func (r *SRSRepo) SaveResidency(ctx context.Context, tenantID string, p DataResidencyPolicy) error {
	if !r.available() {
		return errors.New("postgres unavailable")
	}
	p.UpdatedAt = time.Now().UTC()
	b, _ := json.Marshal(p)
	_, err := r.pool.Exec(ctx, `
INSERT INTO tenant_governance (tenant_id, residency, updated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (tenant_id) DO UPDATE SET residency = EXCLUDED.residency, updated_at = NOW()`,
		tenantID, b)
	return err
}

func (r *SRSRepo) ListAlertPolicies(ctx context.Context, tenantID string) ([]AlertPolicy, error) {
	if !r.available() {
		return nil, errors.New("postgres unavailable")
	}
	rows, err := r.pool.Query(ctx, `
SELECT id, name, service_pattern, severity, enabled, expression, routes, context, dedupe_key
FROM alert_policies WHERE tenant_id = $1`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AlertPolicy
	for rows.Next() {
		var p AlertPolicy
		var routes, ctxRaw []byte
		if err := rows.Scan(&p.ID, &p.Name, &p.ServicePattern, &p.Severity, &p.Enabled, &p.Expression, &routes, &ctxRaw, &p.DedupeKey); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(routes, &p.Routes)
		_ = json.Unmarshal(ctxRaw, &p.Context)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *SRSRepo) SaveAlertPolicy(ctx context.Context, tenantID string, p AlertPolicy) error {
	if !r.available() {
		return errors.New("postgres unavailable")
	}
	routes, _ := json.Marshal(p.Routes)
	ctxB, _ := json.Marshal(p.Context)
	_, err := r.pool.Exec(ctx, `
INSERT INTO alert_policies (id, tenant_id, name, service_pattern, severity, enabled, expression, routes, context, dedupe_key, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW())
ON CONFLICT (id) DO UPDATE SET
  name=EXCLUDED.name, service_pattern=EXCLUDED.service_pattern, severity=EXCLUDED.severity,
  enabled=EXCLUDED.enabled, expression=EXCLUDED.expression, routes=EXCLUDED.routes,
  context=EXCLUDED.context, dedupe_key=EXCLUDED.dedupe_key, updated_at=NOW()`,
		p.ID, tenantID, p.Name, p.ServicePattern, p.Severity, p.Enabled, p.Expression, routes, ctxB, p.DedupeKey)
	return err
}

func (r *SRSRepo) SaveSecurityFinding(ctx context.Context, tenantID string, f SecurityFinding) error {
	if !r.available() {
		return errors.New("postgres unavailable")
	}
	_, err := r.pool.Exec(ctx, `
INSERT INTO security_findings (id, tenant_id, title, category, severity, service, asset, status, exploitability, incident_id, detected_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
ON CONFLICT (id) DO UPDATE SET
  title=EXCLUDED.title, category=EXCLUDED.category, severity=EXCLUDED.severity,
  service=EXCLUDED.service, asset=EXCLUDED.asset, status=EXCLUDED.status,
  exploitability=EXCLUDED.exploitability, incident_id=EXCLUDED.incident_id, detected_at=EXCLUDED.detected_at`,
		f.ID, tenantID, f.Title, f.Category, f.Severity, f.Service, f.Asset, f.Status, f.Exploitability, f.IncidentID, f.DetectedAt)
	return err
}

func (r *SRSRepo) GetSecurityFinding(ctx context.Context, tenantID, id string) (SecurityFinding, error) {
	if !r.available() {
		return SecurityFinding{}, errors.New("postgres unavailable")
	}
	var f SecurityFinding
	err := r.pool.QueryRow(ctx, `
SELECT id, title, category, severity, service, asset, status, exploitability, incident_id, detected_at
FROM security_findings WHERE tenant_id = $1 AND id = $2`, tenantID, id).Scan(
		&f.ID, &f.Title, &f.Category, &f.Severity, &f.Service, &f.Asset, &f.Status, &f.Exploitability, &f.IncidentID, &f.DetectedAt)
	return f, err
}

func (r *SRSRepo) ListSecurityFindings(ctx context.Context, tenantID string) ([]SecurityFinding, error) {
	if !r.available() {
		return nil, errors.New("postgres unavailable")
	}
	rows, err := r.pool.Query(ctx, `
SELECT id, title, category, severity, service, asset, status, exploitability, incident_id, detected_at
FROM security_findings WHERE tenant_id = $1`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SecurityFinding
	for rows.Next() {
		var f SecurityFinding
		if err := rows.Scan(&f.ID, &f.Title, &f.Category, &f.Severity, &f.Service, &f.Asset, &f.Status, &f.Exploitability, &f.IncidentID, &f.DetectedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (r *SRSRepo) GetCollectorAgent(ctx context.Context, tenantID, id string) (CollectorFleetAgent, error) {
	if !r.available() {
		return CollectorFleetAgent{}, errors.New("postgres unavailable")
	}
	var a CollectorFleetAgent
	err := r.pool.QueryRow(ctx, `
SELECT id, name, environment, version, status, last_heartbeat_at, policy_id
FROM collector_fleet_agents WHERE tenant_id = $1 AND id = $2`, tenantID, id).Scan(
		&a.ID, &a.Name, &a.Environment, &a.Version, &a.Status, &a.LastHeartbeatAt, &a.PolicyID)
	return a, err
}

func (r *SRSRepo) SaveCollectorAgent(ctx context.Context, tenantID string, a CollectorFleetAgent) error {
	if !r.available() {
		return errors.New("postgres unavailable")
	}
	_, err := r.pool.Exec(ctx, `
INSERT INTO collector_fleet_agents (id, tenant_id, name, environment, version, status, last_heartbeat_at, policy_id, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NOW())
ON CONFLICT (id) DO UPDATE SET
  name=EXCLUDED.name, environment=EXCLUDED.environment, version=EXCLUDED.version,
  status=EXCLUDED.status, last_heartbeat_at=EXCLUDED.last_heartbeat_at, policy_id=EXCLUDED.policy_id, updated_at=NOW()`,
		a.ID, tenantID, a.Name, a.Environment, a.Version, a.Status, a.LastHeartbeatAt, a.PolicyID)
	return err
}

func (r *SRSRepo) UpgradeCollectorAgent(ctx context.Context, tenantID, id, targetVersion string) (CollectorFleetAgent, error) {
	a, err := r.GetCollectorAgent(ctx, tenantID, id)
	if err != nil {
		return CollectorFleetAgent{}, err
	}
	if targetVersion == "" {
		targetVersion = bumpPatch(a.Version)
	}
	a.Version = targetVersion
	a.Status = "upgrading"
	a.LastHeartbeatAt = time.Now().UTC()
	if err := r.SaveCollectorAgent(ctx, tenantID, a); err != nil {
		return CollectorFleetAgent{}, err
	}
	a.Status = "healthy"
	_ = r.SaveCollectorAgent(ctx, tenantID, a)
	return a, nil
}

func bumpPatch(v string) string {
	if v == "" {
		return "1.0.0"
	}
	return v + "-canary"
}

// SeedFromStore copies in-memory seed data into Postgres for a tenant (idempotent best-effort).
func (r *SRSRepo) SeedFromStore(ctx context.Context, tenantID string, mem *Store) error {
	if !r.available() {
		return errors.New("postgres unavailable")
	}
	_ = r.SaveABAC(ctx, tenantID, mem.GetABACPolicy())
	_ = r.SaveResidency(ctx, tenantID, mem.GetDataResidency())
	for _, p := range mem.ListAlertPolicies() {
		if p.Context.RunbookURL == "" {
			p.Context = defaultAlertContext(p.ServicePattern)
		}
		_ = r.SaveAlertPolicy(ctx, tenantID, p)
	}
	for _, f := range mem.ListSecurityFindings() {
		_, err := r.pool.Exec(ctx, `
INSERT INTO security_findings (id, tenant_id, title, category, severity, service, asset, status, exploitability, incident_id, detected_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
ON CONFLICT (id) DO NOTHING`,
			f.ID, tenantID, f.Title, f.Category, f.Severity, f.Service, f.Asset, f.Status, f.Exploitability, f.IncidentID, f.DetectedAt)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}
	for _, a := range mem.ListCollectorFleet() {
		_ = r.SaveCollectorAgent(ctx, tenantID, a)
	}
	return nil
}
