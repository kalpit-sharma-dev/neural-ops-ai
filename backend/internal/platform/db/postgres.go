package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates a PostgreSQL connection pool.
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}

const migrationAdvisoryLockKey int64 = 83927492834

// RunMigrations executes SQL migration files in order, skipping versions already
// recorded in schema_migrations. An advisory lock serializes concurrent runners
// (gateway, incident, alerting, etc.) so CREATE TABLE IF NOT EXISTS races cannot
// corrupt pg_type.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool, paths ...string) error {
	if _, err := pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`); err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}

	if _, err := pool.Exec(ctx, `SELECT pg_advisory_lock($1)`, migrationAdvisoryLockKey); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer func() {
		_, _ = pool.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, migrationAdvisoryLockKey)
	}()

	for _, path := range paths {
		version := filepath.Base(path)
		var applied bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, version).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %s: %w", version, err)
		}
		if applied {
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", path, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", version, err)
		}
		if _, err := tx.Exec(ctx, string(content)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply migration %s: %w", version, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", version, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", version, err)
		}
	}
	return nil
}

// DefaultMigrationPaths returns standard migration file paths.
func DefaultMigrationPaths() []string {
	candidates := []string{
		"migrations/000001_analysis.up.sql",
		"migrations/000002_correlation_incident.up.sql",
		"migrations/000003_alerting.up.sql",
		"migrations/000004_create_tenants.up.sql",
		"migrations/000005_extend_incidents.up.sql",
		"migrations/000006_create_deployments.up.sql",
		"migrations/000007_create_service_graph.up.sql",
		"migrations/000008_create_error_patterns.up.sql",
		"migrations/000009_phase8_indexes.up.sql",
		"migrations/000010_security.up.sql",
		"migrations/000011_auth_sessions.up.sql",
		"migrations/000012_auth_exchange_codes.up.sql",
		"migrations/000013_observability_ui.up.sql",
		"migrations/000014_collectors.up.sql",
		"migrations/000015_depth_features.up.sql",
		"migrations/000016_p0_p3_features.up.sql",
		"migrations/000017_depth_parity.up.sql",
		"../migrations/000001_analysis.up.sql",
		"../migrations/000002_correlation_incident.up.sql",
		"../migrations/000003_alerting.up.sql",
		"../migrations/000004_create_tenants.up.sql",
		"../migrations/000005_extend_incidents.up.sql",
		"../migrations/000006_create_deployments.up.sql",
		"../migrations/000007_create_service_graph.up.sql",
		"../migrations/000008_create_error_patterns.up.sql",
		"../migrations/000009_phase8_indexes.up.sql",
		"../migrations/000010_security.up.sql",
		"../migrations/000011_auth_sessions.up.sql",
		"../migrations/000012_auth_exchange_codes.up.sql",
		"../migrations/000013_observability_ui.up.sql",
		"../migrations/000014_collectors.up.sql",
		"../migrations/000015_depth_features.up.sql",
		"../migrations/000016_p0_p3_features.up.sql",
		"../migrations/000017_depth_parity.up.sql",
		"../../migrations/000001_analysis.up.sql",
		"../../migrations/000002_correlation_incident.up.sql",
		"../../migrations/000003_alerting.up.sql",
		"../../migrations/000004_create_tenants.up.sql",
		"../../migrations/000005_extend_incidents.up.sql",
		"../../migrations/000006_create_deployments.up.sql",
		"../../migrations/000007_create_service_graph.up.sql",
		"../../migrations/000008_create_error_patterns.up.sql",
		"../../migrations/000009_phase8_indexes.up.sql",
		"../../migrations/000010_security.up.sql",
		"../../migrations/000011_auth_sessions.up.sql",
		"../../migrations/000012_auth_exchange_codes.up.sql",
		"../../migrations/000013_observability_ui.up.sql",
		"../../migrations/000014_collectors.up.sql",
		"../../migrations/000015_depth_features.up.sql",
		"../../migrations/000016_p0_p3_features.up.sql",
		"../../migrations/000017_depth_parity.up.sql",
	}
	var paths []string
	seen := map[string]bool{}
	for _, candidate := range candidates {
		if seen[candidate] {
			continue
		}
		if _, err := os.Stat(candidate); err == nil {
			paths = append(paths, candidate)
			seen[candidate] = true
		}
	}
	return dedupeByBase(paths)
}

func dedupeByBase(paths []string) []string {
	byBase := map[string]string{}
	order := []string{}
	for _, path := range paths {
		base := filepath.Base(path)
		if _, ok := byBase[base]; !ok {
			order = append(order, base)
		}
		byBase[base] = path
	}
	out := make([]string, 0, len(order))
	for _, base := range order {
		out = append(out, byBase[base])
	}
	return out
}

// ParseRedisErrorRate computes error rate from redis hash values.
func ParseRedisErrorRate(totalStr, errorsStr string) float64 {
	var total, errors float64
	_, _ = fmt.Sscan(totalStr, &total)
	_, _ = fmt.Sscan(errorsStr, &errors)
	if total <= 0 {
		return 0
	}
	return errors / total
}

// PaymentServices lists critical banking services.
func PaymentServices() []string {
	return []string{
		"upi-service", "payment-api", "ledger-service", "cbs-adapter",
		"auth-service", "api-gateway",
	}
}

func IsPaymentService(service string) bool {
	service = strings.ToLower(service)
	for _, candidate := range PaymentServices() {
		if service == candidate {
			return true
		}
	}
	return strings.Contains(service, "payment") || strings.Contains(service, "upi")
}
