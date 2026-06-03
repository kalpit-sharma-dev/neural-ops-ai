package observability

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func phase7Router() *gin.Engine {
	gin.SetMode(gin.TestMode)
	store := NewStore()
	h := NewHandler(Deps{Mem: store})
	r := gin.New()
	h.RegisterRoutes(r.Group("/api/v1"))
	return r
}

func TestABACAndResidency(t *testing.T) {
	r := phase7Router()
	for _, path := range []string{"/api/v1/admin/abac-policies", "/api/v1/admin/data-residency", "/api/v1/admin/branding"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s status %d", path, w.Code)
		}
	}
}

func TestMSPTenants(t *testing.T) {
	r := phase7Router()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/msp/tenants", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
}

func TestExports(t *testing.T) {
	r := phase7Router()
	body, _ := json.Marshal(map[string]string{"destination": "s3://neuralops-exports/demo"})
	for _, path := range []string{"/api/v1/exports/warehouse", "/api/v1/exports/bi", "/api/v1/exports/events"} {
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s status %d body %s", path, w.Code, w.Body.String())
		}
	}
}
