package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/gateway/auth"
	"github.com/neuralops/platform/internal/gateway/publicpaths"
)

// DeveloperScope restricts DEVELOPER role access to assigned services.
func DeveloperScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		if publicpaths.IsPublic(c.Request.URL.Path) {
			c.Next()
			return
		}

		principal, ok := auth.PrincipalFromGin(c)
		if !ok || principal.Role != auth.RoleDeveloper || len(principal.Services) == 0 {
			c.Next()
			return
		}

		allowed := strings.Join(principal.Services, ",")
		c.Request.Header.Set("X-Allowed-Services", allowed)

		if service := strings.TrimSpace(c.Query("service")); service != "" {
			if !serviceAllowed(principal.Services, service) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"status": "error", "errorCode": "AUTH006", "message": "service not in developer scope",
				})
				return
			}
		}
		c.Next()
	}
}

func serviceAllowed(allowed []string, service string) bool {
	service = strings.ToLower(strings.TrimSpace(service))
	for _, candidate := range allowed {
		if strings.ToLower(strings.TrimSpace(candidate)) == service {
			return true
		}
	}
	return false
}
