package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/gateway/publicpaths"
)

// MultiRegionRouter enforces active-active region headers for cross-region writes.
func MultiRegionRouter() gin.HandlerFunc {
	local := strings.TrimSpace(os.Getenv("NEURALOPS_REGION"))
	if local == "" {
		local = "us-east-1"
	}
	peers := parsePeerRegions(os.Getenv("NEURALOPS_PEER_REGIONS"))
	enabled := strings.EqualFold(os.Getenv("NEURALOPS_MULTI_REGION"), "true")

	return func(c *gin.Context) {
		if !enabled || publicpaths.IsPublic(c.Request.URL.Path) {
			c.Next()
			return
		}
		c.Header("X-NeuralOps-Region", local)
		target := c.GetHeader("X-Region-Target")
		if target == "" {
			c.Next()
			return
		}
		if strings.EqualFold(target, local) {
			c.Next()
			return
		}
		allowed := false
		for _, p := range peers {
			if strings.EqualFold(p, target) {
				allowed = true
				break
			}
		}
		if !allowed && c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"status": "error", "errorCode": "REG001", "message": "cross-region write blocked for region " + target,
			})
			return
		}
		c.Next()
	}
}

func parsePeerRegions(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
