package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type oauthProvider struct {
	AuthURL  string
	TokenURL string
	Scopes   string
}

var oauthProviders = map[string]oauthProvider{
	"jira": {
		AuthURL:  "https://auth.atlassian.com/authorize",
		TokenURL: "https://auth.atlassian.com/oauth/token",
		Scopes:   "read:jira-work write:jira-work offline_access",
	},
	"slack": {
		AuthURL:  "https://slack.com/oauth/v2/authorize",
		TokenURL: "https://slack.com/api/oauth.v2.access",
		Scopes:   "incoming-webhook,chat:write",
	},
}

// RegisterIntegrationOAuthRoutes mounts OAuth connect flows.
func (h *Handler) RegisterIntegrationOAuthRoutes(v1 *gin.RouterGroup) {
	v1.GET("/integrations/:id/oauth/start", h.StartIntegrationOAuth)
	v1.GET("/integrations/oauth/callback", h.IntegrationOAuthCallback)
	v1.POST("/integrations/:id/disconnect", h.DisconnectIntegration)
}

func (h *Handler) StartIntegrationOAuth(c *gin.Context) {
	key := c.Param("id")
	tid := tenantID(c)
	cfg, err := h.integrationsRepo().GetPartialConfig(c.Request.Context(), tid, key)
	if err != nil {
		writeError(c, http.StatusBadRequest, "integration not found")
		return
	}
	provider, ok := resolveOAuthProvider(key, cfg)
	if !ok {
		writeError(c, http.StatusBadRequest, "oauth not supported for this integration")
		return
	}
	clientID := firstNonEmptyStr(cfg["oauthClientId"], cfg["clientId"])
	if clientID == "" {
		writeError(c, http.StatusBadRequest, "set oauthClientId in integration config first")
		return
	}
	state, _ := randomState()
	redirectURI := absoluteURL(c, "/api/v1/integrations/oauth/callback")
	if h.deps.Pool != nil {
		_, _ = h.deps.Pool.Exec(c.Request.Context(), `
INSERT INTO integration_oauth_states (state, tenant_id, integration_key, redirect_uri, expires_at)
VALUES ($1,$2,$3,$4,NOW() + INTERVAL '15 minutes')`, state, tid, key, redirectURI)
	}
	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("scope", provider.Scopes)
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)
	q.Set("response_type", "code")
	if key == "jira" {
		q.Set("audience", "api.atlassian.com")
		q.Set("prompt", "consent")
	}
	c.Redirect(http.StatusFound, provider.AuthURL+"?"+q.Encode())
}

func resolveOAuthProvider(key string, cfg map[string]string) (oauthProvider, bool) {
	if p, ok := oauthProviders[key]; ok {
		return p, true
	}
	if key == "servicenow" {
		base := strings.TrimRight(cfg["instanceUrl"], "/")
		if base == "" {
			return oauthProvider{}, false
		}
		return oauthProvider{
			AuthURL:  base + "/oauth_auth.do",
			TokenURL: base + "/oauth_token.do",
			Scopes:   "useraccount",
		}, true
	}
	return oauthProvider{}, false
}

func (h *Handler) IntegrationOAuthCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		writeError(c, http.StatusBadRequest, "missing code or state")
		return
	}
	var tenantID, integrationKey, redirectURI string
	if h.deps.Pool != nil {
		err := h.deps.Pool.QueryRow(c.Request.Context(), `
SELECT tenant_id, integration_key, redirect_uri FROM integration_oauth_states
WHERE state = $1 AND expires_at > NOW()`, state).Scan(&tenantID, &integrationKey, &redirectURI)
		if err != nil {
			writeError(c, http.StatusBadRequest, "invalid or expired oauth state")
			return
		}
		_, _ = h.deps.Pool.Exec(c.Request.Context(), `DELETE FROM integration_oauth_states WHERE state = $1`, state)
	}
	cfg, _ := h.integrationsRepo().GetPartialConfig(c.Request.Context(), tenantID, integrationKey)
	if cfg == nil {
		cfg = map[string]string{}
	}
	provider, ok := resolveOAuthProvider(integrationKey, cfg)
	if !ok {
		writeError(c, http.StatusBadGateway, "oauth provider not configured")
		return
	}
	clientID := firstNonEmptyStr(cfg["oauthClientId"], cfg["clientId"])
	clientSecret := cfg["oauthClientSecret"]
	tokenCfg, err := exchangeOAuthCode(c.Request.Context(), provider.TokenURL, clientID, clientSecret, code, redirectURI)
	if err != nil {
		writeError(c, http.StatusBadGateway, err.Error())
		return
	}
	for k, v := range tokenCfg {
		cfg[k] = v
	}
	if webhook := tokenCfg["incoming_webhook_url"]; webhook != "" {
		cfg["webhookUrl"] = webhook
	}
	if _, err := h.integrationsRepo().Connect(c.Request.Context(), tenantID, integrationKey, cfg); err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.Redirect(http.StatusFound, "/integrations?oauth=success")
}

func (h *Handler) DisconnectIntegration(c *gin.Context) {
	key := c.Param("id")
	tid := tenantID(c)
	if !h.integrationsRepo().available() {
		writeError(c, http.StatusServiceUnavailable, "postgres unavailable")
		return
	}
	_, err := h.deps.Pool.Exec(c.Request.Context(), `
UPDATE observability_integrations SET connected = false, config = '{}', updated_at = NOW()
WHERE tenant_id = $1 AND (integration_key = $2 OR id::text = $2)`, tid, key)
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(c, gin.H{"disconnected": true, "integrationKey": key})
}

func exchangeOAuthCode(ctx context.Context, tokenURL, clientID, clientSecret, code, redirectURI string) (map[string]string, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("token exchange: %s", string(body))
	}
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	out := map[string]string{}
	if v, ok := raw["access_token"].(string); ok {
		out["accessToken"] = v
		out["apiToken"] = v
	}
	if v, ok := raw["refresh_token"].(string); ok {
		out["refreshToken"] = v
	}
	if wh, ok := raw["incoming_webhook"].(map[string]any); ok {
		if u, ok := wh["url"].(string); ok {
			out["incoming_webhook_url"] = u
		}
	}
	return out, nil
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return uuid.NewString(), nil
	}
	return hex.EncodeToString(b), nil
}

func absoluteURL(c *gin.Context, path string) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if fwd := c.GetHeader("X-Forwarded-Proto"); fwd != "" {
		scheme = fwd
	}
	host := c.Request.Host
	if host == "" {
		host = "localhost:8080"
	}
	return scheme + "://" + host + path
}

func firstNonEmptyStr(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
