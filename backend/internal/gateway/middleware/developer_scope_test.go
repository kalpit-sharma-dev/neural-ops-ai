package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/gateway/auth"
	"github.com/neuralops/platform/internal/gateway/middleware"
)

func TestDeveloperScopeAllowsAssignedService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("principal", auth.Principal{
			Role:     auth.RoleDeveloper,
			Services: []string{"payments", "platform"},
		})
		c.Next()
	})
	r.Use(middleware.DeveloperScope())
	r.GET("/api/v1/logs/search", func(c *gin.Context) {
		if c.GetHeader("X-Allowed-Services") == "" {
			t.Fatal("expected X-Allowed-Services header")
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/logs/search?service=payments", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("payments service: expected 200 got %d", w.Code)
	}
}

func TestDeveloperScopeDeniesUnassignedService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("principal", auth.Principal{
			Role:     auth.RoleDeveloper,
			Services: []string{"payments"},
		})
		c.Next()
	})
	r.Use(middleware.DeveloperScope())
	r.GET("/api/v1/logs/search", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/logs/search?service=platform", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("platform service: expected 403 got %d", w.Code)
	}
}

func TestDeveloperScopeSkipsNonDeveloper(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("principal", auth.Principal{Role: auth.RoleAdmin})
		c.Next()
	})
	r.Use(middleware.DeveloperScope())
	r.GET("/api/v1/logs/search", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/logs/search?service=any", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("admin: expected 200 got %d", w.Code)
	}
}
