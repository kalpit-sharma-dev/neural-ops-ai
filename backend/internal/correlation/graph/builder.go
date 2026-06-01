package graph

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/neuralops/platform/internal/correlation/repository"
	"github.com/neuralops/platform/internal/ingestion/dto"
	"go.uber.org/zap"
)

// Builder maintains a real-time service dependency graph.
type Builder struct {
	log  *zap.Logger
	repo *repository.Store
	mu   sync.RWMutex
	adj  map[string]map[string]*repository.DependencyEdge
}

// NewBuilder creates a dependency graph builder.
func NewBuilder(log *zap.Logger, repo *repository.Store) *Builder {
	return &Builder{
		log: log,
		repo: repo,
		adj:  make(map[string]map[string]*repository.DependencyEdge),
	}
}

// HandleTraceSpan ingests a distributed trace span.
func (b *Builder) HandleTraceSpan(span dto.TraceSpanRequest) {
	if span.Service == "" {
		return
	}
	target := span.Tags["peer.service"]
	if target == "" {
		target = span.Tags["downstream.service"]
	}
	if target == "" {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.adj[span.Service]; !ok {
		b.adj[span.Service] = make(map[string]*repository.DependencyEdge)
	}
	edge := b.adj[span.Service][target]
	if edge == nil {
		edge = &repository.DependencyEdge{Target: target}
		b.adj[span.Service][target] = edge
	}
	edge.CallCount++
	if float64(span.DurationMs) > edge.P99LatencyMs {
		edge.P99LatencyMs = float64(span.DurationMs)
	}
}

// HandleTraceRaw accepts raw JSON trace payloads.
func (b *Builder) HandleTraceRaw(raw []byte) {
	var span dto.TraceSpanRequest
	if err := json.Unmarshal(raw, &span); err != nil {
		return
	}
	b.HandleTraceSpan(span)
}

// Graph returns a copy of the current adjacency list.
func (b *Builder) Graph() map[string][]repository.DependencyEdge {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make(map[string][]repository.DependencyEdge, len(b.adj))
	for source, targets := range b.adj {
		edges := make([]repository.DependencyEdge, 0, len(targets))
		for _, edge := range targets {
			copyEdge := *edge
			edges = append(edges, copyEdge)
		}
		out[source] = edges
	}
	return out
}

// BlastRadius returns downstream services from a root failure point.
func (b *Builder) BlastRadius(root string) []string {
	return repository.BlastRadius(b.Graph(), root)
}

// Flush persists in-memory graph edges to PostgreSQL.
func (b *Builder) Flush(ctx context.Context) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for source, targets := range b.adj {
		for _, edge := range targets {
			if err := b.repo.UpsertDependency(ctx, source, edge.Target, edge.CallCount, edge.P99LatencyMs); err != nil {
				b.log.Warn("upsert dependency failed", zap.Error(err))
			}
		}
	}
}

// StartFlusher periodically persists the graph.
func (b *Builder) StartFlusher(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				b.Flush(context.Background())
				return
			case <-ticker.C:
				b.Flush(ctx)
			}
		}
	}()
}
