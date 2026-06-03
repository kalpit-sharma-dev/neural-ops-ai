package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/gateway/auth"
	"github.com/neuralops/platform/internal/gateway/chat"
	"github.com/neuralops/platform/internal/gateway/config"
	"github.com/neuralops/platform/internal/gateway/dashboard"
	"github.com/neuralops/platform/internal/gateway/websocket"
	"github.com/neuralops/platform/internal/platform/health"
	"go.uber.org/zap"
)

// Handler exposes gateway-native endpoints.
type Handler struct {
	log         *zap.Logger
	cfg         config.Config
	dashboard   *dashboard.Aggregator
	chat        *chat.Service
	wsHub       *websocket.Hub
	logTailHub  *websocket.LogTailHub
}

// New creates a gateway handler.
func New(log *zap.Logger, cfg config.Config, dashboardAgg *dashboard.Aggregator, chatSvc *chat.Service, wsHub *websocket.Hub, logTailHub *websocket.LogTailHub) *Handler {
	return &Handler{
		log:        log,
		cfg:        cfg,
		dashboard:  dashboardAgg,
		chat:       chatSvc,
		wsHub:      wsHub,
		logTailHub: logTailHub,
	}
}

// RegisterNativeRoutes mounts gateway-owned routes.
func (h *Handler) RegisterNativeRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		v1.GET("/dashboard/overview", h.DashboardOverview)
		v1.POST("/chat/query", h.ChatQuery)
	}
	router.GET("/api/v1/ws/realtime", h.RealtimeWebSocket)
	if h.logTailHub != nil {
		router.GET("/api/v1/logs/tail/ws", h.LogTailWebSocket)
	}
}

// Info returns platform metadata for clients.
func (h *Handler) Info(c *gin.Context) {
	writeSuccess(c, gin.H{
		"service":      "gateway",
		"environment":  h.cfg.Server.Environment,
		"version":      health.DefaultVersion,
		"demoMode":     h.cfg.DemoMode,
		"capabilities": DefaultCapabilities(h.cfg.DemoMode),
	})
}

// DashboardOverview aggregates dashboard metrics.
func (h *Handler) DashboardOverview(c *gin.Context) {
	overview, err := h.dashboard.FetchOverview(c.Request.Context())
	if err != nil {
		h.log.Warn("dashboard aggregation partial failure", zap.Error(err))
	}
	writeSuccess(c, overview)
}

// ChatQuery streams an AI answer using Server-Sent Events.
func (h *Handler) ChatQuery(c *gin.Context) {
	var req chat.QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		writeError(c, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	err := h.chat.StreamAnswer(c.Request.Context(), req.Question, func(chunk string) error {
		_, err := fmt.Fprintf(c.Writer, "data: %s\n\n", chunk)
		if err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}, func(sources []chat.SourceRef) error {
		payload, err := json.Marshal(sources)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(c.Writer, "event: sources\ndata: %s\n\n", payload)
		if err != nil {
			return err
		}
		flusher.Flush()
		return nil
	})
	if err != nil {
		_, _ = fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", err.Error())
		flusher.Flush()
		return
	}
	_, _ = fmt.Fprintf(c.Writer, "event: done\ndata: [DONE]\n\n")
	flusher.Flush()
}

// RealtimeWebSocket upgrades to websocket for dashboard events.
func (h *Handler) RealtimeWebSocket(c *gin.Context) {
	h.wsHub.Handle(c.Writer, c.Request, h.cfg.WebSocket.PingInterval)
}

// LogTailWebSocket streams live logs (LOG-04).
func (h *Handler) LogTailWebSocket(c *gin.Context) {
	if env := h.cfg.Server.Environment; env == "production" || env == "staging" {
		if p, ok := auth.PrincipalFromGin(c); ok && p.Role == auth.RoleReadOnly {
			c.JSON(http.StatusForbidden, gin.H{
				"status": "error", "errorCode": "LOGTAIL001",
				"message": "read-only role cannot use live log tail in regulated environments",
			})
			return
		}
	}
	h.logTailHub.Handle(c.Writer, c.Request, h.cfg.WebSocket.PingInterval)
}

func writeSuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"data":      data,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func writeError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"status":    "error",
		"errorCode": "GWY002",
		"message":   message,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
