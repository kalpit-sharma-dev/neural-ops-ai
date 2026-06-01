package apm

import (
	"context"
	"hash/fnv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RetentionPolicy controls trace sampling and retention.
type RetentionPolicy struct {
	TenantID        string    `json:"tenantId"`
	RetentionDays   int       `json:"retentionDays"`
	HeadSampleRate  float64   `json:"headSampleRate"`
	TailSampleRate  float64   `json:"tailSampleRate"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// PolicyStore reads tenant trace policies with in-memory cache.
type PolicyStore struct {
	pool  *pgxpool.Pool
	cache sync.Map
}

// NewPolicyStore creates a policy store.
func NewPolicyStore(pool *pgxpool.Pool) *PolicyStore {
	return &PolicyStore{pool: pool}
}

// Get returns policy for tenant (defaults if missing).
func (s *PolicyStore) Get(ctx context.Context, tenantID string) RetentionPolicy {
	if tenantID == "" {
		tenantID = "default"
	}
	if v, ok := s.cache.Load(tenantID); ok {
		return v.(RetentionPolicy)
	}
	p := RetentionPolicy{
		TenantID:       tenantID,
		RetentionDays:  30,
		HeadSampleRate: 1.0,
		TailSampleRate: 1.0,
		UpdatedAt:      time.Now().UTC(),
	}
	if s.pool != nil {
		row := s.pool.QueryRow(ctx, `
SELECT tenant_id, retention_days, head_sample_rate, tail_sample_rate, updated_at
FROM trace_retention_policies WHERE tenant_id = $1`, tenantID)
		_ = row.Scan(&p.TenantID, &p.RetentionDays, &p.HeadSampleRate, &p.TailSampleRate, &p.UpdatedAt)
	}
	s.cache.Store(tenantID, p)
	return p
}

// Save upserts a retention policy.
func (s *PolicyStore) Save(ctx context.Context, p RetentionPolicy) error {
	if s.pool == nil {
		return nil
	}
	if p.TenantID == "" {
		p.TenantID = "default"
	}
	if p.RetentionDays <= 0 {
		p.RetentionDays = 30
	}
	if p.HeadSampleRate <= 0 {
		p.HeadSampleRate = 1.0
	}
	if p.TailSampleRate <= 0 {
		p.TailSampleRate = 1.0
	}
	_, err := s.pool.Exec(ctx, `
INSERT INTO trace_retention_policies (tenant_id, retention_days, head_sample_rate, tail_sample_rate, updated_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (tenant_id) DO UPDATE SET
  retention_days = EXCLUDED.retention_days,
  head_sample_rate = EXCLUDED.head_sample_rate,
  tail_sample_rate = EXCLUDED.tail_sample_rate,
  updated_at = NOW()`,
		p.TenantID, p.RetentionDays, p.HeadSampleRate, p.TailSampleRate,
	)
	if err == nil {
		p.UpdatedAt = time.Now().UTC()
		s.cache.Store(p.TenantID, p)
	}
	return err
}

// ShouldSample decides whether to keep a trace span batch.
func ShouldSample(traceID string, isError bool, rate float64) bool {
	if rate >= 1.0 {
		return true
	}
	if rate <= 0 {
		return false
	}
	if isError && rate < 1.0 {
		// Always keep error traces when tail sampling enabled.
		return true
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(traceID))
	bucket := float64(h.Sum32()%10000) / 10000.0
	return bucket < rate
}
