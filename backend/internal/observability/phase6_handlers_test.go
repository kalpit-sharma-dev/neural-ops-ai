package observability

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func phase6Router() *gin.Engine {
	gin.SetMode(gin.TestMode)
	store := NewStore()
	h := NewHandler(Deps{Mem: store})
	r := gin.New()
	h.RegisterRoutes(r.Group("/api/v1"))
	return r
}

func TestRUMFunnels(t *testing.T) {
	r := phase6Router()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rum/funnels", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list status %d", w.Code)
	}
	body, _ := json.Marshal(map[string]interface{}{
		"name": "Signup funnel",
		"steps": []map[string]string{
			{"name": "Landing", "event": "page_landing"},
			{"name": "Signup", "event": "signup_complete"},
		},
	})
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/rum/funnels", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("create status %d body %s", w2.Code, w2.Body.String())
	}
}

func TestSyntheticExpansion(t *testing.T) {
	r := phase6Router()
	cases := []struct {
		method, path string
		body         map[string]interface{}
	}{
		{http.MethodPost, "/api/v1/synthetic/browser-tests", map[string]interface{}{
			"name": "Homepage", "url": "https://example.com", "script": "await page.goto(url);",
		}},
		{http.MethodPost, "/api/v1/synthetic/mobile-tests", map[string]interface{}{
			"name": "Android smoke", "platform": "android", "bundleId": "com.example.app",
		}},
		{http.MethodPost, "/api/v1/synthetic/private-locations", map[string]interface{}{
			"name": "dc-east", "region": "on-prem",
		}},
	}
	for _, tc := range cases {
		b, _ := json.Marshal(tc.body)
		req := httptest.NewRequest(tc.method, tc.path, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s status %d body %s", tc.path, w.Code, w.Body.String())
		}
	}
}

func TestBusinessKPIPacks(t *testing.T) {
	r := phase6Router()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/business/kpi-packs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list status %d", w.Code)
	}
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/business/kpi-packs/kpi-bfsi-payments/enable", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("enable bfsi payments pack status %d", w2.Code)
	}
}
