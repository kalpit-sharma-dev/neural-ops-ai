package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neuralops/platform/internal/domain"
	"github.com/neuralops/platform/internal/platform/db"
)

// IncidentFilter filters incident list queries.
type IncidentFilter struct {
	TenantID  string
	Status    string
	Severity  string
	Service   string
	StartTime *time.Time
	EndTime   *time.Time
	Limit     int
	Offset    int
}

// Store persists incidents and related data.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore creates an incident repository.
func NewStore(ctx context.Context, dsn string) (*Store, error) {
	pool, err := db.NewPool(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := db.RunMigrations(ctx, pool, db.DefaultMigrationPaths()...); err != nil {
		pool.Close()
		return nil, err
	}
	return &Store{pool: pool}, nil
}

// Close closes the pool.
func (s *Store) Close() { s.pool.Close() }

// Pool exposes the underlying connection pool for health checks.
func (s *Store) Pool() *pgxpool.Pool { return s.pool }

// CreateIncident inserts a new incident.
func (s *Store) CreateIncident(ctx context.Context, incident *domain.Incident, fingerprint, rootService, errorCategory, tenantID string) error {
	timelineJSON, _ := json.Marshal(incident.Timeline)
	recsJSON, _ := json.Marshal(incident.Recommendations)
	rcaJSON, _ := json.Marshal(incident.RootCauseAnalysis)

	_, err := s.pool.Exec(ctx, `
INSERT INTO incidents (
  id, title, summary, severity, status, affected_services, root_service, error_category,
  fingerprint, blast_radius, start_time, timeline, recommendations, root_cause_analysis, tenant_id, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,NOW(),NOW())`,
		incident.ID, incident.Title, incident.Summary, incident.Severity, incident.Status,
		incident.AffectedServices, rootService, errorCategory, fingerprint, incident.BlastRadius,
		incident.StartTime, timelineJSON, recsJSON, rcaJSON, tenantID,
	)
	return err
}

// UpdateIncident updates an existing incident.
func (s *Store) UpdateIncident(ctx context.Context, incident *domain.Incident) error {
	timelineJSON, _ := json.Marshal(incident.Timeline)
	recsJSON, _ := json.Marshal(incident.Recommendations)
	rcaJSON, _ := json.Marshal(incident.RootCauseAnalysis)
	var mttr int64
	if incident.MTTR > 0 {
		mttr = int64(incident.MTTR)
	}

	_, err := s.pool.Exec(ctx, `
UPDATE incidents SET
  title=$2, summary=$3, severity=$4, status=$5, affected_services=$6, blast_radius=$7,
  resolved_time=$8, acknowledged_at=$9, mttr_ns=$10, timeline=$11, recommendations=$12,
  root_cause_analysis=$13, updated_at=NOW()
WHERE id=$1`,
		incident.ID, incident.Title, incident.Summary, incident.Severity, incident.Status,
		incident.AffectedServices, incident.BlastRadius, incident.ResolvedTime, incident.AcknowledgedAt,
		mttr, timelineJSON, recsJSON, rcaJSON,
	)
	return err
}

// FindActiveByFingerprint finds an open incident by fingerprint.
func (s *Store) FindActiveByFingerprint(ctx context.Context, fingerprint string) (*domain.Incident, error) {
	row := s.pool.QueryRow(ctx, `
SELECT id, title, summary, severity, status, affected_services, blast_radius, start_time,
       resolved_time, acknowledged_at, mttr_ns, timeline, recommendations, root_cause_analysis, created_at, updated_at
FROM incidents
WHERE fingerprint=$1 AND status IN ('OPEN','INVESTIGATING')
ORDER BY created_at DESC LIMIT 1`, fingerprint)

	return scanIncident(row)
}

// GetIncident returns an incident by ID within a tenant.
func (s *Store) GetIncident(ctx context.Context, tenantID string, id uuid.UUID) (*domain.Incident, error) {
	row := s.pool.QueryRow(ctx, `
SELECT id, title, summary, severity, status, affected_services, blast_radius, start_time,
       resolved_time, acknowledged_at, mttr_ns, timeline, recommendations, root_cause_analysis, created_at, updated_at
FROM incidents WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	return scanIncident(row)
}

// ListIncidents returns incidents matching filters.
func (s *Store) ListIncidents(ctx context.Context, filter IncidentFilter) ([]domain.Incident, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	query := `
SELECT id, title, summary, severity, status, affected_services, blast_radius, start_time,
       resolved_time, acknowledged_at, mttr_ns, timeline, recommendations, root_cause_analysis, created_at, updated_at
FROM incidents WHERE tenant_id=$1`
	args := []any{filter.TenantID}
	argIdx := 2

	if filter.TenantID == "" {
		query = `
SELECT id, title, summary, severity, status, affected_services, blast_radius, start_time,
       resolved_time, acknowledged_at, mttr_ns, timeline, recommendations, root_cause_analysis, created_at, updated_at
FROM incidents WHERE 1=1`
		args = []any{}
		argIdx = 1
	}

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status=$%d", argIdx)
		args = append(args, strings.ToUpper(filter.Status))
		argIdx++
	}
	if filter.Severity != "" {
		query += fmt.Sprintf(" AND severity=$%d", argIdx)
		args = append(args, strings.ToUpper(filter.Severity))
		argIdx++
	}
	if filter.Service != "" {
		query += fmt.Sprintf(" AND $%d = ANY(affected_services)", argIdx)
		args = append(args, filter.Service)
		argIdx++
	}
	if filter.StartTime != nil {
		query += fmt.Sprintf(" AND start_time >= $%d", argIdx)
		args = append(args, *filter.StartTime)
		argIdx++
	}
	if filter.EndTime != nil {
		query += fmt.Sprintf(" AND start_time <= $%d", argIdx)
		args = append(args, *filter.EndTime)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY start_time DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	incidents := make([]domain.Incident, 0)
	for rows.Next() {
		incident, err := scanIncident(rows)
		if err != nil {
			return nil, err
		}
		incidents = append(incidents, *incident)
	}
	return incidents, rows.Err()
}

// AcknowledgeIncident marks an incident acknowledged.
func (s *Store) AcknowledgeIncident(ctx context.Context, tenantID string, id uuid.UUID, at time.Time) error {
	_, err := s.pool.Exec(ctx, `
UPDATE incidents SET status='INVESTIGATING', acknowledged_at=$3, updated_at=NOW()
WHERE id=$1 AND tenant_id=$2`, id, tenantID, at)
	return err
}

// ResolveIncident marks an incident resolved.
func (s *Store) ResolveIncident(ctx context.Context, tenantID string, id uuid.UUID, resolvedAt time.Time, notes string, mttr time.Duration) error {
	_, err := s.pool.Exec(ctx, `
UPDATE incidents SET status='RESOLVED', resolved_time=$3, resolution_notes=$4, mttr_ns=$5, updated_at=NOW()
WHERE id=$1 AND tenant_id=$2`, id, tenantID, resolvedAt, notes, int64(mttr))
	return err
}

// SaveMTTRHistory stores MTTR history for trending.
func (s *Store) SaveMTTRHistory(ctx context.Context, incidentID uuid.UUID, service, team string, mttr time.Duration, periodStart, periodEnd time.Time) error {
	_, err := s.pool.Exec(ctx, `
INSERT INTO mttr_history (id, incident_id, service, team, mttr_ns, period_start, period_end)
VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		uuid.New(), incidentID, service, team, int64(mttr), periodStart, periodEnd,
	)
	return err
}

// IsSuppressed checks tenant suppression rules.
func (s *Store) IsSuppressed(ctx context.Context, tenantID, service string, category domain.ErrorCategory) (bool, error) {
	row := s.pool.QueryRow(ctx, `
SELECT COUNT(*) FROM incident_suppression_rules
WHERE tenant_id=$1 AND enabled=TRUE
  AND (service_pattern IS NULL OR $2 LIKE service_pattern)
  AND (error_category IS NULL OR error_category=$3)`, tenantID, service, category)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListDependencies returns service dependency graph for a tenant.
func (s *Store) ListDependencies(ctx context.Context, tenantID string) (map[string][]DependencyEdge, error) {
	rows, err := s.pool.Query(ctx, `
SELECT source_service, target_service, call_count, p99_latency_ms
FROM service_dependencies WHERE tenant_id = $1`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	graph := make(map[string][]DependencyEdge)
	for rows.Next() {
		var source, target string
		var callCount int64
		var p99 float64
		if err := rows.Scan(&source, &target, &callCount, &p99); err != nil {
			return nil, err
		}
		graph[source] = append(graph[source], DependencyEdge{Target: target, CallCount: callCount, P99LatencyMs: p99})
	}
	return graph, rows.Err()
}

// DependencyEdge represents a service dependency edge.
type DependencyEdge struct {
	Target       string  `json:"target"`
	CallCount    int64   `json:"callCount"`
	P99LatencyMs float64 `json:"p99LatencyMs"`
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanIncident(row rowScanner) (*domain.Incident, error) {
	var incident domain.Incident
	var timelineJSON, recsJSON, rcaJSON []byte
	var mttrNs int64
	var acknowledgedAt *time.Time

	err := row.Scan(
		&incident.ID, &incident.Title, &incident.Summary, &incident.Severity, &incident.Status,
		&incident.AffectedServices, &incident.BlastRadius, &incident.StartTime,
		&incident.ResolvedTime, &acknowledgedAt, &mttrNs, &timelineJSON, &recsJSON, &rcaJSON,
		&incident.CreatedAt, &incident.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.NewNotFoundError("incident", "")
		}
		return nil, err
	}
	if acknowledgedAt != nil {
		incident.AcknowledgedAt = acknowledgedAt
	}
	if mttrNs > 0 {
		incident.MTTR = time.Duration(mttrNs)
	}
	if len(timelineJSON) > 0 {
		_ = json.Unmarshal(timelineJSON, &incident.Timeline)
	}
	if len(recsJSON) > 0 {
		_ = json.Unmarshal(recsJSON, &incident.Recommendations)
	}
	if len(rcaJSON) > 0 {
		var rca domain.RootCause
		if err := json.Unmarshal(rcaJSON, &rca); err == nil {
			incident.RootCauseAnalysis = &rca
		}
	}
	return &incident, nil
}
