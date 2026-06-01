package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/gateway/auth"
	"github.com/neuralops/platform/internal/gateway/config"
	"github.com/neuralops/platform/internal/gateway/publicpaths"
	"github.com/redis/go-redis/v9"
)

// RateLimit applies a Redis sliding-window limiter per tenant.
func RateLimit(cfg config.RateLimitConfig, client *redis.Client) gin.HandlerFunc {
	if !cfg.Enabled || client == nil {
		return func(c *gin.Context) { c.Next() }
	}

	window := time.Minute
	limit := int64(cfg.RequestsPerMinute)
	if limit <= 0 {
		limit = 600
	}

	return func(c *gin.Context) {
		if publicpaths.IsPublic(c.Request.URL.Path) {
			c.Next()
			return
		}

		principal, ok := auth.PrincipalFromGin(c)
		tenantID := "anonymous"
		if ok && principal.TenantID != "" {
			tenantID = principal.TenantID
		}

		now := time.Now().UTC()
		key := fmt.Sprintf("ratelimit:%s:%d", tenantID, now.Unix()/int64(window.Seconds()))
		ctx := c.Request.Context()

		count, err := client.Incr(ctx, key).Result()
		if err != nil {
			c.Next()
			return
		}
		if count == 1 {
			_ = client.Expire(ctx, key, window+time.Second).Err()
		}
		if count > limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"status": "error", "errorCode": "RATE001", "message": "rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}

// NewRedisClient creates a redis client when configured.
func NewRedisClient(url string) (*redis.Client, error) {
	if url == "" {
		return nil, nil
	}
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
