package gateway_test

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/observability"
)

// Regression: observability /alerts/policies must not panic against alerting proxy routes.
func TestAlertRoutesNoWildcardConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	v1 := r.Group("/api/v1")

	h := observability.NewHandler(observability.Deps{Mem: observability.NewStore()})
	h.RegisterRoutes(v1)

	proxy := func(c *gin.Context) { c.Status(http.StatusOK) }
	v1.Any("/alerts", proxy)
	v1.Any("/alerts/rules", proxy)
	v1.Any("/alerts/rules/:id", proxy)
	v1.Any("/alerts/silences", proxy)
	v1.Any("/alerts/:id", proxy)
	v1.Any("/alerts/:id/acknowledge", proxy)
	v1.Any("/alerts/:id/suppress", proxy)
}

// Regression: observability incident extensions must not panic against incident proxy routes.
func TestIncidentRoutesNoWildcardConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	v1 := r.Group("/api/v1")

	h := observability.NewHandler(observability.Deps{Mem: observability.NewStore()})
	h.RegisterRoutes(v1)

	proxy := func(c *gin.Context) { c.Status(http.StatusOK) }
	v1.Any("/incidents", proxy)
	v1.Any("/incidents/:id", proxy)
	v1.Any("/incidents/:id/acknowledge", proxy)
	v1.Any("/incidents/:id/resolve", proxy)
	v1.Any("/incidents/:id/recommendations", proxy)
	v1.Any("/incidents/:id/timeline", proxy)
	v1.Any("/services/dependency-map", proxy)
	v1.Any("/transactions/:txnId", proxy)
}
