package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/domain"
	"github.com/neuralops/platform/internal/ingestion/config"
	"github.com/neuralops/platform/internal/ingestion/dto"
	"github.com/neuralops/platform/internal/ingestion/service"
	pmiddleware "github.com/neuralops/platform/internal/platform/middleware"
	"go.uber.org/zap"
)

// Handler exposes ingestion HTTP endpoints.
type Handler struct {
	cfg     *config.Config
	log     *zap.Logger
	service *service.Service
}

// New creates an ingestion HTTP handler.
func New(cfg *config.Config, log *zap.Logger, svc *service.Service) *Handler {
	return &Handler{cfg: cfg, log: log, service: svc}
}

// RegisterRoutes mounts ingestion routes on the router.
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", h.Health)
		v1.POST("/logs", h.IngestLogs)
		v1.POST("/metrics", h.IngestMetrics)
		v1.POST("/events", h.IngestEvents)
		v1.POST("/traces", h.IngestTraces)
		v1.POST("/webhooks/alerts", h.IngestAlertWebhook)
		v1.POST("/webhooks/dynatrace", h.IngestAlertWebhook)
		v1.POST("/webhooks/datadog", h.IngestAlertWebhook)
		v1.POST("/webhooks/prometheus", h.IngestAlertWebhook)
	}
}

// Health returns ingestion service health.
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"service": "ingestion",
			"healthy": true,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// IngestLogs handles single or batch log ingestion.
func (h *Handler) IngestLogs(c *gin.Context) {
	requests, err := decodeBatch[dto.LogIngestRequest](c, h.cfg.Ingestion.MaxBatchSize)
	if err != nil {
		h.writeError(c, http.StatusBadRequest, err)
		return
	}

	accepted, rejected, err := h.service.IngestLogs(c.Request.Context(), applyTenantLogs(requests, h.tenantID(c)), "http")
	if err != nil {
		h.log.Error("log ingestion failed", zap.Error(err))
		h.writeError(c, http.StatusInternalServerError, err)
		return
	}

	h.writeIngestResponse(c, accepted, rejected)
}

// IngestMetrics handles metric ingestion.
func (h *Handler) IngestMetrics(c *gin.Context) {
	requests, err := decodeBatch[dto.MetricIngestRequest](c, h.cfg.Ingestion.MaxBatchSize)
	if err != nil {
		h.writeError(c, http.StatusBadRequest, err)
		return
	}

	accepted, rejected, err := h.service.IngestMetrics(c.Request.Context(), applyTenantMetrics(requests, h.tenantID(c)))
	if err != nil {
		h.log.Error("metric ingestion failed", zap.Error(err))
		h.writeError(c, http.StatusInternalServerError, err)
		return
	}

	h.writeIngestResponse(c, accepted, rejected)
}

// IngestEvents handles deployment/config event ingestion.
func (h *Handler) IngestEvents(c *gin.Context) {
	requests, err := decodeBatch[dto.EventIngestRequest](c, h.cfg.Ingestion.MaxBatchSize)
	if err != nil {
		h.writeError(c, http.StatusBadRequest, err)
		return
	}

	accepted, rejected, err := h.service.IngestEvents(c.Request.Context(), applyTenantEvents(requests, h.tenantID(c)))
	if err != nil {
		h.log.Error("event ingestion failed", zap.Error(err))
		h.writeError(c, http.StatusInternalServerError, err)
		return
	}

	h.writeIngestResponse(c, accepted, rejected)
}

// IngestTraces handles trace span ingestion.
func (h *Handler) IngestTraces(c *gin.Context) {
	requests, err := decodeBatch[dto.TraceSpanRequest](c, h.cfg.Ingestion.MaxBatchSize)
	if err != nil {
		h.writeError(c, http.StatusBadRequest, err)
		return
	}

	accepted, rejected, err := h.service.IngestTraces(c.Request.Context(), applyTenantTraces(requests, h.tenantID(c)))
	if err != nil {
		h.log.Error("trace ingestion failed", zap.Error(err))
		h.writeError(c, http.StatusInternalServerError, err)
		return
	}

	h.writeIngestResponse(c, accepted, rejected)
}

// IngestAlertWebhook handles third-party alert webhooks.
func (h *Handler) IngestAlertWebhook(c *gin.Context) {
	var req dto.AlertWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, http.StatusBadRequest, domain.NewValidationError("body", err.Error()))
		return
	}

	if req.Source == "" {
		req.Source = "CUSTOM"
	}
	if req.FiredAt.IsZero() {
		req.FiredAt = time.Now().UTC()
	}

	if req.TenantID == "" {
		req.TenantID = h.tenantID(c)
	}

	if err := h.service.IngestAlertWebhook(c.Request.Context(), req); err != nil {
		if domain.IsValidation(err) {
			h.writeError(c, http.StatusTooManyRequests, err)
			return
		}
		h.writeError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"status":    "success",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) tenantID(c *gin.Context) string {
	return pmiddleware.TenantFromGin(c, "default")
}

func applyTenantLogs(requests []dto.LogIngestRequest, tenantID string) []dto.LogIngestRequest {
	for i := range requests {
		if requests[i].TenantID == "" {
			requests[i].TenantID = tenantID
		}
	}
	return requests
}

func applyTenantMetrics(requests []dto.MetricIngestRequest, tenantID string) []dto.MetricIngestRequest {
	for i := range requests {
		if requests[i].TenantID == "" {
			requests[i].TenantID = tenantID
		}
	}
	return requests
}

func applyTenantEvents(requests []dto.EventIngestRequest, tenantID string) []dto.EventIngestRequest {
	for i := range requests {
		if requests[i].TenantID == "" {
			requests[i].TenantID = tenantID
		}
	}
	return requests
}

func applyTenantTraces(requests []dto.TraceSpanRequest, tenantID string) []dto.TraceSpanRequest {
	for i := range requests {
		if requests[i].TenantID == "" {
			requests[i].TenantID = tenantID
		}
	}
	return requests
}

func decodeBatch[T any](c *gin.Context, maxBatch int) ([]T, error) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, domain.NewValidationError("body", "unable to read request body")
	}

	if len(body) == 0 {
		return nil, domain.NewValidationError("body", "must not be empty")
	}

	if body[0] == '[' {
		var batch []T
		if err := json.Unmarshal(body, &batch); err != nil {
			return nil, domain.NewValidationError("body", "invalid JSON array")
		}
		if len(batch) == 0 {
			return nil, domain.NewValidationError("body", "batch must not be empty")
		}
		if len(batch) > maxBatch {
			return nil, domain.NewValidationError("body", "batch exceeds maximum size")
		}
		return batch, nil
	}

	var single T
	if err := json.Unmarshal(body, &single); err != nil {
		return nil, domain.NewValidationError("body", "invalid JSON payload")
	}
	return []T{single}, nil
}

func (h *Handler) writeIngestResponse(c *gin.Context, accepted, rejected int) {
	status := http.StatusAccepted
	if accepted == 0 && rejected > 0 {
		status = http.StatusBadRequest
	}

	c.JSON(status, dto.IngestResponse{
		Status:    "success",
		Accepted:  accepted,
		Rejected:  rejected,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) writeError(c *gin.Context, status int, err error) {
	code := "ING001"
	message := err.Error()
	if domainErr, ok := domain.AsDomainError(err); ok {
		code = string(domainErr.Code)
		message = domainErr.Message
	}

	c.JSON(status, gin.H{
		"status":    "error",
		"errorCode": code,
		"message":   message,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
