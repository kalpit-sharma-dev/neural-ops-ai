package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/gateway/auth"
	"github.com/neuralops/platform/internal/gateway/config"
	"github.com/stretchr/testify/require"
)

func TestAuthenticatorDisabledInjectsAdminPrincipal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authenticator, err := auth.NewAuthenticator(config.Config{
		Auth:   config.AuthConfig{Disabled: true},
		Tenant: config.TenantConfig{DefaultTenant: "tenant-default"},
	}, nil)
	require.NoError(t, err)

	router := gin.New()
	router.Use(authenticator.Middleware())
	router.GET("/api/v1/search/logs", func(c *gin.Context) {
		principal, ok := auth.PrincipalFromGin(c)
		require.True(t, ok)
		require.Equal(t, auth.RoleAdmin, principal.Role)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/search/logs", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthenticatorBearerJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authenticator, err := auth.NewAuthenticator(config.Config{
		Auth: config.AuthConfig{
			JWTIssuer:       "neuralops-test",
			JWTAudience:     "neuralops-api",
			AccessTokenTTL:  time.Hour,
			RefreshTokenTTL: 24 * time.Hour,
		},
		Tenant: config.TenantConfig{DefaultTenant: "tenant-default"},
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, authenticator.Issuer())

	pair, err := authenticator.Issuer().Issue(auth.Principal{
		UserID: "user-1", TenantID: "tenant-1", Email: "demo@neuralops.ai", Role: auth.RoleAdmin, AuthType: "jwt",
	})
	require.NoError(t, err)

	router := gin.New()
	router.Use(authenticator.Middleware())
	router.GET("/api/v1/search/logs", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/search/logs", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthenticatorAPIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authenticator, err := auth.NewAuthenticator(config.Config{
		Auth: config.AuthConfig{
			APIKeys: map[string]config.APIKeyPrincipal{
				"test-key": {UserID: "ingest-1", TenantID: "tenant-1", Role: "SRE"},
			},
		},
		Tenant: config.TenantConfig{DefaultTenant: "tenant-default"},
	}, nil)
	require.NoError(t, err)

	router := gin.New()
	router.Use(authenticator.Middleware())
	router.GET("/api/v1/search/logs", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/search/logs", nil)
	req.Header.Set("X-API-Key", "test-key")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthenticatorMissingCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authenticator, err := auth.NewAuthenticator(config.Config{
		Auth:   config.AuthConfig{},
		Tenant: config.TenantConfig{DefaultTenant: "tenant-default"},
	}, nil)
	require.NoError(t, err)

	router := gin.New()
	router.Use(authenticator.Middleware())
	router.GET("/api/v1/search/logs", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/search/logs", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAPIKeyValidator(t *testing.T) {
	validator := auth.NewAPIKeyValidator(config.AuthConfig{
		APIKeys: map[string]config.APIKeyPrincipal{
			"k1": {UserID: "u1", TenantID: "t1", Role: "ADMIN"},
		},
	})
	principal, err := validator.Validate("k1")
	require.NoError(t, err)
	require.Equal(t, auth.RoleAdmin, principal.Role)
	_, err = validator.Validate("missing")
	require.Error(t, err)
}
