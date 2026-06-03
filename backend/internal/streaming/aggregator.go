package streaming

import (
	"hash/fnv"
	"sync"
	"time"
)

// bucketKey identifies one aggregation bucket.
type bucketKey struct {
	TenantID string
	ID       string
	Service  string
	Bucket   int64 // unix minute
}

func minuteBucket(t time.Time) int64 {
	return t.UTC().Truncate(time.Minute).Unix()
}

// Aggregator shards in-memory windows for high-throughput ingestion before flush.
type Aggregator struct {
	shards []*aggregatorShard
}

type aggregatorShard struct {
	mu       sync.Mutex
	metrics  map[bucketKey]metricAgg
	alerts   map[bucketKey]alertAgg
}

type metricAgg struct {
	sum   float64
	count int64
}

type alertAgg struct {
	signals int64
	score   float64
}

// NewAggregator creates a sharded aggregator (default 64 shards).
func NewAggregator(shards int) *Aggregator {
	if shards <= 0 {
		shards = 64
	}
	a := &Aggregator{shards: make([]*aggregatorShard, shards)}
	for i := range a.shards {
		a.shards[i] = &aggregatorShard{
			metrics: make(map[bucketKey]metricAgg),
			alerts:  make(map[bucketKey]alertAgg),
		}
	}
	return a
}

func (a *Aggregator) shard(tenantID string) *aggregatorShard {
	h := fnv.New32a()
	_, _ = h.Write([]byte(tenantID))
	return a.shards[h.Sum32()%uint32(len(a.shards))]
}

// RecordMetric adds a metric sample to the current minute bucket.
func (a *Aggregator) RecordMetric(tenantID, metricID, service string, value float64, ts time.Time) {
	k := bucketKey{TenantID: tenantID, ID: metricID, Service: service, Bucket: minuteBucket(ts)}
	sh := a.shard(tenantID)
	sh.mu.Lock()
	defer sh.mu.Unlock()
	agg := sh.metrics[k]
	agg.sum += value
	agg.count++
	sh.metrics[k] = agg
}

// RecordAlertSignal updates alert fatigue score with exponential weighting per signal batch.
func (a *Aggregator) RecordAlertSignal(tenantID, policyID, service string, count int64, ts time.Time) {
	k := bucketKey{TenantID: tenantID, ID: policyID, Service: service, Bucket: minuteBucket(ts)}
	sh := a.shard(tenantID)
	sh.mu.Lock()
	defer sh.mu.Unlock()
	agg := sh.alerts[k]
	agg.signals += count
	// Fatigue: each signal batch adds decay-weighted score (tuned for billions/day via sharding).
	agg.score = agg.score*0.85 + float64(count)*1.15
	sh.alerts[k] = agg
}

// Drain returns and clears all pending aggregates for persistence.
func (a *Aggregator) Drain() (metrics []MetricBucket, alerts []AlertBucket) {
	for _, sh := range a.shards {
		sh.mu.Lock()
		for k, v := range sh.metrics {
			val := v.sum
			if v.count > 0 {
				val = v.sum / float64(v.count)
			}
			metrics = append(metrics, MetricBucket{
				TenantID: k.TenantID, MetricID: k.ID, Service: k.Service,
				BucketTS: time.Unix(k.Bucket, 0).UTC(), Value: val, SampleCount: v.count,
			})
		}
		for k, v := range sh.alerts {
			alerts = append(alerts, AlertBucket{
				TenantID: k.TenantID, PolicyID: k.ID, Service: k.Service,
				BucketTS: time.Unix(k.Bucket, 0).UTC(), SignalCount: v.signals, FatigueScore: v.score,
			})
		}
		sh.metrics = make(map[bucketKey]metricAgg)
		sh.alerts = make(map[bucketKey]alertAgg)
		sh.mu.Unlock()
	}
	return metrics, alerts
}

// MetricBucket is a flushed derived metric aggregate.
type MetricBucket struct {
	TenantID    string
	MetricID    string
	Service     string
	BucketTS    time.Time
	Value       float64
	SampleCount int64
}

// AlertBucket is a flushed alert fatigue aggregate.
type AlertBucket struct {
	TenantID     string
	PolicyID     string
	Service      string
	BucketTS     time.Time
	SignalCount  int64
	FatigueScore float64
}
