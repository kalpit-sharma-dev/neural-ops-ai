package tracequery

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/neuralops/platform/internal/apm"
	"github.com/neuralops/platform/internal/ingestion/dto"
	"github.com/neuralops/platform/internal/platform/db"
)

// SpanStore persists and queries distributed trace spans in ClickHouse.
type SpanStore struct {
	conn driver.Conn
}

// NewSpanStore opens a ClickHouse span store.
func NewSpanStore(ctx context.Context, dsn string) (*SpanStore, error) {
	opts, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse clickhouse dsn: %w", err)
	}
	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}
	if err := db.RunClickHouseMigrations(ctx, conn); err != nil {
		return nil, err
	}
	return &SpanStore{conn: conn}, nil
}

// WriteSpan inserts one span row.
func (s *SpanStore) WriteSpan(ctx context.Context, span dto.TraceSpanRequest) error {
	if s == nil || s.conn == nil {
		return nil
	}
	tenantID := span.TenantID
	if tenantID == "" {
		tenantID = "default"
	}
	tags := span.Tags
	if tags == nil {
		tags = map[string]string{}
	}
	return s.conn.Exec(ctx, `
INSERT INTO trace_spans (tenant_id, trace_id, span_id, parent_id, service, operation, start_time, duration_ms, status, tags)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		tenantID, span.TraceID, span.SpanID, span.ParentID, span.Service, span.Operation,
		span.StartTime.UTC(), span.DurationMs, span.Status, tags,
	)
}

// GetTrace loads all spans for a trace ID.
func (s *SpanStore) GetTrace(ctx context.Context, tenantID, traceID string) (apm.TraceDetail, error) {
	if tenantID == "" {
		tenantID = "default"
	}
	rows, err := s.conn.Query(ctx, `
SELECT trace_id, span_id, parent_id, service, operation, start_time, duration_ms, status, tags
FROM trace_spans
WHERE tenant_id = ? AND trace_id = ?
ORDER BY start_time`, tenantID, traceID)
	if err != nil {
		return apm.TraceDetail{}, err
	}
	defer rows.Close()

	spans := make([]apm.Span, 0)
	var total int64
	service := ""
	status := "OK"
	for rows.Next() {
		var sp apm.Span
		var tags map[string]string
		if err := rows.Scan(&sp.TraceID, &sp.SpanID, &sp.ParentID, &sp.Service, &sp.Operation, &sp.StartTime, &sp.DurationMs, &sp.Status, &tags); err != nil {
			return apm.TraceDetail{}, err
		}
		sp.Tags = tags
		spans = append(spans, sp)
		total += sp.DurationMs
		if service == "" {
			service = sp.Service
		}
		if sp.Status == "ERROR" {
			status = "ERROR"
		}
	}
	if len(spans) == 0 {
		return apm.TraceDetail{}, fmt.Errorf("trace not found")
	}
	return apm.TraceDetail{
		TraceID: traceID, RootSpan: spans[0].SpanID, Spans: spans, TotalMs: total,
		Service: service, Status: status, SpanCount: len(spans),
	}, nil
}

// SearchTraces finds trace summaries matching filters.
func (s *SpanStore) SearchTraces(ctx context.Context, tenantID string, req apm.TraceSearchRequest) ([]apm.TraceSummary, error) {
	if tenantID == "" {
		tenantID = "default"
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.conn.Query(ctx, `
SELECT trace_id,
       argMin(service, start_time) AS root_service,
       argMin(operation, start_time) AS root_operation,
       sum(duration_ms) AS total_ms,
       if(countIf(status = 'ERROR') > 0, 'ERROR', 'OK') AS trace_status,
       min(start_time) AS started,
       count() AS span_count
FROM trace_spans
WHERE tenant_id = ?
GROUP BY trace_id
ORDER BY started DESC
LIMIT ?`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]apm.TraceSummary, 0)
	for rows.Next() {
		var summary apm.TraceSummary
		if err := rows.Scan(&summary.TraceID, &summary.Service, &summary.Operation, &summary.DurationMs, &summary.Status, &summary.StartTime, &summary.SpanCount); err != nil {
			return nil, err
		}
		if req.Service != "" && !strings.EqualFold(summary.Service, req.Service) {
			continue
		}
		if req.Status != "" && !strings.EqualFold(summary.Status, req.Status) {
			continue
		}
		out = append(out, summary)
	}
	return out, rows.Err()
}

// ServiceFlow aggregates edges from recent spans.
func (s *SpanStore) ServiceFlow(ctx context.Context, tenantID string) ([]apm.FlowEdge, error) {
	if tenantID == "" {
		tenantID = "default"
	}
	since := time.Now().UTC().Add(-1 * time.Hour)
	rows, err := s.conn.Query(ctx, `
SELECT parent.service AS source, child.service AS target,
       count() AS calls,
       avgIf(child.status = 'ERROR', 1) * 100 AS err_pct,
       quantile(0.5)(child.duration_ms) AS p50,
       quantile(0.95)(child.duration_ms) AS p95
FROM trace_spans AS child
INNER JOIN trace_spans AS parent ON child.parent_id = parent.span_id AND child.trace_id = parent.trace_id AND child.tenant_id = parent.tenant_id
WHERE child.tenant_id = ? AND child.start_time >= ?
GROUP BY source, target
ORDER BY calls DESC
LIMIT 20`, tenantID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]apm.FlowEdge, 0)
	for rows.Next() {
		var edge apm.FlowEdge
		if err := rows.Scan(&edge.Source, &edge.Target, &edge.CallCount, &edge.ErrorRate, &edge.P50Ms, &edge.P95Ms); err != nil {
			return nil, err
		}
		out = append(out, edge)
	}
	return out, rows.Err()
}

// ServiceAnomaly is a derived anomaly signal for one service/metric.
type ServiceAnomaly struct {
	Service    string
	Metric     string // "error_rate" | "latency_p95"
	Score      float64
	Message    string
	DetectedAt time.Time
}

// DetectAnomalies compares the last 15 minutes of spans against the preceding
// baseline window per service and flags elevated error rate or p95 latency.
// This is a lightweight, dependency-free detector over real trace telemetry.
func (s *SpanStore) DetectAnomalies(ctx context.Context, tenantID string) ([]ServiceAnomaly, error) {
	if s == nil || s.conn == nil {
		return nil, fmt.Errorf("span store unavailable")
	}
	if tenantID == "" {
		tenantID = "default"
	}
	rows, err := s.conn.Query(ctx, `
SELECT service,
       countIf(start_time >= now() - INTERVAL 15 MINUTE) AS recent_total,
       countIf(start_time >= now() - INTERVAL 15 MINUTE AND status = 'ERROR') AS recent_err,
       countIf(start_time <  now() - INTERVAL 15 MINUTE) AS base_total,
       countIf(start_time <  now() - INTERVAL 15 MINUTE AND status = 'ERROR') AS base_err,
       quantileIf(0.95)(duration_ms, start_time >= now() - INTERVAL 15 MINUTE) AS recent_p95,
       quantileIf(0.95)(duration_ms, start_time <  now() - INTERVAL 15 MINUTE) AS base_p95
FROM trace_spans
WHERE tenant_id = ? AND start_time >= now() - INTERVAL 75 MINUTE
GROUP BY service`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	now := time.Now().UTC()
	out := make([]ServiceAnomaly, 0)
	for rows.Next() {
		var (
			service                string
			recentTotal, recentErr uint64
			baseTotal, baseErr     uint64
			recentP95, baseP95     float64
		)
		if err := rows.Scan(&service, &recentTotal, &recentErr, &baseTotal, &baseErr, &recentP95, &baseP95); err != nil {
			return nil, err
		}
		if recentTotal < 20 {
			continue // insufficient recent volume to judge
		}
		recentRate := float64(recentErr) / float64(recentTotal)
		baseRate := 0.0
		if baseTotal > 0 {
			baseRate = float64(baseErr) / float64(baseTotal)
		}
		// Error-rate anomaly: meaningfully elevated and at least 2x baseline.
		if recentRate >= 0.02 && recentRate >= 2*maxFloat(baseRate, 0.005) {
			ratio := recentRate / maxFloat(baseRate, 0.005)
			out = append(out, ServiceAnomaly{
				Service: service, Metric: "error_rate",
				Score:      clampScore(recentRate * 100 * 5),
				Message:    fmt.Sprintf("Error rate %.1f%% (%.1fx baseline)", recentRate*100, ratio),
				DetectedAt: now,
			})
		}
		// Latency anomaly: p95 elevated above baseline and above an absolute floor.
		if baseP95 > 0 && recentP95 > 200 && recentP95 >= 1.5*baseP95 {
			out = append(out, ServiceAnomaly{
				Service: service, Metric: "latency_p95",
				Score:      clampScore((recentP95/baseP95 - 1) * 100),
				Message:    fmt.Sprintf("p95 latency %.0fms (%.1fx baseline)", recentP95, recentP95/baseP95),
				DetectedAt: now,
			})
		}
	}
	return out, rows.Err()
}

// DBInstanceStat is a derived database instance summary from client db spans.
type DBInstanceStat struct {
	ID          string
	Name        string
	Engine      string
	SlowQueries int
	QPS         float64
}

// ListDatabases derives database instances from spans carrying OpenTelemetry
// db.* attributes over the last 15 minutes.
func (s *SpanStore) ListDatabases(ctx context.Context, tenantID string) ([]DBInstanceStat, error) {
	if s == nil || s.conn == nil {
		return nil, fmt.Errorf("span store unavailable")
	}
	if tenantID == "" {
		tenantID = "default"
	}
	const windowSeconds = 900.0
	rows, err := s.conn.Query(ctx, `
SELECT tags['db.system'] AS engine,
       tags['db.name'] AS db_name,
       count() AS calls,
       countIf(duration_ms > 100) AS slow
FROM trace_spans
WHERE tenant_id = ? AND tags['db.system'] != '' AND start_time >= now() - INTERVAL 15 MINUTE
GROUP BY engine, db_name
ORDER BY calls DESC
LIMIT 50`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]DBInstanceStat, 0)
	for rows.Next() {
		var engine, dbName string
		var calls, slow uint64
		if err := rows.Scan(&engine, &dbName, &calls, &slow); err != nil {
			return nil, err
		}
		name := dbName
		if name == "" {
			name = engine
		}
		out = append(out, DBInstanceStat{
			ID:          name,
			Name:        name,
			Engine:      engine,
			SlowQueries: int(slow),
			QPS:         round1(float64(calls) / windowSeconds),
		})
	}
	return out, rows.Err()
}

// DBStatementStat is a derived top-statement summary for a database.
type DBStatementStat struct {
	Query   string
	Calls   int64
	AvgMs   float64
	TotalMs float64
}

// DatabaseStatements returns the top db.statement spans for a database instance
// (matched by db.name or db.system) over the last hour.
func (s *SpanStore) DatabaseStatements(ctx context.Context, tenantID, instanceID string) ([]DBStatementStat, error) {
	if s == nil || s.conn == nil {
		return nil, fmt.Errorf("span store unavailable")
	}
	if tenantID == "" {
		tenantID = "default"
	}
	rows, err := s.conn.Query(ctx, `
SELECT tags['db.statement'] AS stmt,
       count() AS calls,
       avg(duration_ms) AS avg_ms,
       sum(duration_ms) AS total_ms
FROM trace_spans
WHERE tenant_id = ? AND tags['db.statement'] != ''
  AND (tags['db.name'] = ? OR tags['db.system'] = ?)
  AND start_time >= now() - INTERVAL 60 MINUTE
GROUP BY stmt
ORDER BY total_ms DESC
LIMIT 20`, tenantID, instanceID, instanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]DBStatementStat, 0)
	for rows.Next() {
		var stmt string
		var calls uint64
		var avgMs, totalMs float64
		if err := rows.Scan(&stmt, &calls, &avgMs, &totalMs); err != nil {
			return nil, err
		}
		out = append(out, DBStatementStat{
			Query: stmt, Calls: int64(calls), AvgMs: round1(avgMs), TotalMs: round1(totalMs),
		})
	}
	return out, rows.Err()
}

// CountTraces returns the number of distinct traces ingested since the given time.
func (s *SpanStore) CountTraces(ctx context.Context, tenantID string, since time.Time) (int64, error) {
	if s == nil || s.conn == nil {
		return 0, fmt.Errorf("span store unavailable")
	}
	if tenantID == "" {
		tenantID = "default"
	}
	var count uint64
	if err := s.conn.QueryRow(ctx, `
SELECT uniqExact(trace_id) FROM trace_spans WHERE tenant_id = ? AND start_time >= ?`,
		tenantID, since.UTC()).Scan(&count); err != nil {
		return 0, err
	}
	return int64(count), nil
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func clampScore(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return float64(int(v*10)) / 10
}

func round1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}

// ListOperations returns distinct operations for a service.
func (s *SpanStore) ListOperations(ctx context.Context, tenantID, service string) ([]string, error) {
	if tenantID == "" {
		tenantID = "default"
	}
	rows, err := s.conn.Query(ctx, `
SELECT DISTINCT operation FROM trace_spans
WHERE tenant_id = ? AND service = ?
ORDER BY operation LIMIT 100`, tenantID, service)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ops := make([]string, 0)
	for rows.Next() {
		var op string
		if err := rows.Scan(&op); err != nil {
			return nil, err
		}
		ops = append(ops, op)
	}
	return ops, rows.Err()
}
