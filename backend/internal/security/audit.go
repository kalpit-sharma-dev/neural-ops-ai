package security

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuditEntry represents an immutable audit log record.
type AuditEntry struct {
	UserID       string
	TenantID     string
	Action       string
	ResourceType string
	ResourceID   string
	IP           string
	UserAgent    string
	Result       string
}

// AuditRepository persists insert-only audit logs to Postgres and optionally ClickHouse.
type AuditRepository struct {
	pool *pgxpool.Pool
	ch   driver.Conn
}

// NewAuditRepository creates an audit log repository.
func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

// WithClickHouse enables dual-write replication to ClickHouse for analytics.
func (r *AuditRepository) WithClickHouse(conn driver.Conn) *AuditRepository {
	if r == nil {
		return nil
	}
	r.ch = conn
	return r
}

// Insert writes an audit log entry. Audit logs are insert-only.
func (r *AuditRepository) Insert(ctx context.Context, entry AuditEntry) error {
	if r == nil || r.pool == nil {
		return nil
	}
	if entry.TenantID == "" {
		entry.TenantID = "default"
	}
	if entry.Action == "" {
		return fmt.Errorf("audit action is required")
	}
	if entry.Result == "" {
		entry.Result = "success"
	}

	createdAt := time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
INSERT INTO audit_logs (user_id, tenant_id, action, resource_type, resource_id, ip, user_agent, result, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		nullIfEmpty(entry.UserID),
		entry.TenantID,
		entry.Action,
		nullIfEmpty(entry.ResourceType),
		nullIfEmpty(entry.ResourceID),
		nullIfEmpty(entry.IP),
		nullIfEmpty(entry.UserAgent),
		entry.Result,
		createdAt,
	)
	if err != nil {
		return err
	}

	if r.ch != nil {
		_ = r.insertClickHouse(ctx, entry, createdAt)
	}
	return nil
}

func (r *AuditRepository) insertClickHouse(ctx context.Context, entry AuditEntry, createdAt time.Time) error {
	return r.ch.Exec(ctx, `
INSERT INTO audit_logs (
	tenant_id, user_id, action, resource_type, resource_id, ip, user_agent, result, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.TenantID,
		entry.UserID,
		entry.Action,
		entry.ResourceType,
		entry.ResourceID,
		entry.IP,
		entry.UserAgent,
		entry.Result,
		createdAt,
	)
}

func nullIfEmpty(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
