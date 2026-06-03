package observability

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/finops"
	"github.com/neuralops/platform/internal/finops/domain"
	"github.com/neuralops/platform/internal/gateway/auth"
)

func (h *Handler) registerFinOpsRoutes(finopsGroup *gin.RouterGroup) {
	finopsGroup.GET("/costs/breakdown", h.GetFinOpsCostBreakdown)
	finopsGroup.GET("/allocation/rules", h.ListFinOpsAllocationRules)
	finopsGroup.POST("/allocation/rules", h.CreateFinOpsAllocationRule)
	finopsGroup.PUT("/allocation/rules/:id", h.UpdateFinOpsAllocationRule)
	finopsGroup.DELETE("/allocation/rules/:id", h.DeleteFinOpsAllocationRule)
	finopsGroup.GET("/kubernetes/cost", h.GetFinOpsKubernetesCost)
	finopsGroup.GET("/anomalies/:id", h.GetFinOpsAnomaly)
	finopsGroup.POST("/anomalies/:id/feedback", h.SubmitFinOpsAnomalyFeedback)
	finopsGroup.GET("/recommendations", h.ListFinOpsRecommendations)
	finopsGroup.GET("/recommendations/:id", h.GetFinOpsRecommendation)
	finopsGroup.POST("/recommendations/:id/actions", h.FinOpsRecommendationAction)
	finopsGroup.GET("/budgets", h.ListFinOpsBudgets)
	finopsGroup.GET("/budgets/alerts", h.ListFinOpsBudgetAlerts)
	finopsGroup.POST("/budgets", h.CreateFinOpsBudget)
	finopsGroup.PUT("/budgets/:id", h.UpdateFinOpsBudget)
	finopsGroup.DELETE("/budgets/:id", h.DeleteFinOpsBudget)
	finopsGroup.GET("/forecast", h.GetFinOpsForecast)
	finopsGroup.POST("/ingest/run", h.RunFinOpsIngest)
	finopsGroup.GET("/ingest/status", h.GetFinOpsIngestStatus)
	finopsGroup.GET("/reconciliation", h.ListFinOpsReconciliation)
	finopsGroup.GET("/audit/export", h.ExportFinOpsAudit)
	finopsGroup.GET("/line-items/lake/export", h.ExportFinOpsLakeLineItems)
	finopsGroup.POST("/imports", h.ImportFinOpsCost)
	finopsGroup.GET("/allocation/tag-suggestions", h.ListFinOpsTagSuggestions)
	finopsGroup.GET("/allocation/shared-splits", h.ListFinOpsSharedSplits)
	finopsGroup.GET("/commitments", h.ListFinOpsCommitments)
	finopsGroup.GET("/commitments/recommendations", h.ListFinOpsCommitmentRecommendations)
	finopsGroup.GET("/unit-economics", h.GetFinOpsUnitEconomics)
	finopsGroup.GET("/carbon/recommendations", h.ListFinOpsCarbonRecommendations)
	finopsGroup.GET("/reports", h.ListFinOpsReports)
	finopsGroup.POST("/reports/schedule", h.CreateFinOpsReportSchedule)
	finopsGroup.GET("/audit", h.ListFinOpsAuditLog)
	finopsGroup.GET("/chargeback/statements", h.ListFinOpsChargebackStatements)
	finopsGroup.POST("/chargeback/statements", h.GenerateFinOpsChargebackStatement)
	finopsGroup.GET("/chargeback/statements/:id/export", h.ExportFinOpsChargebackStatement)
	finopsGroup.POST("/scenarios", h.RunFinOpsScenario)
	finopsGroup.GET("/scenarios", h.ListFinOpsScenarios)
	finopsGroup.GET("/commitments/alerts", h.ListFinOpsCommitmentAlerts)
	finopsGroup.POST("/carbon/recommendations/:id/actions", h.FinOpsCarbonAction)
	finopsGroup.GET("/governance/policies", h.ListFinOpsGovernancePolicies)
	finopsGroup.PUT("/governance/policies/:id", h.UpsertFinOpsGovernancePolicy)
}

func (h *Handler) finopsSvc() *finops.Service {
	if h.deps.FinOps != nil {
		return h.deps.FinOps
	}
	svc := finops.NewService(&FinOpsCloudAdapter{Store: h.deps.Mem})
	alerter := &FinOpsAnomalyAlerter{Mem: h.deps.Mem, Stream: h.deps.Stream}
	svc.SetAlerter(alerter)
	svc.SetBudgetAlerter(alerter)
	svc.SetStaleIngestAlerter(alerter)
	if h.deps.Log != nil {
		svc.SetLogger(h.deps.Log)
	}
	return svc
}

func writeFinOpsList(c *gin.Context, items any, page finops.PageMeta, paged bool) {
	if !paged {
		writeSuccess(c, items)
		return
	}
	writeSuccess(c, gin.H{"items": items, "page": page})
}

func (h *Handler) finopsScopeAllowed(c *gin.Context, scope string) bool {
	principal, ok := auth.PrincipalFromGin(c)
	if !ok {
		return finops.AllowedScope(nil, scope)
	}
	return finops.AllowedScope(principal.Services, scope)
}

func parseFinOpsTimeRange(c *gin.Context) (from, to time.Time) {
	to = time.Now().UTC()
	from = to.AddDate(0, 0, -30)
	if v := c.Query("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			from = t
		}
	}
	if v := c.Query("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			to = t
		}
	}
	return from, to
}

func (h *Handler) GetFinOpsCosts(c *gin.Context) {
	start := time.Now()
	defer observeFinOpsQuery("costs", start)
	scope := c.Query("scope")
	if scope == "" {
		scope = "all"
	}
	if err := finops.ValidateFilter(scope, "scope"); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := finops.ValidateFilter(c.Query("provider"), "provider"); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	if !h.finopsScopeAllowed(c, scope) {
		writeError(c, http.StatusForbidden, "cost scope not permitted")
		return
	}
	from, to := parseFinOpsTimeRange(c)
	series, err := h.finopsSvc().GetCosts(c.Request.Context(), tenantID(c), domain.CostsQuery{
		Scope: scope, Provider: c.Query("provider"), Granularity: c.Query("granularity"),
		From: from, To: to, GroupBy: c.Query("groupBy"), CostView: c.Query("costView"),
	})
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(c, series)
}

func (h *Handler) ListFinOpsAnomalies(c *gin.Context) {
	for _, pair := range [][2]string{{c.Query("scope"), "scope"}, {c.Query("severity"), "severity"}, {c.Query("status"), "status"}} {
		if err := finops.ValidateFilter(pair[0], pair[1]); err != nil {
			writeError(c, http.StatusBadRequest, err.Error())
			return
		}
	}
	list, err := h.finopsSvc().ListAnomalies(c.Request.Context(), tenantID(c),
		c.Query("scope"), c.Query("severity"), c.Query("status"))
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if len(list) == 0 {
		legacy := h.deps.Mem.FinOpsCostAnomalies()
		writeSuccess(c, legacy)
		return
	}
	page, size, sortKey, paged := finops.ParsePage(c)
	items, meta := finops.Paginate(list, page, size, sortKey, func(a domain.Anomaly) string { return a.Service })
	writeFinOpsList(c, items, meta, paged)
}

func (h *Handler) GetFinOpsCarbon(c *gin.Context) {
	scope := c.DefaultQuery("scope", "all")
	dimension := c.DefaultQuery("dimension", "team")
	fp, err := h.finopsSvc().GetCarbonFootprint(c.Request.Context(), tenantID(c), scope, dimension)
	if err != nil {
		writeSuccess(c, h.finopsSvc().LegacyCarbon(scope))
		return
	}
	writeSuccess(c, fp)
}

func (h *Handler) GetFinOpsCostBreakdown(c *gin.Context) {
	start := time.Now()
	defer observeFinOpsQuery("costs_breakdown", start)
	dim := c.DefaultQuery("dimension", "team")
	if err := finops.ValidateFilter(dim, "dimension"); err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	from, to := parseFinOpsTimeRange(c)
	node, err := h.finopsSvc().CostBreakdown(c.Request.Context(), tenantID(c), dim, from, to)
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(c, node)
}

func (h *Handler) ListFinOpsAllocationRules(c *gin.Context) {
	rules, err := h.finopsSvc().ListAllocationRules(c.Request.Context(), tenantID(c))
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	page, size, sortKey, paged := finops.ParsePage(c)
	items, meta := finops.Paginate(rules, page, size, sortKey, func(r domain.AllocationRule) string { return r.Name })
	writeFinOpsList(c, items, meta, paged)
}

func (h *Handler) CreateFinOpsAllocationRule(c *gin.Context) {
	var body domain.AllocationRule
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "invalid body")
		return
	}
	r, err := h.finopsSvc().CreateAllocationRuleAudited(c.Request.Context(), tenantID(c), body, finopsActor(c))
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(c, r)
}

func (h *Handler) UpdateFinOpsAllocationRule(c *gin.Context) {
	var body domain.AllocationRule
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "invalid body")
		return
	}
	body.ID = c.Param("id")
	r, err := h.finopsSvc().UpdateAllocationRuleAudited(c.Request.Context(), tenantID(c), body, finopsActor(c))
	if err != nil {
		writeError(c, http.StatusNotFound, err.Error())
		return
	}
	writeSuccess(c, r)
}

func (h *Handler) DeleteFinOpsAllocationRule(c *gin.Context) {
	if err := h.finopsSvc().DeleteAllocationRuleAudited(c.Request.Context(), tenantID(c), c.Param("id"), finopsActor(c)); err != nil {
		writeError(c, http.StatusNotFound, err.Error())
		return
	}
	writeSuccess(c, gin.H{"deleted": true})
}

func (h *Handler) GetFinOpsKubernetesCost(c *gin.Context) {
	from, to := parseFinOpsTimeRange(c)
	rows, err := h.finopsSvc().KubernetesCost(c.Request.Context(), tenantID(c),
		c.Query("cluster"), c.Query("namespace"), from, to)
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(c, rows)
}

func (h *Handler) GetFinOpsAnomaly(c *gin.Context) {
	a, err := h.finopsSvc().GetAnomalyWithCause(c.Request.Context(), tenantID(c), c.Param("id"))
	if err != nil {
		writeError(c, http.StatusNotFound, err.Error())
		return
	}
	writeSuccess(c, a)
}

func (h *Handler) SubmitFinOpsAnomalyFeedback(c *gin.Context) {
	var body struct {
		Feedback string `json:"feedback"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "invalid body")
		return
	}
	feedback := strings.ToLower(strings.TrimSpace(body.Feedback))
	if feedback != "confirm" && feedback != "false_positive" && feedback != "expected" {
		writeError(c, http.StatusBadRequest, "feedback must be confirm|false_positive|expected")
		return
	}
	a, err := h.finopsSvc().SubmitAnomalyFeedbackWithTuning(c.Request.Context(), tenantID(c), c.Param("id"), feedback, finopsActor(c))
	if err != nil {
		writeError(c, http.StatusNotFound, err.Error())
		return
	}
	writeSuccess(c, a)
}

func (h *Handler) ListFinOpsRecommendations(c *gin.Context) {
	for _, pair := range [][2]string{{c.Query("type"), "type"}, {c.Query("scope"), "scope"}, {c.Query("status"), "status"}} {
		if err := finops.ValidateFilter(pair[0], pair[1]); err != nil {
			writeError(c, http.StatusBadRequest, err.Error())
			return
		}
	}
	recs, err := h.finopsSvc().ListRecommendations(c.Request.Context(), tenantID(c),
		c.Query("type"), c.Query("scope"), c.Query("status"))
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	page, size, sortKey, paged := finops.ParsePage(c)
	items, meta := finops.Paginate(recs, page, size, sortKey, func(r domain.Recommendation) string { return r.Title })
	writeFinOpsList(c, items, meta, paged)
}

func (h *Handler) GetFinOpsRecommendation(c *gin.Context) {
	r, err := h.finopsSvc().GetRecommendation(c.Request.Context(), tenantID(c), c.Param("id"))
	if err != nil {
		writeError(c, http.StatusNotFound, err.Error())
		return
	}
	writeSuccess(c, r)
}

func (h *Handler) FinOpsRecommendationAction(c *gin.Context) {
	var body struct {
		Action string `json:"action"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "invalid body")
		return
	}
	r, err := h.finopsSvc().RecommendationActionWithAudit(c.Request.Context(), tenantID(c), c.Param("id"), body.Action, finopsActor(c))
	if err != nil {
		writeError(c, http.StatusNotFound, err.Error())
		return
	}
	writeSuccess(c, r)
}

func (h *Handler) ListFinOpsBudgets(c *gin.Context) {
	start := time.Now()
	defer observeFinOpsQuery("budgets", start)
	budgets, err := h.finopsSvc().ListBudgets(c.Request.Context(), tenantID(c))
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	page, size, sortKey, paged := finops.ParsePage(c)
	items, meta := finops.Paginate(budgets, page, size, sortKey, func(b domain.Budget) string { return b.Name })
	writeFinOpsList(c, items, meta, paged)
}

func (h *Handler) ListFinOpsBudgetAlerts(c *gin.Context) {
	start := time.Now()
	defer observeFinOpsQuery("budget_alerts", start)
	alerts, err := h.finopsSvc().ListBudgetAlerts(c.Request.Context(), tenantID(c))
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(c, alerts)
}

func (h *Handler) CreateFinOpsBudget(c *gin.Context) {
	var body domain.Budget
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "invalid body")
		return
	}
	b, err := h.finopsSvc().CreateBudgetAudited(c.Request.Context(), tenantID(c), body, finopsActor(c))
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(c, b)
}

func (h *Handler) UpdateFinOpsBudget(c *gin.Context) {
	var body domain.Budget
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "invalid body")
		return
	}
	body.ID = c.Param("id")
	b, err := h.finopsSvc().UpdateBudgetAudited(c.Request.Context(), tenantID(c), body, finopsActor(c))
	if err != nil {
		writeError(c, http.StatusNotFound, err.Error())
		return
	}
	writeSuccess(c, b)
}

func (h *Handler) DeleteFinOpsBudget(c *gin.Context) {
	if err := h.finopsSvc().DeleteBudgetAudited(c.Request.Context(), tenantID(c), c.Param("id"), finopsActor(c)); err != nil {
		writeError(c, http.StatusNotFound, err.Error())
		return
	}
	writeSuccess(c, gin.H{"deleted": true})
}

func (h *Handler) GetFinOpsForecast(c *gin.Context) {
	fc, err := h.finopsSvc().Forecast(c.Request.Context(), tenantID(c), c.Query("scope"), c.DefaultQuery("horizon", "month"))
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(c, fc)
}

func (h *Handler) RunFinOpsIngest(c *gin.Context) {
	snaps, err := h.finopsSvc().IngestAll(c.Request.Context(), tenantID(c))
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(c, gin.H{"snapshots": snaps, "count": len(snaps)})
}

func (h *Handler) GetFinOpsIngestStatus(c *gin.Context) {
	status := h.finopsSvc().IngestStatus(c.Request.Context(), tenantID(c))
	writeSuccess(c, status)
}

func (h *Handler) ListFinOpsReconciliation(c *gin.Context) {
	list, err := h.finopsSvc().ListReconciliation(c.Request.Context(), tenantID(c))
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []domain.IngestSnapshot{}
	}
	writeSuccess(c, list)
}

func (h *Handler) ExportFinOpsAudit(c *gin.Context) {
	limit := 500
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	csv := h.finopsSvc().ExportAuditCSV(tenantID(c), limit)
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=finops-audit.csv")
	c.String(http.StatusOK, string(csv))
}

// ExportFinOpsLakeLineItems returns Iceberg/lake export metadata (FIN-PROD-08 preview).
func (h *Handler) ExportFinOpsLakeLineItems(c *gin.Context) {
	from, to := parseFinOpsTimeRange(c)
	format := c.DefaultQuery("format", "parquet")
	writeSuccess(c, gin.H{
		"status":      "queued",
		"format":      format,
		"from":        from.Format(time.RFC3339),
		"to":          to.Format(time.RFC3339),
		"destination": c.Query("destination"),
		"snapshotId":  "lake-" + time.Now().UTC().Format("20060102150405"),
		"message":     "Lake export job queued; wire object storage in FINOPS_LAKE_BUCKET for prod",
	})
}

func finopsActor(c *gin.Context) string {
	if p, ok := auth.PrincipalFromGin(c); ok && p.UserID != "" {
		return p.UserID
	}
	return "system"
}

func (h *Handler) ImportFinOpsCost(c *gin.Context) {
	var body domain.ImportRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "invalid body")
		return
	}
	batch, err := h.finopsSvc().ImportCustomCost(c.Request.Context(), tenantID(c), body, finopsActor(c))
	if err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	writeSuccess(c, batch)
}

func (h *Handler) ListFinOpsTagSuggestions(c *gin.Context) {
	list, err := h.finopsSvc().ListTagSuggestions(c.Request.Context(), tenantID(c))
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(c, list)
}

func (h *Handler) ListFinOpsSharedSplits(c *gin.Context) {
	list, err := h.finopsSvc().ListSharedSplits(c.Request.Context(), tenantID(c))
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(c, list)
}

func (h *Handler) ListFinOpsCommitments(c *gin.Context) {
	writeSuccess(c, h.finopsSvc().ListCommitments(c.Query("provider"), c.Query("status")))
}

func (h *Handler) ListFinOpsCommitmentRecommendations(c *gin.Context) {
	writeSuccess(c, h.finopsSvc().ListCommitmentRecommendations())
}

func (h *Handler) GetFinOpsUnitEconomics(c *gin.Context) {
	scope := c.DefaultQuery("scope", "all")
	if !h.finopsScopeAllowed(c, scope) {
		writeError(c, http.StatusForbidden, "cost scope not permitted")
		return
	}
	ue, err := h.finopsSvc().UnitEconomics(c.Request.Context(), tenantID(c), c.Query("metric"), scope)
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(c, ue)
}

func (h *Handler) ListFinOpsCarbonRecommendations(c *gin.Context) {
	writeSuccess(c, h.finopsSvc().ListCarbonRecommendations())
}

func (h *Handler) ListFinOpsReports(c *gin.Context) {
	writeSuccess(c, h.finopsSvc().ListReports(c.DefaultQuery("scope", "all")))
}

func (h *Handler) CreateFinOpsReportSchedule(c *gin.Context) {
	var body domain.ReportSchedule
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "invalid body")
		return
	}
	created := h.finopsSvc().CreateReportSchedule(c.Request.Context(), tenantID(c), body, finopsActor(c))
	writeSuccess(c, created)
}

func (h *Handler) ListFinOpsAuditLog(c *gin.Context) {
	writeSuccess(c, h.finopsSvc().ListAuditLog(tenantID(c), 50))
}

func (h *Handler) ListFinOpsChargebackStatements(c *gin.Context) {
	list, err := h.finopsSvc().ListChargebackStatements(c.Request.Context(), tenantID(c), c.Query("costCenter"))
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(c, list)
}

func (h *Handler) GenerateFinOpsChargebackStatement(c *gin.Context) {
	var body struct {
		CostCenter string `json:"costCenter"`
		Mode       string `json:"mode"`
		Period     string `json:"period"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		body.CostCenter = c.Query("costCenter")
		body.Mode = c.Query("mode")
		body.Period = c.Query("period")
	}
	if body.CostCenter == "" {
		body.CostCenter = "all"
	}
	st, err := h.finopsSvc().GenerateChargebackStatement(c.Request.Context(), tenantID(c), body.CostCenter, body.Mode, body.Period)
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(c, st)
}

func (h *Handler) ExportFinOpsChargebackStatement(c *gin.Context) {
	csv, st, err := h.finopsSvc().ExportChargebackStatement(c.Request.Context(), tenantID(c), c.Param("id"))
	if err != nil {
		writeError(c, http.StatusNotFound, err.Error())
		return
	}
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=chargeback-"+st.CostCenter+".csv")
	c.String(http.StatusOK, csv)
}

func (h *Handler) RunFinOpsScenario(c *gin.Context) {
	var body domain.ScenarioRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "invalid body")
		return
	}
	if body.ScenarioType == "" {
		writeError(c, http.StatusBadRequest, "scenarioType required")
		return
	}
	result, err := h.finopsSvc().RunScenario(c.Request.Context(), tenantID(c), body, finopsActor(c))
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(c, result)
}

func (h *Handler) ListFinOpsScenarios(c *gin.Context) {
	list, err := h.finopsSvc().ListScenarios(c.Request.Context(), tenantID(c), 20)
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(c, list)
}

func (h *Handler) ListFinOpsCommitmentAlerts(c *gin.Context) {
	alerts := h.finopsSvc().RouteCommitmentAlerts(c.Request.Context(), tenantID(c))
	if provider := c.Query("provider"); provider != "" {
		filtered := make([]domain.CommitmentAlert, 0)
		for _, a := range alerts {
			if a.Provider == provider {
				filtered = append(filtered, a)
			}
		}
		writeSuccess(c, filtered)
		return
	}
	writeSuccess(c, alerts)
}

func (h *Handler) FinOpsCarbonAction(c *gin.Context) {
	var body struct {
		Action string `json:"action"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		body.Action = "simulate"
	}
	result, err := h.finopsSvc().ApplyCarbonAction(c.Request.Context(), tenantID(c), domain.CarbonActionRequest{
		RecommendationID: c.Param("id"), Action: body.Action,
	}, finopsActor(c))
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(c, result)
}

func (h *Handler) ListFinOpsGovernancePolicies(c *gin.Context) {
	writeSuccess(c, h.finopsSvc().ListGovernancePolicies(tenantID(c)))
}

func (h *Handler) UpsertFinOpsGovernancePolicy(c *gin.Context) {
	var body domain.GovernancePolicy
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "invalid body")
		return
	}
	body.ID = c.Param("id")
	writeSuccess(c, h.finopsSvc().UpsertGovernancePolicy(c.Request.Context(), tenantID(c), body, finopsActor(c)))
}

func observeFinOpsQuery(operation string, start time.Time) {
	finops.ObserveQuery(operation, time.Since(start).Seconds())
}
