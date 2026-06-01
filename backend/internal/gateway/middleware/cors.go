package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/gateway/config"
)

// CORS configures cross-origin access for the React frontend.
func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	origins := cfg.AllowedOrigins
	methods := cfg.AllowedMethods
	headers := cfg.AllowedHeaders
	if len(methods) == 0 {
		methods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}
	if len(headers) == 0 {
		headers = []string{"Authorization", "Content-Type", "X-API-Key", "X-Tenant-ID", "X-Request-ID"}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && originAllowed(origins, origin) {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if len(origins) == 1 {
			c.Header("Access-Control-Allow-Origin", origins[0])
		}
		c.Header("Access-Control-Allow-Methods", strings.Join(methods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(headers, ", "))
		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Vary", "Origin")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func originAllowed(allowed []string, origin string) bool {
	for _, item := range allowed {
		if item == "*" || item == origin {
			return true
		}
	}
	return false
}
