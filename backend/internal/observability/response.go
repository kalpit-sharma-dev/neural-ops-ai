package observability

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func tenantID(c *gin.Context) string {
	if v, ok := c.Get("tenant_id"); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	if h := c.GetHeader("X-Tenant-ID"); h != "" {
		return h
	}
	return "default"
}

func withAnalyticsTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, 3*time.Second)
}

func writeSuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"data":      data,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func writeError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"status":    "error",
		"errorCode": "OBS001",
		"message":   message,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
