package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func phase5Router() *gin.Engine {
	gin.SetMode(gin.TestMode)
	store := NewStore()
	h := NewHandler(Deps{Mem: store})
	r := gin.New()
	h.RegisterRoutes(r.Group("/api/v1"))
	return r
}

func TestCloudAssetsAndTopology(t *testing.T) {
	r := phase5Router()
	for _, path := range []string{"/api/v1/cloud/assets", "/api/v1/cloud/topology"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s status %d", path, w.Code)
		}
	}
}

func TestFinOpsEndpoints(t *testing.T) {
	r := phase5Router()
	for _, path := range []string{
		"/api/v1/finops/costs?scope=payments",
		"/api/v1/finops/anomalies",
		"/api/v1/finops/carbon",
		"/api/v1/finops/costs/breakdown?dimension=team",
		"/api/v1/finops/allocation/rules",
		"/api/v1/finops/kubernetes/cost",
		"/api/v1/finops/recommendations",
		"/api/v1/finops/budgets",
		"/api/v1/finops/budgets/alerts",
		"/api/v1/finops/forecast?scope=all",
		"/api/v1/finops/commitments",
		"/api/v1/finops/commitments/recommendations",
		"/api/v1/finops/unit-economics?metric=request&scope=all",
		"/api/v1/finops/carbon/recommendations",
		"/api/v1/finops/reports",
		"/api/v1/finops/allocation/tag-suggestions",
		"/api/v1/finops/allocation/shared-splits",
		"/api/v1/finops/audit",
		"/api/v1/finops/chargeback/statements",
		"/api/v1/finops/scenarios",
		"/api/v1/finops/commitments/alerts",
		"/api/v1/finops/ingest/status",
		"/api/v1/finops/governance/policies",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s status %d body %s", path, w.Code, w.Body.String())
		}
	}
}

func TestNetworkEndpoints(t *testing.T) {
	r := phase5Router()
	for _, path := range []string{
		"/api/v1/network/flows",
		"/api/v1/network/devices",
		"/api/v1/network/topology",
		"/api/v1/network/anomalies",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s status %d", path, w.Code)
		}
	}
}
