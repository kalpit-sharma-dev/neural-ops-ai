package observability

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetLogPatterns(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(Deps{Mem: NewStore()})
	r := gin.New()
	v1 := r.Group("/api/v1")
	h.registerLogRoutes(v1)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/logs/patterns?limit=10", nil)
	req.Header.Set("X-Tenant-ID", "default")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var body struct {
		Status string              `json:"status"`
		Data   []LogPatternCluster `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Data) < 2 {
		t.Fatalf("expected patterns, got %d", len(body.Data))
	}
	if body.Data[0].Template == "" || body.Data[0].Count <= 0 {
		t.Fatalf("invalid cluster: %+v", body.Data[0])
	}
}
