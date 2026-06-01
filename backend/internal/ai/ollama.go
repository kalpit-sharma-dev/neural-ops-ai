package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/neuralops/platform/internal/domain"
)

// OllamaClient implements LLMClient using a local Ollama instance.
type OllamaClient struct {
	openai *OpenAIClient
	cfg    ClientConfig
}

// NewOllamaClient creates an Ollama-backed client via OpenAI-compatible API.
func NewOllamaClient(cfg ClientConfig) *OllamaClient {
	baseURL := cfg.OllamaBaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434/v1"
	}
	if !strings.HasSuffix(baseURL, "/v1") {
		baseURL = strings.TrimRight(baseURL, "/") + "/v1"
	}

	openaiCfg := cfg
	openaiCfg.OpenAIBaseURL = baseURL
	openaiCfg.OpenAIModel = firstNonEmpty(cfg.OllamaModel, "llama3.1")
	openaiCfg.EmbeddingModel = firstNonEmpty(cfg.EmbeddingModel, "nomic-embed-text")

	return &OllamaClient{
		openai: NewOpenAIClient(openaiCfg),
		cfg:    cfg,
	}
}

func (c *OllamaClient) Classify(ctx context.Context, logMsg string) (*domain.ErrorClassification, error) {
	return c.openai.Classify(ctx, logMsg)
}

func (c *OllamaClient) Explain(ctx context.Context, logMsg string, classification *domain.ErrorClassification) (*LogExplanation, error) {
	return c.openai.Explain(ctx, logMsg, classification)
}

func (c *OllamaClient) SummarizeStackTrace(ctx context.Context, trace domain.StackTrace) (*StackSummary, error) {
	return c.openai.SummarizeStackTrace(ctx, trace)
}

func (c *OllamaClient) GenerateRCA(ctx context.Context, incident IncidentContext) (*domain.RootCause, error) {
	return c.openai.GenerateRCA(ctx, incident)
}

func (c *OllamaClient) GenerateRecommendation(ctx context.Context, rca domain.RootCause) ([]domain.Recommendation, error) {
	return c.openai.GenerateRecommendation(ctx, rca)
}

func (c *OllamaClient) AnswerQuery(ctx context.Context, query string, logs []domain.LogEntry) (string, error) {
	return c.openai.AnswerQuery(ctx, query, logs)
}

func (c *OllamaClient) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	return c.openai.Embed(ctx, texts)
}

// NewLLMClient selects the configured LLM provider.
func NewLLMClient(cfg ClientConfig) (LLMClient, error) {
	provider := strings.ToLower(cfg.Provider)
	if provider == "" {
		provider = "openai"
	}

	var client LLMClient

	switch provider {
	case "anthropic", "claude":
		if cfg.AnthropicAPIKey == "" {
			return NewNoopLLMClient(), nil
		}
		client = NewAnthropicClient(cfg)
	case "ollama":
		client = NewOllamaClient(cfg)
	case "openai", "azure":
		if cfg.OpenAIAPIKey == "" && cfg.OpenAIBaseURL == "" {
			return NewNoopLLMClient(), nil
		}
		client = NewOpenAIClient(cfg)
	default:
		return nil, fmt.Errorf("unsupported llm provider: %s", cfg.Provider)
	}

	return instrumentClient(provider, client), nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

// NoopLLMClient returns empty AI responses when no provider is configured.
type NoopLLMClient struct{}

// NewNoopLLMClient creates a no-op LLM client for local development.
func NewNoopLLMClient() *NoopLLMClient {
	return &NoopLLMClient{}
}

func (n *NoopLLMClient) Classify(context.Context, string) (*domain.ErrorClassification, error) {
	return nil, fmt.Errorf("llm not configured")
}

func (n *NoopLLMClient) Explain(context.Context, string, *domain.ErrorClassification) (*LogExplanation, error) {
	return nil, fmt.Errorf("llm not configured")
}

func (n *NoopLLMClient) SummarizeStackTrace(context.Context, domain.StackTrace) (*StackSummary, error) {
	return nil, fmt.Errorf("llm not configured")
}

func (n *NoopLLMClient) GenerateRCA(context.Context, IncidentContext) (*domain.RootCause, error) {
	return nil, fmt.Errorf("llm not configured")
}

func (n *NoopLLMClient) GenerateRecommendation(context.Context, domain.RootCause) ([]domain.Recommendation, error) {
	return nil, fmt.Errorf("llm not configured")
}

func (n *NoopLLMClient) AnswerQuery(context.Context, string, []domain.LogEntry) (string, error) {
	return "", fmt.Errorf("llm not configured")
}

func (n *NoopLLMClient) Embed(context.Context, []string) ([][]float32, error) {
	return nil, fmt.Errorf("llm not configured")
}
