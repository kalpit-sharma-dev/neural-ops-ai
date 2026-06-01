package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/neuralops/platform/internal/ai"
	"github.com/neuralops/platform/internal/search/dto"
)

// Parser converts natural language queries into structured search filters.
type Parser struct {
	llm ai.LLMClient
}

// NewParser creates a natural language query parser.
func NewParser(llm ai.LLMClient) *Parser {
	return &Parser{llm: llm}
}

const parsePrompt = `Extract structured log search filters from the user question.
Return ONLY valid JSON with keys:
query, service, severity, classification, startTime, endTime, explanation.
Use RFC3339 timestamps in UTC. If time is relative (yesterday, last hour), resolve against now.
Severity must be uppercase when present (ERROR, WARN, INFO, DEBUG, FATAL).
If a field is unknown, use empty string or omit it.`

// Parse converts a natural language question into a structured query.
func (p *Parser) Parse(ctx context.Context, question string) (*dto.ParsedQuery, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return nil, fmt.Errorf("question is required")
	}

	parsed, err := p.parseWithLLM(ctx, question)
	if err == nil && parsed != nil {
		return parsed, nil
	}
	return p.parseHeuristic(question), nil
}

func (p *Parser) parseWithLLM(ctx context.Context, question string) (*dto.ParsedQuery, error) {
	if p.llm == nil {
		return nil, fmt.Errorf("llm unavailable")
	}

	prompt := fmt.Sprintf("%s\n\nQuestion: %s\nNow: %s", parsePrompt, question, time.Now().UTC().Format(time.RFC3339))
	raw, err := p.llm.AnswerQuery(ctx, prompt, nil)
	if err != nil || strings.TrimSpace(raw) == "" {
		return nil, err
	}

	jsonBody := extractJSON(raw)
	var parsed dto.ParsedQuery
	if err := json.Unmarshal([]byte(jsonBody), &parsed); err != nil {
		return nil, err
	}
	if parsed.StartTime != nil {
		t := parsed.StartTime.UTC()
		parsed.StartTime = &t
	}
	if parsed.EndTime != nil {
		t := parsed.EndTime.UTC()
		parsed.EndTime = &t
	}
	if parsed.Explanation == "" {
		parsed.Explanation = "Parsed via LLM"
	}
	return &parsed, nil
}

func (p *Parser) parseHeuristic(question string) *dto.ParsedQuery {
	lower := strings.ToLower(question)
	parsed := &dto.ParsedQuery{
		Query:       question,
		Explanation: "Parsed via heuristics",
	}

	if strings.Contains(lower, "error") || strings.Contains(lower, "failure") || strings.Contains(lower, "fail") {
		parsed.Severity = "ERROR"
	}
	if strings.Contains(lower, "warn") {
		parsed.Severity = "WARN"
	}

	servicePatterns := map[string]string{
		"upi": "upi-service", "neft": "neft-service", "rtgs": "rtgs-service",
		"payment": "payment-service", "auth": "auth-service",
	}
	for keyword, service := range servicePatterns {
		if strings.Contains(lower, keyword) {
			parsed.Service = service
			break
		}
	}

	now := time.Now().UTC()
	switch {
	case strings.Contains(lower, "yesterday"):
		start := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC)
		parsed.StartTime = &start
		end := start.Add(24 * time.Hour)
		parsed.EndTime = &end
	case strings.Contains(lower, "last hour"):
		start := now.Add(-1 * time.Hour)
		parsed.StartTime = &start
		parsed.EndTime = &now
	case strings.Contains(lower, "today"):
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		parsed.StartTime = &start
		parsed.EndTime = &now
	}

	timeRange := regexp.MustCompile(`(\d{1,2})(?::(\d{2}))?\s*(am|pm)?`).FindStringSubmatch(lower)
	if len(timeRange) >= 2 && parsed.StartTime != nil {
		hour := parseHour(timeRange[1], timeRange[3])
		start := time.Date(parsed.StartTime.Year(), parsed.StartTime.Month(), parsed.StartTime.Day(), hour, 0, 0, 0, time.UTC)
		parsed.StartTime = &start
		parsed.EndTime = ptrTime(start.Add(time.Hour))
	}

	parsed.Query = strings.TrimSpace(parsed.Query)
	return parsed
}

func parseHour(value, meridiem string) int {
	hour := 0
	fmt.Sscanf(value, "%d", &hour)
	switch strings.ToLower(meridiem) {
	case "pm":
		if hour < 12 {
			hour += 12
		}
	case "am":
		if hour == 12 {
			hour = 0
		}
	}
	return hour
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

var jsonBlockPattern = regexp.MustCompile(`(?s)\{.*\}`)

func extractJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "```") {
		raw = strings.TrimPrefix(raw, "```json")
		raw = strings.TrimPrefix(raw, "```")
		raw = strings.TrimSuffix(raw, "```")
		raw = strings.TrimSpace(raw)
	}
	if match := jsonBlockPattern.FindString(raw); match != "" {
		return match
	}
	return raw
}

// ToLogSearchRequest converts a parsed query to a log search request.
func ToLogSearchRequest(parsed *dto.ParsedQuery, size int) dto.LogSearchRequest {
	if parsed == nil {
		return dto.LogSearchRequest{Size: size}
	}
	return dto.LogSearchRequest{
		Query:          parsed.Query,
		Service:        parsed.Service,
		Severity:       parsed.Severity,
		Classification: parsed.Classification,
		StartTime:      parsed.StartTime,
		EndTime:        parsed.EndTime,
		Size:           size,
	}
}
