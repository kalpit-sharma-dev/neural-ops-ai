package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MarketplaceConfigField declares an input required to install an extension.
type MarketplaceConfigField struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Secret bool   `json:"secret,omitempty"`
}

// MarketplaceExtension is a catalog entry merged with per-tenant install state.
type MarketplaceExtension struct {
	Key          string                   `json:"key"`
	Name         string                   `json:"name"`
	Category     string                   `json:"category"`
	Description  string                   `json:"description"`
	Publisher    string                   `json:"publisher"`
	Version      string                   `json:"version"`
	Installed    bool                     `json:"installed"`
	ConfigFields []MarketplaceConfigField `json:"configFields,omitempty"`
	ConfigPublic map[string]string        `json:"configPublic,omitempty"`
}

// marketplaceCatalog is the curated, code-defined set of installable add-ons.
// Install state is stored per tenant; the catalog itself is immutable config.
func marketplaceCatalog() []MarketplaceExtension {
	return []MarketplaceExtension{
		{
			Key: "grafana-panels", Name: "Grafana Panels", Category: "visualization",
			Publisher: "Grafana Labs", Version: "2.4.0",
			Description: "Embed Grafana dashboards and panels directly inside NeuralOps views.",
			ConfigFields: []MarketplaceConfigField{
				{Key: "grafanaUrl", Label: "Grafana base URL"},
				{Key: "apiToken", Label: "API token", Secret: true},
			},
		},
		{
			Key: "datadog-importer", Name: "Datadog Importer", Category: "datasource",
			Publisher: "NeuralOps", Version: "1.1.2",
			Description: "Import dashboards, monitors and metrics from an existing Datadog account.",
			ConfigFields: []MarketplaceConfigField{
				{Key: "site", Label: "Datadog site (e.g. datadoghq.eu)"},
				{Key: "apiKey", Label: "API key", Secret: true},
				{Key: "appKey", Label: "Application key", Secret: true},
			},
		},
		{
			Key: "otel-collector", Name: "OpenTelemetry Collector", Category: "datasource",
			Publisher: "CNCF", Version: "0.102.0",
			Description: "Managed OTLP collector pipeline for traces, metrics and logs ingestion.",
		},
		{
			Key: "statuspage-publisher", Name: "Statuspage Publisher", Category: "app",
			Publisher: "Atlassian", Version: "3.0.1",
			Description: "Automatically publish incident updates to a public status page.",
			ConfigFields: []MarketplaceConfigField{
				{Key: "pageId", Label: "Status page ID"},
				{Key: "apiKey", Label: "API key", Secret: true},
			},
		},
		{
			Key: "slack-digest", Name: "Slack Daily Digest", Category: "app",
			Publisher: "NeuralOps", Version: "1.0.4",
			Description: "Post a daily reliability digest (SLOs, incidents, top errors) to a Slack channel.",
			ConfigFields: []MarketplaceConfigField{
				{Key: "channel", Label: "Slack channel"},
			},
		},
		{
			Key: "snowflake-export", Name: "Snowflake Exporter", Category: "datasource",
			Publisher: "NeuralOps", Version: "0.9.0",
			Description: "Stream long-term logs and metrics into a Snowflake warehouse for analytics.",
			ConfigFields: []MarketplaceConfigField{
				{Key: "account", Label: "Snowflake account"},
				{Key: "warehouse", Label: "Warehouse"},
				{Key: "credentials", Label: "Credentials JSON", Secret: true},
			},
		},
	}
}

// catalogByKey indexes the catalog for validation and field lookups.
func catalogByKey() map[string]MarketplaceExtension {
	out := make(map[string]MarketplaceExtension)
	for _, e := range marketplaceCatalog() {
		out[e.Key] = e
	}
	return out
}

// maskExtensionConfig returns config with secret fields masked for display.
func maskExtensionConfig(ext MarketplaceExtension, cfg map[string]string) map[string]string {
	if len(cfg) == 0 {
		return nil
	}
	secret := make(map[string]bool)
	for _, f := range ext.ConfigFields {
		if f.Secret {
			secret[f.Key] = true
		}
	}
	out := make(map[string]string, len(cfg))
	for k, v := range cfg {
		if secret[k] {
			if len(v) > 4 {
				out[k] = v[:2] + "…" + v[len(v)-2:]
			} else if v != "" {
				out[k] = "••••"
			}
			continue
		}
		out[k] = v
	}
	return out
}

// MarketplaceRepo persists per-tenant extension installs.
type MarketplaceRepo struct {
	pool *pgxpool.Pool
}

// NewMarketplaceRepo creates a marketplace repository.
func NewMarketplaceRepo(pool *pgxpool.Pool) *MarketplaceRepo {
	return &MarketplaceRepo{pool: pool}
}

func (r *MarketplaceRepo) available() bool { return r != nil && r.pool != nil }

// List returns the full catalog merged with this tenant's install state.
func (r *MarketplaceRepo) List(ctx context.Context, tenantID string) ([]MarketplaceExtension, error) {
	installs := map[string]map[string]string{}
	if r.available() {
		rows, err := r.pool.Query(ctx, `
SELECT extension_key, config FROM marketplace_installs WHERE tenant_id = $1`, tenantID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var key string
			var cfgJSON []byte
			if err := rows.Scan(&key, &cfgJSON); err != nil {
				return nil, err
			}
			cfg := map[string]string{}
			_ = json.Unmarshal(cfgJSON, &cfg)
			installs[key] = cfg
		}
	}
	return mergeCatalog(installs), nil
}

// Install upserts an extension's install state and config for a tenant.
func (r *MarketplaceRepo) Install(ctx context.Context, tenantID, key string, cfg map[string]string) error {
	if !r.available() {
		return fmt.Errorf("postgres unavailable")
	}
	cfgJSON, _ := json.Marshal(cfg)
	_, err := r.pool.Exec(ctx, `
INSERT INTO marketplace_installs (tenant_id, extension_key, config, installed_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (tenant_id, extension_key) DO UPDATE SET config = EXCLUDED.config`,
		tenantID, key, cfgJSON)
	return err
}

// Uninstall removes an extension install for a tenant.
func (r *MarketplaceRepo) Uninstall(ctx context.Context, tenantID, key string) error {
	if !r.available() {
		return fmt.Errorf("postgres unavailable")
	}
	_, err := r.pool.Exec(ctx, `
DELETE FROM marketplace_installs WHERE tenant_id = $1 AND extension_key = $2`, tenantID, key)
	return err
}

// mergeCatalog overlays install state (key -> config) onto the static catalog.
func mergeCatalog(installs map[string]map[string]string) []MarketplaceExtension {
	catalog := marketplaceCatalog()
	out := make([]MarketplaceExtension, 0, len(catalog))
	for _, ext := range catalog {
		if cfg, ok := installs[ext.Key]; ok {
			ext.Installed = true
			ext.ConfigPublic = maskExtensionConfig(ext, cfg)
		}
		out = append(out, ext)
	}
	return out
}

// --- HTTP handlers -------------------------------------------------------

func (h *Handler) marketplaceRepo() *MarketplaceRepo {
	return NewMarketplaceRepo(h.deps.Pool)
}

// ListMarketplace returns the catalog with install state (PG, falling back to mem).
func (h *Handler) ListMarketplace(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.Pool != nil {
		if list, err := h.marketplaceRepo().List(c.Request.Context(), tid); err == nil {
			writeSuccess(c, list)
			return
		}
	}
	writeSuccess(c, mergeCatalog(h.deps.Mem.ListExtensionInstalls()))
}

// InstallExtension installs (or reconfigures) a catalog extension for the tenant.
func (h *Handler) InstallExtension(c *gin.Context) {
	key := c.Param("key")
	ext, ok := catalogByKey()[key]
	if !ok {
		writeError(c, http.StatusNotFound, "extension not found")
		return
	}
	var body struct {
		Config map[string]string `json:"config"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.Config == nil {
		body.Config = map[string]string{}
	}
	if h.deps.Pool != nil {
		if err := h.marketplaceRepo().Install(c.Request.Context(), tenantID(c), key, body.Config); err == nil {
			writeSuccess(c, installedResponse(ext, body.Config))
			return
		}
	}
	h.deps.Mem.InstallExtension(key, body.Config)
	writeSuccess(c, installedResponse(ext, body.Config))
}

// UninstallExtension removes an extension install for the tenant.
func (h *Handler) UninstallExtension(c *gin.Context) {
	key := c.Param("key")
	if _, ok := catalogByKey()[key]; !ok {
		writeError(c, http.StatusNotFound, "extension not found")
		return
	}
	if h.deps.Pool != nil {
		if err := h.marketplaceRepo().Uninstall(c.Request.Context(), tenantID(c), key); err == nil {
			writeSuccess(c, gin.H{"uninstalled": true, "key": key})
			return
		}
	}
	h.deps.Mem.UninstallExtension(key)
	writeSuccess(c, gin.H{"uninstalled": true, "key": key})
}

func installedResponse(ext MarketplaceExtension, cfg map[string]string) MarketplaceExtension {
	ext.Installed = true
	ext.ConfigPublic = maskExtensionConfig(ext, cfg)
	return ext
}
