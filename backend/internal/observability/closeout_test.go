package observability

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestTracePropagationThroughObservabilityAPI verifies traceId continuity on search (APM-01).
func TestTracePropagationThroughObservabilityAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(Deps{Mem: NewStore()})
	r := gin.New()
	v1 := r.Group("/api/v1")
	h.RegisterRoutes(v1)

	traceID := "trace-prop-test-001"
	body, _ := json.Marshal(map[string]any{
		"query":   "service:payment-service",
		"from":    "logs",
		"traceId": traceID,
		"limit":   10,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/query/unified", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "default")
	req.Header.Set("traceparent", "00-"+traceID+"-0011223344556677-01")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("unified query: %d %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["status"] != "success" {
		t.Fatalf("expected success envelope")
	}
}

func TestSLOBurnStatusEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(Deps{Mem: NewStore()})
	r := gin.New()
	v1 := r.Group("/api/v1")
	h.registerExtensionRoutes(v1)
	slos := h.deps.Mem.ListSLOs()
	if len(slos) == 0 {
		t.Fatal("expected seeded SLOs")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/slos/"+slos[0].ID+"/burn-status", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("burn status: %d", w.Code)
	}
}

func TestNetFlowIngest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(Deps{Mem: NewStore()})
	r := gin.New()
	v1 := r.Group("/api/v1")
	h.registerNPMExtendedRoutes(v1)
	raw, _ := json.Marshal(map[string]any{
		"records": []map[string]any{
			{"srcIp": "10.0.0.1", "dstIp": "10.0.0.2", "protocol": "TCP", "bytes": 1000, "packets": 10},
		},
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/network/flows/netflow/ingest", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("netflow ingest: %d %s", w.Code, w.Body.String())
	}
}
