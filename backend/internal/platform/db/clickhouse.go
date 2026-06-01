package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// DefaultClickHouseMigrationPaths returns ClickHouse migration SQL files.
func DefaultClickHouseMigrationPaths() []string {
	candidates := []string{
		"migrations/clickhouse/000001_init.up.sql",
		"migrations/clickhouse/000002_audit_logs.up.sql",
		"migrations/clickhouse/000003_metrics.up.sql",
		"migrations/clickhouse/000004_spans.up.sql",
		"../migrations/clickhouse/000001_init.up.sql",
		"../migrations/clickhouse/000002_audit_logs.up.sql",
		"../migrations/clickhouse/000003_metrics.up.sql",
		"../migrations/clickhouse/000004_spans.up.sql",
		"../../migrations/clickhouse/000001_init.up.sql",
		"../../migrations/clickhouse/000002_audit_logs.up.sql",
		"../../migrations/clickhouse/000003_metrics.up.sql",
		"../../migrations/clickhouse/000004_spans.up.sql",
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

// RunClickHouseMigrations executes ClickHouse DDL migration files.
func RunClickHouseMigrations(ctx context.Context, conn driver.Conn, paths ...string) error {
	if len(paths) == 0 {
		paths = DefaultClickHouseMigrationPaths()
	}
	sort.Strings(paths)
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read clickhouse migration %s: %w", path, err)
		}
		for _, statement := range splitSQLStatements(string(content)) {
			if err := conn.Exec(ctx, statement); err != nil {
				return fmt.Errorf("apply clickhouse migration %s: %w", filepath.Base(path), err)
			}
		}
	}
	return nil
}

func splitSQLStatements(content string) []string {
	parts := strings.Split(content, ";")
	statements := make([]string, 0, len(parts))
	for _, part := range parts {
		stmt := strings.TrimSpace(part)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		statements = append(statements, stmt)
	}
	return statements
}
