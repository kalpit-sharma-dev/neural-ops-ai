package ai

import (
	"context"
	"time"

	"github.com/neuralops/platform/internal/domain"
	"github.com/neuralops/platform/pkg/metrics"
)

type instrumentedClient struct {
	provider string
	delegate LLMClient
}

func instrumentClient(provider string, delegate LLMClient) LLMClient {
	if delegate == nil {
		return delegate
	}
	return &instrumentedClient{provider: provider, delegate: delegate}
}

func (c *instrumentedClient) observe(operation string, err error, started time.Time) {
	status := "success"
	if err != nil {
		status = "error"
	}
	metrics.RecordAIRequest(c.provider, operation, status, time.Since(started))
}

func (c *instrumentedClient) Classify(ctx context.Context, logMsg string) (*domain.ErrorClassification, error) {
	start := time.Now()
	out, err := c.delegate.Classify(ctx, logMsg)
	c.observe("classify", err, start)
	return out, err
}

func (c *instrumentedClient) Explain(ctx context.Context, logMsg string, classification *domain.ErrorClassification) (*LogExplanation, error) {
	start := time.Now()
	out, err := c.delegate.Explain(ctx, logMsg, classification)
	c.observe("explain", err, start)
	return out, err
}

func (c *instrumentedClient) SummarizeStackTrace(ctx context.Context, trace domain.StackTrace) (*StackSummary, error) {
	start := time.Now()
	out, err := c.delegate.SummarizeStackTrace(ctx, trace)
	c.observe("summarize_stack_trace", err, start)
	return out, err
}

func (c *instrumentedClient) GenerateRCA(ctx context.Context, incident IncidentContext) (*domain.RootCause, error) {
	start := time.Now()
	out, err := c.delegate.GenerateRCA(ctx, incident)
	c.observe("generate_rca", err, start)
	return out, err
}

func (c *instrumentedClient) GenerateRecommendation(ctx context.Context, rca domain.RootCause) ([]domain.Recommendation, error) {
	start := time.Now()
	out, err := c.delegate.GenerateRecommendation(ctx, rca)
	c.observe("generate_recommendation", err, start)
	return out, err
}

func (c *instrumentedClient) AnswerQuery(ctx context.Context, query string, logs []domain.LogEntry) (string, error) {
	start := time.Now()
	out, err := c.delegate.AnswerQuery(ctx, query, logs)
	c.observe("answer_query", err, start)
	return out, err
}

func (c *instrumentedClient) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	start := time.Now()
	out, err := c.delegate.Embed(ctx, texts)
	c.observe("embed", err, start)
	return out, err
}
