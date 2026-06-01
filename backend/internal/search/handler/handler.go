package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	platmiddleware "github.com/neuralops/platform/internal/platform/middleware"
	"github.com/neuralops/platform/internal/search/dto"
	"github.com/neuralops/platform/internal/search/service"
	"go.uber.org/zap"
)

// Handler exposes search REST endpoints.
type Handler struct {
	log           *zap.Logger
	service       *service.Service
	defaultTenant string
}

// New creates a search handler.
func New(log *zap.Logger, svc *service.Service, defaultTenant string) *Handler {
	return &Handler{log: log, service: svc, defaultTenant: defaultTenant}
}

// RegisterRoutes mounts search API routes.
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.Use(platmiddleware.ExtractTenant(h.defaultTenant))
	v1 := router.Group("/api/v1")
	{
		v1.POST("/search/logs", h.SearchLogs)
		v1.POST("/search/semantic", h.SemanticSearch)
		v1.GET("/search/trace/:traceId", h.SearchTrace)
		v1.GET("/search/txn/:txnId", h.GetTransaction)
		v1.POST("/search/ai", h.AISearch)
		v1.POST("/search/transactions", h.SearchTransactions)
	}
}

func (h *Handler) tenantID(c *gin.Context) string {
	return platmiddleware.TenantFromGin(c, h.defaultTenant)
}

// SearchLogs handles structured log search.
func (h *Handler) SearchLogs(c *gin.Context) {
	var req dto.LogSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	req.TenantID = h.tenantID(c)
	allowed := platmiddleware.AllowedServicesFromGin(c)
	if platmiddleware.AbortIfServiceForbidden(c, allowed, req.Service) {
		return
	}
	req.AllowedServices = allowed

	result, err := h.service.SearchLogs(c.Request.Context(), req)
	if err != nil {
		h.writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeSuccess(c, result)
}

// SemanticSearch handles vector and hybrid search.
func (h *Handler) SemanticSearch(c *gin.Context) {
	var req dto.SemanticSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	req.TenantID = h.tenantID(c)

	result, err := h.service.SemanticSearch(c.Request.Context(), req)
	if err != nil {
		h.writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeSuccess(c, result)
}

// SearchTrace returns logs for a trace ID.
func (h *Handler) SearchTrace(c *gin.Context) {
	traceID := c.Param("traceId")
	size, _ := strconv.Atoi(c.DefaultQuery("size", "200"))

	result, err := h.service.SearchTrace(c.Request.Context(), traceID, h.tenantID(c), size)
	if err != nil {
		h.writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	h.writeSuccess(c, result)
}

// GetTransaction returns a transaction journey.
func (h *Handler) GetTransaction(c *gin.Context) {
	txnID := c.Param("txnId")
	txn, err := h.service.GetTransaction(c.Request.Context(), txnID)
	if err != nil {
		h.writeError(c, http.StatusNotFound, err.Error())
		return
	}
	h.writeSuccess(c, txn)
}

// SearchTransactions searches transaction journeys with filters.
func (h *Handler) SearchTransactions(c *gin.Context) {
	var req dto.TransactionSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	results, err := h.service.SearchTransactions(c.Request.Context(), req)
	if err != nil {
		h.writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeSuccess(c, results)
}

// AISearch handles natural language search.
func (h *Handler) AISearch(c *gin.Context) {
	var req dto.AISearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	parsed, result, err := h.service.AISearch(c.Request.Context(), req, h.tenantID(c), platmiddleware.AllowedServicesFromGin(c))
	if err != nil {
		h.writeError(c, http.StatusInternalServerError, err.Error())
		return
	}

	h.writeSuccess(c, gin.H{
		"parsed":  parsed,
		"results": result,
	})
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
		"errorCode": "SRCH001",
		"message":   message,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
