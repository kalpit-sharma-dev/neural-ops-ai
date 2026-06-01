package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/gateway/auth"
	"github.com/stretchr/testify/require"
)

func TestRBACMiddlewareAllowsReadOnlyGet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(attachPrincipal(auth.RoleReadOnly))
	router.Use(RBAC())
	router.GET("/api/v1/search/logs", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/search/logs", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestRBACMiddlewareBlocksReadOnlyPost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(attachPrincipal(auth.RoleReadOnly))
	router.Use(RBAC())
	router.POST("/api/v1/settings", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestRBACMiddlewareMissingPrincipal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RBAC())
	router.GET("/api/v1/search/logs", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/search/logs", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestDeveloperScopeSetsAllowedServicesHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(attachPrincipalWithServices(auth.RoleDeveloper, []string{"upi-service"}))
	router.Use(DeveloperScope())
	router.GET("/api/v1/search/logs", func(c *gin.Context) {
		require.Equal(t, "upi-service", c.Request.Header.Get("X-Allowed-Services"))
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/search/logs?service=ledger-service", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestDeveloperScopeAllowsMatchingServiceQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(attachPrincipalWithServices(auth.RoleDeveloper, []string{"upi-service"}))
	router.Use(DeveloperScope())
	router.GET("/api/v1/search/logs", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/search/logs?service=upi-service", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestRBACFineGrainedPermissionDeny(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(attachPrincipal(auth.RoleDeveloper))
	router.Use(RBAC())
	router.POST("/api/v1/settings", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodPost, "/api/v1/settings", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestContainsAnyHelper(t *testing.T) {
	require.True(t, containsAny("/api/v1/search/logs", "/search", "/logs"))
	require.False(t, containsAny("/api/v1/health", "/search"))
}

func TestAllowedRoleMatrix(t *testing.T) {
	require.True(t, allowed(auth.RoleAdmin, http.MethodPost, "/api/v1/settings"))
	require.True(t, allowed(auth.RoleSRE, http.MethodPost, "/api/v1/incidents/abc/acknowledge"))
	require.False(t, allowed(auth.RoleDeveloper, http.MethodPost, "/api/v1/settings"))
	require.True(t, allowed(auth.RoleAlertManager, http.MethodGet, "/api/v1/alerts"))
}

func attachPrincipal(role auth.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth.AttachPrincipal(c, auth.Principal{
			UserID:   "user-1",
			TenantID: "00000000-0000-0000-0000-000000000002",
			Email:    "demo@neuralops.ai",
			Role:     role,
		})
		c.Next()
	}
}

func attachPrincipalWithServices(role auth.Role, services []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth.AttachPrincipal(c, auth.Principal{
			UserID:   "user-1",
			TenantID: "00000000-0000-0000-0000-000000000002",
			Email:    "demo@neuralops.ai",
			Role:     role,
			Services: services,
		})
		c.Next()
	}
}
