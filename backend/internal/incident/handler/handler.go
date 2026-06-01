package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/domain"
	"github.com/neuralops/platform/internal/incident/engine"
	platmiddleware "github.com/neuralops/platform/internal/platform/middleware"
	"github.com/neuralops/platform/internal/incident/repository"
	"go.uber.org/zap"
)

// Handler exposes incident REST endpoints.
type Handler struct {
	log           *zap.Logger
	service       *engine.Service
	repo          *repository.Store
	txnStore      *repository.TransactionStore
	defaultTenant string
}

// New creates an incident handler.
func New(log *zap.Logger, service *engine.Service, repo *repository.Store, txnStore *repository.TransactionStore, defaultTenant string) *Handler {
	return &Handler{log: log, service: service, repo: repo, txnStore: txnStore, defaultTenant: defaultTenant}
}

// RegisterRoutes mounts incident API routes.
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.Use(platmiddleware.ExtractTenant(h.defaultTenant))
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", h.Health)
		v1.GET("/incidents", h.ListIncidents)
		v1.GET("/incidents/:id", h.GetIncident)
		v1.POST("/incidents/:id/acknowledge", h.AcknowledgeIncident)
		v1.POST("/incidents/:id/resolve", h.ResolveIncident)
		v1.GET("/incidents/:id/recommendations", h.GetRecommendations)
		v1.GET("/incidents/:id/timeline", h.GetTimeline)
		v1.GET("/services/dependency-map", h.GetDependencyMap)
		v1.GET("/transactions/:txnId", h.GetTransaction)
	}
}

func (h *Handler) tenantID(c *gin.Context) string {
	return platmiddleware.TenantFromGin(c, h.defaultTenant)
}

// Health returns service health.
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{"service": "incident", "healthy": true},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// ListIncidents lists incidents with filters.
func (h *Handler) ListIncidents(c *gin.Context) {
	filter := repository.IncidentFilter{
		TenantID: h.tenantID(c),
		Status:   c.Query("status"),
		Severity: c.Query("severity"),
		Service:  c.Query("service"),
	}
	if limit, err := strconv.Atoi(c.DefaultQuery("size", "50")); err == nil {
		filter.Limit = limit
	}
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && page > 0 {
		filter.Offset = (page - 1) * filter.Limit
	}
	if start := c.Query("startTime"); start != "" {
		if parsed, err := time.Parse(time.RFC3339, start); err == nil {
			filter.StartTime = &parsed
		}
	}
	if end := c.Query("endTime"); end != "" {
		if parsed, err := time.Parse(time.RFC3339, end); err == nil {
			filter.EndTime = &parsed
		}
	}

	incidents, err := h.repo.ListIncidents(c.Request.Context(), filter)
	if err != nil {
		h.writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": incidents, "timestamp": time.Now().UTC().Format(time.RFC3339)})
}

// GetIncident returns incident detail.
func (h *Handler) GetIncident(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.writeError(c, http.StatusBadRequest, domain.NewValidationError("id", "invalid uuid"))
		return
	}
	incident, err := h.repo.GetIncident(c.Request.Context(), h.tenantID(c), id)
	if err != nil {
		h.writeError(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": incident, "timestamp": time.Now().UTC().Format(time.RFC3339)})
}

type resolveRequest struct {
	Notes string `json:"notes"`
}

// AcknowledgeIncident acknowledges an incident.
func (h *Handler) AcknowledgeIncident(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.writeError(c, http.StatusBadRequest, domain.NewValidationError("id", "invalid uuid"))
		return
	}
	incident, err := h.service.Acknowledge(c.Request.Context(), h.tenantID(c), id)
	if err != nil {
		h.writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": incident, "timestamp": time.Now().UTC().Format(time.RFC3339)})
}

// ResolveIncident resolves an incident.
func (h *Handler) ResolveIncident(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.writeError(c, http.StatusBadRequest, domain.NewValidationError("id", "invalid uuid"))
		return
	}
	var req resolveRequest
	_ = c.ShouldBindJSON(&req)
	incident, err := h.service.Resolve(c.Request.Context(), h.tenantID(c), id, req.Notes)
	if err != nil {
		h.writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": incident, "timestamp": time.Now().UTC().Format(time.RFC3339)})
}

// GetRecommendations returns AI recommendations.
func (h *Handler) GetRecommendations(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.writeError(c, http.StatusBadRequest, domain.NewValidationError("id", "invalid uuid"))
		return
	}
	recs, err := h.service.GetRecommendations(c.Request.Context(), h.tenantID(c), id)
	if err != nil {
		h.writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": recs, "timestamp": time.Now().UTC().Format(time.RFC3339)})
}

// GetTimeline returns incident timeline events.
func (h *Handler) GetTimeline(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		h.writeError(c, http.StatusBadRequest, domain.NewValidationError("id", "invalid uuid"))
		return
	}
	incident, err := h.repo.GetIncident(c.Request.Context(), h.tenantID(c), id)
	if err != nil {
		h.writeError(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"timeline":  incident.Timeline,
			"narrative": engine.BuildNarrative(*incident),
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// GetDependencyMap returns the service dependency graph.
func (h *Handler) GetDependencyMap(c *gin.Context) {
	graph, err := h.repo.ListDependencies(c.Request.Context(), h.tenantID(c))
	if err != nil {
		h.writeError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": graph, "timestamp": time.Now().UTC().Format(time.RFC3339)})
}

// GetTransaction returns a banking transaction journey.
func (h *Handler) GetTransaction(c *gin.Context) {
	if h.txnStore == nil {
		h.writeError(c, http.StatusServiceUnavailable, domain.NewDomainError(domain.ErrCodeExternalDependency, "transaction store unavailable"))
		return
	}
	txn, err := h.txnStore.GetTransaction(c.Request.Context(), c.Param("txnId"))
	if err != nil {
		h.writeError(c, http.StatusNotFound, domain.NewNotFoundError("transaction", c.Param("txnId")))
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": txn, "timestamp": time.Now().UTC().Format(time.RFC3339)})
}

func (h *Handler) writeError(c *gin.Context, status int, err error) {
	code := "INC001"
	message := err.Error()
	if domainErr, ok := domain.AsDomainError(err); ok {
		code = string(domainErr.Code)
		message = domainErr.Message
	}
	c.JSON(status, gin.H{"status": "error", "errorCode": code, "message": message, "timestamp": time.Now().UTC().Format(time.RFC3339)})
}
