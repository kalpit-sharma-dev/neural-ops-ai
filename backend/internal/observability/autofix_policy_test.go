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

func TestAutoFixDisabledByDefaultInProd(t *testing.T) {
	t.Setenv("AUTOFIX_ENABLED", "false")
	gin.SetMode(gin.TestMode)
	store := observability.NewStore()
	h := observability.NewHandler(observability.Deps{Mem: store})
	r := gin.New()
	h.RegisterRoutes(r.Group("/api/v1"))

	body, _ := json.Marshal(map[string]string{"incidentId": "inc-1"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/autofix/plan", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 when disabled, got %d", w.Code)
	}
}

func TestAutoFixRequiresApprovalFlag(t *testing.T) {
	t.Setenv("AUTOFIX_ENABLED", "true")
	t.Setenv("AUTOFIX_REQUIRE_APPROVAL", "true")
	gin.SetMode(gin.TestMode)
	store := observability.NewStore()
	h := observability.NewHandler(observability.Deps{Mem: store})
	r := gin.New()
	h.RegisterRoutes(r.Group("/api/v1"))

	planBody, _ := json.Marshal(map[string]string{"incidentId": "inc-1"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/autofix/plan", bytes.NewReader(planBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("plan status %d", w.Code)
	}
	var planResp struct {
		Data observability.AutoFixPlan `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &planResp)

	execBody, _ := json.Marshal(map[string]interface{}{"planId": planResp.Data.ID, "approved": false})
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/ai/autofix/execute", bytes.NewReader(execBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without approval, got %d", w2.Code)
	}
}
