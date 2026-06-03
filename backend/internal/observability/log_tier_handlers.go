package observability

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// LogTierPolicy defines hot/warm/cold retention per signal (LOG-05).
type LogTierPolicy struct {
	TenantID       string    `json:"tenantId"`
	HotRetentionDays  int    `json:"hotRetentionDays"`
	WarmRetentionDays int    `json:"warmRetentionDays"`
	ColdRetentionDays int    `json:"coldRetentionDays"`
	RestoreSLAHours   int    `json:"restoreSlaHours"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (h *Handler) registerLogRoutes(v1 *gin.RouterGroup) {
	logs := v1.Group("/logs")
	{
		logs.GET("/tiering", h.GetLogTierPolicy)
		logs.PUT("/tiering", h.UpdateLogTierPolicy)
	}
}

func (h *Handler) GetLogTierPolicy(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.LogTierPolicy(tenantID(c)))
}

func (h *Handler) UpdateLogTierPolicy(c *gin.Context) {
	var body LogTierPolicy
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.HotRetentionDays <= 0 {
		body.HotRetentionDays = 7
	}
	if body.WarmRetentionDays <= body.HotRetentionDays {
		body.WarmRetentionDays = 30
	}
	if body.ColdRetentionDays <= body.WarmRetentionDays {
		body.ColdRetentionDays = 90
	}
	if body.RestoreSLAHours <= 0 {
		body.RestoreSLAHours = 4
	}
	writeSuccess(c, h.deps.Mem.SaveLogTierPolicy(tenantID(c), body))
}
