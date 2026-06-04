package observability

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// alert_feedback integrates with gap_closure_store fatigue weights (GAP-MET-004).

// AlertQualityFeedback is user feedback on alert noise/usefulness.
type AlertQualityFeedback struct {
	ID        string    `json:"id"`
	PolicyID  string    `json:"policyId"`
	Service   string    `json:"service"`
	Helpful   bool      `json:"helpful"`
	Comment   string    `json:"comment,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

var (
	alertFeedbackMu sync.RWMutex
	alertFeedback   []AlertQualityFeedback
)

func (h *Handler) SubmitAlertPolicyFeedback(c *gin.Context) {
	var body struct {
		Service string `json:"service"`
		Helpful bool   `json:"helpful"`
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, 400, "invalid body")
		return
	}
	fb := AlertQualityFeedback{
		ID: uuid.New().String()[:8], PolicyID: c.Param("id"),
		Service: body.Service, Helpful: body.Helpful, Comment: body.Comment,
		CreatedAt: time.Now().UTC(),
	}
	alertFeedbackMu.Lock()
	alertFeedback = append(alertFeedback, fb)
	alertFeedbackMu.Unlock()
	score := h.deps.Mem.ApplyAlertFeedbackToFatigue(c.Param("id"), body.Helpful)
	writeSuccess(c, gin.H{"feedback": fb, "fatigueWeight": score})
}

func (h *Handler) ListAlertPolicyFeedback(c *gin.Context) {
	alertFeedbackMu.RLock()
	defer alertFeedbackMu.RUnlock()
	id := c.Param("id")
	out := make([]AlertQualityFeedback, 0)
	for _, f := range alertFeedback {
		if f.PolicyID == id {
			out = append(out, f)
		}
	}
	writeSuccess(c, out)
}
