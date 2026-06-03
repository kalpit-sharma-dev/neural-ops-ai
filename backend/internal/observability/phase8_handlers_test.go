package observability

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func phase8Router() *gin.Engine {
	gin.SetMode(gin.TestMode)
	store := NewStore()
	h := NewHandler(Deps{Mem: store})
	r := gin.New()
	h.RegisterRoutes(r.Group("/api/v1"))
	return r
}

func TestNFREndpoints(t *testing.T) {
	r := phase8Router()
	for _, path := range []string{
		"/api/v1/nfr/benchmarks",
		"/api/v1/nfr/reliability",
		"/api/v1/nfr/accessibility",
		"/api/v1/nfr/i18n/locales",
		"/api/v1/nfr/certification",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s status %d", path, w.Code)
		}
	}
}
