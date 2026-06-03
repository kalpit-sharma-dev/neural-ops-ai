package observability

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DerivedMetricSample is a materialized derived metric bucket.
type DerivedMetricSample struct {
	MetricID    string    `json:"metricId"`
	Service     string    `json:"service"`
	BucketTS    time.Time `json:"bucketTs"`
	Value       float64   `json:"value"`
	SampleCount int64     `json:"sampleCount"`
}

// AlertScoreBucket is materialized alert fatigue for a policy/service.
type AlertScoreBucket struct {
	PolicyID     string    `json:"policyId"`
	Service      string    `json:"service"`
	BucketTS     time.Time `json:"bucketTs"`
	SignalCount  int64     `json:"signalCount"`
	FatigueScore float64   `json:"fatigueScore"`
}

func (h *Handler) ListDerivedMetricSamples(c *gin.Context) {
	metricID := c.Param("id")
	if h.deps.Pool == nil {
		writeSuccess(c, []DerivedMetricSample{})
		return
	}
	rows, err := h.deps.Pool.Query(c.Request.Context(), `
SELECT metric_id, service, bucket_ts, value, sample_count
FROM derived_metric_samples
WHERE tenant_id = $1 AND metric_id = $2
ORDER BY bucket_ts DESC LIMIT 120`, tenantID(c), metricID)
	if err != nil {
		writeError(c, 500, err.Error())
		return
	}
	defer rows.Close()
	var out []DerivedMetricSample
	for rows.Next() {
		var s DerivedMetricSample
		if err := rows.Scan(&s.MetricID, &s.Service, &s.BucketTS, &s.Value, &s.SampleCount); err != nil {
			writeError(c, 500, err.Error())
			return
		}
		out = append(out, s)
	}
	writeSuccess(c, out)
}

func (h *Handler) ListAlertPolicyScores(c *gin.Context) {
	policyID := c.Param("id")
	if h.deps.Pool == nil {
		writeSuccess(c, []AlertScoreBucket{})
		return
	}
	out, err := listAlertScores(c.Request.Context(), h.deps.Pool, tenantID(c), policyID)
	if err != nil {
		writeError(c, 500, err.Error())
		return
	}
	writeSuccess(c, out)
}

func listAlertScores(ctx context.Context, pool *pgxpool.Pool, tenantID, policyID string) ([]AlertScoreBucket, error) {
	rows, err := pool.Query(ctx, `
SELECT policy_id, service, bucket_ts, signal_count, fatigue_score
FROM alert_score_buckets
WHERE tenant_id = $1 AND policy_id = $2
ORDER BY bucket_ts DESC LIMIT 120`, tenantID, policyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AlertScoreBucket
	for rows.Next() {
		var b AlertScoreBucket
		if err := rows.Scan(&b.PolicyID, &b.Service, &b.BucketTS, &b.SignalCount, &b.FatigueScore); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}
