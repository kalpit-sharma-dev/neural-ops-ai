package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/gateway/config"
	"github.com/redis/go-redis/v9"
)

var authRateLimitPaths = map[string]struct{}{
	"/api/v1/auth/oidc/start":    {},
	"/api/v1/auth/oidc/exchange": {},
	"/api/v1/auth/refresh":       {},
	"/api/v1/auth/dev/login":     {},
}

// AuthRateLimit applies stricter IP-based limits on authentication endpoints.
func AuthRateLimit(cfg config.AuthRateLimitConfig, client *redis.Client) gin.HandlerFunc {
	if !cfg.Enabled || client == nil {
		return func(c *gin.Context) { c.Next() }
	}

	window := time.Minute
	limit := int64(cfg.RequestsPerMinute)
	if limit <= 0 {
		limit = 20
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if _, ok := authRateLimitPaths[path]; !ok {
			c.Next()
			return
		}

		clientIP := strings.TrimSpace(c.ClientIP())
		if clientIP == "" {
			clientIP = "unknown"
		}

		now := time.Now().UTC()
		key := fmt.Sprintf("auth-ratelimit:%s:%s:%d", path, clientIP, now.Unix()/int64(window.Seconds()))
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
				"status": "error", "errorCode": "RATE002", "message": "authentication rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}
