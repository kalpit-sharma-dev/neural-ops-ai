package observability

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func aiRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	store := NewStore()
	h := NewHandler(Deps{Mem: store})
	r := gin.New()
	h.RegisterRoutes(r.Group("/api/v1"))
	return r
}

func TestGetIncidentRCA(t *testing.T) {
	r := aiRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai/rca/inc-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}

func TestAIForecast(t *testing.T) {
	r := aiRouter()
	body, _ := json.Marshal(map[string]string{"metric": "cpu.utilization", "service": "payment-service", "horizon": "24h"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/forecast", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}

func TestAutoFixLifecycle(t *testing.T) {
	t.Setenv("AUTOFIX_ENABLED", "true")
	t.Setenv("AUTOFIX_REQUIRE_APPROVAL", "false")
	r := aiRouter()
	planBody, _ := json.Marshal(map[string]string{"incidentId": "inc-1"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/autofix/plan", bytes.NewReader(planBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("plan status %d", w.Code)
	}
	var planResp struct {
		Data AutoFixPlan `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &planResp); err != nil {
		t.Fatal(err)
	}
	execBody, _ := json.Marshal(map[string]interface{}{"planId": planResp.Data.ID, "approved": true})
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/ai/autofix/execute", bytes.NewReader(execBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("execute status %d body %s", w2.Code, w2.Body.String())
	}
	var actionResp struct {
		Data AutoFixActionRecord `json:"data"`
	}
	if err := json.Unmarshal(w2.Body.Bytes(), &actionResp); err != nil {
		t.Fatal(err)
	}
	rbBody, _ := json.Marshal(map[string]string{"actionId": actionResp.Data.ID})
	req3 := httptest.NewRequest(http.MethodPost, "/api/v1/ai/autofix/rollback", bytes.NewReader(rbBody))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("rollback status %d", w3.Code)
	}
}
