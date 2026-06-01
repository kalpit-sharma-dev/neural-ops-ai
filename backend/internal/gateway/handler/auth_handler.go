package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/gateway/auth"
	"github.com/neuralops/platform/internal/gateway/config"
	platmiddleware "github.com/neuralops/platform/internal/platform/middleware"
	"github.com/neuralops/platform/internal/security"
	"go.uber.org/zap"
)

// AuthHandler exposes authentication endpoints.
type AuthHandler struct {
	log           *zap.Logger
	cfg           config.Config
	issuer        *auth.JWTIssuer
	store         *auth.IdentityStore
	ssoManager    *auth.SSOManager
	saml          auth.SAMLProvider
	authenticator *auth.Authenticator
	audit         *security.AuditRepository
}

// NewAuthHandler creates auth routes handler.
func NewAuthHandler(
	log *zap.Logger,
	cfg config.Config,
	issuer *auth.JWTIssuer,
	store *auth.IdentityStore,
	ssoManager *auth.SSOManager,
	saml auth.SAMLProvider,
	authenticator *auth.Authenticator,
	audit *security.AuditRepository,
) *AuthHandler {
	return &AuthHandler{
		log:           log,
		cfg:           cfg,
		issuer:        issuer,
		store:         store,
		ssoManager:    ssoManager,
		saml:          saml,
		authenticator: authenticator,
		audit:         audit,
	}
}

func (h *AuthHandler) oidcFlow(c *gin.Context) *auth.OIDCFlow {
	if h.ssoManager == nil {
		return nil
	}
	tid := platmiddleware.TenantFromGin(c, h.cfg.Tenant.DefaultTenant)
	return h.ssoManager.Flow(c.Request.Context(), tid)
}

// RegisterRoutes mounts auth endpoints.
func (h *AuthHandler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		v1.GET("/auth/config", h.AuthConfig)
		v1.POST("/auth/oidc/start", h.OIDCStart)
		v1.GET("/auth/oidc/callback", h.OIDCCallback)
		v1.POST("/auth/oidc/exchange", h.OIDCExchange)
		v1.POST("/auth/refresh", h.Refresh)
		v1.GET("/auth/me", h.Me)
		v1.POST("/auth/logout", h.Logout)
		v1.POST("/auth/dev/login", h.DevLogin)
		v1.GET("/auth/saml/metadata", h.SAMLMetadata)
		v1.GET("/auth/saml/login", h.SAMLLogin)
		v1.POST("/auth/saml/acs", h.SAMLACS)
	}
}

type authConfigResponse struct {
	AuthEnabled   bool   `json:"authEnabled"`
	OIDCEnabled   bool   `json:"oidcEnabled"`
	SAMLEnabled   bool   `json:"samlEnabled"`
	DevLogin      bool   `json:"devLoginEnabled"`
	DefaultTenant string `json:"defaultTenant"`
}

// AuthConfig returns client auth configuration.
func (h *AuthHandler) AuthConfig(c *gin.Context) {
	writeSuccess(c, authConfigResponse{
		AuthEnabled:   !h.cfg.Auth.Disabled,
		OIDCEnabled:   h.cfg.Auth.OIDC.Enabled && h.oidcFlow(c) != nil,
		SAMLEnabled:   h.cfg.Auth.SAML.Enabled && h.saml != nil,
		DevLogin:      h.cfg.Auth.AllowDevLogin && h.cfg.Server.Environment != "production",
		DefaultTenant: h.cfg.Tenant.DefaultTenant,
	})
}

type oidcStartResponse struct {
	AuthorizationURL string `json:"authorizationUrl"`
	State            string `json:"state"`
}

// OIDCStart begins PKCE login and returns the IdP authorization URL.
func (h *AuthHandler) OIDCStart(c *gin.Context) {
	if h.oidcFlow(c) == nil || h.store == nil {
		writeError(c, http.StatusServiceUnavailable, "OIDC not configured")
		return
	}
	state, err := auth.GenerateState()
	if err != nil {
		writeError(c, http.StatusInternalServerError, "failed to generate state")
		return
	}
	verifier, challenge, err := auth.GeneratePKCE()
	if err != nil {
		writeError(c, http.StatusInternalServerError, "failed to generate pkce")
		return
	}
	expires := time.Now().UTC().Add(10 * time.Minute)
	if err := h.store.SaveOIDCState(c.Request.Context(), state, verifier, h.cfg.Auth.OIDC.RedirectURL, expires); err != nil {
		writeError(c, http.StatusInternalServerError, "failed to persist oidc state")
		return
	}
	writeSuccess(c, oidcStartResponse{
		AuthorizationURL: h.oidcFlow(c).AuthorizationURL(state, challenge),
		State:            state,
	})
}

// OIDCCallback completes OIDC login and redirects to the SPA with a one-time exchange code.
func (h *AuthHandler) OIDCCallback(c *gin.Context) {
	if h.oidcFlow(c) == nil || h.store == nil {
		writeError(c, http.StatusServiceUnavailable, "OIDC not configured")
		return
	}
	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		writeError(c, http.StatusBadRequest, "missing code or state")
		return
	}
	verifier, _, err := h.store.ConsumeOIDCState(c.Request.Context(), state)
	if err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	email, sub, role, tenantID, err := h.oidcFlow(c).ExchangeCode(c.Request.Context(), code, verifier)
	if err != nil {
		h.log.Warn("oidc exchange failed", zap.Error(err))
		writeError(c, http.StatusUnauthorized, "oidc authentication failed")
		return
	}
	user, err := h.store.UpsertOIDCUser(c.Request.Context(), tenantID, email, role, sub)
	if err != nil {
		h.recordAuthAudit(c, security.AuditActionLoginFailure, "failure", "oidc", tenantID, "")
		writeError(c, http.StatusForbidden, "tenant access denied")
		return
	}
	exchangeCode, err := auth.GenerateExchangeCode()
	if err != nil {
		writeError(c, http.StatusInternalServerError, "failed to generate exchange code")
		return
	}
	expires := time.Now().UTC().Add(2 * time.Minute)
	if err := h.store.SaveExchangeCode(c.Request.Context(), exchangeCode, user.ID, expires); err != nil {
		writeError(c, http.StatusInternalServerError, "failed to persist exchange code")
		return
	}
	redirect := auth.FrontendRedirect(h.cfg.Auth.FrontendURL, exchangeCode)
	h.log.Info("oidc login success", zap.String("email", user.Email), zap.String("tenant", user.TenantID))
	c.Redirect(http.StatusFound, redirect)
}

type oidcExchangeRequest struct {
	Code string `json:"code"`
}

// OIDCExchange swaps a one-time callback code for platform tokens.
func (h *AuthHandler) OIDCExchange(c *gin.Context) {
	if h.store == nil || h.issuer == nil {
		writeError(c, http.StatusServiceUnavailable, "auth store unavailable")
		return
	}
	var req oidcExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Code) == "" {
		writeError(c, http.StatusBadRequest, "code required")
		return
	}
	userID, err := h.store.ConsumeExchangeCode(c.Request.Context(), req.Code)
	if err != nil {
		h.recordAuthAudit(c, security.AuditActionOIDCExchange, "failure", "oidc", h.cfg.Tenant.DefaultTenant, "")
		writeError(c, http.StatusUnauthorized, "invalid exchange code")
		return
	}
	user, err := h.store.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		writeError(c, http.StatusUnauthorized, "user not found")
		return
	}
	tenant, _ := h.store.GetTenant(c.Request.Context(), user.TenantID)
	plan := ""
	tenantName := user.TenantID
	if tenant != nil {
		plan = tenant.PlanTier
		tenantName = tenant.Name
	}
	pair, err := h.issueSession(c, auth.Principal{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Email:    user.Email,
		Role:     user.Role,
		Plan:     plan,
		AuthType: "oidc",
	})
	if err != nil {
		writeError(c, http.StatusInternalServerError, "failed to issue token")
		return
	}
	h.recordAuthAudit(c, security.AuditActionOIDCExchange, "success", "oidc", user.TenantID, user.ID)
	writeSuccess(c, gin.H{
		"accessToken":  pair.AccessToken,
		"refreshToken": pair.RefreshToken,
		"expiresIn":    pair.ExpiresIn,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  displayName(user.Email),
			"role":  user.Role,
		},
		"tenantId":   user.TenantID,
		"tenantName": tenantName,
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// Refresh exchanges a refresh token for a new access token.
func (h *AuthHandler) Refresh(c *gin.Context) {
	if h.store == nil || h.issuer == nil {
		writeError(c, http.StatusServiceUnavailable, "auth store unavailable")
		return
	}
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		writeError(c, http.StatusBadRequest, "refreshToken required")
		return
	}
	userID, err := h.store.ConsumeRefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.recordAuthAudit(c, security.AuditActionTokenRefresh, "failure", "refresh", h.cfg.Tenant.DefaultTenant, "")
		writeError(c, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	user, err := h.store.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		writeError(c, http.StatusUnauthorized, "user not found")
		return
	}
	tenant, _ := h.store.GetTenant(c.Request.Context(), user.TenantID)
	plan := ""
	if tenant != nil {
		plan = tenant.PlanTier
	}
	pair, err := h.issueSession(c, auth.Principal{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Email:    user.Email,
		Role:     user.Role,
		Plan:     plan,
		AuthType: "jwt",
	})
	if err != nil {
		writeError(c, http.StatusInternalServerError, "failed to issue token")
		return
	}
	h.recordAuthAudit(c, security.AuditActionTokenRefresh, "success", "refresh", user.TenantID, user.ID)
	writeSuccess(c, gin.H{
		"accessToken":  pair.AccessToken,
		"refreshToken": pair.RefreshToken,
		"expiresIn":    pair.ExpiresIn,
	})
}

// Me returns the authenticated principal profile.
func (h *AuthHandler) Me(c *gin.Context) {
	principal, ok := auth.PrincipalFromGin(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "authentication required")
		return
	}
	tenantName := principal.TenantID
	if h.store != nil {
		if tenant, err := h.store.GetTenant(c.Request.Context(), principal.TenantID); err == nil {
			tenantName = tenant.Name
		}
	}
	writeSuccess(c, gin.H{
		"user": gin.H{
			"id":    principal.UserID,
			"email": principal.Email,
			"name":  displayName(principal.Email),
			"role":  principal.Role,
		},
		"tenantId":   principal.TenantID,
		"tenantName": tenantName,
		"plan":       principal.Plan,
	})
}

type devLoginRequest struct {
	Email    string `json:"email"`
	TenantID string `json:"tenantId"`
}

// DevLogin issues tokens for a seeded user in non-production environments.
func (h *AuthHandler) DevLogin(c *gin.Context) {
	if !h.cfg.Auth.AllowDevLogin || h.cfg.Server.Environment == "production" {
		writeError(c, http.StatusForbidden, "dev login disabled")
		return
	}
	if h.store == nil || h.issuer == nil {
		writeError(c, http.StatusServiceUnavailable, "auth store unavailable")
		return
	}
	var req devLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
		writeError(c, http.StatusBadRequest, "email required")
		return
	}
	tenantID := req.TenantID
	if tenantID == "" {
		tenantID = h.cfg.Tenant.DefaultTenant
	}
	user, err := h.store.DevLogin(c.Request.Context(), req.Email, tenantID)
	if err != nil {
		h.recordAuthAudit(c, security.AuditActionLoginFailure, "failure", "dev:"+req.Email, tenantID, "")
		writeError(c, http.StatusUnauthorized, "invalid credentials")
		return
	}
	tenant, _ := h.store.GetTenant(c.Request.Context(), user.TenantID)
	plan := ""
	tenantName := tenantID
	if tenant != nil {
		plan = tenant.PlanTier
		tenantName = tenant.Name
	}
	pair, err := h.issueSession(c, auth.Principal{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Email:    user.Email,
		Role:     user.Role,
		Plan:     plan,
		AuthType: "dev",
	})
	if err != nil {
		writeError(c, http.StatusInternalServerError, "failed to issue token")
		return
	}
	h.recordAuthAudit(c, security.AuditActionLoginSuccess, "success", "dev", user.TenantID, user.ID)
	writeSuccess(c, gin.H{
		"accessToken":  pair.AccessToken,
		"refreshToken": pair.RefreshToken,
		"expiresIn":    pair.ExpiresIn,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  displayName(user.Email),
			"role":  user.Role,
		},
		"tenantId":   user.TenantID,
		"tenantName": tenantName,
	})
}

type logoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// Logout revokes refresh tokens for the current session.
func (h *AuthHandler) Logout(c *gin.Context) {
	if h.store == nil {
		writeSuccess(c, gin.H{"loggedOut": true})
		return
	}

	var req logoutRequest
	_ = c.ShouldBindJSON(&req)

	if strings.TrimSpace(req.RefreshToken) != "" {
		if err := h.store.RevokeRefreshToken(c.Request.Context(), req.RefreshToken); err != nil {
			h.log.Warn("refresh token revocation failed", zap.Error(err))
		}
	}

	if principal, ok := auth.PrincipalFromGin(c); ok && principal.UserID != "" {
		if err := h.store.RevokeUserRefreshTokens(c.Request.Context(), principal.UserID); err != nil {
			h.log.Warn("user refresh token revocation failed", zap.Error(err))
		}
		tenantID := principal.TenantID
		if tenantID == "" {
			tenantID = h.cfg.Tenant.DefaultTenant
		}
		h.recordAuthAudit(c, security.AuditActionLogout, "success", "session", tenantID, principal.UserID)
	}

	writeSuccess(c, gin.H{"loggedOut": true})
}

func (h *AuthHandler) recordAuthAudit(c *gin.Context, action, result, resourceID, tenantID, userID string) {
	if h.audit == nil {
		return
	}
	if tenantID == "" {
		tenantID = h.cfg.Tenant.DefaultTenant
	}
	if err := h.audit.Insert(c.Request.Context(), security.AuditEntry{
		UserID:       userID,
		TenantID:     tenantID,
		Action:       action,
		ResourceType: "auth",
		ResourceID:   resourceID,
		IP:           c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
		Result:       result,
	}); err != nil {
		h.log.Warn("auth audit insert failed", zap.Error(err), zap.String("action", action))
	}
}

func (h *AuthHandler) issueSession(c *gin.Context, principal auth.Principal) (auth.TokenPair, error) {
	pair, err := h.issuer.Issue(principal)
	if err != nil {
		return auth.TokenPair{}, err
	}
	if h.store != nil {
		expires := time.Now().UTC().Add(h.issuer.RefreshTokenTTL())
		if err := h.store.SaveRefreshToken(c.Request.Context(), principal.UserID, pair.RefreshToken, expires); err != nil {
			return auth.TokenPair{}, err
		}
	}
	return pair, nil
}

// SAMLMetadata returns SP metadata for IdP federation.
func (h *AuthHandler) SAMLMetadata(c *gin.Context) {
	if h.saml == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "SAML not enabled"})
		return
	}
	h.saml.WriteMetadata(c.Writer)
}

// SAMLLogin redirects to the configured IdP SSO URL.
func (h *AuthHandler) SAMLLogin(c *gin.Context) {
	if h.saml == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "SAML not enabled"})
		return
	}
	url, err := h.saml.LoginURL(c.Query("relayState"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.Redirect(http.StatusFound, url)
}

type samlACSRequest struct {
	SAMLResponse string `form:"SAMLResponse" json:"SAMLResponse"`
	RelayState   string `form:"RelayState" json:"RelayState"`
}

// SAMLACS consumes the IdP POST and issues NeuralOps session tokens.
func (h *AuthHandler) SAMLACS(c *gin.Context) {
	if h.saml == nil || h.store == nil {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "SAML not enabled"})
		return
	}

	var req samlACSRequest
	if err := c.ShouldBind(&req); err != nil || strings.TrimSpace(req.SAMLResponse) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "SAMLResponse required"})
		return
	}

	identity, err := h.saml.ParseACS(c.Request)
	if err != nil {
		h.recordAuthAudit(c, security.AuditActionLoginFailure, "failure", "saml", h.cfg.Tenant.DefaultTenant, "")
		c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "message": err.Error()})
		return
	}

	user, err := h.store.UpsertOIDCUser(c.Request.Context(), identity.TenantID, identity.Email, string(identity.Role), identity.NameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "user provisioning failed"})
		return
	}

	pair, err := h.issueSession(c, auth.Principal{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Email:    user.Email,
		Role:     user.Role,
		AuthType: "saml",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "token issue failed"})
		return
	}

	h.recordAuthAudit(c, security.AuditActionLoginSuccess, "success", "saml", user.TenantID, user.ID)
	redirect := h.cfg.Auth.FrontendURL + "/auth/callback"
	if req.RelayState != "" && strings.HasPrefix(req.RelayState, h.cfg.Auth.FrontendURL) {
		redirect = req.RelayState
	}
	c.Redirect(http.StatusFound, redirect+"?code="+pair.AccessToken)
}

func displayName(email string) string {
	if email == "" {
		return "User"
	}
	at := 0
	for i, ch := range email {
		if ch == '@' {
			at = i
			break
		}
	}
	if at == 0 {
		return email
	}
	return email[:at]
}
