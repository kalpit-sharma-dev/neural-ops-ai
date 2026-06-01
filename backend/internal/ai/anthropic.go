package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/neuralops/platform/internal/domain"
)

// AnthropicClient implements LLMClient using Anthropic Messages API.
type AnthropicClient struct {
	cfg    ClientConfig
	client *http.Client
}

// NewAnthropicClient creates an Anthropic Claude client.
func NewAnthropicClient(cfg ClientConfig) *AnthropicClient {
	return &AnthropicClient{
		cfg: cfg,
		client: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

func (c *AnthropicClient) Classify(ctx context.Context, logMsg string) (*domain.ErrorClassification, error) {
	prompt := fmt.Sprintf(`Classify this log message. Return JSON only with category, confidence, reasoning.

Message:
%s`, logMsg)
	var result domain.ErrorClassification
	if err := c.completeJSON(ctx, prompt, 30*time.Second, &result); err != nil {
		return nil, err
	}
	if !result.Category.IsValid() {
		result.Category = domain.ErrorCategoryUnknown
	}
	return &result, nil
}

func (c *AnthropicClient) Explain(ctx context.Context, logMsg string, classification *domain.ErrorClassification) (*LogExplanation, error) {
	classificationJSON, _ := json.Marshal(classification)
	prompt := fmt.Sprintf(`Explain this log. Return JSON with plain_english, technical_summary, possible_causes, recommended_actions, severity_assessment.

Classification: %s
Log: %s`, string(classificationJSON), logMsg)

	timeout := c.cfg.ExplainTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	var explanation LogExplanation
	if err := c.completeJSON(ctx, prompt, timeout, &explanation); err != nil {
		return nil, err
	}
	return &explanation, nil
}

func (c *AnthropicClient) SummarizeStackTrace(ctx context.Context, trace domain.StackTrace) (*StackSummary, error) {
	payload, _ := json.Marshal(trace)
	prompt := fmt.Sprintf(`Summarize this stack trace as JSON with language, hotspot, summary, likely_cause, recurrence_hint, affected_modules.

%s`, string(payload))
	var summary StackSummary
	if err := c.completeJSON(ctx, prompt, 30*time.Second, &summary); err != nil {
		return nil, err
	}
	return &summary, nil
}

func (c *AnthropicClient) GenerateRCA(ctx context.Context, incident IncidentContext) (*domain.RootCause, error) {
	payload, _ := json.Marshal(incident)
	timeout := c.cfg.RCATimeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	prompt := fmt.Sprintf(`Generate root cause analysis JSON for this incident context: %s`, string(payload))
	var rca domain.RootCause
	if err := c.completeJSON(ctx, prompt, timeout, &rca); err != nil {
		return nil, err
	}
	return &rca, nil
}

func (c *AnthropicClient) GenerateRecommendation(ctx context.Context, rca domain.RootCause) ([]domain.Recommendation, error) {
	payload, _ := json.Marshal(rca)
	prompt := fmt.Sprintf(`Generate remediation recommendations JSON array for: %s`, string(payload))
	var recommendations []domain.Recommendation
	if err := c.completeJSON(ctx, prompt, 60*time.Second, &recommendations); err != nil {
		return nil, err
	}
	return recommendations, nil
}

func (c *AnthropicClient) AnswerQuery(ctx context.Context, query string, logs []domain.LogEntry) (string, error) {
	payload, _ := json.Marshal(logs)
	prompt := fmt.Sprintf(`Query: %s\nLogs: %s`, query, string(payload))
	var answer string
	if err := c.completeText(ctx, prompt, 60*time.Second, &answer); err != nil {
		return "", err
	}
	return answer, nil
}

func (c *AnthropicClient) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	openai := NewOpenAIClient(c.cfg)
	return openai.Embed(ctx, texts)
}

func (c *AnthropicClient) completeJSON(ctx context.Context, prompt string, timeout time.Duration, out any) error {
	content, err := c.completeRaw(ctx, prompt, timeout)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(extractJSON(content)), out)
}

func (c *AnthropicClient) completeText(ctx context.Context, prompt string, timeout time.Duration, out *string) error {
	content, err := c.completeRaw(ctx, prompt, timeout)
	if err != nil {
		return err
	}
	*out = content
	return nil
}

func (c *AnthropicClient) completeRaw(ctx context.Context, prompt string, timeout time.Duration) (string, error) {
	model := c.cfg.AnthropicModel
	if model == "" {
		model = "claude-3-5-sonnet-latest"
	}

	type message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type request struct {
		Model     string    `json:"model"`
		MaxTokens int       `json:"max_tokens"`
		Messages  []message `json:"messages"`
	}
	type contentBlock struct {
		Text string `json:"text"`
	}
	type response struct {
		Content []contentBlock `json:"content"`
	}

	body, _ := json.Marshal(request{
		Model:     model,
		MaxTokens: 4096,
		Messages:  []message{{Role: "user", Content: prompt}},
	})

	var resp response
	err := withRetry(ctx, c.cfg.MaxRetries, c.cfg.InitialBackoff, func(callCtx context.Context) error {
		callCtx, cancel := withTimeout(callCtx, timeout)
		defer cancel()

		req, reqErr := http.NewRequestWithContext(callCtx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
		if reqErr != nil {
			return reqErr
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("x-api-key", c.cfg.AnthropicAPIKey)
		req.Header.Set("anthropic-version", "2023-06-01")

		res, doErr := c.client.Do(req)
		if doErr != nil {
			return doErr
		}
		defer res.Body.Close()

		raw, readErr := io.ReadAll(res.Body)
		if readErr != nil {
			return readErr
		}
		if res.StatusCode >= 400 {
			return fmt.Errorf("anthropic api error: status=%d body=%s", res.StatusCode, string(raw))
		}
		return json.Unmarshal(raw, &resp)
	})
	if err != nil {
		return "", err
	}
	if len(resp.Content) == 0 {
		return "", fmt.Errorf("anthropic: empty response")
	}
	return strings.TrimSpace(resp.Content[0].Text), nil
}
