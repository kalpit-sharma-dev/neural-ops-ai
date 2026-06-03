package observability

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ServerlessFunction is a monitored FaaS workload (INF-04).
type ServerlessFunction struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Provider     string    `json:"provider"`
	Runtime      string    `json:"runtime"`
	Region       string    `json:"region"`
	Invocations24h int64   `json:"invocations24h"`
	ErrorRatePct float64   `json:"errorRatePct"`
	P95DurationMs float64  `json:"p95DurationMs"`
	ColdStartPct float64   `json:"coldStartPct"`
	MemoryMB     int       `json:"memoryMb"`
	Status       string    `json:"status"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (h *Handler) registerServerlessRoutes(infra *gin.RouterGroup) {
	infra.GET("/serverless/functions", h.ListServerlessFunctions)
	infra.GET("/serverless/functions/:id", h.GetServerlessFunction)
}

func (h *Handler) ListServerlessFunctions(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListServerlessFunctions())
}

func (h *Handler) GetServerlessFunction(c *gin.Context) {
	fn, ok := h.deps.Mem.GetServerlessFunction(c.Param("id"))
	if !ok {
		writeError(c, http.StatusNotFound, "function not found")
		return
	}
	writeSuccess(c, fn)
}
