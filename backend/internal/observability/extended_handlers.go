package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/cloud"
	"github.com/neuralops/platform/internal/notebook"
)

// RegisterExtendedRoutes mounts P0-P3 production routes.
func (h *Handler) RegisterExtendedRoutes(v1 *gin.RouterGroup) {
	v1.GET("/cloud/metrics", h.QueryCloudMetrics)
	v1.POST("/notebooks/:id/execute", h.ExecuteNotebook)
	v1.POST("/mobile/push/register", h.RegisterPushToken)
	h.RegisterIntegrationOAuthRoutes(v1)

	admin := v1.Group("/admin")
	{
		admin.GET("/sso", h.GetSSOConfig)
		admin.PUT("/sso", h.PutSSOConfig)
		admin.GET("/tenant-policies", h.GetTenantPolicies)
		admin.PUT("/tenant-policies", h.PutTenantPolicies)
		admin.GET("/oncall", h.ListOncallSchedules)
		admin.POST("/oncall", h.CreateOncallSchedule)
		admin.PUT("/oncall/:id", h.UpdateOncallSchedule)
		admin.DELETE("/oncall/:id", h.DeleteOncallSchedule)
		admin.POST("/oncall/:id/sync-pagerduty", h.SyncOncallPagerDuty)
	}
}

func (h *Handler) integrationsRepo() *IntegrationsRepo {
	return NewIntegrationsRepo(h.deps.Pool)
}

func (h *Handler) QueryCloudMetrics(c *gin.Context) {
	provider := c.Query("provider")
	metric := c.Query("metric")
	region := c.Query("region")
	tid := tenantID(c)
	cfg := map[string]string{}
	if repo := h.integrationsRepo(); repo.available() {
		if icfg, err := repo.GetConfig(c.Request.Context(), tid, provider); err == nil {
			cfg = icfg
		}
	}
	series, err := cloud.QueryMetrics(c.Request.Context(), provider, metric, region, cfg)
	if err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	writeSuccess(c, series)
}

func (h *Handler) ExecuteNotebook(c *gin.Context) {
	id := c.Param("id")
	tid := tenantID(c)
	logs := notebook.NewLogBackend(h.deps.SearchURL)
	exec := notebook.NewExecutor(h.deps.Pool, logs, NotebookPromQuerier(h.deps.Prom))
	results, err := exec.ExecuteNotebook(c.Request.Context(), tid, id)
	if err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	writeSuccess(c, results)
}

func (h *Handler) RegisterPushToken(c *gin.Context) {
	var body struct {
		UserID        string `json:"userId"`
		ExpoPushToken string `json:"expoPushToken"`
		Platform      string `json:"platform"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.ExpoPushToken == "" {
		writeError(c, http.StatusBadRequest, "expoPushToken required")
		return
	}
	tid := tenantID(c)
	if h.deps.Pool != nil {
		_, _ = h.deps.Pool.Exec(c.Request.Context(), `
INSERT INTO mobile_push_tokens (tenant_id, user_id, expo_push_token, platform, updated_at)
VALUES ($1,$2,$3,$4,NOW())
ON CONFLICT (tenant_id, expo_push_token) DO UPDATE SET user_id=EXCLUDED.user_id, platform=EXCLUDED.platform, updated_at=NOW()`,
			tid, body.UserID, body.ExpoPushToken, body.Platform)
	}
	writeSuccess(c, gin.H{"registered": true})
}

func (h *Handler) GetSSOConfig(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.Pool == nil {
		writeSuccess(c, gin.H{"provider": "oidc", "issuerUrl": "http://localhost:8088/realms/neuralops"})
		return
	}
	var provider, metadataURL, clientID, issuerURL, clientSecret string
	_ = h.deps.Pool.QueryRow(c.Request.Context(), `
SELECT sso_provider, COALESCE(sso_metadata_url,''), COALESCE(sso_client_id,''), COALESCE(sso_issuer_url,''),
       COALESCE(sso_client_secret,'')
FROM tenant_policies WHERE tenant_id = $1`, tid).Scan(&provider, &metadataURL, &clientID, &issuerURL, &clientSecret)
	writeSuccess(c, gin.H{
		"provider": provider, "metadataUrl": metadataURL, "clientId": clientID, "issuerUrl": issuerURL,
		"clientSecretSet": clientSecret != "",
	})
}

func (h *Handler) PutSSOConfig(c *gin.Context) {
	var body struct {
		Provider     string `json:"provider"`
		MetadataURL  string `json:"metadataUrl"`
		ClientID     string `json:"clientId"`
		IssuerURL    string `json:"issuerUrl"`
		ClientSecret string `json:"clientSecret"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "invalid body")
		return
	}
	tid := tenantID(c)
	if h.deps.Pool != nil {
		if body.ClientSecret == "" {
			_ = h.deps.Pool.QueryRow(c.Request.Context(), `
SELECT COALESCE(sso_client_secret,'') FROM tenant_policies WHERE tenant_id = $1`, tid).Scan(&body.ClientSecret)
		}
		_, _ = h.deps.Pool.Exec(c.Request.Context(), `
INSERT INTO tenant_policies (tenant_id, sso_provider, sso_metadata_url, sso_client_id, sso_issuer_url, sso_client_secret, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,NOW())
ON CONFLICT (tenant_id) DO UPDATE SET
  sso_provider=EXCLUDED.sso_provider, sso_metadata_url=EXCLUDED.sso_metadata_url,
  sso_client_id=EXCLUDED.sso_client_id, sso_issuer_url=EXCLUDED.sso_issuer_url,
  sso_client_secret=CASE WHEN EXCLUDED.sso_client_secret <> '' THEN EXCLUDED.sso_client_secret ELSE tenant_policies.sso_client_secret END,
  updated_at=NOW()`,
			tid, body.Provider, body.MetadataURL, body.ClientID, body.IssuerURL, body.ClientSecret)
	}
	if h.deps.SSOManager != nil {
		_ = h.deps.SSOManager.ReloadTenant(c.Request.Context(), tid)
	}
	writeSuccess(c, gin.H{
		"provider": body.Provider, "metadataUrl": body.MetadataURL, "clientId": body.ClientID, "issuerUrl": body.IssuerURL,
	})
}

func (h *Handler) GetTenantPolicies(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.Pool == nil {
		writeSuccess(c, gin.H{"logRetentionDays": 90, "ingestionRateLimit": 100000})
		return
	}
	var logDays, rateLimit int
	_ = h.deps.Pool.QueryRow(c.Request.Context(), `
SELECT log_retention_days, ingestion_rate_limit FROM tenant_policies WHERE tenant_id = $1`, tid).Scan(&logDays, &rateLimit)
	writeSuccess(c, gin.H{"logRetentionDays": logDays, "ingestionRateLimit": rateLimit})
}

func (h *Handler) PutTenantPolicies(c *gin.Context) {
	var body struct {
		LogRetentionDays   int `json:"logRetentionDays"`
		IngestionRateLimit int `json:"ingestionRateLimit"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "invalid body")
		return
	}
	tid := tenantID(c)
	if h.deps.Pool != nil {
		_, _ = h.deps.Pool.Exec(c.Request.Context(), `
INSERT INTO tenant_policies (tenant_id, log_retention_days, ingestion_rate_limit, updated_at)
VALUES ($1,$2,$3,NOW())
ON CONFLICT (tenant_id) DO UPDATE SET log_retention_days=EXCLUDED.log_retention_days,
  ingestion_rate_limit=EXCLUDED.ingestion_rate_limit, updated_at=NOW()`,
			tid, body.LogRetentionDays, body.IngestionRateLimit)
	}
	writeSuccess(c, body)
}

type oncallSchedule struct {
	ID       string `json:"id"`
	Team     string `json:"team"`
	Timezone string `json:"timezone"`
	Enabled  bool   `json:"enabled"`
	Rotation []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		After string `json:"after"`
	} `json:"rotation"`
}

func (h *Handler) ListOncallSchedules(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.Pool == nil {
		writeSuccess(c, []oncallSchedule{})
		return
	}
	rows, err := h.deps.Pool.Query(c.Request.Context(), `
SELECT id::text, team, timezone, enabled, rotation FROM oncall_schedules WHERE tenant_id = $1`, tid)
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	out := make([]oncallSchedule, 0)
	for rows.Next() {
		var s oncallSchedule
		var rotJSON []byte
		if err := rows.Scan(&s.ID, &s.Team, &s.Timezone, &s.Enabled, &rotJSON); err != nil {
			continue
		}
		_ = json.Unmarshal(rotJSON, &s.Rotation)
		out = append(out, s)
	}
	writeSuccess(c, out)
}

func (h *Handler) CreateOncallSchedule(c *gin.Context) {
	var body oncallSchedule
	if err := c.ShouldBindJSON(&body); err != nil || body.Team == "" {
		writeError(c, http.StatusBadRequest, "team required")
		return
	}
	tid := tenantID(c)
	id := uuid.New().String()
	if h.deps.Pool != nil {
		rotJSON, _ := json.Marshal(body.Rotation)
		_, _ = h.deps.Pool.Exec(c.Request.Context(), `
INSERT INTO oncall_schedules (id, tenant_id, team, rotation, timezone, enabled)
VALUES ($1,$2,$3,$4,$5,$6)`, id, tid, body.Team, rotJSON, body.Timezone, body.Enabled)
	}
	body.ID = id
	writeSuccess(c, body)
}

func (h *Handler) UpdateOncallSchedule(c *gin.Context) {
	id := c.Param("id")
	var body oncallSchedule
	if err := c.ShouldBindJSON(&body); err != nil || body.Team == "" {
		writeError(c, http.StatusBadRequest, "team required")
		return
	}
	tid := tenantID(c)
	if h.deps.Pool == nil {
		writeSuccess(c, body)
		return
	}
	rotJSON, _ := json.Marshal(body.Rotation)
	_, err := h.deps.Pool.Exec(c.Request.Context(), `
UPDATE oncall_schedules SET team=$3, rotation=$4, timezone=$5, enabled=$6, updated_at=NOW()
WHERE tenant_id=$1 AND id=$2`, tid, id, body.Team, rotJSON, body.Timezone, body.Enabled)
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	body.ID = id
	writeSuccess(c, body)
}

func (h *Handler) DeleteOncallSchedule(c *gin.Context) {
	id := c.Param("id")
	tid := tenantID(c)
	if h.deps.Pool != nil {
		_, _ = h.deps.Pool.Exec(c.Request.Context(), `
DELETE FROM oncall_schedules WHERE tenant_id=$1 AND id=$2`, tid, id)
	}
	writeSuccess(c, gin.H{"deleted": true, "id": id})
}

func (h *Handler) SyncOncallPagerDuty(c *gin.Context) {
	id := c.Param("id")
	tid := tenantID(c)
	var body struct {
		ScheduleID string `json:"scheduleId"`
	}
	_ = c.ShouldBindJSON(&body)
	cfg, err := h.integrationsRepo().GetConfig(c.Request.Context(), tid, "pagerduty")
	if err != nil {
		writeError(c, http.StatusBadRequest, "connect PagerDuty first")
		return
	}
	token := firstNonEmptyStr(cfg["apiToken"], cfg["accessToken"])
	if token == "" {
		writeError(c, http.StatusBadRequest, "pagerduty api token required")
		return
	}
	scheduleID := body.ScheduleID
	if scheduleID == "" && h.deps.Pool != nil {
		_ = h.deps.Pool.QueryRow(c.Request.Context(), `
SELECT COALESCE(pagerduty_schedule_id,'') FROM oncall_schedules WHERE tenant_id=$1 AND id=$2`, tid, id).Scan(&scheduleID)
	}
	if scheduleID == "" {
		writeError(c, http.StatusBadRequest, "scheduleId required")
		return
	}
	rotation, err := fetchPagerDutyRotation(c.Request.Context(), token, scheduleID)
	if err != nil {
		writeError(c, http.StatusBadGateway, err.Error())
		return
	}
	if h.deps.Pool != nil {
		rotJSON, _ := json.Marshal(rotation)
		_, _ = h.deps.Pool.Exec(c.Request.Context(), `
UPDATE oncall_schedules SET rotation=$3, pagerduty_schedule_id=$4, pagerduty_synced_at=NOW(), updated_at=NOW()
WHERE tenant_id=$1 AND id=$2`, tid, id, rotJSON, scheduleID)
	}
	writeSuccess(c, gin.H{"id": id, "scheduleId": scheduleID, "rotation": rotation})
}

func fetchPagerDutyRotation(ctx context.Context, token, scheduleID string) ([]struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	After string `json:"after"`
}, error) {
	url := fmt.Sprintf("https://api.pagerduty.com/schedules/%s/users", scheduleID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Token token="+token)
	req.Header.Set("Accept", "application/vnd.pagerduty+json;version=2")
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("pagerduty: %s", string(body))
	}
	var payload struct {
		Users []struct {
			User struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"user"`
		} `json:"users"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	out := make([]struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		After string `json:"after"`
	}, 0, len(payload.Users))
	for i, u := range payload.Users {
		after := "0m"
		if i > 0 {
			after = fmt.Sprintf("%dm", i*30)
		}
		out = append(out, struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			After string `json:"after"`
		}{Name: u.User.Name, Email: u.User.Email, After: after})
	}
	return out, nil
}
