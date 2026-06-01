package anomaly

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"github.com/neuralops/platform/internal/domain"
)

// MetricStatsReader reads rolling metric statistics.
type MetricStatsReader interface {
	UpdateMetricStats(ctx context.Context, service string, metricType domain.MetricType, value float64) (count int64, mean float64, m2 float64, err error)
	GetMetricStats(ctx context.Context, service string, metricType domain.MetricType) (count int64, mean, m2, last float64, err error)
	SetAnomalyScore(ctx context.Context, service string, metricType domain.MetricType, score float64) error
}

// Publisher publishes anomaly events.
type Publisher interface {
	PublishAnomaly(ctx context.Context, anomaly domain.AnomalyDetection) error
}

// Detector performs z-score and EWMA based anomaly detection.
type Detector struct {
	zThreshold float64
	alpha      float64
	stats      MetricStatsReader
	publisher  Publisher
}

// NewDetector creates an anomaly detector.
func NewDetector(zThreshold float64, stats MetricStatsReader, publisher Publisher) *Detector {
	if zThreshold <= 0 {
		zThreshold = 3
	}
	return &Detector{
		zThreshold: zThreshold,
		alpha:      0.2,
		stats:      stats,
		publisher:  publisher,
	}
}

// AnalyzeMetric evaluates a metric sample for anomalies.
func (d *Detector) AnalyzeMetric(ctx context.Context, metric domain.Metric) (*domain.AnomalyDetection, error) {
	count, mean, m2, err := d.stats.UpdateMetricStats(ctx, metric.ServiceName, metric.MetricType, metric.Value)
	if err != nil {
		return nil, err
	}

	if count < 5 {
		return &domain.AnomalyDetection{
			ServiceName: metric.ServiceName,
			MetricType:  metric.MetricType,
			Score:       0,
			Detected:    false,
			Baseline:    mean,
			Actual:      metric.Value,
			DetectedAt:  time.Now().UTC(),
			AnomalyType: domain.AnomalyTypeTrend,
		}, nil
	}

	variance := m2 / float64(count)
	if variance < 0 {
		variance = 0
	}
	stddev := math.Sqrt(variance)
	zScore := 0.0
	if stddev > 0 {
		zScore = (metric.Value - mean) / stddev
	}

	ewma := mean
	if _, _, _, last, err := d.stats.GetMetricStats(ctx, metric.ServiceName, metric.MetricType); err == nil && last != 0 {
		ewma = d.alpha*metric.Value + (1-d.alpha)*last
	}

	score := math.Min(1, math.Abs(zScore)/d.zThreshold)
	detected := math.Abs(zScore) > d.zThreshold

	anomalyType := domain.AnomalyTypeTrend
	switch {
	case zScore > d.zThreshold:
		anomalyType = domain.AnomalyTypeSpike
	case zScore < -d.zThreshold:
		anomalyType = domain.AnomalyTypeDrop
	}

	result := &domain.AnomalyDetection{
		ServiceName: metric.ServiceName,
		MetricType:  metric.MetricType,
		Score:       score,
		Detected:    detected,
		Baseline:    ewma,
		Actual:      metric.Value,
		DetectedAt:  time.Now().UTC(),
		AnomalyType: anomalyType,
	}

	_ = d.stats.SetAnomalyScore(ctx, metric.ServiceName, metric.MetricType, score)

	if detected && d.publisher != nil {
		if err := d.publisher.PublishAnomaly(ctx, *result); err != nil {
			return result, err
		}
	}

	return result, nil
}

// AnomalyEventJSON serializes anomaly events for Kafka.
func AnomalyEventJSON(anomaly domain.AnomalyDetection) ([]byte, error) {
	return json.Marshal(anomaly)
}
