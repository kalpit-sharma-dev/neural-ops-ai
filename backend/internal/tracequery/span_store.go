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
