package audit

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neuralops/platform/internal/finops/domain"
)

// Logger records FinOps mutation audit entries (REQ-FINOPS-072).
type Logger struct {
	mu      sync.RWMutex
	entries map[string][]domain.AuditEntry
	pool    *pgxpool.Pool
}

// NewLogger creates an audit logger.
func NewLogger() *Logger {
	return &Logger{entries: map[string][]domain.AuditEntry{}}
}

// SetPostgres enables durable audit persistence (FIN-PROD-09).
func (l *Logger) SetPostgres(pool *pgxpool.Pool) {
	l.pool = pool
}

// Record appends an audit entry for tenant.
func (l *Logger) Record(ctx context.Context, tenantID, actor, action, entityType, entityID string, detail map[string]string) {
	entry := domain.AuditEntry{
		ID: uuid.NewString(), Actor: actor, Action: action,
		EntityType: entityType, EntityID: entityID, Detail: detail,
		CreatedAt: time.Now().UTC(),
	}
	l.mu.Lock()
	l.entries[tenantID] = append(l.entries[tenantID], entry)
	l.mu.Unlock()
	if l.pool != nil {
		detailJSON, _ := json.Marshal(detail)
		if detailJSON == nil {
			detailJSON = []byte("{}")
		}
		_, _ = l.pool.Exec(ctx, `
INSERT INTO finops_audit_log (id, tenant_id, actor, action, entity_type, entity_id, detail, created_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			entry.ID, tenantID, actor, action, entityType, entityID, detailJSON, entry.CreatedAt)
	}
}

// List returns recent audit entries for tenant.
func (l *Logger) List(tenantID string, limit int) []domain.AuditEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	list := l.entries[tenantID]
	if limit <= 0 || limit > len(list) {
		limit = len(list)
	}
	start := len(list) - limit
	if start < 0 {
		start = 0
	}
	out := make([]domain.AuditEntry, limit)
	copy(out, list[start:])
	// reverse to newest first
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}
