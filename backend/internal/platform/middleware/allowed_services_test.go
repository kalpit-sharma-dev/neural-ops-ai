package middleware_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	platmiddleware "github.com/neuralops/platform/internal/platform/middleware"
	"github.com/stretchr/testify/require"
)

func TestAllowedServicesFromGin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("X-Allowed-Services", "upi-service, payment-api")

	services := platmiddleware.AllowedServicesFromGin(c)
	require.Equal(t, []string{"upi-service", "payment-api"}, services)
	require.True(t, platmiddleware.ServiceAllowed(services, "upi-service"))
	require.False(t, platmiddleware.ServiceAllowed(services, "ledger-service"))
}
