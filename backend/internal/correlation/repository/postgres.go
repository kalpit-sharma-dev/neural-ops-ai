package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neuralops/platform/internal/platform/db"
)

// CorrelationLevel indicates deployment correlation strength.
type CorrelationLevel string

const (
	CorrelationHigh   CorrelationLevel = "HIGH"
	CorrelationMedium CorrelationLevel = "MEDIUM"
	CorrelationLow    CorrelationLevel = "LOW"
)

// Store persists correlation artifacts.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore creates a correlation repository.
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

// SaveDeploymentCorrelation stores a deployment correlation record.
func (s *Store) SaveDeploymentCorrelation(ctx context.Context, deploymentID uuid.UUID, service, version string, deployedAt time.Time, level CorrelationLevel, before, after float64) error {
	_, err := s.pool.Exec(ctx, `
INSERT INTO deployment_correlations (id, deployment_id, service, version, deployed_at, correlation_level, error_rate_before, error_rate_after)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		uuid.New(), deploymentID, service, version, deployedAt, level, before, after,
	)
	return err
}

// UpsertDependency upserts a service dependency edge.
func (s *Store) UpsertDependency(ctx context.Context, source, target string, callCount int64, p99LatencyMs float64) error {
	_, err := s.pool.Exec(ctx, `
INSERT INTO service_dependencies (id, source_service, target_service, call_count, p99_latency_ms, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (source_service, target_service)
DO UPDATE SET call_count = service_dependencies.call_count + EXCLUDED.call_count,
              p99_latency_ms = GREATEST(service_dependencies.p99_latency_ms, EXCLUDED.p99_latency_ms),
              updated_at = NOW()`,
		uuid.New(), source, target, callCount, p99LatencyMs,
	)
	return err
}

// ListDependencies returns all dependency edges.
func (s *Store) ListDependencies(ctx context.Context) (map[string][]DependencyEdge, error) {
	rows, err := s.pool.Query(ctx, `
SELECT source_service, target_service, call_count, p99_latency_ms
FROM service_dependencies`)
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
		graph[source] = append(graph[source], DependencyEdge{
			Target:       target,
			CallCount:    callCount,
			P99LatencyMs: p99,
		})
	}
	return graph, rows.Err()
}

// SaveTemporalCorrelation stores a temporal correlation window.
func (s *Store) SaveTemporalCorrelation(ctx context.Context, windowStart, windowEnd time.Time, services []string, hint string, infraEvents []string) error {
	_, err := s.pool.Exec(ctx, `
INSERT INTO temporal_correlations (id, window_start, window_end, affected_services, common_cause_hint, infra_events)
VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(), windowStart, windowEnd, services, hint, infraEvents,
	)
	return err
}

// DependencyEdge represents a directed service dependency.
type DependencyEdge struct {
	Target       string  `json:"target"`
	CallCount    int64   `json:"callCount"`
	P99LatencyMs float64 `json:"p99LatencyMs"`
}

// BlastRadius performs BFS from a failing service over dependency graph.
func BlastRadius(graph map[string][]DependencyEdge, root string) []string {
	visited := map[string]bool{root: true}
	queue := []string{root}
	out := []string{root}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, edge := range graph[current] {
			if visited[edge.Target] {
				continue
			}
			visited[edge.Target] = true
			out = append(out, edge.Target)
			queue = append(queue, edge.Target)
		}
	}
	return out
}
