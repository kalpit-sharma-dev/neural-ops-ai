package observability_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/observability"
)

func TestIncidentUnifiedContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mem := observability.NewStore()
	// sf-1 is seeded in ListSecurityFindings; link it to the incident under test.
	mem.LinkSecurityFindingIncident("sf-1", "inc-99")
	h := observability.NewHandler(observability.Deps{Mem: mem})
	r := gin.New()
	v1 := r.Group("/api/v1")
	h.RegisterRoutes(v1)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/unified/incidents/inc-99/context?service=payment-service", nil)
	req.Header.Set("X-Tenant-ID", "default")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var body struct {
		Status string                           `json:"status"`
		Data   observability.IncidentUnifiedContext `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.PrimaryService != "payment-service" {
		t.Fatalf("service %q", body.Data.PrimaryService)
	}
	if len(body.Data.SecurityFindings) == 0 {
		t.Fatal("expected correlated findings")
	}
	if body.Data.DeepLinks.Logs == "" || body.Data.DeepLinks.Traces == "" {
		t.Fatal("expected deep links")
	}
}
