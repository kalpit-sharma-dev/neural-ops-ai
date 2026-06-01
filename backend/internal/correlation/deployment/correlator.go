package deployment

import (
	"context"
	"sync"
	"time"

	"github.com/neuralops/platform/internal/correlation/repository"
	"github.com/neuralops/platform/internal/domain"
	"github.com/neuralops/platform/internal/platform/db"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type watch struct {
	deployment       domain.Deployment
	baselineErrorRate float64
	errorCount       int64
	totalCount       int64
	expiresAt        time.Time
}

// Correlator correlates deployments with post-deploy error spikes.
type Correlator struct {
	log         *zap.Logger
	repo        *repository.Store
	redis       *redis.Client
	watchWindow time.Duration
	evalWindow  time.Duration
	mu          sync.Mutex
	watches     []watch
}

// NewCorrelator creates a deployment correlator.
func NewCorrelator(log *zap.Logger, repo *repository.Store, redisClient *redis.Client, watchWindow, evalWindow time.Duration) *Correlator {
	return &Correlator{
		log:         log,
		repo:        repo,
		redis:       redisClient,
		watchWindow: watchWindow,
		evalWindow:  evalWindow,
		watches:     make([]watch, 0),
	}
}

// HandleDeployment registers a deployment for correlation monitoring.
func (c *Correlator) HandleDeployment(ctx context.Context, deployment domain.Deployment) {
	baseline := c.currentErrorRate(ctx, deployment.Service)
	c.mu.Lock()
	c.watches = append(c.watches, watch{
		deployment:        deployment,
		baselineErrorRate: baseline,
		expiresAt:         deployment.DeployedAt.Add(c.watchWindow),
	})
	c.mu.Unlock()
	c.log.Info("deployment watch started",
		zap.String("service", deployment.Service),
		zap.String("version", deployment.Version),
		zap.Float64("baseline_error_rate", baseline),
	)
}

// HandleLog updates post-deployment error counters.
func (c *Correlator) HandleLog(entry domain.LogEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UTC()
	for i := range c.watches {
		if now.After(c.watches[i].expiresAt) {
			continue
		}
		if c.watches[i].deployment.Service != entry.Service {
			continue
		}
		c.watches[i].totalCount++
		if entry.IsErrorSeverity() {
			c.watches[i].errorCount++
		}
	}
}

// Evaluate checks active watches and persists correlations.
func (c *Correlator) Evaluate(ctx context.Context) {
	c.mu.Lock()
	active := make([]watch, 0, len(c.watches))
	now := time.Now().UTC()
	for _, item := range c.watches {
		if now.Before(item.expiresAt) {
			active = append(active, item)
			continue
		}
		c.persistWatch(ctx, item)
	}
	c.watches = active
	c.mu.Unlock()
}

func (c *Correlator) persistWatch(ctx context.Context, item watch) {
	if item.totalCount == 0 {
		return
	}
	after := float64(item.errorCount) / float64(item.totalCount)
	before := item.baselineErrorRate
	if before <= 0 {
		before = 0.01
	}
	ratio := after / before

	level := repository.CorrelationLow
	switch {
	case ratio > 2.0 && time.Since(item.deployment.DeployedAt) <= c.evalWindow:
		level = repository.CorrelationHigh
	case ratio >= 1.5:
		level = repository.CorrelationMedium
	default:
		return
	}

	if err := c.repo.SaveDeploymentCorrelation(ctx, item.deployment.ID, item.deployment.Service, item.deployment.Version, item.deployment.DeployedAt, level, before, after); err != nil {
		c.log.Warn("save deployment correlation failed", zap.Error(err))
		return
	}
	c.log.Info("deployment correlation detected",
		zap.String("service", item.deployment.Service),
		zap.String("level", string(level)),
		zap.Float64("ratio", ratio),
	)
}

func (c *Correlator) currentErrorRate(ctx context.Context, service string) float64 {
	if c.redis == nil {
		return 0
	}
	values, err := c.redis.HGetAll(ctx, "analysis:stats:"+service).Result()
	if err != nil {
		return 0
	}
	return db.ParseRedisErrorRate(values["total"], values["errors"])
}

// StartEvaluator runs periodic deployment correlation evaluation.
func (c *Correlator) StartEvaluator(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.Evaluate(ctx)
			}
		}
	}()
}
