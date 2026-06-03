package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/gateway/auth"
	"github.com/neuralops/platform/internal/gateway/config"
	"github.com/redis/go-redis/v9"
)

// quotaWindow is the rolling accounting window for per-tenant quotas. Daily
// windows align tenant plans to billing-style limits while keeping Redis keys
// bounded (one counter pair per tenant per day, expired after 48h).
const quotaWindow = 48 * time.Hour

// TenantQuota enforces per-tenant daily ingest budgets (request count and bytes)
// on the configured ingest path prefixes. It is the multi-tenant fairness
// control that prevents a single noisy tenant from exhausting shared capacity —
// a prerequisite for safe multi-tenant scale.
//
// Counters are stored in Redis so the limit is enforced consistently across all
// horizontally-scaled gateway replicas. When Redis is unavailable the middleware
// fails open (availability over strict enforcement) but logs nothing here to
// avoid hot-path overhead; Redis health is surfaced by its own probes.
func TenantQuota(cfg config.TenantQuotaConfig, client *redis.Client) gin.HandlerFunc {
	if !cfg.Enabled || client == nil {
		return func(c *gin.Context) { c.Next() }
	}
	prefixes := cfg.Paths
	if len(prefixes) == 0 {
		prefixes = defaultQuotaPaths()
	}

	return func(c *gin.Context) {
		if !quotaPathMatches(c.Request.URL.Path, prefixes) {
			c.Next()
			return
		}

		tenantID := "anonymous"
		if principal, ok := auth.PrincipalFromGin(c); ok && principal.TenantID != "" {
			tenantID = principal.TenantID
		}

		day := time.Now().UTC().Format("20060102")
		ctx := c.Request.Context()

		reqKey := fmt.Sprintf("quota:req:%s:%s", tenantID, day)
		reqCount, err := client.Incr(ctx, reqKey).Result()
		if err != nil {
			c.Next() // fail open on Redis error
			return
		}
		if reqCount == 1 {
			_ = client.Expire(ctx, reqKey, quotaWindow).Err()
		}

		var byteCount int64
		if c.Request.ContentLength > 0 {
			bytesKey := fmt.Sprintf("quota:bytes:%s:%s", tenantID, day)
			byteCount, err = client.IncrBy(ctx, bytesKey, c.Request.ContentLength).Result()
			if err == nil && byteCount == c.Request.ContentLength {
				_ = client.Expire(ctx, bytesKey, quotaWindow).Err()
			}
		}

		blocked, code, msg := evaluateQuota(reqCount, byteCount, cfg)
		setQuotaHeaders(c, cfg, reqCount, byteCount)
		if blocked {
			c.Header("Retry-After", retryAfterSeconds())
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"status":    "error",
				"errorCode": code,
				"message":   msg,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
			return
		}
		c.Next()
	}
}

// evaluateQuota is the pure decision function: given the post-increment counters
// and the tenant's configured budgets, it reports whether the request must be
// rejected and with which error code. A zero budget means "unlimited".
func evaluateQuota(reqCount, byteCount int64, cfg config.TenantQuotaConfig) (blocked bool, code, msg string) {
	if cfg.DailyRequestQuota > 0 && reqCount > cfg.DailyRequestQuota {
		return true, "QUOTA001", fmt.Sprintf("daily request quota exceeded (%d/day)", cfg.DailyRequestQuota)
	}
	if cfg.DailyBytesQuota > 0 && byteCount > cfg.DailyBytesQuota {
		return true, "QUOTA002", fmt.Sprintf("daily ingest byte quota exceeded (%d bytes/day)", cfg.DailyBytesQuota)
	}
	return false, "", ""
}

// quotaPathMatches reports whether path falls under any enforced prefix.
func quotaPathMatches(path string, prefixes []string) bool {
	for _, p := range prefixes {
		if p != "" && strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// defaultQuotaPaths is the set of ingest-heavy endpoints quotas guard by default.
func defaultQuotaPaths() []string {
	return []string{
		"/api/v1/events",
		"/api/v1/logs/ingest",
		"/api/v1/logs/bulk",
		"/api/v1/metrics/ingest",
	}
}

func setQuotaHeaders(c *gin.Context, cfg config.TenantQuotaConfig, reqCount, byteCount int64) {
	if cfg.DailyRequestQuota > 0 {
		c.Header("X-Quota-Requests-Limit", fmt.Sprintf("%d", cfg.DailyRequestQuota))
		remaining := cfg.DailyRequestQuota - reqCount
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-Quota-Requests-Remaining", fmt.Sprintf("%d", remaining))
	}
	if cfg.DailyBytesQuota > 0 {
		c.Header("X-Quota-Bytes-Limit", fmt.Sprintf("%d", cfg.DailyBytesQuota))
		remaining := cfg.DailyBytesQuota - byteCount
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-Quota-Bytes-Remaining", fmt.Sprintf("%d", remaining))
	}
}

// retryAfterSeconds returns seconds until the next UTC day boundary.
func retryAfterSeconds() string {
	now := time.Now().UTC()
	next := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour)
	return fmt.Sprintf("%d", int(next.Sub(now).Seconds()))
}
