package observability_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/observability"
)

func TestSRSRoutesEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mem := observability.NewStore()
	h := observability.NewHandler(observability.Deps{Mem: mem})
	r := gin.New()
	v1 := r.Group("/api/v1")
	h.RegisterRoutes(v1)

	tests := []struct {
		method, path string
		body         []byte
	}{
		{http.MethodGet, "/api/v1/collectors/fleet/agent-1", nil},
		{http.MethodPost, "/api/v1/collectors/fleet/agent-1/upgrade", []byte(`{"targetVersion":"1.3.0"}`)},
		{http.MethodGet, "/api/v1/security/findings/sf-1", nil},
		{http.MethodPost, "/api/v1/alerts/policies/ap-1/trigger", []byte(`{"service":"payment-service","severity":"P1"}`)},
		{http.MethodPost, "/api/v1/alerts/policies/ap-1/feedback", []byte(`{"service":"payment-service","helpful":true}`)},
		{http.MethodGet, "/api/v1/alerts/policies/ap-1/feedback", nil},
		{http.MethodGet, "/api/v1/metrics/derived", nil},
		{http.MethodGet, "/api/v1/admin/regions", nil},
		{http.MethodPost, "/api/v1/nfr/benchmark/run", []byte(`{}`)},
		{http.MethodGet, "/api/v1/network/flows/netflow", nil},
		{http.MethodGet, "/api/v1/apm/sampling/policies", nil},
	}
	for _, tc := range tests {
		req := httptest.NewRequest(tc.method, tc.path, bytes.NewReader(tc.body))
		req.Header.Set("X-Tenant-ID", "default")
		if tc.body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
			t.Fatalf("%s %s: status %d body %s", tc.method, tc.path, w.Code, w.Body.String())
		}
		if w.Code == http.StatusOK {
			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("%s: invalid json", tc.path)
			}
			if body["status"] != "success" {
				t.Fatalf("%s: expected success envelope", tc.path)
			}
		}
	}
}
