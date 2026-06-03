package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/finops"
	"github.com/neuralops/platform/internal/gateway/auth"
)

func finOpsRouterWithPrincipal(services []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	store := NewStore()
	h := NewHandler(Deps{Mem: store})
	r := gin.New()
	r.Use(func(c *gin.Context) {
		auth.AttachPrincipal(c, auth.Principal{
			UserID:   "dev-user",
			TenantID: "default",
			Services: services,
		})
		c.Next()
	})
	h.RegisterRoutes(r.Group("/api/v1"))
	return r
}

func TestFinOpsScopeForbidden(t *testing.T) {
	r := finOpsRouterWithPrincipal([]string{"platform"})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/finops/costs?scope=payments", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for disallowed scope, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestFinOpsScopeAllowed(t *testing.T) {
	r := finOpsRouterWithPrincipal([]string{"payments"})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/finops/costs?scope=payments", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for allowed scope, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestFinOpsFilterInjectionRejected(t *testing.T) {
	r := phase5Router()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/finops/costs?scope=foo%20bar", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for injection attempt, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestFinOpsTenantIsolation(t *testing.T) {
	t.Setenv("FINOPS_BILLING_MODE", "live")
	t.Setenv("FINOPS_ALLOW_SIMULATION_FALLBACK", "false")

	gin.SetMode(gin.TestMode)
	store := NewStore()
	svc := finops.NewService(&FinOpsCloudAdapter{Store: store})
	if _, err := svc.IngestAll(context.Background(), "tenant-a"); err != nil {
		t.Fatalf("ingest tenant-a: %v", err)
	}
	h := NewHandler(Deps{Mem: store, FinOps: svc})
	r := gin.New()
	h.RegisterRoutes(r.Group("/api/v1"))

	reqA := httptest.NewRequest(http.MethodGet, "/api/v1/finops/costs?scope=all", nil)
	reqA.Header.Set("X-Tenant-ID", "tenant-a")
	wA := httptest.NewRecorder()
	r.ServeHTTP(wA, reqA)
	if wA.Code != http.StatusOK {
		t.Fatalf("tenant-a costs: %d", wA.Code)
	}

	reqB := httptest.NewRequest(http.MethodGet, "/api/v1/finops/costs?scope=all", nil)
	reqB.Header.Set("X-Tenant-ID", "tenant-b")
	wB := httptest.NewRecorder()
	r.ServeHTTP(wB, reqB)
	if wB.Code != http.StatusOK {
		t.Fatalf("tenant-b costs: %d", wB.Code)
	}
	if wA.Body.String() == wB.Body.String() && wA.Body.Len() > 50 {
		t.Fatal("tenant-b should not receive tenant-a cost data")
	}
}

func TestFinOpsIngestStatus(t *testing.T) {
	r := phase5Router()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/finops/ingest/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", w.Code, w.Body.String())
	}
}
