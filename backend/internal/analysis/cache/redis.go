package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/neuralops/platform/internal/ai"
	"github.com/neuralops/platform/internal/domain"
	"github.com/redis/go-redis/v9"
)

// RedisCache provides Redis-backed caching for analysis results and stats.
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a Redis cache client.
func NewRedisCache(url string) (*RedisCache, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return &RedisCache{client: client}, nil
}

// Close closes the Redis client.
func (c *RedisCache) Close() error {
	return c.client.Close()
}

func (c *RedisCache) GetClassification(ctx context.Context, key string) (*domain.ErrorClassification, error) {
	raw, err := c.client.Get(ctx, classificationKey(key)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var result domain.ErrorClassification
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *RedisCache) SetClassification(ctx context.Context, key string, value *domain.ErrorClassification, ttlSeconds int) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, classificationKey(key), raw, time.Duration(ttlSeconds)*time.Second).Err()
}

// GetExplanation returns a cached log explanation.
func (c *RedisCache) GetExplanation(ctx context.Context, serviceType, cacheKey string) (*ai.LogExplanation, error) {
	raw, err := c.client.Get(ctx, explanationKey(serviceType, cacheKey)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var explanation ai.LogExplanation
	if err := json.Unmarshal([]byte(raw), &explanation); err != nil {
		return nil, err
	}
	return &explanation, nil
}

// SetExplanation caches a log explanation.
func (c *RedisCache) SetExplanation(ctx context.Context, serviceType, cacheKey string, explanation *ai.LogExplanation, ttl time.Duration) error {
	raw, err := json.Marshal(explanation)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, explanationKey(serviceType, cacheKey), raw, ttl).Err()
}

// TrackException stores recent exception fingerprints for recurrence detection.
func (c *RedisCache) TrackException(ctx context.Context, service, fingerprint string, window int) (count int64, err error) {
	key := exceptionKey(service)
	pipe := c.client.Pipeline()
	pipe.LPush(ctx, key, fingerprint)
	pipe.LTrim(ctx, key, 0, int64(window-1))
	countCmd := pipe.LRange(ctx, key, 0, -1)
	if _, err = pipe.Exec(ctx); err != nil {
		return 0, err
	}
	values, err := countCmd.Result()
	if err != nil {
		return 0, err
	}
	var matches int64
	for _, value := range values {
		if value == fingerprint {
			matches++
		}
	}
	return matches, nil
}

// UpdateServiceErrorRate increments error counters for a service.
func (c *RedisCache) UpdateServiceErrorRate(ctx context.Context, service string, isError bool) error {
	key := serviceStatsKey(service)
	pipe := c.client.Pipeline()
	pipe.HIncrBy(ctx, key, "total", 1)
	if isError {
		pipe.HIncrBy(ctx, key, "errors", 1)
	}
	pipe.Expire(ctx, key, 24*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

// UpdateMetricStats updates rolling metric statistics for anomaly detection.
func (c *RedisCache) UpdateMetricStats(ctx context.Context, service string, metricType domain.MetricType, value float64) (count int64, mean float64, m2 float64, err error) {
	key := metricStatsKey(service, metricType)
	raw, err := c.client.HGetAll(ctx, key).Result()
	if err != nil {
		return 0, 0, 0, err
	}

	n := parseFloat(raw["count"])
	mean = parseFloat(raw["mean"])
	m2 = parseFloat(raw["m2"])

	n++
	delta := value - mean
	mean += delta / n
	delta2 := value - mean
	m2 += delta * delta2

	pipe := c.client.Pipeline()
	pipe.HSet(ctx, key, map[string]interface{}{
		"count": n,
		"mean":  mean,
		"m2":    m2,
		"last":  value,
	})
	pipe.Expire(ctx, key, 7*24*time.Hour)
	if _, err = pipe.Exec(ctx); err != nil {
		return 0, 0, 0, err
	}
	return int64(n), mean, m2, nil
}

// GetMetricStats returns stored metric statistics.
func (c *RedisCache) GetMetricStats(ctx context.Context, service string, metricType domain.MetricType) (count int64, mean, m2, last float64, err error) {
	raw, err := c.client.HGetAll(ctx, metricStatsKey(service, metricType)).Result()
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return int64(parseFloat(raw["count"])), parseFloat(raw["mean"]), parseFloat(raw["m2"]), parseFloat(raw["last"]), nil
}

// SetAnomalyScore stores the latest anomaly score for a service metric.
func (c *RedisCache) SetAnomalyScore(ctx context.Context, service string, metricType domain.MetricType, score float64) error {
	return c.client.Set(ctx, anomalyScoreKey(service, metricType), score, time.Hour).Err()
}

func classificationKey(key string) string  { return "analysis:class:" + key }
func explanationKey(serviceType, key string) string {
	return "analysis:explain:" + serviceType + ":" + key
}
func exceptionKey(service string) string { return "analysis:exceptions:" + service }
func serviceStatsKey(service string) string { return "analysis:stats:" + service }
func metricStatsKey(service string, metricType domain.MetricType) string {
	return fmt.Sprintf("analysis:metric:%s:%s", service, metricType)
}
func anomalyScoreKey(service string, metricType domain.MetricType) string {
	return fmt.Sprintf("analysis:anomaly:%s:%s", service, metricType)
}

func parseFloat(value string) float64 {
	if value == "" {
		return 0
	}
	var out float64
	_, _ = fmt.Sscan(value, &out)
	return out
}
