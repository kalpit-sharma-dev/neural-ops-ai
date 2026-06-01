package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const DefaultVersion = "1.0.0"

// CheckFunc probes a dependency and returns an error when unhealthy.
type CheckFunc func(ctx context.Context) error

// Checker evaluates service health with optional dependency checks.
type Checker struct {
	Service     string
	Environment string
	Version     string
	startTime   time.Time
	checks      map[string]CheckFunc
}

// NewChecker creates a health checker for a service.
func NewChecker(service, environment, version string) *Checker {
	if version == "" {
		version = DefaultVersion
	}
	return &Checker{
		Service:     service,
		Environment: environment,
		Version:     version,
		startTime:   time.Now().UTC(),
		checks:      make(map[string]CheckFunc),
	}
}

// Register adds a named dependency check.
func (c *Checker) Register(name string, fn CheckFunc) {
	if fn == nil {
		return
	}
	c.checks[name] = fn
}

// Result is the structured health response payload.
type Result struct {
	Status        string            `json:"status"`
	Version       string            `json:"version"`
	UptimeSeconds int64             `json:"uptime_seconds"`
	Checks        map[string]string `json:"checks"`
}

// Evaluate runs all dependency checks and returns aggregate health.
func (c *Checker) Evaluate(ctx context.Context) Result {
	checks := make(map[string]string, len(c.checks))
	overall := "healthy"

	for name, fn := range c.checks {
		if err := fn(ctx); err != nil {
			checks[name] = "down"
			overall = "degraded"
			continue
		}
		checks[name] = "healthy"
	}

	return Result{
		Status:        overall,
		Version:       c.Version,
		UptimeSeconds: int64(time.Since(c.startTime).Seconds()),
		Checks:        checks,
	}
}

// GinHandler returns a Gin handler for GET /health style endpoints.
func (c *Checker) GinHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result := c.Evaluate(ctx.Request.Context())
		statusCode := http.StatusOK
		if result.Status == "down" {
			statusCode = http.StatusServiceUnavailable
		}
		ctx.JSON(statusCode, gin.H{
			"status": "success",
			"data": gin.H{
				"service": c.Service,
				"health":  result,
			},
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}

// PingCheck wraps any type exposing Ping(context.Context) error.
func PingCheck(pinger interface{ Ping(context.Context) error }) CheckFunc {
	if pinger == nil {
		return nil
	}
	return func(ctx context.Context) error {
		return pinger.Ping(ctx)
	}
}

// RedisCheck pings a Redis client.
func RedisCheck(client *redis.Client) CheckFunc {
	if client == nil {
		return nil
	}
	return func(ctx context.Context) error {
		return client.Ping(ctx).Err()
	}
}
