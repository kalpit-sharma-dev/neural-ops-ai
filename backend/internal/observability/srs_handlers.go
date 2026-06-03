package observability

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/streaming"
)

func (h *Handler) registerSRSRoutes(v1 *gin.RouterGroup) {
	collectors := v1.Group("/collectors")
	{
		collectors.GET("/fleet/:id", h.GetCollectorFleetAgent)
		collectors.POST("/fleet/:id/upgrade", h.UpgradeCollectorFleetAgent)
	}
	v1.GET("/security/findings/:id", h.GetSecurityFinding)
	v1.POST("/security/findings/:id/correlate", h.CorrelateSecurityFinding)
	v1.POST("/alerts/policies/:id/trigger", h.TriggerAlertPolicy)
	v1.GET("/metrics/derived", h.ListDerivedMetrics)
	v1.POST("/metrics/derived", h.CreateDerivedMetric)
	v1.GET("/metrics/derived/:id/samples", h.ListDerivedMetricSamples)
	v1.GET("/alerts/policies/:id/scores", h.ListAlertPolicyScores)
	v1.POST("/alerts/policies/:id/feedback", h.SubmitAlertPolicyFeedback)
	v1.GET("/alerts/policies/:id/feedback", h.ListAlertPolicyFeedback)
}

func (h *Handler) GetCollectorFleetAgent(c *gin.Context) {
	id := c.Param("id")
	tid := tenantID(c)
	if h.deps.SRS != nil && h.deps.SRS.available() {
		if a, err := h.deps.SRS.GetCollectorAgent(c.Request.Context(), tid, id); err == nil {
			writeSuccess(c, a)
			return
		}
	}
	for _, a := range h.deps.Mem.ListCollectorFleet() {
		if a.ID == id {
			writeSuccess(c, a)
			return
		}
	}
	writeError(c, http.StatusNotFound, "collector agent not found")
}

type upgradeCollectorRequest struct {
	TargetVersion string `json:"targetVersion"`
}

func (h *Handler) UpgradeCollectorFleetAgent(c *gin.Context) {
	id := c.Param("id")
	tid := tenantID(c)
	var req upgradeCollectorRequest
	_ = c.ShouldBindJSON(&req)
	if h.deps.SRS != nil && h.deps.SRS.available() {
		if a, err := h.deps.SRS.UpgradeCollectorAgent(c.Request.Context(), tid, id, req.TargetVersion); err == nil {
			writeSuccess(c, a)
			return
		}
	}
	agents := h.deps.Mem.ListCollectorFleet()
	for i, a := range agents {
		if a.ID == id {
			if req.TargetVersion != "" {
				a.Version = req.TargetVersion
			} else {
				a.Version = bumpPatch(a.Version)
			}
			a.Status = "healthy"
			a.LastHeartbeatAt = time.Now().UTC()
			updated, _ := h.deps.Mem.UpdateCollectorAgent(id, a)
			writeSuccess(c, updated)
			return
		}
		_ = i
	}
	writeError(c, http.StatusNotFound, "collector agent not found")
}

func (h *Handler) CorrelateSecurityFinding(c *gin.Context) {
	id := c.Param("id")
	tid := tenantID(c)
	if h.deps.SecCorr == nil {
		writeError(c, http.StatusServiceUnavailable, "correlation service unavailable")
		return
	}
	f, err := h.deps.SecCorr.CorrelateFinding(c.Request.Context(), tid, id)
	if err != nil {
		writeError(c, http.StatusNotFound, err.Error())
		return
	}
	writeSuccess(c, f)
}

func (h *Handler) GetSecurityFinding(c *gin.Context) {
	id := c.Param("id")
	tid := tenantID(c)
	if h.deps.SRS != nil && h.deps.SRS.available() {
		if f, err := h.deps.SRS.GetSecurityFinding(c.Request.Context(), tid, id); err == nil {
			writeSuccess(c, f)
			return
		}
	}
	for _, f := range h.deps.Mem.ListSecurityFindings() {
		if f.ID == id {
			writeSuccess(c, f)
			return
		}
	}
	writeError(c, http.StatusNotFound, "finding not found")
}

type triggerAlertRequest struct {
	Service  string `json:"service"`
	Severity string `json:"severity"`
}

func (h *Handler) TriggerAlertPolicy(c *gin.Context) {
	var req triggerAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Service == "" {
		req.Service = "payment-service"
		req.Severity = "P1"
	}
	if req.Severity == "" {
		req.Severity = "P1"
	}
	res := EvaluateAlertPolicy(h.deps.Mem, c.Param("id"), req.Service, req.Severity)
	recordAlertTrigger(tenantID(c), res.Matched)
	if h.deps.Stream != nil && res.Matched {
		_ = h.deps.Stream.PublishAlertSignal(c.Request.Context(), tenantID(c), streaming.AlertSignalPayload{
			PolicyID: c.Param("id"), Service: req.Service, Severity: req.Severity, Count: 1,
		})
	}
	writeSuccess(c, res)
}

func (h *Handler) ListDerivedMetrics(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListDerivedMetrics())
}

func (h *Handler) CreateDerivedMetric(c *gin.Context) {
	var body DerivedMetric
	if err := c.ShouldBindJSON(&body); err != nil || body.Name == "" {
		writeError(c, http.StatusBadRequest, "name and expression required")
		return
	}
	saved := h.deps.Mem.SaveDerivedMetric(body)
	if h.deps.Stream != nil {
		_ = h.deps.Stream.PublishMetricSample(c.Request.Context(), tenantID(c), streaming.MetricSamplePayload{
			MetricID: saved.ID, Service: "platform", Value: saved.Value,
		})
	}
	writeSuccess(c, saved)
}
