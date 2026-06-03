package observability

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// TailSamplingPolicy controls trace tail-based sampling.
type TailSamplingPolicy struct {
	ID              string    `json:"id"`
	ServicePattern  string    `json:"servicePattern"`
	ErrorSamplePct  float64   `json:"errorSamplePct"`
	LatencyThresholdMs int    `json:"latencyThresholdMs"`
	SlowSamplePct   float64   `json:"slowSamplePct"`
	Enabled         bool      `json:"enabled"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

var (
	samplingMu       sync.RWMutex
	samplingPolicies = map[string]TailSamplingPolicy{
		"ts-1": {
			ID: "ts-1", ServicePattern: "payment-*", ErrorSamplePct: 100,
			LatencyThresholdMs: 500, SlowSamplePct: 25, Enabled: true, UpdatedAt: time.Now().UTC(),
		},
	}
)

func (h *Handler) registerAPMSamplingRoutes(v1 *gin.RouterGroup) {
	apm := v1.Group("/apm")
	{
		apm.GET("/sampling/policies", h.ListTailSamplingPolicies)
		apm.POST("/sampling/policies", h.CreateTailSamplingPolicy)
		apm.GET("/sampling/policies/export", h.ExportTailSamplingPolicies)
	}
}

func (h *Handler) ListTailSamplingPolicies(c *gin.Context) {
	samplingMu.RLock()
	defer samplingMu.RUnlock()
	out := make([]TailSamplingPolicy, 0, len(samplingPolicies))
	for _, p := range samplingPolicies {
		out = append(out, p)
	}
	writeSuccess(c, out)
}

func (h *Handler) CreateTailSamplingPolicy(c *gin.Context) {
	var p TailSamplingPolicy
	if err := c.ShouldBindJSON(&p); err != nil || p.ServicePattern == "" {
		writeError(c, 400, "servicePattern required")
		return
	}
	if p.ID == "" {
		p.ID = "ts-" + p.ServicePattern
	}
	p.UpdatedAt = time.Now().UTC()
	samplingMu.Lock()
	samplingPolicies[p.ID] = p
	samplingMu.Unlock()
	writeSuccess(c, p)
}

func (h *Handler) ExportTailSamplingPolicies(c *gin.Context) {
	writeSuccess(c, listTailSamplingPolicies())
}

func listTailSamplingPolicies() []TailSamplingPolicy {
	samplingMu.RLock()
	defer samplingMu.RUnlock()
	out := make([]TailSamplingPolicy, 0, len(samplingPolicies))
	for _, p := range samplingPolicies {
		out = append(out, p)
	}
	return out
}
