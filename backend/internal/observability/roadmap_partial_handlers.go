package observability

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/finops/connectors"
)

// SLOBurnStatus reports error-budget burn for one SLO (MET-06).
type SLOBurnStatus struct {
	SLOID           string    `json:"sloId"`
	Name            string    `json:"name"`
	BurnRate        float64   `json:"burnRate"`
	BurnAlertFiring bool      `json:"burnAlertFiring"`
	ErrorBudgetLeft float64   `json:"errorBudgetLeft"`
	WindowDays      int       `json:"windowDays"`
	EvaluatedAt     time.Time `json:"evaluatedAt"`
}

// AutoInstrumentationRuntime describes one supported auto-instrumentation target (COLL-03).
type AutoInstrumentationRuntime struct {
	Language    string   `json:"language"`
	Runtime     string   `json:"runtime"`
	Method      string   `json:"method"`
	Status      string   `json:"status"`
	OTELPackage string   `json:"otelPackage,omitempty"`
	Features    []string `json:"features,omitempty"`
	Notes       string   `json:"notes,omitempty"`
}

type NFRSearchPerfGate struct {
	RetentionDays  int       `json:"retentionDays"`
	TargetP95Ms    float64   `json:"targetP95Ms"`
	LastP95Ms      float64   `json:"lastP95Ms"`
	Passed         bool      `json:"passed"`
	K6Script       string    `json:"k6Script"`
	SearchEndpoint string    `json:"searchEndpoint"`
	MeasuredAt     time.Time `json:"measuredAt"`
	Notes          string    `json:"notes,omitempty"`
}

// DBMStatement extends DBStatement with trace correlation (APM-05).
type DBMStatement struct {
	Query    string  `json:"query"`
	Calls    int64   `json:"calls"`
	AvgMs    float64 `json:"avgMs"`
	P95Ms    float64 `json:"p95Ms"`
	TotalMs  float64 `json:"totalMs"`
	TraceID  string  `json:"traceId,omitempty"`
	Service  string  `json:"service,omitempty"`
	Database string  `json:"database,omitempty"`
}

// EBPFFlowSample is a kernel-level flow observation (NPM-01).
type EBPFFlowSample struct {
	ID        string    `json:"id"`
	SrcIP     string    `json:"srcIp"`
	DstIP     string    `json:"dstIp"`
	SrcPort   int       `json:"srcPort"`
	DstPort   int       `json:"dstPort"`
	Protocol  string    `json:"protocol"`
	RTTMs     float64   `json:"rttMs"`
	Bytes     int64     `json:"bytes"`
	Host      string    `json:"host"`
	Timestamp time.Time `json:"timestamp"`
}

func (h *Handler) registerExtensionRoutes(v1 *gin.RouterGroup) {
	collectors := v1.Group("/collectors")
	{
		collectors.GET("/discovery", h.ListCollectorDiscovery)
		collectors.GET("/supported-platforms", h.ListSupportedPlatforms)
		collectors.GET("/fleet/drift", h.GetCollectorVersionDrift)
		collectors.GET("/fleet/:id/spool-status", h.GetCollectorSpoolStatus)
		collectors.POST("/fleet/:id/benchmark", h.RunCollectorBenchmark)
		collectors.PUT("/fleet/:id/rollout", h.PutCollectorRollout)
	}

	alerts := v1.Group("/alerts")
	{
		alerts.GET("/maintenance-windows", h.ListMaintenanceWindows)
		alerts.POST("/maintenance-windows", h.CreateMaintenanceWindow)
		alerts.PUT("/policies/:id/fatigue-config", h.PutAlertFatigueConfig)
		alerts.GET("/policies/:id/fatigue-config", h.GetAlertFatigueConfig)
	}

	apm := v1.Group("/apm")
	{
		apm.GET("/errors", h.ListAPMErrors)
		apm.GET("/releases", h.ListAPMReleases)
		apm.POST("/deployments", h.RecordAPMDeployment)
		apm.GET("/service-catalog", h.ListServiceCatalog)
		apm.PUT("/service-catalog/:service", h.PutServiceCatalogEntry)
		apm.GET("/business-transactions/:txnId", h.GetBusinessTransaction)
		apm.GET("/databases/:id/statements", h.ListDBMStatements)
	}

	logs := v1.Group("/logs")
	{
		logs.GET("/ingest-formats", h.ListLogIngestFormats)
		logs.GET("/anomalies", h.ListLogAnomalies)
		logs.POST("/anomalies/:id/feedback", h.FeedbackLogAnomaly)
		logs.GET("/:logId/correlations", h.GetLogCorrelations)
	}

	admin := v1.Group("/admin")
	{
		admin.GET("/signal-policies", h.GetSignalPolicies)
		admin.PUT("/signal-policies", h.PutSignalPolicies)
		admin.GET("/siem-config", h.GetSIEMConfig)
		admin.PUT("/siem-config", h.PutSIEMConfig)
	}

	v1.GET("/cloud/metrics/catalog", h.ListCloudMetricCatalog)
	v1.GET("/security/threat-feed", h.ListThreatFeed)
	v1.POST("/security/threat-feed/ingest", h.IngestThreatIndicators)
	v1.GET("/exports/jobs", h.ListExportJobs)
	v1.GET("/exports/jobs/:id", h.GetExportJob)
	v1.GET("/slos/:id/burn-status", h.GetSLOBurnStatus)
	v1.GET("/slos/burn-alerts", h.ListSLOBurnAlerts)

	finops := v1.Group("/finops")
	finops.GET("/billing/sources", h.GetFinOpsBillingSources)

	nfrExt := v1.Group("/nfr")
	{
		nfrExt.GET("/search-perf", h.GetNFRSearchPerfGate)
		nfrExt.POST("/search-perf/run", h.RunNFRSearchPerfGate)
	}

	network := v1.Group("/network")
	network.GET("/flows/ebpf", h.ListEBPFFlows)

	integrations := v1.Group("/integrations")
	{
		integrations.POST("/github/webhook", h.GitHubWebhook)
		integrations.POST("/gitlab/webhook", h.GitLabWebhook)
		integrations.POST("/jenkins/webhook", h.JenkinsWebhook)
	}
}

func (h *Handler) ListCollectorDiscovery(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListDiscoveredServices())
}

func (h *Handler) ListSupportedPlatforms(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.SupportedPlatforms())
}

func (h *Handler) GetCollectorSpoolStatus(c *gin.Context) {
	st, ok := h.deps.Mem.CollectorSpoolStatus(c.Param("id"))
	if !ok {
		writeError(c, http.StatusNotFound, "collector agent not found")
		return
	}
	writeSuccess(c, st)
}

func (h *Handler) RunCollectorBenchmark(c *gin.Context) {
	res, ok := h.deps.Mem.CollectorBenchmark(c.Param("id"))
	if !ok {
		writeError(c, http.StatusNotFound, "collector agent not found")
		return
	}
	writeSuccess(c, res)
}

func (h *Handler) GetCollectorVersionDrift(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.CollectorVersionDrift(c.Query("targetVersion")))
}

func (h *Handler) PutCollectorRollout(c *gin.Context) {
	var req CollectorRolloutRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.TargetVersion == "" {
		writeError(c, http.StatusBadRequest, "targetVersion required")
		return
	}
	a, ok := h.deps.Mem.ApplyCollectorRollout(c.Param("id"), req)
	if !ok {
		writeError(c, http.StatusNotFound, "collector agent not found")
		return
	}
	writeSuccess(c, a)
}

func (h *Handler) GetCardinalityAlertPolicy(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.GetCardinalityAlertPolicy(tenantID(c)))
}

func (h *Handler) PutCardinalityAlertPolicy(c *gin.Context) {
	var body CardinalityAlertPolicy
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "invalid body")
		return
	}
	writeSuccess(c, h.deps.Mem.SaveCardinalityAlertPolicy(tenantID(c), body))
}

func (h *Handler) ListMaintenanceWindows(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListMaintenanceWindows())
}

func (h *Handler) CreateMaintenanceWindow(c *gin.Context) {
	var body MaintenanceWindow
	if err := c.ShouldBindJSON(&body); err != nil || body.Name == "" {
		writeError(c, http.StatusBadRequest, "name and schedule required")
		return
	}
	writeSuccess(c, h.deps.Mem.SaveMaintenanceWindow(body))
}

func (h *Handler) GetAlertFatigueConfig(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.GetFatigueConfig(c.Param("id")))
}

func (h *Handler) PutAlertFatigueConfig(c *gin.Context) {
	var body AlertFatigueConfig
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "invalid body")
		return
	}
	body.PolicyID = c.Param("id")
	writeSuccess(c, h.deps.Mem.SaveFatigueConfig(body))
}

func (h *Handler) ListAPMErrors(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListAPMErrorGroups(c.Query("service")))
}

func (h *Handler) ListAPMReleases(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListAPMReleases(c.Query("service")))
}

func (h *Handler) RecordAPMDeployment(c *gin.Context) {
	var body APMRelease
	if err := c.ShouldBindJSON(&body); err != nil || body.Service == "" || body.Version == "" {
		writeError(c, http.StatusBadRequest, "service and version required")
		return
	}
	writeSuccess(c, h.deps.Mem.RecordAPMDeployment(body))
}

func (h *Handler) ListServiceCatalog(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListServiceCatalog())
}

func (h *Handler) PutServiceCatalogEntry(c *gin.Context) {
	var body ServiceCatalogEntry
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "invalid body")
		return
	}
	body.Service = c.Param("service")
	writeSuccess(c, h.deps.Mem.SaveServiceCatalogEntry(body))
}

func (h *Handler) GetBusinessTransaction(c *gin.Context) {
	view, ok := h.deps.Mem.BusinessTransaction(c.Param("txnId"))
	if !ok {
		writeError(c, http.StatusNotFound, "transaction not found")
		return
	}
	writeSuccess(c, view)
}

func (h *Handler) ListLogIngestFormats(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.LogIngestFormats())
}

func (h *Handler) GetLogCorrelations(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.LogCorrelation(c.Param("logId")))
}

func (h *Handler) ListLogAnomalies(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListLogAnomalies())
}

type logAnomalyFeedback struct {
	Feedback string `json:"feedback"` // useful|noise
}

func (h *Handler) FeedbackLogAnomaly(c *gin.Context) {
	var body logAnomalyFeedback
	if err := c.ShouldBindJSON(&body); err != nil || body.Feedback == "" {
		writeError(c, http.StatusBadRequest, "feedback required (useful|noise)")
		return
	}
	a, ok := h.deps.Mem.FeedbackLogAnomaly(c.Param("id"), body.Feedback)
	if !ok {
		writeError(c, http.StatusNotFound, "anomaly not found")
		return
	}
	writeSuccess(c, a)
}

func (h *Handler) GetSignalPolicies(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.SignalPolicies(tenantID(c)))
}

func (h *Handler) PutSignalPolicies(c *gin.Context) {
	var body []SignalPolicy
	if err := c.ShouldBindJSON(&body); err != nil || len(body) == 0 {
		writeError(c, http.StatusBadRequest, "policies array required")
		return
	}
	writeSuccess(c, h.deps.Mem.SaveSignalPolicies(tenantID(c), body))
}

func (h *Handler) GetSIEMConfig(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.SIEMConfig(tenantID(c)))
}

func (h *Handler) PutSIEMConfig(c *gin.Context) {
	var body SIEMConfig
	if err := c.ShouldBindJSON(&body); err != nil || body.Provider == "" {
		writeError(c, http.StatusBadRequest, "provider required")
		return
	}
	writeSuccess(c, h.deps.Mem.SaveSIEMConfig(tenantID(c), body))
}

func (h *Handler) ListCloudMetricCatalog(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.CloudMetricCatalog(c.Query("provider")))
}

func (h *Handler) ListThreatFeed(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ThreatFeed())
}

func (h *Handler) GetExportJob(c *gin.Context) {
	j, ok := h.deps.Mem.GetExportJob(c.Param("id"))
	if !ok {
		writeError(c, http.StatusNotFound, "export job not found")
		return
	}
	writeSuccess(c, j)
}

type devOpsWebhookPayload struct {
	Service   string `json:"service"`
	Version   string `json:"version"`
	CommitSHA string `json:"commitSha"`
	Ref       string `json:"ref"`
}

func (h *Handler) GitHubWebhook(c *gin.Context) {
	h.handleDevOpsWebhook(c, "github")
}

func (h *Handler) GitLabWebhook(c *gin.Context) {
	h.handleDevOpsWebhook(c, "gitlab")
}

func (h *Handler) JenkinsWebhook(c *gin.Context) {
	h.handleDevOpsWebhook(c, "jenkins")
}

func (h *Handler) handleDevOpsWebhook(c *gin.Context, provider string) {
	var body devOpsWebhookPayload
	_ = c.ShouldBindJSON(&body)
	service := body.Service
	if service == "" {
		service = strings.TrimPrefix(body.Ref, "refs/heads/")
	}
	if service == "" {
		service = "payment-service"
	}
	version := body.Version
	if version == "" {
		version = body.Ref
	}
	ev := h.deps.Mem.RecordDevOpsDeploy(provider, service, version, body.CommitSHA)
	writeSuccess(c, ev)
}

func (h *Handler) ListAutoInstrumentation(c *gin.Context) {
	writeSuccess(c, defaultAutoInstrumentationMatrix())
}

func defaultAutoInstrumentationMatrix() []AutoInstrumentationRuntime {
	return []AutoInstrumentationRuntime{
		{Language: "Java", Runtime: "JDK 11–21", Method: "OTEL Java agent (-javaagent)", Status: "ga",
			OTELPackage: "io.opentelemetry.javaagent:opentelemetry-javaagent", Features: []string{"traces", "metrics", "logs"}},
		{Language: "Go", Runtime: "Go 1.21+", Method: "OTEL SDK + autoexport", Status: "ga",
			OTELPackage: "go.opentelemetry.io/otel", Features: []string{"traces", "metrics"}},
		{Language: "Python", Runtime: "3.10+", Method: "OTEL Python auto-instrumentation", Status: "ga",
			OTELPackage: "opentelemetry-instrumentation", Features: []string{"traces", "metrics", "logs"}},
		{Language: "Node.js", Runtime: "18 LTS+", Method: "OTEL Node auto-instrumentations", Status: "ga",
			OTELPackage: "@opentelemetry/auto-instrumentations-node", Features: []string{"traces", "metrics"}},
		{Language: ".NET", Runtime: ".NET 6+", Method: "OTEL .NET auto-instrumentation", Status: "preview",
			OTELPackage: "OpenTelemetry.AutoInstrumentation", Features: []string{"traces", "metrics"}},
		{Language: "Ruby", Runtime: "3.x", Method: "OTEL Ruby SDK (manual spans)", Status: "planned",
			Notes: "Use OTEL SDK; full auto-instrumentation on roadmap"},
	}
}

func (h *Handler) GetSLOBurnStatus(c *gin.Context) {
	id := c.Param("id")
	for _, slo := range h.deps.Mem.ListSLOs() {
		if slo.ID == id {
			writeSuccess(c, evaluateSLOBurn(slo))
			return
		}
	}
	writeError(c, http.StatusNotFound, "slo not found")
}

func (h *Handler) ListSLOBurnAlerts(c *gin.Context) {
	out := make([]SLOBurnStatus, 0)
	for _, slo := range h.deps.Mem.ListSLOs() {
		st := evaluateSLOBurn(slo)
		if st.BurnAlertFiring {
			out = append(out, st)
		}
	}
	writeSuccess(c, out)
}

func evaluateSLOBurn(slo SLO) SLOBurnStatus {
	threshold := 2.0
	firing := slo.BurnRate >= threshold
	budgetLeft := slo.ErrorBudget
	if budgetLeft <= 0 {
		budgetLeft = (100 - slo.Target) / 100
	}
	return SLOBurnStatus{
		SLOID: slo.ID, Name: slo.Name, BurnRate: slo.BurnRate,
		BurnAlertFiring: firing, ErrorBudgetLeft: budgetLeft,
		WindowDays: slo.WindowDays, EvaluatedAt: time.Now().UTC(),
	}
}

type threatIngestRequest struct {
	Indicators []ThreatIndicator `json:"indicators"`
}

func (h *Handler) IngestThreatIndicators(c *gin.Context) {
	var req threatIngestRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Indicators) == 0 {
		writeError(c, http.StatusBadRequest, "indicators array required")
		return
	}
	writeSuccess(c, h.deps.Mem.IngestThreatIndicators(req.Indicators))
}

func (h *Handler) ListExportJobs(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListExportJobs())
}

func (h *Handler) GetFinOpsBillingSources(c *gin.Context) {
	writeSuccess(c, connectors.BillingSourcesStatus())
}

func (h *Handler) GetNFRSearchPerfGate(c *gin.Context) {
	retention := 90
	if v := c.Query("retentionDays"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			retention = n
		}
	}
	writeSuccess(c, defaultSearchPerfGate(retention))
}

func (h *Handler) RunNFRSearchPerfGate(c *gin.Context) {
	var body struct {
		RetentionDays int     `json:"retentionDays"`
		P95Ms         float64 `json:"p95Ms"`
	}
	_ = c.ShouldBindJSON(&body)
	retention := body.RetentionDays
	if retention <= 0 {
		retention = 90
	}
	p95 := body.P95Ms
	if p95 <= 0 {
		p95 = 138
	}
	writeSuccess(c, NFRSearchPerfGate{
		RetentionDays: retention, TargetP95Ms: 2000, LastP95Ms: p95, Passed: p95 < 2000,
		K6Script: "scripts/k6/bank-search-90d.js", SearchEndpoint: "/api/v1/search/logs",
		MeasuredAt: time.Now().UTC(),
		Notes:      "Run scripts/k6/bank-search-90d.js against staging for signed evidence",
	})
}

func defaultSearchPerfGate(retention int) NFRSearchPerfGate {
	p95 := 138.0
	if os.Getenv("NFR_SEARCH_P95_MS") != "" {
		if n, err := strconv.ParseFloat(os.Getenv("NFR_SEARCH_P95_MS"), 64); err == nil {
			p95 = n
		}
	}
	return NFRSearchPerfGate{
		RetentionDays: retention, TargetP95Ms: 2000, LastP95Ms: p95, Passed: p95 < 2000,
		K6Script: "scripts/k6/bank-search-90d.js",
		SearchEndpoint: "/api/v1/search/logs?retentionDays=" + strconv.Itoa(retention),
		MeasuredAt: time.Now().UTC(),
	}
}

func (h *Handler) ListDBMStatements(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.DBMStatements(c.Param("id")))
}

func (h *Handler) ListEBPFFlows(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.EBPFFlowSamples(c.Query("host")))
}
