package graph

import (
	"testing"

	"github.com/neuralops/platform/internal/ingestion/dto"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBuilderHandleTraceSpanBuildsEdge(t *testing.T) {
	b := NewBuilder(zap.NewNop(), nil)
	b.HandleTraceSpan(dto.TraceSpanRequest{
		Service:    "api-gateway",
		DurationMs: 120,
		Tags:       map[string]string{"peer.service": "payment-api"},
	})
	b.HandleTraceSpan(dto.TraceSpanRequest{
		Service:    "api-gateway",
		DurationMs: 250,
		Tags:       map[string]string{"peer.service": "payment-api"},
	})

	b.mu.RLock()
	defer b.mu.RUnlock()
	require.Contains(t, b.adj, "api-gateway")
	require.Contains(t, b.adj["api-gateway"], "payment-api")
	edge := b.adj["api-gateway"]["payment-api"]
	require.Equal(t, int64(2), edge.CallCount)
	require.Equal(t, float64(250), edge.P99LatencyMs)
}

func TestBuilderIgnoresMissingPeer(t *testing.T) {
	b := NewBuilder(zap.NewNop(), nil)
	b.HandleTraceSpan(dto.TraceSpanRequest{Service: "solo-service", DurationMs: 10})
	b.mu.RLock()
	defer b.mu.RUnlock()
	require.Empty(t, b.adj)
}
