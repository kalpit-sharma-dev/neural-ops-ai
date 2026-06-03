package observability

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// IntegrationsRepo persists third-party integration connections.
type IntegrationsRepo struct {
	pool *pgxpool.Pool
}

// NewIntegrationsRepo creates an integrations repository.
func NewIntegrationsRepo(pool *pgxpool.Pool) *IntegrationsRepo {
	return &IntegrationsRepo{pool: pool}
}

func (r *IntegrationsRepo) available() bool {
	return r != nil && r.pool != nil
}

// IntegrationRecord is a stored integration with optional public config keys.
type IntegrationRecord struct {
	ID              string            `json:"id"`
	IntegrationKey  string            `json:"integrationKey"`
	Name            string            `json:"name"`
	Type            string            `json:"type"`
	Status          string            `json:"status"`
	Connected       bool              `json:"connected"`
	ConfigPublic    map[string]string `json:"configPublic,omitempty"`
}

// ConnectRequest is the connect/update payload (secrets stored server-side).
type ConnectRequest struct {
	Config map[string]string `json:"config"`
}

// List returns integrations for tenant, seeding defaults when empty.
func (r *IntegrationsRepo) List(ctx context.Context, tenantID string) ([]IntegrationRecord, error) {
	if !r.available() {
		return nil, fmt.Errorf("postgres unavailable")
	}
	if err := r.seedDefaults(ctx, tenantID); err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `
SELECT id::text, integration_key, name, integration_type, connected, config
FROM observability_integrations WHERE tenant_id = $1 ORDER BY name`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]IntegrationRecord, 0)
	for rows.Next() {
		var rec IntegrationRecord
		var configJSON []byte
		var connected bool
		if err := rows.Scan(&rec.ID, &rec.IntegrationKey, &rec.Name, &rec.Type, &connected, &configJSON); err != nil {
			return nil, err
		}
		rec.Connected = connected
		rec.Status = statusLabel(connected)
		rec.ConfigPublic = publicConfig(rec.IntegrationKey, configJSON)
		out = append(out, rec)
	}
	return out, rows.Err()
}

// Connect upserts integration credentials.
func (r *IntegrationsRepo) Connect(ctx context.Context, tenantID, key string, cfg map[string]string) (IntegrationRecord, error) {
	if !r.available() {
		return IntegrationRecord{}, fmt.Errorf("postgres unavailable")
	}
	if err := r.seedDefaults(ctx, tenantID); err != nil {
		return IntegrationRecord{}, err
	}
	configJSON, _ := json.Marshal(cfg)
	var rec IntegrationRecord
	var configOut []byte
	err := r.pool.QueryRow(ctx, `
UPDATE observability_integrations
SET connected = true, config = $3, updated_at = NOW()
WHERE tenant_id = $1 AND (integration_key = $2 OR id::text = $2)
RETURNING id::text, integration_key, name, integration_type, connected, config`,
		tenantID, key, configJSON,
	).Scan(&rec.ID, &rec.IntegrationKey, &rec.Name, &rec.Type, &rec.Connected, &configOut)
	if err != nil {
		return IntegrationRecord{}, fmt.Errorf("integration not found: %s", key)
	}
	rec.Status = statusLabel(rec.Connected)
	rec.ConfigPublic = publicConfig(rec.IntegrationKey, configOut)
	return rec, nil
}

// GetConfig returns full config for internal use (workflows, cloud).
func (r *IntegrationsRepo) GetConfig(ctx context.Context, tenantID, key string) (map[string]string, error) {
	if !r.available() {
		return nil, fmt.Errorf("postgres unavailable")
	}
	var configJSON []byte
	var connected bool
	err := r.pool.QueryRow(ctx, `
SELECT config, connected FROM observability_integrations
WHERE tenant_id = $1 AND integration_key = $2`, tenantID, key).Scan(&configJSON, &connected)
	if err != nil || !connected {
		return nil, fmt.Errorf("integration not connected")
	}
	var cfg map[string]string
	_ = json.Unmarshal(configJSON, &cfg)
	return cfg, nil
}

func (r *IntegrationsRepo) seedDefaults(ctx context.Context, tenantID string) error {
	defaults := []struct{ key, name, typ string }{
		{"jira", "Jira", "jira"},
		{"slack", "Slack", "slack"},
		{"pagerduty", "PagerDuty", "pagerduty"},
		{"servicenow", "ServiceNow", "servicenow"},
		{"github", "GitHub", "github"},
		{"gitlab", "GitLab", "gitlab"},
		{"jenkins", "Jenkins", "jenkins"},
		{"aws", "Amazon Web Services", "cloud"},
		{"azure", "Microsoft Azure", "cloud"},
		{"gcp", "Google Cloud Platform", "cloud"},
	}
	for _, d := range defaults {
		_, err := r.pool.Exec(ctx, `
INSERT INTO observability_integrations (id, tenant_id, integration_key, name, integration_type, config, connected)
VALUES ($1, $2, $3, $4, $5, '{}', false)
ON CONFLICT (tenant_id, integration_key) DO NOTHING`,
			uuid.New(), tenantID, d.key, d.name, d.typ)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetPartialConfig returns config without requiring connected status (for OAuth setup).
func (r *IntegrationsRepo) GetPartialConfig(ctx context.Context, tenantID, key string) (map[string]string, error) {
	if !r.available() {
		return nil, fmt.Errorf("postgres unavailable")
	}
	_ = r.seedDefaults(ctx, tenantID)
	var configJSON []byte
	err := r.pool.QueryRow(ctx, `
SELECT config FROM observability_integrations WHERE tenant_id = $1 AND integration_key = $2`, tenantID, key).Scan(&configJSON)
	if err != nil {
		return nil, err
	}
	var cfg map[string]string
	_ = json.Unmarshal(configJSON, &cfg)
	return cfg, nil
}

func statusLabel(connected bool) string {
	if connected {
		return "connected"
	}
	return "disconnected"
}

func publicConfig(key string, raw []byte) map[string]string {
	var cfg map[string]string
	_ = json.Unmarshal(raw, &cfg)
	if cfg == nil {
		return nil
	}
	out := map[string]string{}
	secretKeys := map[string]bool{
		"apiToken": true, "api_token": true, "secretAccessKey": true,
		"clientSecret": true, "privateKey": true, "webhookUrl": true, "webhook_url": true,
	}
	for k, v := range cfg {
		if secretKeys[k] {
			if len(v) > 4 {
				out[k] = v[:2] + "…" + v[len(v)-2:]
			} else {
				out[k] = "••••"
			}
		} else {
			out[k] = v
		}
	}
	if key == "jira" && out["baseUrl"] == "" && cfg["baseUrl"] != "" {
		out["baseUrl"] = cfg["baseUrl"]
	}
	return out
}
