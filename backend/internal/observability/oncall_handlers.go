package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// UpdateOncallSchedule updates an on-call schedule.
func (h *Handler) UpdateOncallSchedule(c *gin.Context) {
	id := c.Param("id")
	var body oncallSchedule
	if err := c.ShouldBindJSON(&body); err != nil || body.Team == "" {
		writeError(c, http.StatusBadRequest, "team required")
		return
	}
	tid := tenantID(c)
	if h.deps.Pool == nil {
		writeSuccess(c, body)
		return
	}
	rotJSON, _ := json.Marshal(body.Rotation)
	_, err := h.deps.Pool.Exec(c.Request.Context(), `
UPDATE oncall_schedules SET team=$3, rotation=$4, timezone=$5, enabled=$6, updated_at=NOW()
WHERE tenant_id=$1 AND id=$2`, tid, id, body.Team, rotJSON, body.Timezone, body.Enabled)
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	body.ID = id
	writeSuccess(c, body)
}

// DeleteOncallSchedule removes an on-call schedule.
func (h *Handler) DeleteOncallSchedule(c *gin.Context) {
	id := c.Param("id")
	tid := tenantID(c)
	if h.deps.Pool != nil {
		_, _ = h.deps.Pool.Exec(c.Request.Context(), `
DELETE FROM oncall_schedules WHERE tenant_id=$1 AND id=$2`, tid, id)
	}
	writeSuccess(c, gin.H{"deleted": true, "id": id})
}

// SyncOncallPagerDuty imports rotation from PagerDuty schedule API.
func (h *Handler) SyncOncallPagerDuty(c *gin.Context) {
	id := c.Param("id")
	tid := tenantID(c)
	var body struct {
		ScheduleID string `json:"scheduleId"`
	}
	_ = c.ShouldBindJSON(&body)
	cfg, err := h.integrationsRepo().GetConfig(c.Request.Context(), tid, "pagerduty")
	if err != nil {
		writeError(c, http.StatusBadRequest, "connect PagerDuty first")
		return
	}
	token := firstNonEmptyStr(cfg["apiToken"], cfg["accessToken"])
	if token == "" {
		writeError(c, http.StatusBadRequest, "pagerduty api token required")
		return
	}
	scheduleID := body.ScheduleID
	if scheduleID == "" && h.deps.Pool != nil {
		_ = h.deps.Pool.QueryRow(c.Request.Context(), `
SELECT COALESCE(pagerduty_schedule_id,'') FROM oncall_schedules WHERE tenant_id=$1 AND id=$2`, tid, id).Scan(&scheduleID)
	}
	if scheduleID == "" {
		writeError(c, http.StatusBadRequest, "scheduleId required")
		return
	}
	rotation, err := fetchPagerDutyRotation(c.Request.Context(), token, scheduleID)
	if err != nil {
		writeError(c, http.StatusBadGateway, err.Error())
		return
	}
	if h.deps.Pool != nil {
		rotJSON, _ := json.Marshal(rotation)
		_, _ = h.deps.Pool.Exec(c.Request.Context(), `
UPDATE oncall_schedules SET rotation=$3, pagerduty_schedule_id=$4, pagerduty_synced_at=NOW(), updated_at=NOW()
WHERE tenant_id=$1 AND id=$2`, tid, id, rotJSON, scheduleID)
	}
	writeSuccess(c, gin.H{"id": id, "scheduleId": scheduleID, "rotation": rotation})
}

func fetchPagerDutyRotation(ctx context.Context, token, scheduleID string) ([]struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	After string `json:"after"`
}, error) {
	url := fmt.Sprintf("https://api.pagerduty.com/schedules/%s/users", scheduleID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Token token="+token)
	req.Header.Set("Accept", "application/vnd.pagerduty+json;version=2")
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("pagerduty: %s", string(body))
	}
	var payload struct {
		Users []struct {
			User struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"user"`
		} `json:"users"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	out := make([]struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		After string `json:"after"`
	}, 0, len(payload.Users))
	for i, u := range payload.Users {
		after := "0m"
		if i > 0 {
			after = fmt.Sprintf("%dm", i*30)
		}
		out = append(out, struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			After string `json:"after"`
		}{Name: u.User.Name, Email: u.User.Email, After: after})
	}
	return out, nil
}

