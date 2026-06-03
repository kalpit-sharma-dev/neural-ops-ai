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

func TestExplainUnifiedQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := observability.NewHandler(observability.Deps{Mem: observability.NewStore()})
	r := gin.New()
	r.POST("/api/v1/query/explain", h.ExplainUnifiedQuery)

	body := map[string]any{"query": "error rate trace_id:abc", "from": "all", "service": "payment-service"}
	raw, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/query/explain", bytes.NewReader(raw)))
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestSavedQueryLifecycle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := observability.NewStore()
	h := observability.NewHandler(observability.Deps{Mem: store})
	r := gin.New()
	r.GET("/api/v1/query/saved", h.ListSavedQueries)
	r.POST("/api/v1/query/saved", h.CreateSavedQuery)
	r.DELETE("/api/v1/query/saved/:id", h.DeleteSavedQuery)

	createBody := map[string]any{"name": "payments errors", "query": "error payment-service"}
	raw, _ := json.Marshal(createBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/query/saved", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "bank-a")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create status=%d", w.Code)
	}
}

func TestAutoInstrumentationMatrix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := observability.NewHandler(observability.Deps{Mem: observability.NewStore()})
	r := gin.New()
	r.GET("/api/v1/collectors/autoinstrumentation", h.ListAutoInstrumentation)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/collectors/autoinstrumentation", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
}
