package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/gateway/auth"
	"github.com/neuralops/platform/internal/gateway/middleware"
	"github.com/neuralops/platform/internal/observability"
)

func TestGovernanceEnforcementAllowsRead(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mem := observability.NewStore()
	mem.PutABACPolicy(observability.ABACPolicy{
		Enabled: true,
		Rules: []observability.ABACPolicyRule{
			{ID: "r1", Effect: "allow", Action: "read", Resource: "/api/v1/*"},
		},
	})
	gs := observability.NewGovernanceService(nil, mem)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("principal", auth.Principal{TenantID: "default", Role: auth.RoleDeveloper, Email: "dev@neuralops.ai"})
		c.Next()
	})
	r.Use(middleware.GovernanceEnforcement(gs))
	r.GET("/api/v1/alerts/policies", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts/policies", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
}

func TestGovernanceEnforcementDeniesExport(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mem := observability.NewStore()
	mem.PutABACPolicy(observability.ABACPolicy{
		Enabled: true,
		Rules: []observability.ABACPolicyRule{
			{ID: "deny-export", Effect: "deny", Action: "export", Resource: "/api/v1/exports/*"},
		},
	})
	gs := observability.NewGovernanceService(nil, mem)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("principal", auth.Principal{TenantID: "default", Role: auth.RoleDeveloper, Email: "dev@neuralops.ai"})
		c.Next()
	})
	r.Use(middleware.GovernanceEnforcement(gs))
	r.POST("/api/v1/exports/logs", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodPost, "/api/v1/exports/logs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 got %d", w.Code)
	}
}
