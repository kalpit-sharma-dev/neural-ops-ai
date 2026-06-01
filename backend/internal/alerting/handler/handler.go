package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/alerting/model"
	"github.com/neuralops/platform/internal/alerting/normalizer"
	"github.com/neuralops/platform/internal/alerting/repository"
	"github.com/neuralops/platform/internal/alerting/service"
	"github.com/neuralops/platform/internal/domain"
	"go.uber.org/zap"
)

// Handler exposes alerting HTTP endpoints.
type Handler struct {
	log     *zap.Logger
	service *service.Service
	tenant  string
}

// New creates an alerting handler.
func New(log *zap.Logger, svc *service.Service, defaultTenant string) *Handler {
	return &Handler{log: log, service: svc, tenant: defaultTenant}
}

// RegisterRoutes mounts alerting routes.
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.POST("/webhook/prometheus", h.PrometheusWebhook)
	router.POST("/webhook/dynatrace", h.DynatraceWebhook)
	router.POST("/webhook/aws", h.CloudWatchWebhook)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/alerts", h.ListAlerts)
		v1.GET("/alerts/rules", h.ListRules)
		v1.POST("/alerts/rules", h.CreateRule)
		v1.PUT("/alerts/rules/:id", h.UpdateRule)
		v1.DELETE("/alerts/rules/:id", h.DeleteRule)
		v1.GET("/alerts/silences", h.ListSilences)
		v1.POST("/alerts/silences", h.CreateSilence)
		v1.GET("/alerts/:id", h.GetAlert)
		v1.POST("/alerts/:id/acknowledge", h.AcknowledgeAlert)
		v1.POST("/alerts/:id/suppress", h.SuppressAlert)
		v1.GET("/notifications/channels", h.ListChannels)
		v1.POST("/notifications/channels", h.CreateChannel)
		v1.PUT("/notifications/channels/:id", h.UpdateChannel)
		v1.DELETE("/notifications/channels/:id", h.DeleteChannel)
		v1.POST("/notifications/channels/:id/test", h.TestChannel)
		v1.GET("/escalation/policies", h.ListEscalationPolicies)
		v1.POST("/escalation/policies", h.CreateEscalationPolicy)
		v1.PUT("/escalation/policies/:id", h.UpdateEscalationPolicy)
		v1.DELETE("/escalation/policies/:id", h.DeleteEscalationPolicy)
	}
}

// PrometheusWebhook ingests Alertmanager alerts.
func (h *Handler) PrometheusWebhook(c *gin.Context) {
	var payload normalizer.PrometheusPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid payload")
		return
	}
	h.ingest(c, normalizer.FromPrometheus(payload))
}

// DynatraceWebhook ingests Dynatrace problem notifications.
func (h *Handler) DynatraceWebhook(c *gin.Context) {
	var payload normalizer.DynatracePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid payload")
		return
	}
	h.ingest(c, normalizer.FromDynatrace(payload))
}

// CloudWatchWebhook ingests AWS CloudWatch alarm notifications.
func (h *Handler) CloudWatchWebhook(c *gin.Context) {
	var payload normalizer.CloudWatchPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid payload")
		return
	}
	h.ingest(c, normalizer.FromCloudWatch(payload))
}

func (h *Handler) ingest(c *gin.Context, incoming []model.IncomingAlert) {
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		tenantID = h.tenant
	}
	alerts, err := h.service.Ingest(c.Request.Context(), tenantID, incoming)
	if err != nil {
		h.log.Error("alert ingestion failed", zap.Error(err))
		h.writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeSuccess(c, alerts)
}

// ListAlerts lists alerts with filters.
func (h *Handler) ListAlerts(c *gin.Context) {
	filter := repository.AlertFilter{
		TenantID: c.DefaultQuery("tenantId", h.tenant),
		Status:   c.Query("status"),
		Severity: c.Query("severity"),
		Service:  c.Query("service"),
		Source:   c.Query("source"),
	}
	if limit, err := strconv.Atoi(c.DefaultQuery("size", "50")); err == nil {
		filter.Limit = limit
	}
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && page > 0 {
		filter.Offset = (page - 1) * filter.Limit
	}
	alerts, err := h.service.ListAlerts(c.Request.Context(), filter)
	if err != nil {
		h.writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeSuccess(c, alerts)
}

// GetAlert returns alert detail.
func (h *Handler) GetAlert(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid alert id")
		return
	}
	alert, err := h.service.GetAlert(c.Request.Context(), id)
	if err != nil {
		h.writeError(c, http.StatusNotFound, err.Error())
		return
	}
	h.writeSuccess(c, alert)
}

// AcknowledgeAlert acknowledges an alert.
func (h *Handler) AcknowledgeAlert(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid alert id")
		return
	}
	alert, err := h.service.AcknowledgeAlert(c.Request.Context(), id)
	if err != nil {
		h.writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeSuccess(c, alert)
}

type suppressRequest struct {
	Duration string `json:"duration"`
	Reason   string `json:"reason"`
}

// SuppressAlert suppresses an alert for a duration.
func (h *Handler) SuppressAlert(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid alert id")
		return
	}
	var req suppressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		duration = 30 * time.Minute
	}
	alert, err := h.service.SuppressAlert(c.Request.Context(), id, duration, req.Reason)
	if err != nil {
		h.writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeSuccess(c, alert)
}

type createRuleRequest struct {
	Name           string                 `json:"name"`
	Source         domain.AlertSource     `json:"source"`
	ServicePattern string                 `json:"servicePattern"`
	Severity       domain.IncidentSeverity `json:"severity"`
	Enabled        bool                   `json:"enabled"`
	Labels         map[string]string      `json:"labels"`
}

// CreateRule creates an alert rule.
func (h *Handler) CreateRule(c *gin.Context) {
	var req createRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	rule := &model.AlertRule{
		ID:             uuid.New(),
		TenantID:       h.tenant,
		Name:           req.Name,
		Source:         req.Source,
		ServicePattern: req.ServicePattern,
		Severity:       req.Severity,
		Enabled:        req.Enabled,
		Labels:         req.Labels,
	}
	if err := h.service.CreateRule(c.Request.Context(), rule); err != nil {
		h.writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeSuccess(c, rule)
}

// ListRules lists alert rules.
func (h *Handler) ListRules(c *gin.Context) {
	rules, err := h.service.ListRules(c.Request.Context(), h.tenant)
	if err != nil {
		h.writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeSuccess(c, rules)
}

// UpdateRule updates an alert rule.
func (h *Handler) UpdateRule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid rule id")
		return
	}
	var req createRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	rule := &model.AlertRule{
		ID: id, TenantID: h.tenant, Name: req.Name, Source: req.Source,
		ServicePattern: req.ServicePattern, Severity: req.Severity, Enabled: req.Enabled, Labels: req.Labels,
	}
	if err := h.service.UpdateRule(c.Request.Context(), rule); err != nil {
		h.writeError(c, http.StatusNotFound, err.Error())
		return
	}
	h.writeSuccess(c, rule)
}

// DeleteRule deletes an alert rule.
func (h *Handler) DeleteRule(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid rule id")
		return
	}
	if err := h.service.DeleteRule(c.Request.Context(), h.tenant, id); err != nil {
		h.writeError(c, http.StatusNotFound, err.Error())
		return
	}
	h.writeSuccess(c, gin.H{"deleted": true})
}

type createSilenceRequest struct {
	ServicePattern   string `json:"servicePattern"`
	AlertNamePattern string `json:"alertNamePattern"`
	Reason           string `json:"reason"`
	Duration         string `json:"duration"`
}

// ListSilences lists alert silences.
func (h *Handler) ListSilences(c *gin.Context) {
	silences, err := h.service.ListSilences(c.Request.Context(), h.tenant)
	if err != nil {
		h.writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeSuccess(c, silences)
}

// CreateSilence creates a silence window.
func (h *Handler) CreateSilence(c *gin.Context) {
	var req createSilenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		duration = 1 * time.Hour
	}
	now := time.Now().UTC()
	silence := &model.Silence{
		TenantID: h.tenant, ServicePattern: req.ServicePattern, AlertNamePattern: req.AlertNamePattern,
		Reason: req.Reason, StartsAt: now, EndsAt: now.Add(duration),
	}
	if err := h.service.CreateSilence(c.Request.Context(), silence); err != nil {
		h.writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeSuccess(c, silence)
}

type createChannelRequest struct {
	Name        string         `json:"name"`
	ChannelType string         `json:"channelType"`
	Config      map[string]any `json:"config"`
	Enabled     bool           `json:"enabled"`
}

// CreateChannel creates a notification channel.
func (h *Handler) CreateChannel(c *gin.Context) {
	var req createChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	channel := &model.NotificationChannel{
		ID:          uuid.New(),
		TenantID:    h.tenant,
		Name:        req.Name,
		ChannelType: req.ChannelType,
		Config:      req.Config,
		Enabled:     req.Enabled,
	}
	if err := h.service.CreateChannel(c.Request.Context(), channel); err != nil {
		h.writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeSuccess(c, channel)
}

// ListChannels lists notification channels.
func (h *Handler) ListChannels(c *gin.Context) {
	channels, err := h.service.ListChannels(c.Request.Context(), h.tenant)
	if err != nil {
		h.writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeSuccess(c, channels)
}

// UpdateChannel updates a notification channel.
func (h *Handler) UpdateChannel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid channel id")
		return
	}
	var req createChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	channel := &model.NotificationChannel{
		ID: id, TenantID: h.tenant, Name: req.Name, ChannelType: req.ChannelType, Config: req.Config, Enabled: req.Enabled,
	}
	if err := h.service.UpdateChannel(c.Request.Context(), channel); err != nil {
		h.writeError(c, http.StatusNotFound, err.Error())
		return
	}
	h.writeSuccess(c, channel)
}

// DeleteChannel deletes a notification channel.
func (h *Handler) DeleteChannel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid channel id")
		return
	}
	if err := h.service.DeleteChannel(c.Request.Context(), h.tenant, id); err != nil {
		h.writeError(c, http.StatusNotFound, err.Error())
		return
	}
	h.writeSuccess(c, gin.H{"deleted": true})
}

// TestChannel sends a test notification (dry-run).
func (h *Handler) TestChannel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid channel id")
		return
	}
	channels, err := h.service.ListChannels(c.Request.Context(), h.tenant)
	if err != nil {
		h.writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	for _, ch := range channels {
		if ch.ID == id {
			h.writeSuccess(c, gin.H{"sent": true, "channelId": id.String(), "message": "Test notification dispatched"})
			return
		}
	}
	h.writeError(c, http.StatusNotFound, "channel not found")
}

func (h *Handler) ListEscalationPolicies(c *gin.Context) {
	policies, err := h.service.ListEscalationPolicies(c.Request.Context(), h.tenant)
	if err != nil {
		h.writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeSuccess(c, policies)
}

func (h *Handler) CreateEscalationPolicy(c *gin.Context) {
	var body model.EscalationPolicy
	if err := c.ShouldBindJSON(&body); err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid body")
		return
	}
	body.TenantID = h.tenant
	if body.ID == uuid.Nil {
		body.ID = uuid.New()
	}
	if err := h.service.CreateEscalationPolicy(c.Request.Context(), &body); err != nil {
		h.writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeSuccess(c, body)
}

func (h *Handler) UpdateEscalationPolicy(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid id")
		return
	}
	var body model.EscalationPolicy
	if err := c.ShouldBindJSON(&body); err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid body")
		return
	}
	body.ID = id
	body.TenantID = h.tenant
	if err := h.service.UpdateEscalationPolicy(c.Request.Context(), &body); err != nil {
		h.writeError(c, http.StatusNotFound, err.Error())
		return
	}
	h.writeSuccess(c, body)
}

func (h *Handler) DeleteEscalationPolicy(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.service.DeleteEscalationPolicy(c.Request.Context(), h.tenant, id); err != nil {
		h.writeError(c, http.StatusNotFound, err.Error())
		return
	}
	h.writeSuccess(c, gin.H{"deleted": true})
}

func (h *Handler) writeSuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"data":      data,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) writeError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"status":    "error",
		"errorCode": "ALRT001",
		"message":   message,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
