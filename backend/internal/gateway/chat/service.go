package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/neuralops/platform/internal/ai"
	"github.com/neuralops/platform/internal/domain"
	"github.com/neuralops/platform/internal/gateway/config"
)

// Service answers AI chat queries with operational context.
type Service struct {
	cfg    config.Config
	llm    ai.LLMClient
	client *http.Client
}

// NewService creates an AI chat service.
func NewService(cfg config.Config, llm ai.LLMClient) *Service {
	return &Service{
		cfg:    cfg,
		llm:    llm,
		client: &http.Client{Timeout: cfg.Dashboard.FetchTimeout},
	}
}

// QueryRequest is an AI chat request.
type QueryRequest struct {
	Question string `json:"question"`
}

// SourceRef is a citation returned with AI chat answers.
type SourceRef struct {
	Service   string `json:"service"`
	Message   string `json:"message"`
	Severity  string `json:"severity"`
	Timestamp string `json:"timestamp"`
}

// StreamAnswer fetches context and streams an LLM answer via callback chunks.
func (s *Service) StreamAnswer(ctx context.Context, question string, emit func(chunk string) error, emitSources func([]SourceRef) error) error {
	question = strings.TrimSpace(question)
	if question == "" {
		return fmt.Errorf("question is required")
	}

	contextText, logs := s.buildContext(ctx)
	prompt := fmt.Sprintf(`You are NeuralOps SRE assistant. Use the context below to answer the operator question clearly and concisely.

Context:
%s

Question: %s`, contextText, question)

	if emitSources != nil {
		sources := make([]SourceRef, 0, len(logs))
		for _, entry := range logs {
			sources = append(sources, SourceRef{
				Service:   entry.Service,
				Message:   truncate(entry.Message, 240),
				Severity:  string(entry.Severity),
				Timestamp: entry.Timestamp.UTC().Format(time.RFC3339),
			})
		}
		if err := emitSources(sources); err != nil {
			return err
		}
	}

	answer, err := s.llm.AnswerQuery(ctx, prompt, logs)
	if err != nil {
		return err
	}

	for _, chunk := range chunkText(answer, 120) {
		if err := emit(chunk); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) buildContext(ctx context.Context) (string, []domain.LogEntry) {
	end := time.Now().UTC()
	start := end.Add(-24 * time.Hour)

	incidentsBody, _ := s.get(ctx, s.cfg.Services.Incident+"/api/v1/incidents?size=20")
	searchPayload, _ := json.Marshal(map[string]any{
		"severity":  "ERROR",
		"startTime": start.Format(time.RFC3339),
		"endTime":   end.Format(time.RFC3339),
		"size":      20,
	})
	searchBody, _ := s.post(ctx, s.cfg.Services.Search+"/api/v1/search/logs", searchPayload)

	contextParts := []string{
		"Incidents (last 24h): " + truncate(string(incidentsBody), 4000),
		"Error logs (last 24h): " + truncate(string(searchBody), 4000),
	}

	logs := make([]domain.LogEntry, 0)
	var searchResp struct {
		Data struct {
			Hits []struct {
				Message   string `json:"message"`
				Service   string `json:"service"`
				Severity  string `json:"severity"`
				Timestamp string `json:"timestamp"`
			} `json:"hits"`
		} `json:"data"`
	}
	if err := json.Unmarshal(searchBody, &searchResp); err == nil {
		for _, hit := range searchResp.Data.Hits {
			ts, _ := time.Parse(time.RFC3339, hit.Timestamp)
			logs = append(logs, domain.LogEntry{
				Message:   hit.Message,
				Service:   hit.Service,
				Severity:  domain.LogSeverity(hit.Severity),
				Timestamp: ts,
			})
		}
	}
	return strings.Join(contextParts, "\n\n"), logs
}

func (s *Service) get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return io.ReadAll(res.Body)
}

func (s *Service) post(ctx context.Context, url string, payload []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return io.ReadAll(res.Body)
}

func chunkText(value string, size int) []string {
	if size <= 0 {
		return []string{value}
	}
	chunks := make([]string, 0)
	for len(value) > 0 {
		if len(value) <= size {
			chunks = append(chunks, value)
			break
		}
		chunks = append(chunks, value[:size])
		value = value[size:]
	}
	return chunks
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max] + "..."
}
