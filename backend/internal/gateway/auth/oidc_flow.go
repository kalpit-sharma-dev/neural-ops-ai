package auth

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/neuralops/platform/internal/gateway/config"
	"golang.org/x/oauth2"
)

// OIDCFlow implements authorization-code flow with PKCE.
type OIDCFlow struct {
	cfg      config.OIDCConfig
	provider *oidc.Provider
	oauth    oauth2.Config
}

// NewOIDCFlow creates an OIDC PKCE flow handler.
func NewOIDCFlow(ctx context.Context, cfg config.OIDCConfig) (*OIDCFlow, error) {
	if !cfg.Enabled || cfg.ClientID == "" {
		return nil, fmt.Errorf("oidc not configured")
	}
	discoveryIssuer := cfg.Issuer
	if cfg.IssuerInternal != "" {
		discoveryIssuer = cfg.IssuerInternal
	}
	if discoveryIssuer == "" {
		return nil, fmt.Errorf("oidc issuer not configured")
	}
	oidcCtx := ctx
	if cfg.Issuer != "" && cfg.Issuer != discoveryIssuer {
		oidcCtx = oidc.InsecureIssuerURLContext(ctx, cfg.Issuer)
	}
	provider, err := oidc.NewProvider(oidcCtx, discoveryIssuer)
	if err != nil {
		return nil, fmt.Errorf("oidc provider: %w", err)
	}
	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = []string{oidc.ScopeOpenID, "profile", "email"}
	}
	oauthCfg := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       scopes,
	}
	return &OIDCFlow{cfg: cfg, provider: provider, oauth: oauthCfg}, nil
}

// AuthorizationURL builds the IdP authorize URL with PKCE parameters.
func (f *OIDCFlow) AuthorizationURL(state, codeChallenge string) string {
	url := f.oauth.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
	return f.rewriteBrowserURL(url)
}

func (f *OIDCFlow) rewriteBrowserURL(raw string) string {
	if f.cfg.BrowserIssuer == "" {
		return raw
	}
	discovery := f.cfg.IssuerInternal
	if discovery == "" {
		discovery = f.cfg.Issuer
	}
	if discovery != "" {
		raw = strings.ReplaceAll(raw, discovery, f.cfg.BrowserIssuer)
	}
	return strings.ReplaceAll(raw, "http://keycloak:", "http://localhost:")
}

// ExchangeCode exchanges an authorization code for tokens and returns user claims.
func (f *OIDCFlow) ExchangeCode(ctx context.Context, code, codeVerifier string) (email, sub, role, tenantID string, err error) {
	token, err := f.oauth.Exchange(
		ctx,
		code,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
	if err != nil {
		return "", "", "", "", fmt.Errorf("token exchange: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return "", "", "", "", fmt.Errorf("missing id_token")
	}
	verifier := f.provider.Verifier(&oidc.Config{ClientID: f.cfg.ClientID})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return "", "", "", "", fmt.Errorf("verify id_token: %w", err)
	}

	claims := map[string]any{}
	if err := idToken.Claims(&claims); err != nil {
		return "", "", "", "", err
	}

	email = stringClaimFromMap(claims, "email")
	sub = stringClaimFromMap(claims, "sub")
	role = stringClaimFromMap(claims, "role")
	tenantID = stringClaimFromMap(claims, "tenant_id", "tenantId", "tid")
	if tenantID == "" {
		tenantID = f.cfg.DefaultTenantID
	}
	return email, sub, role, tenantID, nil
}

// FrontendRedirect builds the SPA callback URL with a one-time exchange code.
func FrontendRedirect(frontendURL, exchangeCode string) string {
	base := strings.TrimRight(frontendURL, "/")
	values := url.Values{}
	values.Set("code", exchangeCode)
	return base + "/auth/callback?" + values.Encode()
}

func stringClaimFromMap(claims map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := claims[key].(string); ok && value != "" {
			return value
		}
	}
	return ""
}

// CleanupExpiredStates removes stale PKCE state rows (best-effort).
func (s *IdentityStore) CleanupExpiredStates(ctx context.Context) {
	if s == nil || s.pool == nil {
		return
	}
	_, _ = s.pool.Exec(ctx, `DELETE FROM oidc_auth_states WHERE expires_at < NOW()`)
	_, _ = s.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE expires_at < NOW()`)
	_, _ = s.pool.Exec(ctx, `DELETE FROM auth_exchange_codes WHERE expires_at < NOW()`)
}

// StartBackgroundCleanup periodically purges expired auth session rows.
func StartBackgroundCleanup(ctx context.Context, store *IdentityStore, interval time.Duration) {
	if store == nil {
		return
	}
	if interval <= 0 {
		interval = 15 * time.Minute
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				store.CleanupExpiredStates(context.Background())
			}
		}
	}()
}
