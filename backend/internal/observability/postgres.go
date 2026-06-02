package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepo persists observability UI config in Postgres.
type PostgresRepo struct {
	pool *pgxpool.Pool
}

// NewPostgresRepo creates a postgres observability repository.
func NewPostgresRepo(pool *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{pool: pool}
}

func (r *PostgresRepo) available() bool {
	return r != nil && r.pool != nil
}

// ListDashboards loads dashboards for tenant.
func (r *PostgresRepo) ListDashboards(ctx context.Context, tenantID string) ([]Dashboard, error) {
	if !r.available() {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `
SELECT id::text, tenant_id, name, COALESCE(description, ''), shared, tiles, created_at, updated_at
FROM observability_dashboards WHERE tenant_id = $1 ORDER BY updated_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDashboards(rows)
}

// GetDashboard loads one dashboard.
func (r *PostgresRepo) GetDashboard(ctx context.Context, tenantID, id string) (Dashboard, error) {
	if !r.available() {
		return Dashboard{}, fmt.Errorf("dashboard not found")
	}
	row := r.pool.QueryRow(ctx, `
SELECT id::text, tenant_id, name, COALESCE(description, ''), shared, tiles, created_at, updated_at
FROM observability_dashboards WHERE tenant_id = $1 AND id = $2`, tenantID, id)
	return scanDashboard(row)
}

// SaveDashboard upserts a dashboard.
func (r *PostgresRepo) SaveDashboard(ctx context.Context, d Dashboard) (Dashboard, error) {
	if !r.available() {
		return d, fmt.Errorf("postgres unavailable")
	}
	tilesJSON, _ := json.Marshal(d.Tiles)
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	var desc string
	err := r.pool.QueryRow(ctx, `
INSERT INTO observability_dashboards (id, tenant_id, name, description, shared, tiles, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name, description = EXCLUDED.description, shared = EXCLUDED.shared,
  tiles = EXCLUDED.tiles, updated_at = EXCLUDED.updated_at
RETURNING id::text, tenant_id, name, COALESCE(description, ''), shared, tiles, created_at, updated_at`,
		d.ID, d.TenantID, d.Name, d.Description, d.Shared, tilesJSON, now,
	).Scan(&d.ID, &d.TenantID, &d.Name, &desc, &d.Shared, &tilesJSON, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return d, err
	}
	d.Description = desc
	_ = json.Unmarshal(tilesJSON, &d.Tiles)
	return d, nil
}

// DeleteDashboard removes a dashboard.
func (r *PostgresRepo) DeleteDashboard(ctx context.Context, tenantID, id string) error {
	if !r.available() {
		return fmt.Errorf("postgres unavailable")
	}
	tag, err := r.pool.Exec(ctx, `DELETE FROM observability_dashboards WHERE tenant_id = $1 AND id = $2`, tenantID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("dashboard not found")
	}
	return nil
}

// ListSLOs returns SLOs for tenant.
func (r *PostgresRepo) ListSLOs(ctx context.Context, tenantID string) ([]SLO, error) {
	if !r.available() {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `
SELECT id::text, tenant_id, name, service, sli_query, target, window_days, error_budget, burn_rate, status, created_at
FROM observability_slos WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]SLO, 0)
	for rows.Next() {
		var slo SLO
		if err := rows.Scan(&slo.ID, &slo.TenantID, &slo.Name, &slo.Service, &slo.SLIQuery, &slo.Target, &slo.WindowDays, &slo.ErrorBudget, &slo.BurnRate, &slo.Status, &slo.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, slo)
	}
	return out, rows.Err()
}

// SaveSLO upserts an SLO.
func (r *PostgresRepo) SaveSLO(ctx context.Context, slo SLO) (SLO, error) {
	if !r.available() {
		return slo, fmt.Errorf("postgres unavailable")
	}
	if slo.ID == "" {
		slo.ID = uuid.New().String()
	}
	if slo.CreatedAt.IsZero() {
		slo.CreatedAt = time.Now().UTC()
	}
	err := r.pool.QueryRow(ctx, `
INSERT INTO observability_slos (id, tenant_id, name, service, sli_query, target, window_days, error_budget, burn_rate, status, created_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
ON CONFLICT (id) DO UPDATE SET
  name=EXCLUDED.name, service=EXCLUDED.service, sli_query=EXCLUDED.sli_query, target=EXCLUDED.target,
  window_days=EXCLUDED.window_days, error_budget=EXCLUDED.error_budget, burn_rate=EXCLUDED.burn_rate, status=EXCLUDED.status
RETURNING id::text`, slo.ID, slo.TenantID, slo.Name, slo.Service, slo.SLIQuery, slo.Target, slo.WindowDays, slo.ErrorBudget, slo.BurnRate, slo.Status, slo.CreatedAt,
	).Scan(&slo.ID)
	return slo, err
}

func scanDashboards(rows pgx.Rows) ([]Dashboard, error) {
	out := make([]Dashboard, 0)
	for rows.Next() {
		d, err := scanDashboard(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func scanDashboard(row pgx.Row) (Dashboard, error) {
	var d Dashboard
	var tilesJSON []byte
	if err := row.Scan(&d.ID, &d.TenantID, &d.Name, &d.Description, &d.Shared, &tilesJSON, &d.CreatedAt, &d.UpdatedAt); err != nil {
		return d, err
	}
	_ = json.Unmarshal(tilesJSON, &d.Tiles)
	return d, nil
}

// QueryMetrics loads metric series from ClickHouse.
func QueryMetrics(ctx context.Context, conn driver.Conn, tenantID, metricType, service string, start, end time.Time) (MetricSeries, error) {
	if conn == nil {
		return MetricSeries{}, fmt.Errorf("clickhouse unavailable")
	}
	if tenantID == "" {
		tenantID = "default"
	}
	if end.IsZero() {
		end = time.Now().UTC()
	}
	if start.IsZero() {
		start = end.Add(-1 * time.Hour)
	}
	rows, err := conn.Query(ctx, `
SELECT toStartOfMinute(timestamp) AS bucket, avg(value) AS v
FROM metrics
WHERE tenant_id = ? AND metric_type = ? AND service_name = ? AND timestamp >= ? AND timestamp <= ?
GROUP BY bucket ORDER BY bucket`, tenantID, metricType, service, start.UTC(), end.UTC())
	if err != nil {
		return MetricSeries{}, err
	}
	defer rows.Close()
	points := make([]MetricSeriesPoint, 0)
	for rows.Next() {
		var ts time.Time
		var v float64
		if err := rows.Scan(&ts, &v); err != nil {
			return MetricSeries{}, err
		}
		points = append(points, MetricSeriesPoint{Timestamp: ts, Value: v})
	}
	if len(points) == 0 {
		return MetricSeries{}, fmt.Errorf("no data")
	}
	return MetricSeries{Name: metricType, Labels: map[string]string{"service": service}, Points: points}, nil
}

// QueryUsageFromClickHouse derives ingestion/usage counters from analytics tables.
// It returns an error when there is no meaningful live data so handlers can
// fall back to seeded demo stats.
func QueryUsageFromClickHouse(ctx context.Context, conn driver.Conn, tenantID string, since time.Time) (UsageStats, error) {
	if conn == nil {
		return UsageStats{}, fmt.Errorf("clickhouse unavailable")
	}
	if tenantID == "" {
		tenantID = "default"
	}
	if since.IsZero() {
		since = time.Now().UTC().Add(-24 * time.Hour)
	}
	var logsCount uint64
	var logsBytes uint64
	if err := conn.QueryRow(ctx, `
SELECT count(), coalesce(sum(length(message)), 0)
FROM logs
WHERE tenant_id = ? AND timestamp >= ?`, tenantID, since.UTC()).Scan(&logsCount, &logsBytes); err != nil {
		return UsageStats{}, err
	}
	var tracesCount uint64
	if err := conn.QueryRow(ctx, `
SELECT uniqExact(trace_id)
FROM trace_spans
WHERE tenant_id = ? AND start_time >= ?`, tenantID, since.UTC()).Scan(&tracesCount); err != nil {
		return UsageStats{}, err
	}
	var metricsCount uint64
	if err := conn.QueryRow(ctx, `
SELECT count()
FROM metrics
WHERE tenant_id = ? AND timestamp >= ?`, tenantID, since.UTC()).Scan(&metricsCount); err != nil {
		return UsageStats{}, err
	}
	if logsCount == 0 && tracesCount == 0 && metricsCount == 0 {
		return UsageStats{}, fmt.Errorf("no usage data")
	}
	return UsageStats{
		LogsIngestedGB:  float64(logsBytes) / 1_000_000_000.0,
		TracesIngested:  int64(tracesCount),
		MetricsIngested: int64(metricsCount),
		// No first-class sources yet for these counters.
		AITokensUsed: 0,
		ActiveUsers:  0,
	}, nil
}

// QuerySecurityVulnerabilities derives likely vuln findings from recent logs.
func QuerySecurityVulnerabilities(ctx context.Context, conn driver.Conn, tenantID string) ([]SecurityVulnerability, error) {
	if conn == nil {
		return nil, fmt.Errorf("clickhouse unavailable")
	}
	if tenantID == "" {
		tenantID = "default"
	}
	rows, err := conn.Query(ctx, `
SELECT service,
       any(message) AS sample_message,
       max(timestamp) AS detected_at
FROM logs
WHERE tenant_id = ?
  AND timestamp >= now() - INTERVAL 7 DAY
  AND (
    positionCaseInsensitive(message, 'cve-') > 0
    OR positionCaseInsensitive(classification, 'vulnerab') > 0
    OR positionCaseInsensitive(message, 'critical patch') > 0
    OR positionCaseInsensitive(message, 'dependency') > 0
  )
GROUP BY service
ORDER BY detected_at DESC
LIMIT 50`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]SecurityVulnerability, 0)
	i := 0
	for rows.Next() {
		var service, msg string
		var at time.Time
		if err := rows.Scan(&service, &msg, &at); err != nil {
			return nil, err
		}
		i++
		sev := "medium"
		lmsg := strings.ToLower(msg)
		if strings.Contains(lmsg, "critical") {
			sev = "critical"
		} else if strings.Contains(lmsg, "high") {
			sev = "high"
		}
		cve := extractCVE(msg)
		if cve == "" {
			cve = fmt.Sprintf("CVE-LIVE-%04d", i)
		}
		out = append(out, SecurityVulnerability{
			ID:          fmt.Sprintf("vuln-%d", i),
			CVE:         cve,
			Severity:    sev,
			Service:     service,
			Description: truncateForUI(msg, 180),
			DetectedAt:  at,
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no vulnerabilities")
	}
	return out, nil
}

// QuerySecurityAttacks derives attack events from runtime logs.
func QuerySecurityAttacks(ctx context.Context, conn driver.Conn, tenantID string) ([]SecurityAttack, error) {
	if conn == nil {
		return nil, fmt.Errorf("clickhouse unavailable")
	}
	if tenantID == "" {
		tenantID = "default"
	}
	rows, err := conn.Query(ctx, `
SELECT service,
       any(message) AS sample_message,
       max(timestamp) AS detected_at,
       any(classification) AS classn
FROM logs
WHERE tenant_id = ?
  AND timestamp >= now() - INTERVAL 24 HOUR
  AND (
    positionCaseInsensitive(message, 'sql injection') > 0
    OR positionCaseInsensitive(message, 'xss') > 0
    OR positionCaseInsensitive(message, 'path traversal') > 0
    OR positionCaseInsensitive(message, 'ssrf') > 0
    OR positionCaseInsensitive(message, 'rce') > 0
    OR positionCaseInsensitive(classification, 'attack') > 0
    OR positionCaseInsensitive(classification, 'threat') > 0
  )
GROUP BY service
ORDER BY detected_at DESC
LIMIT 100`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]SecurityAttack, 0)
	i := 0
	for rows.Next() {
		var service, msg, classn string
		var at time.Time
		if err := rows.Scan(&service, &msg, &at, &classn); err != nil {
			return nil, err
		}
		i++
		attackType := inferAttackType(msg, classn)
		sourceIP := extractIP(msg)
		if sourceIP == "" {
			sourceIP = "unknown"
		}
		blocked := strings.Contains(strings.ToLower(msg), "blocked") || strings.Contains(strings.ToLower(classn), "blocked")
		out = append(out, SecurityAttack{
			ID:         fmt.Sprintf("atk-%d", i),
			Type:       attackType,
			SourceIP:   sourceIP,
			Service:    service,
			Blocked:    blocked,
			DetectedAt: at,
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no attacks")
	}
	return out, nil
}

func inferAttackType(msg, classn string) string {
	l := strings.ToLower(msg + " " + classn)
	switch {
	case strings.Contains(l, "sql injection"):
		return "SQL injection"
	case strings.Contains(l, "xss"):
		return "Cross-site scripting"
	case strings.Contains(l, "path traversal"):
		return "Path traversal"
	case strings.Contains(l, "ssrf"):
		return "SSRF attempt"
	case strings.Contains(l, "rce"):
		return "Remote code execution attempt"
	default:
		return "Suspicious request"
	}
}

func extractCVE(s string) string {
	u := strings.ToUpper(s)
	i := strings.Index(u, "CVE-")
	if i < 0 {
		return ""
	}
	j := i
	for j < len(u) {
		ch := u[j]
		if (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-' {
			j++
			continue
		}
		break
	}
	return u[i:j]
}

func extractIP(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return !(r == '.' || (r >= '0' && r <= '9'))
	})
	for _, p := range parts {
		if p == "" {
			continue
		}
		octets := strings.Split(p, ".")
		if len(octets) != 4 {
			continue
		}
		ok := true
		for _, o := range octets {
			if o == "" || len(o) > 3 {
				ok = false
				break
			}
		}
		if ok {
			return p
		}
	}
	return ""
}

func truncateForUI(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// LoadTopologyFromPostgres builds topology from service graph tables.
func LoadTopologyFromPostgres(ctx context.Context, pool *pgxpool.Pool, tenantID string) (TopologyGraph, error) {
	if pool == nil {
		return TopologyGraph{}, fmt.Errorf("postgres unavailable")
	}
	if tenantID == "" {
		tenantID = "default"
	}
	nodeRows, err := pool.Query(ctx, `
SELECT sh.service_name, COALESCE(sh.health_score, 100), COALESCE(sh.error_rate, 0)
FROM service_health sh WHERE sh.tenant_id = $1`, tenantID)
	if err != nil {
		return TopologyGraph{}, err
	}
	defer nodeRows.Close()
	nodes := make([]TopologyNode, 0)
	nodeSet := map[string]bool{}
	for nodeRows.Next() {
		var name string
		var healthScore, errorRate float64
		if err := nodeRows.Scan(&name, &healthScore, &errorRate); err != nil {
			return TopologyGraph{}, err
		}
		health := "healthy"
		if healthScore < 50 {
			health = "critical"
		} else if healthScore < 80 {
			health = "degraded"
		}
		nodes = append(nodes, TopologyNode{
			ID: name, DisplayName: name, Type: "service", Health: health,
			ErrorRate: errorRate, Throughput: 1000,
		})
		nodeSet[name] = true
	}
	edgeRows, err := pool.Query(ctx, `
SELECT source_service, target_service, call_count, COALESCE(error_rate, 0), COALESCE(p99_latency_ms, 0)
FROM service_dependencies WHERE tenant_id = $1`, tenantID)
	if err != nil {
		return TopologyGraph{}, err
	}
	defer edgeRows.Close()
	edges := make([]TopologyEdge, 0)
	for edgeRows.Next() {
		var e TopologyEdge
		if err := rowsScanEdge(edgeRows, &e); err != nil {
			return TopologyGraph{}, err
		}
		edges = append(edges, e)
		ensureNode(&nodes, nodeSet, e.Source)
		ensureNode(&nodes, nodeSet, e.Target)
	}
	if len(nodes) == 0 {
		return TopologyGraph{}, fmt.Errorf("no topology")
	}
	return TopologyGraph{Nodes: nodes, Edges: edges, At: time.Now().UTC()}, nil
}

func rowsScanEdge(rows pgx.Rows, e *TopologyEdge) error {
	return rows.Scan(&e.Source, &e.Target, &e.CallCount, &e.ErrorRate, &e.P95Ms)
}

func ensureNode(nodes *[]TopologyNode, set map[string]bool, id string) {
	if set[id] {
		return
	}
	*nodes = append(*nodes, TopologyNode{ID: id, DisplayName: id, Type: "service", Health: "healthy"})
	set[id] = true
}

// ListAuditEntries reads audit logs.
func ListAuditEntries(ctx context.Context, pool *pgxpool.Pool, tenantID string, limit int) ([]AuditEntry, error) {
	if pool == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 100
	}
	rows, err := pool.Query(ctx, `
SELECT id::text, COALESCE(user_id, ''), action, COALESCE(resource_type, ''), COALESCE(resource_id, ''), created_at
FROM audit_logs WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AuditEntry, 0)
	for rows.Next() {
		var e AuditEntry
		if err := rows.Scan(&e.ID, &e.UserID, &e.Action, &e.Resource, &e.Detail, &e.Timestamp); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ListLogMetricRules loads log metric rules for tenant.
func (r *PostgresRepo) ListLogMetricRules(ctx context.Context, tenantID string) ([]LogMetricRule, error) {
	if !r.available() {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `
SELECT id::text, tenant_id, name, pattern, COALESCE(service, ''), enabled
FROM observability_log_metric_rules WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]LogMetricRule, 0)
	for rows.Next() {
		var rule LogMetricRule
		if err := rows.Scan(&rule.ID, &rule.TenantID, &rule.Name, &rule.Pattern, &rule.Service, &rule.Enabled); err != nil {
			return nil, err
		}
		out = append(out, rule)
	}
	return out, rows.Err()
}

// SaveLogMetricRule inserts a log metric rule.
func (r *PostgresRepo) SaveLogMetricRule(ctx context.Context, rule LogMetricRule) (LogMetricRule, error) {
	if !r.available() {
		return rule, fmt.Errorf("postgres unavailable")
	}
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}
	err := r.pool.QueryRow(ctx, `
INSERT INTO observability_log_metric_rules (id, tenant_id, name, pattern, service, enabled)
VALUES ($1,$2,$3,$4,NULLIF($5,''),$6)
RETURNING id::text`, rule.ID, rule.TenantID, rule.Name, rule.Pattern, rule.Service, rule.Enabled,
	).Scan(&rule.ID)
	return rule, err
}

// ListLogParsingRules loads log parsing rules for tenant.
func (r *PostgresRepo) ListLogParsingRules(ctx context.Context, tenantID string) ([]LogParsingRule, error) {
	if !r.available() {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `
SELECT id::text, tenant_id, name, pattern, field, enabled
FROM observability_log_parsing_rules WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]LogParsingRule, 0)
	for rows.Next() {
		var rule LogParsingRule
		if err := rows.Scan(&rule.ID, &rule.TenantID, &rule.Name, &rule.Pattern, &rule.Field, &rule.Enabled); err != nil {
			return nil, err
		}
		out = append(out, rule)
	}
	return out, rows.Err()
}

// SaveLogParsingRule inserts a log parsing rule.
func (r *PostgresRepo) SaveLogParsingRule(ctx context.Context, rule LogParsingRule) (LogParsingRule, error) {
	if !r.available() {
		return rule, fmt.Errorf("postgres unavailable")
	}
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}
	if rule.Field == "" {
		rule.Field = "message"
	}
	err := r.pool.QueryRow(ctx, `
INSERT INTO observability_log_parsing_rules (id, tenant_id, name, pattern, field, enabled)
VALUES ($1,$2,$3,$4,$5,$6)
RETURNING id::text`, rule.ID, rule.TenantID, rule.Name, rule.Pattern, rule.Field, rule.Enabled,
	).Scan(&rule.ID)
	return rule, err
}

// ListWorkflows loads workflows for tenant.
func (r *PostgresRepo) ListWorkflows(ctx context.Context, tenantID string) ([]Workflow, error) {
	if !r.available() {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `
SELECT id::text, name, trigger_name, enabled, steps, graph
FROM observability_workflows WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Workflow, 0)
	for rows.Next() {
		var w Workflow
		var stepsJSON []byte
		var graphJSON []byte
		if err := rows.Scan(&w.ID, &w.Name, &w.Trigger, &w.Enabled, &stepsJSON, &graphJSON); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(stepsJSON, &w.Steps)
		if g := unmarshalGraph(graphJSON); g != nil {
			w.Graph = g
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// ListNotebooks loads notebooks for tenant.
func (r *PostgresRepo) ListNotebooks(ctx context.Context, tenantID string) ([]Notebook, error) {
	if !r.available() {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `
SELECT id::text, name, cells, updated_at
FROM observability_notebooks WHERE tenant_id = $1 ORDER BY updated_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Notebook, 0)
	for rows.Next() {
		var n Notebook
		var cellsJSON []byte
		if err := rows.Scan(&n.ID, &n.Name, &cellsJSON, &n.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(cellsJSON, &n.Cells)
		out = append(out, n)
	}
	return out, rows.Err()
}

// SaveWorkflow inserts a workflow.
func (r *PostgresRepo) SaveWorkflow(ctx context.Context, tenantID string, w Workflow) (Workflow, error) {
	if !r.available() {
		return w, fmt.Errorf("postgres unavailable")
	}
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	stepsJSON, _ := json.Marshal(w.Steps)
	graphJSON := marshalGraph(w.Graph)
	err := r.pool.QueryRow(ctx, `
INSERT INTO observability_workflows (id, tenant_id, name, trigger_name, enabled, steps, graph)
VALUES ($1,$2,$3,$4,$5,$6,$7)
RETURNING id::text`, w.ID, tenantID, w.Name, w.Trigger, w.Enabled, stepsJSON, graphJSON,
	).Scan(&w.ID)
	return w, err
}

// UpdateWorkflow updates an existing workflow scoped to a tenant.
func (r *PostgresRepo) UpdateWorkflow(ctx context.Context, tenantID string, w Workflow) (Workflow, error) {
	if !r.available() {
		return w, fmt.Errorf("postgres unavailable")
	}
	stepsJSON, _ := json.Marshal(w.Steps)
	graphJSON := marshalGraph(w.Graph)
	tag, err := r.pool.Exec(ctx, `
UPDATE observability_workflows
SET name = $3, trigger_name = $4, enabled = $5, steps = $6, graph = $7
WHERE id = $1 AND tenant_id = $2`,
		w.ID, tenantID, w.Name, w.Trigger, w.Enabled, stepsJSON, graphJSON,
	)
	if err != nil {
		return w, err
	}
	if tag.RowsAffected() == 0 {
		return w, fmt.Errorf("workflow not found")
	}
	return w, nil
}

// marshalGraph serializes a workflow graph to JSONB, normalizing nil to '{}'.
func marshalGraph(g *WorkflowGraph) []byte {
	if g == nil {
		return []byte("{}")
	}
	b, err := json.Marshal(g)
	if err != nil {
		return []byte("{}")
	}
	return b
}

// unmarshalGraph parses a JSONB graph column, returning nil when empty.
func unmarshalGraph(raw []byte) *WorkflowGraph {
	if len(raw) == 0 {
		return nil
	}
	var g WorkflowGraph
	if err := json.Unmarshal(raw, &g); err != nil {
		return nil
	}
	if len(g.Nodes) == 0 {
		return nil
	}
	return &g
}

// DeleteWorkflow removes a workflow (and cascades its runs) for a tenant.
func (r *PostgresRepo) DeleteWorkflow(ctx context.Context, tenantID, id string) error {
	if !r.available() {
		return fmt.Errorf("postgres unavailable")
	}
	tag, err := r.pool.Exec(ctx, `
DELETE FROM observability_workflows WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("workflow not found")
	}
	return nil
}

// SaveNotebook inserts a notebook.
func (r *PostgresRepo) SaveNotebook(ctx context.Context, tenantID string, n Notebook) (Notebook, error) {
	if !r.available() {
		return n, fmt.Errorf("postgres unavailable")
	}
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	cellsJSON, _ := json.Marshal(n.Cells)
	now := time.Now().UTC()
	err := r.pool.QueryRow(ctx, `
INSERT INTO observability_notebooks (id, tenant_id, name, cells, updated_at)
VALUES ($1,$2,$3,$4,$5)
RETURNING id::text`, n.ID, tenantID, n.Name, cellsJSON, now,
	).Scan(&n.ID)
	if err == nil {
		n.UpdatedAt = now
	}
	return n, err
}
