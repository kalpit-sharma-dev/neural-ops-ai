//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neuralops/platform/internal/gateway/auth"
	gatewayconfig "github.com/neuralops/platform/internal/gateway/config"
	gatewayhandler "github.com/neuralops/platform/internal/gateway/handler"
	"github.com/neuralops/platform/internal/platform/db"
	"github.com/neuralops/platform/internal/security"
	"github.com/neuralops/platform/internal/seed"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

const demoUserID = "00000000-0000-0000-0000-000000000101"

type authTestEnv struct {
	pool   *pgxpool.Pool
	router *gin.Engine
	cfg    gatewayconfig.Config
}

func setupAuthPostgres(t *testing.T, ctx context.Context) (*pgxpool.Pool, string) {
	t.Helper()
	skipUnlessDocker(t)

	pgContainer, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("neuralops"),
		postgres.WithUsername("neuralops"),
		postgres.WithPassword("neuralops"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp").WithStartupTimeout(2*time.Minute)),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = pgContainer.Terminate(ctx) })

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := db.NewPool(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	require.NoError(t, db.RunMigrations(ctx, pool, db.DefaultMigrationPaths()...))
	require.NoError(t, seedDemoAuthData(ctx, pool))
	return pool, dsn
}

func seedDemoAuthData(ctx context.Context, pool *pgxpool.Pool) error {
	tenantID := uuid.MustParse(seed.DemoTenantID)
	userID := uuid.MustParse(demoUserID)
	if _, err := pool.Exec(ctx, `
INSERT INTO tenants (id, name, plan_tier, subscription_status)
VALUES ($1, $2, 'enterprise', 'active')
ON CONFLICT (id) DO UPDATE SET subscription_status = 'active'`, tenantID, seed.DemoTenantName); err != nil {
		return err
	}
	if _, err := pool.Exec(ctx, `
INSERT INTO users (id, tenant_id, email, role)
VALUES ($1, $2, $3, 'ADMIN')
ON CONFLICT (tenant_id, email) DO UPDATE SET role = EXCLUDED.role`,
		userID, tenantID, seed.DemoEmail,
	); err != nil {
		return err
	}
	keyHash := security.HashAPIKey("demo-api-key")
	_, err := pool.Exec(ctx, `
INSERT INTO api_keys (id, tenant_id, key_hash, name, role)
VALUES ($1, $2, $3, 'Demo API Key', 'ADMIN')
ON CONFLICT (key_hash) DO NOTHING`,
		uuid.MustParse("00000000-0000-0000-0000-000000000201"), tenantID, keyHash,
	)
	return err
}

func newAuthTestEnv(t *testing.T, pool *pgxpool.Pool) *authTestEnv {
	t.Helper()
	gin.SetMode(gin.TestMode)

	cfg := gatewayconfig.Config{
		Server: gatewayconfig.ServerConfig{
			Environment: "development",
		},
		Auth: gatewayconfig.AuthConfig{
			Disabled:      false,
			AllowDevLogin: true,
			JWTIssuer:     "neuralops-test",
			JWTAudience:   "neuralops-api",
			AccessTokenTTL:  time.Hour,
			RefreshTokenTTL: 24 * time.Hour,
		},
		Tenant: gatewayconfig.TenantConfig{
			DefaultTenant: seed.DemoTenantID,
		},
	}

	store := auth.NewIdentityStore(pool)
	authenticator, err := auth.NewAuthenticator(cfg, store)
	require.NoError(t, err)

	auditRepo := security.NewAuditRepository(pool)
	handler := gatewayhandler.NewAuthHandler(
		zap.NewNop(),
		cfg,
		authenticator.Issuer(),
		store,
		nil,
		nil,
		authenticator,
		auditRepo,
	)

	router := gin.New()
	router.Use(authenticator.Middleware())
	handler.RegisterRoutes(router)

	return &authTestEnv{pool: pool, router: router, cfg: cfg}
}

func TestAuthDevLoginSessionLifecycle(t *testing.T) {
	ctx := context.Background()
	pool, _ := setupAuthPostgres(t, ctx)
	env := newAuthTestEnv(t, pool)

	beforeFailures := countAuditLogs(ctx, pool, security.AuditActionLoginFailure, "failure")
	resp := env.postJSON(t, "/api/v1/auth/dev/login", map[string]string{
		"email": "unknown@example.com",
	})
	require.Equal(t, http.StatusUnauthorized, resp.Code)
	require.Equal(t, beforeFailures+1, countAuditLogs(ctx, pool, security.AuditActionLoginFailure, "failure"))

	login := env.postJSON(t, "/api/v1/auth/dev/login", map[string]string{
		"email": seed.DemoEmail,
	})
	require.Equal(t, http.StatusOK, login.Code)

	var loginBody struct {
		Data struct {
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(login.Body.Bytes(), &loginBody))
	require.NotEmpty(t, loginBody.Data.AccessToken)
	require.NotEmpty(t, loginBody.Data.RefreshToken)
	require.GreaterOrEqual(t, countAuditLogs(ctx, pool, security.AuditActionLoginSuccess, "success"), 1)

	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+loginBody.Data.AccessToken)
	meResp := httptest.NewRecorder()
	env.router.ServeHTTP(meResp, meReq)
	require.Equal(t, http.StatusOK, meResp.Code)

	refresh := env.postJSON(t, "/api/v1/auth/refresh", map[string]string{
		"refreshToken": loginBody.Data.RefreshToken,
	})
	require.Equal(t, http.StatusOK, refresh.Code)

	var refreshBody struct {
		Data struct {
			AccessToken  string `json:"accessToken"`
			RefreshToken string `json:"refreshToken"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(refresh.Body.Bytes(), &refreshBody))
	require.NotEmpty(t, refreshBody.Data.RefreshToken)
	require.GreaterOrEqual(t, countAuditLogs(ctx, pool, security.AuditActionTokenRefresh, "success"), 1)

	staleRefresh := env.postJSON(t, "/api/v1/auth/refresh", map[string]string{
		"refreshToken": loginBody.Data.RefreshToken,
	})
	require.Equal(t, http.StatusUnauthorized, staleRefresh.Code)

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewBufferString(`{"refreshToken":"`+refreshBody.Data.RefreshToken+`"}`))
	logoutReq.Header.Set("Content-Type", "application/json")
	logoutReq.Header.Set("Authorization", "Bearer "+refreshBody.Data.AccessToken)
	logoutResp := httptest.NewRecorder()
	env.router.ServeHTTP(logoutResp, logoutReq)
	require.Equal(t, http.StatusOK, logoutResp.Code)
	require.GreaterOrEqual(t, countAuditLogs(ctx, pool, security.AuditActionLogout, "success"), 1)

	afterLogout := env.postJSON(t, "/api/v1/auth/refresh", map[string]string{
		"refreshToken": refreshBody.Data.RefreshToken,
	})
	require.Equal(t, http.StatusUnauthorized, afterLogout.Code)
}

func TestAuthOIDCExchangeCodeFlow(t *testing.T) {
	ctx := context.Background()
	pool, _ := setupAuthPostgres(t, ctx)
	env := newAuthTestEnv(t, pool)
	store := auth.NewIdentityStore(pool)

	code, err := auth.GenerateExchangeCode()
	require.NoError(t, err)
	require.NoError(t, store.SaveExchangeCode(ctx, code, demoUserID, time.Now().UTC().Add(2*time.Minute)))

	exchange := env.postJSON(t, "/api/v1/auth/oidc/exchange", map[string]string{"code": code})
	require.Equal(t, http.StatusOK, exchange.Code)

	var exchangeBody struct {
		Data struct {
			AccessToken string `json:"accessToken"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(exchange.Body.Bytes(), &exchangeBody))
	require.NotEmpty(t, exchangeBody.Data.AccessToken)
	require.GreaterOrEqual(t, countAuditLogs(ctx, pool, security.AuditActionOIDCExchange, "success"), 1)

	replay := env.postJSON(t, "/api/v1/auth/oidc/exchange", map[string]string{"code": code})
	require.Equal(t, http.StatusUnauthorized, replay.Code)
	require.GreaterOrEqual(t, countAuditLogs(ctx, pool, security.AuditActionOIDCExchange, "failure"), 1)
}

func TestAuthValidateAPIKeyAgainstPostgres(t *testing.T) {
	ctx := context.Background()
	pool, _ := setupAuthPostgres(t, ctx)
	store := auth.NewIdentityStore(pool)

	principal, err := store.ValidateAPIKey(ctx, "demo-api-key")
	require.NoError(t, err)
	require.Equal(t, seed.DemoTenantID, principal.TenantID)
	require.Equal(t, auth.RoleAdmin, principal.Role)

	_, err = store.ValidateAPIKey(ctx, "invalid-key")
	require.Error(t, err)
}

func TestAuthExchangeCodeStore(t *testing.T) {
	ctx := context.Background()
	pool, _ := setupAuthPostgres(t, ctx)
	store := auth.NewIdentityStore(pool)

	code, err := auth.GenerateExchangeCode()
	require.NoError(t, err)
	require.NoError(t, store.SaveExchangeCode(ctx, code, demoUserID, time.Now().UTC().Add(time.Minute)))

	userID, err := store.ConsumeExchangeCode(ctx, code)
	require.NoError(t, err)
	require.Equal(t, demoUserID, userID)

	_, err = store.ConsumeExchangeCode(ctx, code)
	require.Error(t, err)
}

func (env *authTestEnv) postJSON(t *testing.T, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	env.router.ServeHTTP(resp, req)
	return resp
}

func countAuditLogs(ctx context.Context, pool *pgxpool.Pool, action, result string) int {
	var count int
	_ = pool.QueryRow(ctx, `
SELECT COUNT(*) FROM audit_logs WHERE action = $1 AND result = $2`, action, result).Scan(&count)
	return count
}
