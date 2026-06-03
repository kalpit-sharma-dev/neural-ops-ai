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

func TestBatchImportVulnerabilities(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := observability.NewStore()
	h := observability.NewHandler(observability.Deps{Mem: store})
	r := gin.New()
	r.POST("/api/v1/security/vulnerabilities/batch", h.BatchImportVulnerabilities)

	body := map[string]any{
		"source": "trivy",
		"findings": []map[string]any{
			{"cveId": "CVE-2026-0001", "severity": "high", "assetId": "neuralops/gateway:1.0"},
		},
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/security/vulnerabilities/batch", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	list := store.ListVulnerabilities()
	found := false
	for _, v := range list {
		if v.CVE == "CVE-2026-0001" {
			found = true
		}
	}
	if !found {
		t.Fatal("ingested CVE not listed")
	}
}
