package observability

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/gateway/auth"
)

func (h *Handler) registerAIRoutes(v1 *gin.RouterGroup) {
	ai := v1.Group("/ai")
	{
		ai.GET("/rca/:incidentId", h.GetIncidentRCA)
		ai.GET("/explanations/:id", h.GetAIExplanation)
		ai.POST("/forecast", h.PostAIForecast)
		ai.POST("/autofix/plan", h.PostAutoFixPlan)
		ai.POST("/autofix/execute", h.PostAutoFixExecute)
		ai.POST("/autofix/rollback", h.PostAutoFixRollback)
		ai.GET("/llm/workloads", h.ListLLMWorkloads)
		ai.GET("/llm/workloads/:id/traces", h.ListLLMWorkloadTraces)
		ai.GET("/llm/usage", h.GetLLMUsageSummary)
	}
}

func (h *Handler) autofixAllowed(c *gin.Context) bool {
	p := LoadAutoFixPolicy()
	if !p.Enabled {
		writeError(c, http.StatusForbidden, "AutoFix is disabled for this environment")
		return false
	}
	if pr, ok := auth.PrincipalFromGin(c); ok {
		if !p.RoleMayExecute(string(pr.Role)) {
			writeError(c, http.StatusForbidden, "role not permitted for AutoFix")
			return false
		}
	}
	return true
}

func (h *Handler) GetIncidentRCA(c *gin.Context) {
	incidentID := c.Param("incidentId")
	if incidentID == "" {
		writeError(c, http.StatusBadRequest, "incidentId required")
		return
	}
	rca, ok := h.deps.Mem.RCAForIncident(incidentID)
	if !ok {
		writeError(c, http.StatusNotFound, "rca not found for incident")
		return
	}
	writeSuccess(c, rca)
}

func (h *Handler) GetAIExplanation(c *gin.Context) {
	exp, ok := h.deps.Mem.GetAIExplanation(c.Param("id"))
	if !ok {
		writeError(c, http.StatusNotFound, "explanation not found")
		return
	}
	writeSuccess(c, exp)
}

type forecastRequest struct {
	Metric  string `json:"metric"`
	Service string `json:"service"`
	Horizon string `json:"horizon"`
}

func (h *Handler) PostAIForecast(c *gin.Context) {
	var req forecastRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Metric == "" {
		writeError(c, http.StatusBadRequest, "metric is required")
		return
	}
	writeSuccess(c, h.deps.Mem.Forecast(req.Metric, req.Service, req.Horizon))
}

type autofixPlanRequest struct {
	IncidentID string `json:"incidentId"`
}

func (h *Handler) PostAutoFixPlan(c *gin.Context) {
	if !h.autofixAllowed(c) {
		return
	}
	var req autofixPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.IncidentID == "" {
		writeError(c, http.StatusBadRequest, "incidentId is required")
		return
	}
	plan := h.deps.Mem.CreateAutoFixPlan(req.IncidentID)
	pol := LoadAutoFixPolicy()
	if pol.RequireApproval {
		plan.RequiresApproval = true
		if plan.PolicyReason == "" {
			plan.PolicyReason = "Tenant policy requires explicit human approval before execution."
		}
	}
	writeSuccess(c, plan)
}

type autofixExecuteRequest struct {
	PlanID   string `json:"planId"`
	Approved bool   `json:"approved"`
}

func (h *Handler) PostAutoFixExecute(c *gin.Context) {
	if !h.autofixAllowed(c) {
		return
	}
	var req autofixExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.PlanID == "" {
		writeError(c, http.StatusBadRequest, "planId is required")
		return
	}
	if LoadAutoFixPolicy().RequireApproval && !req.Approved {
		writeError(c, http.StatusForbidden, "approved=true required for AutoFix execution")
		return
	}
	action, err := h.deps.Mem.ExecuteAutoFix(req.PlanID, req.Approved)
	if err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	writeSuccess(c, action)
}

type autofixRollbackRequest struct {
	ActionID string `json:"actionId"`
}

func (h *Handler) ListLLMWorkloads(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListLLMWorkloads())
}

func (h *Handler) GetLLMUsageSummary(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.LLMUsageSummary(c.Query("window")))
}

func (h *Handler) ListLLMWorkloadTraces(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.LLMWorkloadTraces(c.Param("id")))
}

func (h *Handler) PostAutoFixRollback(c *gin.Context) {
	if !h.autofixAllowed(c) {
		return
	}
	var req autofixRollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ActionID == "" {
		writeError(c, http.StatusBadRequest, "actionId is required")
		return
	}
	action, err := h.deps.Mem.RollbackAutoFix(req.ActionID)
	if err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	writeSuccess(c, action)
}
