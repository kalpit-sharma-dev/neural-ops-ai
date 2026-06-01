package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (e *Executor) notifySlack(ctx context.Context, tenantID string, step Step, ctxData map[string]string) (string, error) {
	cfg, err := e.integrationConfig(ctx, tenantID, "slack")
	if err != nil {
		return "", err
	}
	url := cfg["webhookUrl"]
	if url == "" {
		url = cfg["webhook_url"]
	}
	if url == "" {
		return "", fmt.Errorf("slack webhook not configured")
	}
	text := ctxData["title"]
	if text == "" {
		text = step.Label
	}
	if desc := ctxData["description"]; desc != "" {
		text += "\n" + desc
	}
	payload := map[string]string{"text": text}
	raw, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("slack: %s", resp.Status)
	}
	return "Slack notification sent", nil
}

func (e *Executor) notifyPagerDuty(ctx context.Context, tenantID string, step Step, ctxData map[string]string) (string, error) {
	cfg, err := e.integrationConfig(ctx, tenantID, "pagerduty")
	if err != nil {
		return "", err
	}
	key := cfg["routingKey"]
	if key == "" {
		key = cfg["integrationKey"]
	}
	if key == "" {
		return "", fmt.Errorf("pagerduty routing key not configured")
	}
	summary := ctxData["title"]
	if summary == "" {
		summary = step.Label
	}
	payload := map[string]any{
		"routing_key":  key,
		"event_action": "trigger",
		"payload": map[string]string{
			"summary":  summary,
			"severity": "critical",
			"source":   "neuralops",
		},
	}
	raw, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://events.pagerduty.com/v2/enqueue", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("pagerduty: %s", resp.Status)
	}
	return "PagerDuty incident triggered", nil
}

func (e *Executor) integrationConfig(ctx context.Context, tenantID, key string) (map[string]string, error) {
	var configJSON []byte
	var connected bool
	err := e.pool.QueryRow(ctx, `
SELECT config, connected FROM observability_integrations
WHERE tenant_id = $1 AND integration_key = $2`, tenantID, key).Scan(&configJSON, &connected)
	if err != nil || !connected {
		return nil, fmt.Errorf("%s not connected", key)
	}
	var cfg map[string]string
	_ = json.Unmarshal(configJSON, &cfg)
	return cfg, nil
}

func (e *Executor) createJiraTicketImpl(ctx context.Context, tenantID string, step Step, ctxData map[string]string) (string, error) {
	cfg, err := e.integrationConfig(ctx, tenantID, "jira")
	if err != nil {
		return "", err
	}
	baseURL := cfg["baseUrl"]
	email := cfg["email"]
	token := cfg["apiToken"]
	project := cfg["projectKey"]
	if project == "" {
		project = "OPS"
	}
	if baseURL == "" || token == "" {
		return "", fmt.Errorf("jira config incomplete")
	}
	summary := ctxData["title"]
	if summary == "" {
		summary = step.Config["summary"]
	}
	if summary == "" {
		summary = "NeuralOps incident"
	}
	body := ctxData["description"]
	if body == "" {
		body = step.Config["description"]
	}
	payload := map[string]any{
		"fields": map[string]any{
			"project":     map[string]string{"key": project},
			"summary":     summary,
			"description": body,
			"issuetype":   map[string]string{"name": "Task"},
		},
	}
	raw, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/rest/api/2/issue", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if email != "" {
		req.SetBasicAuth(email, token)
	} else {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("jira api: %s", resp.Status)
	}
	var out struct {
		Key string `json:"key"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return fmt.Sprintf("Jira ticket created: %s", out.Key), nil
}
