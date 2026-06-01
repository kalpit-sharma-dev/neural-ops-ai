package contract_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/observability"
)

// TestObservabilityEnvelope verifies API contract envelope on depth routes.
func TestObservabilityEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := observability.NewHandler(observability.Deps{Mem: observability.NewStore()})
	r := gin.New()
	v1 := r.Group("/api/v1")
	h.RegisterRoutes(v1)

	paths := []string{
		"/api/v1/apm/retention",
		"/api/v1/cloud/dashboards",
		"/api/v1/cloud/metrics?provider=aws&metric=CPUUtilization",
		"/api/v1/infra/k8s/namespaces",
		"/api/v1/admin/sso",
		"/api/v1/admin/tenant-policies",
		"/api/v1/admin/oncall",
	}
	for _, path := range paths {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("X-Tenant-ID", "default")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s: expected 200 got %d", path, w.Code)
		}
		var body map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s: invalid json: %v", path, err)
		}
		if body["status"] != "success" {
			t.Fatalf("%s: expected success envelope", path)
		}
		if body["data"] == nil {
			t.Fatalf("%s: missing data field", path)
		}
	}
}
