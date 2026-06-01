package auth

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neuralops/platform/internal/gateway/config"
)

// SSOManager holds per-tenant OIDC flows with env fallback.
type SSOManager struct {
	mu            sync.RWMutex
	flows         map[string]*OIDCFlow
	pool          *pgxpool.Pool
	fallback      config.OIDCConfig
	defaultTenant string
}

// NewSSOManager creates a multi-tenant SSO manager.
func NewSSOManager(pool *pgxpool.Pool, fallback config.OIDCConfig, defaultTenant string, initial *OIDCFlow) *SSOManager {
	m := &SSOManager{
		pool:          pool,
		fallback:      fallback,
		defaultTenant: defaultTenant,
		flows:         map[string]*OIDCFlow{},
	}
	if initial != nil && fallback.Enabled && defaultTenant != "" {
		m.flows[defaultTenant] = initial
	}
	return m
}

// Flow returns OIDC flow for a tenant (lazy-load from DB).
func (m *SSOManager) Flow(ctx context.Context, tenantID string) *OIDCFlow {
	if tenantID == "" {
		tenantID = m.defaultTenant
	}
	m.mu.RLock()
	if f, ok := m.flows[tenantID]; ok {
		m.mu.RUnlock()
		return f
	}
	m.mu.RUnlock()
	_ = m.ReloadTenant(ctx, tenantID)
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.flows[tenantID]
}

// ReloadTenant rebuilds OIDC for one tenant org.
func (m *SSOManager) ReloadTenant(ctx context.Context, tenantID string) error {
	cfg := m.fallback
	if m.pool != nil && tenantID != "" {
		var provider, metadataURL, clientID, issuerURL, clientSecret string
		err := m.pool.QueryRow(ctx, `
SELECT COALESCE(sso_provider,''), COALESCE(sso_metadata_url,''), COALESCE(sso_client_id,''),
       COALESCE(sso_issuer_url,''), COALESCE(sso_client_secret,'')
FROM tenant_policies WHERE tenant_id = $1`, tenantID).Scan(&provider, &metadataURL, &clientID, &issuerURL, &clientSecret)
		if err == nil && issuerURL != "" {
			cfg = config.OIDCConfig{
				Enabled:        true,
				Issuer:         issuerURL,
				ClientID:       clientID,
				ClientSecret:   clientSecret,
				RedirectURL:    m.fallback.RedirectURL,
				Scopes:         m.fallback.Scopes,
				BrowserIssuer:  m.fallback.BrowserIssuer,
				IssuerInternal: m.fallback.IssuerInternal,
			}
			if metadataURL != "" && cfg.Issuer == "" {
				cfg.Issuer = metadataURL
			}
			_ = provider
		}
	}
	var flow *OIDCFlow
	if cfg.Enabled && cfg.ClientID != "" && cfg.Issuer != "" {
		if f, err := NewOIDCFlow(ctx, cfg); err == nil {
			flow = f
		}
	}
	m.mu.Lock()
	if flow != nil {
		m.flows[tenantID] = flow
	} else {
		delete(m.flows, tenantID)
	}
	m.mu.Unlock()
	return nil
}

// ReloadAll loads SSO config for every tenant in tenant_policies.
func (m *SSOManager) ReloadAll(ctx context.Context) error {
	if m.pool == nil {
		return m.ReloadTenant(ctx, m.defaultTenant)
	}
	rows, err := m.pool.Query(ctx, `SELECT tenant_id FROM tenant_policies`)
	if err != nil {
		return m.ReloadTenant(ctx, m.defaultTenant)
	}
	defer rows.Close()
	for rows.Next() {
		var tid string
		if err := rows.Scan(&tid); err == nil {
			_ = m.ReloadTenant(ctx, tid)
		}
	}
	return rows.Err()
}

// Reload is an alias for ReloadAll (backward compatible).
func (m *SSOManager) Reload(ctx context.Context) error {
	return m.ReloadAll(ctx)
}
