package streaming

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store persists materialized buckets to Postgres.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore creates a streaming store.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) available() bool {
	return s != nil && s.pool != nil
}

// FlushMetricBuckets upserts derived metric samples.
func (s *Store) FlushMetricBuckets(ctx context.Context, buckets []MetricBucket) error {
	if !s.available() || len(buckets) == 0 {
		return nil
	}
	for _, b := range buckets {
		_, err := s.pool.Exec(ctx, `
INSERT INTO derived_metric_samples (tenant_id, metric_id, service, bucket_ts, value, sample_count)
VALUES ($1,$2,$3,$4,$5,$6)
ON CONFLICT (tenant_id, metric_id, service, bucket_ts) DO UPDATE SET
  value = (derived_metric_samples.value * derived_metric_samples.sample_count + EXCLUDED.value * EXCLUDED.sample_count)
        / (derived_metric_samples.sample_count + EXCLUDED.sample_count),
  sample_count = derived_metric_samples.sample_count + EXCLUDED.sample_count`,
			b.TenantID, b.MetricID, b.Service, b.BucketTS, b.Value, b.SampleCount)
		if err != nil {
			return err
		}
	}
	return nil
}

// FlushAlertBuckets upserts alert score buckets.
func (s *Store) FlushAlertBuckets(ctx context.Context, buckets []AlertBucket) error {
	if !s.available() || len(buckets) == 0 {
		return nil
	}
	for _, b := range buckets {
		_, err := s.pool.Exec(ctx, `
INSERT INTO alert_score_buckets (tenant_id, policy_id, service, bucket_ts, signal_count, fatigue_score)
VALUES ($1,$2,$3,$4,$5,$6)
ON CONFLICT (tenant_id, policy_id, service, bucket_ts) DO UPDATE SET
  signal_count = alert_score_buckets.signal_count + EXCLUDED.signal_count,
  fatigue_score = GREATEST(alert_score_buckets.fatigue_score, EXCLUDED.fatigue_score)`,
			b.TenantID, b.PolicyID, b.Service, b.BucketTS, b.SignalCount, b.FatigueScore)
		if err != nil {
			return err
		}
	}
	return nil
}
