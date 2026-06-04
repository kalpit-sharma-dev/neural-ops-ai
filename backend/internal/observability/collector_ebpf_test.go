package observability

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetCollectorEBPFMatrix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(Deps{Mem: NewStore()})
	r := gin.New()
	v1 := r.Group("/api/v1")
	h.RegisterRoutes(v1)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/collectors/fleet/ebpf-matrix", nil)
	req.Header.Set("X-Tenant-ID", "default")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	var body struct {
		Status string `json:"status"`
		Data   struct {
			Passed bool `json:"passed"`
			Checks []struct {
				Name string `json:"name"`
			} `json:"checks"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Data.Checks) < 4 {
		t.Fatalf("expected matrix checks, got %+v", body.Data)
	}
}
