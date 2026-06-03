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

func TestLogTierPolicyCRUD(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := observability.NewStore()
	h := observability.NewHandler(observability.Deps{Mem: store})
	r := gin.New()
	r.GET("/api/v1/logs/tiering", h.GetLogTierPolicy)
	r.PUT("/api/v1/logs/tiering", h.UpdateLogTierPolicy)

	body := map[string]any{"hotRetentionDays": 3, "warmRetentionDays": 14, "coldRetentionDays": 60}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/logs/tiering", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "bank-a")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("put status=%d", w.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/logs/tiering", nil)
	req2.Header.Set("X-Tenant-ID", "bank-a")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("get status=%d", w2.Code)
	}
}

func TestListServerlessFunctions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := observability.NewHandler(observability.Deps{Mem: observability.NewStore()})
	r := gin.New()
	r.GET("/api/v1/infra/serverless/functions", h.ListServerlessFunctions)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/infra/serverless/functions", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
}
