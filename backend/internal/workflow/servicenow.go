package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (e *Executor) createServiceNowIncident(ctx context.Context, tenantID string, step Step, ctxData map[string]string) (string, error) {
	cfg, err := e.integrationConfig(ctx, tenantID, "servicenow")
	if err != nil {
		return "", err
	}
	baseURL := cfg["baseUrl"]
	if baseURL == "" {
		baseURL = cfg["instanceUrl"]
	}
	user := cfg["username"]
	if user == "" {
		user = cfg["user"]
	}
	pass := cfg["password"]
	if pass == "" {
		pass = cfg["apiToken"]
	}
	table := cfg["table"]
	if table == "" {
		table = "incident"
	}
	if baseURL == "" || user == "" || pass == "" {
		return "", fmt.Errorf("servicenow config incomplete (baseUrl, username, password required)")
	}

	shortDesc := ctxData["title"]
	if shortDesc == "" {
		shortDesc = step.Config["short_description"]
	}
	if shortDesc == "" {
		shortDesc = step.Label
	}
	desc := ctxData["description"]
	if desc == "" {
		desc = step.Config["description"]
	}
	urgency := step.Config["urgency"]
	if urgency == "" {
		urgency = "2"
	}
	impact := step.Config["impact"]
	if impact == "" {
		impact = "2"
	}

	payload := map[string]any{
		"short_description": shortDesc,
		"description":       desc,
		"urgency":           urgency,
		"impact":            impact,
	}
	if cat := step.Config["category"]; cat != "" {
		payload["category"] = cat
	}
	if assignment := step.Config["assignment_group"]; assignment != "" {
		payload["assignment_group"] = assignment
	}

	raw, _ := json.Marshal(payload)
	url := strings.TrimRight(baseURL, "/") + "/api/now/table/" + table
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(user, pass)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("servicenow api: %s", resp.Status)
	}
	var out struct {
		Result struct {
			Number string `json:"number"`
			SysID  string `json:"sys_id"`
		} `json:"result"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	if out.Result.Number != "" {
		return fmt.Sprintf("ServiceNow incident created: %s", out.Result.Number), nil
	}
	if out.Result.SysID != "" {
		return fmt.Sprintf("ServiceNow incident created: %s", out.Result.SysID), nil
	}
	return "ServiceNow incident created", nil
}
