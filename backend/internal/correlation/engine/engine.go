package engine

import (
	"context"
	"encoding/json"

	"github.com/neuralops/platform/internal/correlation/deployment"
	"github.com/neuralops/platform/internal/correlation/graph"
	"github.com/neuralops/platform/internal/correlation/temporal"
	"github.com/neuralops/platform/internal/correlation/transaction"
	"github.com/neuralops/platform/internal/domain"
	"github.com/neuralops/platform/internal/ingestion/dto"
	"go.uber.org/zap"
)

// Engine orchestrates correlation subsystems.
type Engine struct {
	log         *zap.Logger
	deployment  *deployment.Correlator
	graph       *graph.Builder
	transaction *transaction.Correlator
	temporal    *temporal.Analyzer
	spans       SpanWriter
}

// SpanWriter persists trace spans for APM queries.
type SpanWriter interface {
	WriteSpan(ctx context.Context, span dto.TraceSpanRequest) error
}

// New creates a correlation engine.
func New(
	log *zap.Logger,
	deploymentCorrelator *deployment.Correlator,
	graphBuilder *graph.Builder,
	txnCorrelator *transaction.Correlator,
	temporalAnalyzer *temporal.Analyzer,
	spanWriter SpanWriter,
) *Engine {
	return &Engine{
		log:         log,
		deployment:  deploymentCorrelator,
		graph:       graphBuilder,
		transaction: txnCorrelator,
		temporal:    temporalAnalyzer,
		spans:       spanWriter,
	}
}

// HandleEvent processes deployment/config events.
func (e *Engine) HandleEvent(ctx context.Context, raw []byte) error {
	var event domain.Deployment
	if err := json.Unmarshal(raw, &event); err != nil {
		return err
	}
	e.deployment.HandleDeployment(ctx, event)
	return nil
}

// HandleTrace processes distributed trace spans.
func (e *Engine) HandleTrace(ctx context.Context, raw []byte) error {
	if e.spans != nil {
		var span dto.TraceSpanRequest
		if err := json.Unmarshal(raw, &span); err == nil && span.TraceID != "" {
			_ = e.spans.WriteSpan(ctx, span)
		}
	}
	e.graph.HandleTraceRaw(raw)
	return nil
}

// HandleLog processes enriched logs for transaction and temporal correlation.
func (e *Engine) HandleLog(_ context.Context, raw []byte) error {
	entry, err := decodeLogEntry(raw)
	if err != nil {
		return err
	}
	e.deployment.HandleLog(entry)
	e.transaction.HandleLog(entry)
	e.temporal.HandleLog(entry)

	if entry.Pod != "" && entry.IsErrorSeverity() {
		e.temporal.HandleInfraEvent("pod:" + entry.Pod + " error")
	}
	return nil
}

func decodeLogEntry(raw []byte) (domain.LogEntry, error) {
	var enriched dto.EnrichedLog
	if err := json.Unmarshal(raw, &enriched); err == nil && enriched.Service != "" {
		return enriched.LogEntry, nil
	}
	var wrapper struct {
		Log domain.LogEntry `json:"log"`
	}
	if err := json.Unmarshal(raw, &wrapper); err == nil && wrapper.Log.Service != "" {
		return wrapper.Log, nil
	}
	var entry domain.LogEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		return domain.LogEntry{}, err
	}
	return entry, nil
}
