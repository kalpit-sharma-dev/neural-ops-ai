package observability

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPreviewAlertSilence(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(Deps{Mem: NewStore()})
	r := gin.New()
	v1 := r.Group("/api/v1")
	h.registerGapClosureRoutes(v1)

	body := []byte(`{"servicePattern":"payment-*"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/alerts/silences/preview", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "default")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	var resp struct {
		Data struct {
			MatchedAlerts int  `json:"matchedAlerts"`
			WouldSuppress bool `json:"wouldSuppress"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.MatchedAlerts < 1 || !resp.Data.WouldSuppress {
		t.Fatalf("expected payment-* to match firing alerts, got %+v", resp.Data)
	}
}

func TestMatchServicePattern(t *testing.T) {
	if !MatchServicePattern("payment-*", "payment-service") {
		t.Fatal("wildcard should match")
	}
	if MatchServicePattern("auth-*", "payment-service") {
		t.Fatal("wildcard should not match unrelated service")
	}
}

func TestRestoreLogTierArchive(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(Deps{Mem: NewStore()})
	r := gin.New()
	v1 := r.Group("/api/v1")
	h.registerGapClosureRoutes(v1)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/logs/tiering/restore", bytes.NewReader([]byte(`{"tier":"cold"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "default")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
}

func TestApplyAlertFeedbackFatigue(t *testing.T) {
	mem := NewStore()
	before := mem.ApplyAlertFeedbackToFatigue("pol-1", false)
	after := mem.ApplyAlertFeedbackToFatigue("pol-1", true)
	if after >= before {
		t.Fatalf("helpful feedback should reduce fatigue weight: before=%v after=%v", before, after)
	}
}

func TestCSPMDriftRun(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(Deps{Mem: NewStore()})
	r := gin.New()
	v1 := r.Group("/api/v1")
	h.registerGapClosureRoutes(v1)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/security/cspm/drift/run", nil)
	req.Header.Set("X-Tenant-ID", "default")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	var resp struct {
		Data struct {
			Drift int `json:"driftCount"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
}
