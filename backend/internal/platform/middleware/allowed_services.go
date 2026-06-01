package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const allowedServicesHeader = "X-Allowed-Services"

// AllowedServicesFromGin parses comma-separated service scope from gateway proxy headers.
func AllowedServicesFromGin(c *gin.Context) []string {
	raw := strings.TrimSpace(c.GetHeader(allowedServicesHeader))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if s := strings.TrimSpace(part); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// ServiceAllowed reports whether a service is within the allowed list (empty list = unrestricted).
func ServiceAllowed(allowed []string, service string) bool {
	if len(allowed) == 0 {
		return true
	}
	service = strings.ToLower(strings.TrimSpace(service))
	for _, candidate := range allowed {
		if strings.ToLower(strings.TrimSpace(candidate)) == service {
			return true
		}
	}
	return false
}

// AbortIfServiceForbidden rejects requests targeting a service outside developer scope.
func AbortIfServiceForbidden(c *gin.Context, allowed []string, service string) bool {
	if ServiceAllowed(allowed, service) {
		return false
	}
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"status": "error", "errorCode": "AUTH006", "message": "service not in developer scope",
	})
	return true
}
