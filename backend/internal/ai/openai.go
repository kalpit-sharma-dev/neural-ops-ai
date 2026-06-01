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

// OpenAIClient implements LLMClient using OpenAI-compatible APIs.
type OpenAIClient struct {
	cfg    ClientConfig
	client *http.Client
}

// NewOpenAIClient creates an OpenAI-compatible client.
func NewOpenAIClient(cfg ClientConfig) *OpenAIClient {
	baseURL := cfg.OpenAIBaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &OpenAIClient{
		cfg: cfg,
		client: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

func (c *OpenAIClient) Classify(ctx context.Context, logMsg string) (*domain.ErrorClassification, error) {
	prompt := fmt.Sprintf(`Classify the following log/error message into one category: TIMEOUT, DB_DEADLOCK, MEMORY_LEAK, CONNECTION_EXHAUSTION, DNS_ISSUE, SSL_ISSUE, RETRY_STORM, KAFKA_LAG, DEPLOYMENT_ISSUE, DEPENDENCY_FAILURE, THREAD_STARVATION, CIRCUIT_BREAKER_OPEN, RATE_LIMITING, AUTH_FAILURE, UNKNOWN.
Return JSON: {"category":"...","confidence":0.0,"reasoning":"..."}

Message:
%s`, logMsg)

	var result domain.ErrorClassification
	if err := c.chatJSON(ctx, c.cfg.OpenAIModel, prompt, 30*time.Second, &result); err != nil {
		return nil, err
	}
	if !result.Category.IsValid() {
		result.Category = domain.ErrorCategoryUnknown
	}
	return &result, nil
}

func (c *OpenAIClient) Explain(ctx context.Context, logMsg string, classification *domain.ErrorClassification) (*LogExplanation, error) {
	classificationJSON, _ := json.Marshal(classification)
	prompt := fmt.Sprintf(`Explain this log for SRE engineers. Return JSON with keys plain_english, technical_summary, possible_causes, recommended_actions, severity_assessment.

Classification: %s
Log:
%s`, string(classificationJSON), logMsg)

	var explanation LogExplanation
	timeout := c.cfg.ExplainTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	if err := c.chatJSON(ctx, c.cfg.OpenAIModel, prompt, timeout, &explanation); err != nil {
		return nil, err
	}
	return &explanation, nil
}

func (c *OpenAIClient) SummarizeStackTrace(ctx context.Context, trace domain.StackTrace) (*StackSummary, error) {
	payload, _ := json.Marshal(trace)
	prompt := fmt.Sprintf(`Summarize this stack trace. Return JSON with keys language, hotspot, summary, likely_cause, recurrence_hint, affected_modules.

Stack trace:
%s`, string(payload))

	var summary StackSummary
	if err := c.chatJSON(ctx, c.cfg.OpenAIModel, prompt, 30*time.Second, &summary); err != nil {
		return nil, err
	}
	return &summary, nil
}

func (c *OpenAIClient) GenerateRCA(ctx context.Context, incident IncidentContext) (*domain.RootCause, error) {
	payload, _ := json.Marshal(incident)
	prompt := fmt.Sprintf(`Generate root cause analysis. Return JSON matching RootCause fields: firstFailingService, rootCauseDescription, confidence, evidence array.

Context:
%s`, string(payload))

	timeout := c.cfg.RCATimeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	var rca domain.RootCause
	if err := c.chatJSON(ctx, c.cfg.OpenAIModel, prompt, timeout, &rca); err != nil {
		return nil, err
	}
	return &rca, nil
}

func (c *OpenAIClient) GenerateRecommendation(ctx context.Context, rca domain.RootCause) ([]domain.Recommendation, error) {
	payload, _ := json.Marshal(rca)
	prompt := fmt.Sprintf(`Generate remediation recommendations. Return JSON array with type, description, codeSnippet, priority.

Root cause:
%s`, string(payload))

	var recommendations []domain.Recommendation
	if err := c.chatJSON(ctx, c.cfg.OpenAIModel, prompt, 60*time.Second, &recommendations); err != nil {
		return nil, err
	}
	return recommendations, nil
}

func (c *OpenAIClient) AnswerQuery(ctx context.Context, query string, logs []domain.LogEntry) (string, error) {
	payload, _ := json.Marshal(logs)
	prompt := fmt.Sprintf(`Answer the observability query using the provided logs.

Query: %s
Logs: %s`, query, string(payload))

	var answer string
	if err := c.chatText(ctx, c.cfg.OpenAIModel, prompt, 60*time.Second, &answer); err != nil {
		return "", err
	}
	return answer, nil
}

func (c *OpenAIClient) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	model := c.cfg.EmbeddingModel
	if model == "" {
		model = "text-embedding-3-small"
	}

	type embedRequest struct {
		Model string   `json:"model"`
		Input []string `json:"input"`
	}
	type embedData struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	}
	type embedResponse struct {
		Data []embedData `json:"data"`
	}

	body, _ := json.Marshal(embedRequest{Model: model, Input: texts})
	url := strings.TrimRight(c.cfg.OpenAIBaseURL, "/") + "/embeddings"
	if c.cfg.OpenAIBaseURL == "" {
		url = "https://api.openai.com/v1/embeddings"
	}

	var resp embedResponse
	if err := c.doRequest(ctx, http.MethodPost, url, body, 60*time.Second, &resp); err != nil {
		return nil, err
	}

	vectors := make([][]float32, len(texts))
	for _, item := range resp.Data {
		if item.Index >= 0 && item.Index < len(vectors) {
			vectors[item.Index] = item.Embedding
		}
	}
	return vectors, nil
}

func (c *OpenAIClient) chatJSON(ctx context.Context, model, prompt string, timeout time.Duration, out any) error {
	content, err := c.chatRaw(ctx, model, prompt, timeout)
	if err != nil {
		return err
	}
	content = extractJSON(content)
	return json.Unmarshal([]byte(content), out)
}

func (c *OpenAIClient) chatText(ctx context.Context, model, prompt string, timeout time.Duration, out *string) error {
	content, err := c.chatRaw(ctx, model, prompt, timeout)
	if err != nil {
		return err
	}
	*out = content
	return nil
}

func (c *OpenAIClient) chatRaw(ctx context.Context, model, prompt string, timeout time.Duration) (string, error) {
	if model == "" {
		model = "gpt-4o-mini"
	}

	type message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type chatRequest struct {
		Model    string    `json:"model"`
		Messages []message `json:"messages"`
	}
	type chatChoice struct {
		Message message `json:"message"`
	}
	type chatResponse struct {
		Choices []chatChoice `json:"choices"`
	}

	body, _ := json.Marshal(chatRequest{
		Model: model,
		Messages: []message{
			{Role: "system", Content: "You are an expert SRE assistant. Respond with valid JSON when asked."},
			{Role: "user", Content: prompt},
		},
	})

	url := strings.TrimRight(c.cfg.OpenAIBaseURL, "/") + "/chat/completions"
	if c.cfg.OpenAIBaseURL == "" {
		url = "https://api.openai.com/v1/chat/completions"
	}

	var resp chatResponse
	err := withRetry(ctx, c.cfg.MaxRetries, c.cfg.InitialBackoff, func(callCtx context.Context) error {
		return c.doRequest(callCtx, http.MethodPost, url, body, timeout, &resp)
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("openai: empty response")
	}
	return resp.Choices[0].Message.Content, nil
}

func (c *OpenAIClient) doRequest(ctx context.Context, method, url string, body []byte, timeout time.Duration, out any) error {
	callCtx, cancel := withTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(callCtx, method, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.cfg.OpenAIAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.OpenAIAPIKey)
	}

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		return fmt.Errorf("openai api error: status=%d body=%s", res.StatusCode, string(raw))
	}
	return json.Unmarshal(raw, out)
}

func extractJSON(content string) string {
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		return content[start : end+1]
	}
	start = strings.Index(content, "[")
	end = strings.LastIndex(content, "]")
	if start >= 0 && end > start {
		return content[start : end+1]
	}
	return content
}
