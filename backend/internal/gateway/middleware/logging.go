package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/gateway/auth"
	"go.uber.org/zap"
)

// RequestLogger logs request metadata using structured JSON logging.
func RequestLogger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		tenantID := ""
		if principal, ok := auth.PrincipalFromGin(c); ok {
			tenantID = principal.TenantID
		}

		log.Info("gateway request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("tenantId", tenantID),
			zap.String("requestId", c.GetHeader("X-Request-ID")),
		)
	}
}
