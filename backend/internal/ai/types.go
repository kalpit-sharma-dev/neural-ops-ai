package ai

import (
	"context"
	"time"

	"github.com/neuralops/platform/internal/domain"
)

// LogExplanation is a human-readable explanation of a log entry.
type LogExplanation struct {
	PlainEnglish       string   `json:"plain_english"`
	TechnicalSummary   string   `json:"technical_summary"`
	PossibleCauses     []string `json:"possible_causes"`
	RecommendedActions []string `json:"recommended_actions"`
	SeverityAssessment string   `json:"severity_assessment"`
}

// StackSummary summarizes a parsed stack trace.
type StackSummary struct {
	Language        string   `json:"language"`
	Hotspot         string   `json:"hotspot"`
	Summary         string   `json:"summary"`
	LikelyCause     string   `json:"likely_cause"`
	RecurrenceHint  string   `json:"recurrence_hint,omitempty"`
	AffectedModules []string `json:"affected_modules,omitempty"`
}

// IncidentContext provides context for RCA generation.
type IncidentContext struct {
	Incident    domain.Incident   `json:"incident"`
	Logs        []domain.LogEntry `json:"logs"`
	Deployments []domain.Deployment `json:"deployments,omitempty"`
	Metrics     []domain.Metric   `json:"metrics,omitempty"`
}

// LLMClient defines AI operations used by the analysis engine.
type LLMClient interface {
	Classify(ctx context.Context, logMsg string) (*domain.ErrorClassification, error)
	Explain(ctx context.Context, logMsg string, classification *domain.ErrorClassification) (*LogExplanation, error)
	SummarizeStackTrace(ctx context.Context, trace domain.StackTrace) (*StackSummary, error)
	GenerateRCA(ctx context.Context, incident IncidentContext) (*domain.RootCause, error)
	GenerateRecommendation(ctx context.Context, rca domain.RootCause) ([]domain.Recommendation, error)
	AnswerQuery(ctx context.Context, query string, logs []domain.LogEntry) (string, error)
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}

// ClientConfig configures LLM backends.
type ClientConfig struct {
	Provider        string
	OpenAIBaseURL   string
	OpenAIAPIKey    string
	OpenAIModel     string
	EmbeddingModel  string
	AnthropicAPIKey string
	AnthropicModel  string
	OllamaBaseURL   string
	OllamaModel     string
	ExplainTimeout  time.Duration
	RCATimeout      time.Duration
	MaxRetries      int
	InitialBackoff  time.Duration
}
