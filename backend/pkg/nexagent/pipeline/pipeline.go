// Package pipeline wires NEXAGENT collectors to the durable disk spool and the
// gateway ingest endpoint, providing scrape scheduling, batching, and backoff.
package pipeline

import (
	"context"
	"log/slog"
	"time"

	"github.com/neuralops/platform/pkg/nexagent/buffer"
	"github.com/neuralops/platform/pkg/nexagent/collector"
)

// Config controls scrape cadence and batch shaping.
type Config struct {
	Hostname     string
	Gateway      string
	ScrapeEvery  time.Duration
	FlushEvery   time.Duration
	AgentVersion string
}

// Pipeline scrapes registered collectors, enqueues batches to the spool, and
// flushes them to the gateway. The spool guarantees at-least-once delivery
// across restarts and network partitions (air-gap friendly).
type Pipeline struct {
	cfg        Config
	collectors []collector.Collector
	spool      *buffer.DiskSpool
	log        *slog.Logger
}

// New constructs a pipeline.
func New(cfg Config, spool *buffer.DiskSpool, log *slog.Logger, collectors ...collector.Collector) *Pipeline {
	if cfg.ScrapeEvery <= 0 {
		cfg.ScrapeEvery = 15 * time.Second
	}
	if cfg.FlushEvery <= 0 {
		cfg.FlushEvery = 30 * time.Second
	}
	if log == nil {
		log = slog.Default()
	}
	return &Pipeline{cfg: cfg, collectors: collectors, spool: spool, log: log}
}

// Run blocks until ctx is cancelled, scraping and flushing on independent timers.
func (p *Pipeline) Run(ctx context.Context) error {
	scrape := time.NewTicker(p.cfg.ScrapeEvery)
	defer scrape.Stop()
	flush := time.NewTicker(p.cfg.FlushEvery)
	defer flush.Stop()

	// Prime one scrape immediately so the first batch is not delayed a full tick.
	p.scrapeOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			// Best-effort final flush so buffered data is not stranded on shutdown.
			flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = p.spool.FlushPending(flushCtx, p.cfg.Gateway)
			return ctx.Err()
		case <-scrape.C:
			p.scrapeOnce(ctx)
		case <-flush.C:
			if err := p.spool.FlushPending(ctx, p.cfg.Gateway); err != nil {
				p.log.Warn("nexagent flush failed", "error", err)
			}
		}
	}
}

// scrapeOnce collects from every collector and persists a single batch.
func (p *Pipeline) scrapeOnce(ctx context.Context) {
	samples := make([]collector.Sample, 0, 32)
	for _, c := range p.collectors {
		scoped, cancel := context.WithTimeout(ctx, p.cfg.ScrapeEvery)
		batch, err := c.Collect(scoped)
		cancel()
		if err != nil {
			p.log.Warn("nexagent collector error", "collector", c.Name(), "error", err)
			continue
		}
		samples = append(samples, batch...)
	}
	if len(samples) == 0 {
		return
	}

	payload := map[string]any{
		"source":       "nexagent",
		"agentVersion": p.cfg.AgentVersion,
		"host":         p.cfg.Hostname,
		"collectedAt":  time.Now().UTC().Format(time.RFC3339Nano),
		"sampleCount":  len(samples),
		"samples":      samples,
	}
	if err := p.spool.Enqueue(payload); err != nil {
		p.log.Error("nexagent spool enqueue failed", "error", err)
		return
	}
	p.log.Debug("nexagent batch spooled", "samples", len(samples))
}
