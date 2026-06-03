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
		"/api/v1/alerts/policies",
		"/api/v1/security/findings",
		"/api/v1/collectors/fleet",
		"/api/v1/ai/rca/inc-1",
		"/api/v1/admin/abac-policies",
		"/api/v1/nfr/benchmarks",
		"/api/v1/nfr/security-evidence",
		"/api/v1/metrics/derived",
		"/api/v1/admin/regions",
		"/api/v1/network/flows/netflow",
		"/api/v1/network/sdwan/tunnels",
		"/api/v1/network/wireless/links",
		"/api/v1/apm/sampling/policies",
		"/api/v1/finops/costs?scope=all",
		"/api/v1/finops/costs/breakdown?dimension=team",
		"/api/v1/finops/allocation/rules",
		"/api/v1/finops/budgets",
		"/api/v1/finops/budgets/alerts",
		"/api/v1/finops/ingest/status",
		"/api/v1/finops/anomalies",
		"/api/v1/finops/recommendations",
		"/api/v1/finops/forecast?scope=all",
		"/api/v1/finops/commitments",
		"/api/v1/finops/carbon?scope=all",
		"/api/v1/finops/chargeback/statements",
		"/api/v1/finops/scenarios",
		"/api/v1/finops/governance/policies",
		"/api/v1/finops/reconciliation",
		"/api/v1/ai/llm/workloads",
		"/api/v1/ai/llm/usage",
		"/api/v1/collectors/autoinstrumentation",
		"/api/v1/query/saved",
		"/api/v1/logs/tiering",
		"/api/v1/infra/serverless/functions",
		"/api/v1/admin/msp/tenants",
		"/api/v1/collectors/discovery",
		"/api/v1/collectors/supported-platforms",
		"/api/v1/logs/ingest-formats",
		"/api/v1/apm/service-catalog",
		"/api/v1/apm/errors",
		"/api/v1/admin/signal-policies",
		"/api/v1/security/threat-feed",
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
