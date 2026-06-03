// NEXAGENT — unified telemetry agent.
//
// It runs an always-on cross-platform host collector (CPU/memory/network/load)
// plus, on Linux, a low-overhead kernel TCP counter collector read from procfs.
// Samples are batched into a durable on-disk spool and flushed to the gateway,
// giving at-least-once delivery across restarts and network partitions.
//
// The eBPF collector (per-flow RTT, syscall latency) lives in
// pkg/nexagent/ebpf and is an optional, toolchain-gated upgrade to the procfs
// kernel collector; see that package's README for the libbpf/clang build.
package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/neuralops/platform/pkg/nexagent/buffer"
	"github.com/neuralops/platform/pkg/nexagent/collector"
	"github.com/neuralops/platform/pkg/nexagent/pipeline"
)

// version is overridable at build time via -ldflags "-X main.version=...".
var version = "0.1.0"

func main() {
	gateway := flag.String("gateway", envOr("NEXAGENT_GATEWAY", "http://localhost:8080"), "gateway base URL")
	bufferDir := flag.String("buffer", envOr("NEXAGENT_BUFFER_DIR", "./data/nexagent-buffer"), "offline spool directory")
	scrapeEvery := flag.Duration("scrape-interval", envDuration("NEXAGENT_SCRAPE_INTERVAL", 15*time.Second), "collector scrape interval")
	flushEvery := flag.Duration("flush-interval", envDuration("NEXAGENT_FLUSH_INTERVAL", 30*time.Second), "spool flush interval")
	replay := flag.Bool("replay", false, "replay buffered batches on startup")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "unknown-host"
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	spool := buffer.NewDiskSpool(*bufferDir)
	if *replay {
		n, rerr := spool.Replay(ctx, *gateway)
		if rerr != nil {
			logger.Warn("nexagent replay failed", "error", rerr)
		} else {
			logger.Info("nexagent replayed buffered batches", "count", n)
		}
	}

	collectors := []collector.Collector{
		collector.NewHostCollector(hostname),
		collector.NewKernelCollector(hostname),
	}

	p := pipeline.New(pipeline.Config{
		Hostname:     hostname,
		Gateway:      *gateway,
		ScrapeEvery:  *scrapeEvery,
		FlushEvery:   *flushEvery,
		AgentVersion: version,
	}, spool, logger, collectors...)

	logger.Info("nexagent started",
		"version", version, "gateway", *gateway, "buffer", *bufferDir,
		"host", hostname, "collectors", len(collectors),
		"scrapeInterval", scrapeEvery.String(), "flushInterval", flushEvery.String(),
	)

	if rerr := p.Run(ctx); rerr != nil && rerr != context.Canceled {
		log.Fatalf("nexagent pipeline stopped: %v", rerr)
	}
	logger.Info("nexagent stopped")
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envDuration(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
