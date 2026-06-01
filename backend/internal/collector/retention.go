package collector

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neuralops/platform/internal/apm"
)

// RetentionSync applies ClickHouse TTL from Postgres trace retention policies.
type RetentionSync struct {
	chDSN string
	pool  *pgxpool.Pool
}

// NewRetentionSync creates a retention synchronizer.
func NewRetentionSync(chDSN string, pool *pgxpool.Pool) *RetentionSync {
	return &RetentionSync{chDSN: chDSN, pool: pool}
}

// Sync reads tenant policies and applies MODIFY TTL on trace_spans.
func (r *RetentionSync) Sync(ctx context.Context, tenantID string) error {
	if r.pool == nil || r.chDSN == "" {
		return fmt.Errorf("retention sync unavailable")
	}
	policy := apm.NewPolicyStore(r.pool).Get(ctx, tenantID)
	if policy.RetentionDays <= 0 {
		policy.RetentionDays = 30
	}
	opts, err := clickhouse.ParseDSN(r.chDSN)
	if err != nil {
		return err
	}
	conn, err := clickhouse.Open(opts)
	if err != nil {
		return err
	}
	defer conn.Close()

	ttlExpr := fmt.Sprintf("start_time + INTERVAL %d DAY", policy.RetentionDays)
	if err := conn.Exec(ctx, fmt.Sprintf(`ALTER TABLE trace_spans MODIFY TTL %s`, ttlExpr)); err != nil {
		return fmt.Errorf("apply clickhouse ttl: %w", err)
	}
	return nil
}
