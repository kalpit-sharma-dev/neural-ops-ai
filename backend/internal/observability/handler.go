package observability

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/neuralops/platform/internal/gateway/auth"
)

// Handler exposes observability UI APIs on the gateway.
type Handler struct {
	deps Deps
}

// NewHandler creates an observability handler.
func NewHandler(deps Deps) *Handler {
	if deps.Mem == nil {
		deps.Mem = NewStore()
	}
	return &Handler{deps: deps}
}

// RegisterRoutes mounts observability routes under /api/v1.
func (h *Handler) RegisterRoutes(v1 *gin.RouterGroup) {
	// Phase 2 — APM / tracing (uses /apm prefix to avoid ingestion proxy conflict)
	apm := v1.Group("/apm")
	{
		apm.GET("/traces/:traceId", h.GetTrace)
		apm.POST("/traces/search", h.SearchTraces)
		apm.GET("/flow", h.ServiceFlow)
		apm.GET("/services/:service/operations", h.ListOperations)
	}

	// Phase 3 — Metrics & dashboards
	v1.GET("/metrics/catalog", h.MetricCatalog)
	v1.GET("/metrics/query", h.QueryMetric)
	v1.GET("/dashboards", h.ListDashboards)
	v1.POST("/dashboards", h.CreateDashboard)
	v1.GET("/dashboards/:id", h.GetDashboard)
	v1.PUT("/dashboards/:id", h.UpdateDashboard)
	v1.DELETE("/dashboards/:id", h.DeleteDashboard)

	// Phase 4 — Topology
	v1.GET("/topology", h.Topology)
	v1.GET("/zones", h.ListZones)

	// Phase 6 — Log settings
	v1.GET("/logs/metric-rules", h.ListLogMetricRules)
	v1.POST("/logs/metric-rules", h.CreateLogMetricRule)
	v1.GET("/logs/parsing-rules", h.ListLogParsingRules)
	v1.POST("/logs/parsing-rules", h.CreateLogParsingRule)

	// Phase 7 — SLOs & anomalies
	v1.GET("/slos", h.ListSLOs)
	v1.POST("/slos", h.CreateSLO)
	v1.GET("/anomalies/entities", h.ListAnomalies)

	// Phase 8 — Infra & K8s
	infra := v1.Group("/infra")
	{
		infra.GET("/hosts", h.ListHosts)
		infra.GET("/k8s/clusters", h.ListK8sClusters)
		infra.GET("/k8s/pods", h.ListK8sPods)
	}

	// Phase 9 — DB & middleware
	v1.GET("/databases", h.ListDatabases)
	v1.GET("/databases/:id/statements", h.DBStatements)
	v1.GET("/middleware/kafka/lag", h.KafkaLag)

	// Phase 10 — RUM & synthetic
	v1.GET("/rum/sessions", h.ListRUMSessions)
	v1.GET("/rum/sessions/:sessionId/replay", h.GetSessionReplay)
	v1.POST("/rum/beacon", h.IngestRUMBeacon)
	v1.POST("/rum/replay", h.IngestRUMReplay)
	v1.GET("/synthetic/monitors", h.ListSyntheticMonitors)
	v1.POST("/synthetic/monitors", h.CreateSyntheticMonitor)
	v1.GET("/synthetic/monitors/:id/runs", h.SyntheticRuns)

	// PromQL proxy (fallback when ClickHouse empty)
	v1.GET("/metrics/promql", h.QueryPromQL)

	// Phase 11 — Workflows & notebooks
	v1.GET("/workflows", h.ListWorkflows)
	v1.POST("/workflows", h.CreateWorkflow)
	v1.GET("/notebooks", h.ListNotebooks)
	v1.POST("/notebooks", h.CreateNotebook)

	// Phase 12 — Security & integrations
	v1.GET("/security/vulnerabilities", h.ListVulnerabilities)
	v1.GET("/security/attacks", h.ListAttacks)
	v1.GET("/security/attacks/:id", h.GetAttack)
	v1.GET("/integrations", h.ListIntegrations)
	v1.POST("/integrations/:id/connect", h.ConnectIntegration)

	// Phase 5 — Admin (demo data; production uses IdentityStore extensions)
	admin := v1.Group("/admin")
	{
		admin.GET("/users", h.ListAdminUsers)
		admin.GET("/api-keys", h.ListAPIKeys)
		admin.POST("/api-keys", h.CreateAPIKey)
		admin.GET("/audit", h.ListAudit)
		admin.GET("/usage", h.Usage)
	}

	h.RegisterDepthRoutes(v1)
	h.RegisterExtendedRoutes(v1)
}

func (h *Handler) GetTrace(c *gin.Context) {
	traceID := c.Param("traceId")
	tid := tenantID(c)
	if h.deps.Spans != nil {
		if trace, err := h.deps.Spans.GetTrace(c.Request.Context(), tid, traceID); err == nil {
			writeSuccess(c, trace)
			return
		}
	}
	trace, ok := h.deps.Mem.GetTrace(traceID)
	if !ok {
		writeError(c, http.StatusNotFound, "trace not found")
		return
	}
	writeSuccess(c, trace)
}

func (h *Handler) SearchTraces(c *gin.Context) {
	var req TraceSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	tid := tenantID(c)
	if h.deps.Spans != nil {
		if results, err := h.deps.Spans.SearchTraces(c.Request.Context(), tid, req); err == nil && len(results) > 0 {
			writeSuccess(c, results)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.SearchTraces(req))
}

func (h *Handler) ServiceFlow(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.Spans != nil {
		if flow, err := h.deps.Spans.ServiceFlow(c.Request.Context(), tid); err == nil && len(flow) > 0 {
			writeSuccess(c, flow)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.ServiceFlow())
}

func (h *Handler) ListOperations(c *gin.Context) {
	service := c.Param("service")
	tid := tenantID(c)
	if h.deps.Spans != nil {
		if ops, err := h.deps.Spans.ListOperations(c.Request.Context(), tid, service); err == nil && len(ops) > 0 {
			writeSuccess(c, ops)
			return
		}
	}
	writeSuccess(c, []string{})
}

func (h *Handler) MetricCatalog(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.MetricCatalog())
}

func (h *Handler) QueryMetric(c *gin.Context) {
	name := c.Query("name")
	service := c.DefaultQuery("service", "payment-service")
	start, _ := time.Parse(time.RFC3339, c.Query("start"))
	end, _ := time.Parse(time.RFC3339, c.Query("end"))
	tid := tenantID(c)
	if h.deps.CH != nil && name != "" {
		if series, err := QueryMetrics(c.Request.Context(), h.deps.CH, tid, name, service, start, end); err == nil {
			writeSuccess(c, series)
			return
		}
	}
	if h.deps.Prom != nil && name != "" {
		query := MetricToPromQL(name, service)
		if points, err := h.deps.Prom.QueryRange(c.Request.Context(), query, start, end, time.Minute); err == nil {
			writeSuccess(c, MetricSeries{
				Name: name, Labels: map[string]string{"service": service, "source": "prometheus"}, Points: points,
			})
			return
		}
	}
	writeSuccess(c, h.deps.Mem.QueryMetric(name, service, start, end))
}

func (h *Handler) QueryPromQL(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		writeError(c, http.StatusBadRequest, "query parameter required")
		return
	}
	if h.deps.Prom == nil {
		writeError(c, http.StatusServiceUnavailable, "prometheus unavailable")
		return
	}
	start, _ := time.Parse(time.RFC3339, c.Query("start"))
	end, _ := time.Parse(time.RFC3339, c.Query("end"))
	points, err := h.deps.Prom.QueryRange(c.Request.Context(), query, start, end, time.Minute)
	if err != nil {
		writeError(c, http.StatusBadGateway, err.Error())
		return
	}
	writeSuccess(c, MetricSeries{Name: query, Labels: map[string]string{"source": "prometheus"}, Points: points})
}

func (h *Handler) ListDashboards(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.PG != nil {
		if list, err := h.deps.PG.ListDashboards(c.Request.Context(), tid); err == nil && len(list) > 0 {
			writeSuccess(c, list)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.ListDashboards(tid))
}

func (h *Handler) GetDashboard(c *gin.Context) {
	tid := tenantID(c)
	id := c.Param("id")
	if h.deps.PG != nil {
		if d, err := h.deps.PG.GetDashboard(c.Request.Context(), tid, id); err == nil {
			writeSuccess(c, d)
			return
		}
	}
	d, ok := h.deps.Mem.GetDashboard(id)
	if !ok {
		writeError(c, http.StatusNotFound, "dashboard not found")
		return
	}
	writeSuccess(c, d)
}

func (h *Handler) CreateDashboard(c *gin.Context) {
	var d Dashboard
	if err := c.ShouldBindJSON(&d); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	d.TenantID = tenantID(c)
	if h.deps.PG != nil {
		if saved, err := h.deps.PG.SaveDashboard(c.Request.Context(), d); err == nil {
			writeSuccess(c, saved)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.SaveDashboard(d))
}

func (h *Handler) UpdateDashboard(c *gin.Context) {
	var d Dashboard
	if err := c.ShouldBindJSON(&d); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	d.ID = c.Param("id")
	d.TenantID = tenantID(c)
	if h.deps.PG != nil {
		if saved, err := h.deps.PG.SaveDashboard(c.Request.Context(), d); err == nil {
			writeSuccess(c, saved)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.SaveDashboard(d))
}

func (h *Handler) DeleteDashboard(c *gin.Context) {
	tid := tenantID(c)
	id := c.Param("id")
	if h.deps.PG != nil {
		if err := h.deps.PG.DeleteDashboard(c.Request.Context(), tid, id); err == nil {
			writeSuccess(c, gin.H{"deleted": true})
			return
		}
	}
	if !h.deps.Mem.DeleteDashboard(id) {
		writeError(c, http.StatusNotFound, "dashboard not found")
		return
	}
	writeSuccess(c, gin.H{"deleted": true})
}

func (h *Handler) Topology(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.Pool != nil {
		if graph, err := LoadTopologyFromPostgres(c.Request.Context(), h.deps.Pool, tid); err == nil {
			writeSuccess(c, graph)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.Topology(c.Query("zone")))
}

func (h *Handler) ListZones(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListZones())
}

func (h *Handler) ListLogMetricRules(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.PG != nil {
		if list, err := h.deps.PG.ListLogMetricRules(c.Request.Context(), tid); err == nil && len(list) > 0 {
			writeSuccess(c, list)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.ListLogMetricRules(tid))
}

func (h *Handler) CreateLogMetricRule(c *gin.Context) {
	var r LogMetricRule
	if err := c.ShouldBindJSON(&r); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	r.TenantID = tenantID(c)
	if h.deps.PG != nil {
		if saved, err := h.deps.PG.SaveLogMetricRule(c.Request.Context(), r); err == nil {
			writeSuccess(c, saved)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.SaveLogMetricRule(r))
}

func (h *Handler) ListLogParsingRules(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.PG != nil {
		if list, err := h.deps.PG.ListLogParsingRules(c.Request.Context(), tid); err == nil && len(list) > 0 {
			writeSuccess(c, list)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.ListLogParsingRules(tid))
}

func (h *Handler) CreateLogParsingRule(c *gin.Context) {
	var r LogParsingRule
	if err := c.ShouldBindJSON(&r); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	r.TenantID = tenantID(c)
	if h.deps.PG != nil {
		if saved, err := h.deps.PG.SaveLogParsingRule(c.Request.Context(), r); err == nil {
			writeSuccess(c, saved)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.SaveLogParsingRule(r))
}

func (h *Handler) ListSLOs(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.PG != nil {
		if list, err := h.deps.PG.ListSLOs(c.Request.Context(), tid); err == nil && len(list) > 0 {
			writeSuccess(c, list)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.ListSLOs())
}

func (h *Handler) CreateSLO(c *gin.Context) {
	var slo SLO
	if err := c.ShouldBindJSON(&slo); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	slo.TenantID = tenantID(c)
	if h.deps.PG != nil {
		if saved, err := h.deps.PG.SaveSLO(c.Request.Context(), slo); err == nil {
			writeSuccess(c, saved)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.SaveSLO(slo))
}

func (h *Handler) ListAnomalies(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListAnomalies())
}

func (h *Handler) ListHosts(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.Collectors != nil {
		if list, err := h.deps.Collectors.ListHosts(c.Request.Context(), tid); err == nil && len(list) > 0 {
			writeSuccess(c, list)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.ListHosts())
}

func (h *Handler) ListK8sClusters(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.Collectors != nil {
		if list, err := h.deps.Collectors.ListK8sClusters(c.Request.Context(), tid); err == nil && len(list) > 0 {
			writeSuccess(c, list)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.ListK8sClusters())
}

func (h *Handler) ListK8sPods(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.Collectors != nil {
		if list, err := h.deps.Collectors.ListK8sPods(c.Request.Context(), tid, c.Query("namespace")); err == nil && len(list) > 0 {
			writeSuccess(c, list)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.ListK8sPods(c.Query("namespace")))
}

func (h *Handler) ListDatabases(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListDatabases())
}

func (h *Handler) DBStatements(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.DBStatements(c.Param("id")))
}

func (h *Handler) KafkaLag(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.KafkaLag())
}

func (h *Handler) ListRUMSessions(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.Collectors != nil {
		if list, err := h.deps.Collectors.ListRUMSessions(c.Request.Context(), tid, 100); err == nil && len(list) > 0 {
			writeSuccess(c, list)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.ListRUMSessions())
}

type rumBeaconRequest struct {
	SessionID  string  `json:"sessionId"`
	UserID     string  `json:"userId"`
	Page       string  `json:"page"`
	Device     string  `json:"device"`
	Country    string  `json:"country"`
	DurationMs int64   `json:"durationMs"`
	Errors     int     `json:"errors"`
	LCP        float64 `json:"lcp"`
}

func (h *Handler) IngestRUMBeacon(c *gin.Context) {
	if h.deps.Collectors == nil {
		writeError(c, http.StatusServiceUnavailable, "rum storage unavailable")
		return
	}
	var req rumBeaconRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.SessionID == "" {
		writeError(c, http.StatusBadRequest, "invalid beacon payload")
		return
	}
	session := RUMSession{
		ID: req.SessionID, UserID: req.UserID, Page: req.Page, Device: req.Device,
		Country: req.Country, DurationMs: req.DurationMs, Errors: req.Errors, LCP: req.LCP,
		StartedAt: time.Now().UTC(),
	}
	if err := h.deps.Collectors.UpsertRUMSession(c.Request.Context(), tenantID(c), session); err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(c, gin.H{"accepted": true})
}

type rumReplayRequest struct {
	SessionID string         `json:"sessionId"`
	Seq       int            `json:"seq"`
	Type      string         `json:"type"`
	Payload   map[string]any `json:"payload"`
}

func (h *Handler) IngestRUMReplay(c *gin.Context) {
	if h.deps.Collectors == nil {
		writeError(c, http.StatusServiceUnavailable, "replay storage unavailable")
		return
	}
	var req rumReplayRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.SessionID == "" {
		writeError(c, http.StatusBadRequest, "invalid replay payload")
		return
	}
	if err := h.deps.Collectors.AppendReplayEvent(c.Request.Context(), tenantID(c), req.SessionID, req.Seq, req.Type, req.Payload); err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(c, gin.H{"accepted": true})
}

func (h *Handler) GetSessionReplay(c *gin.Context) {
	if h.deps.Collectors == nil {
		writeError(c, http.StatusServiceUnavailable, "replay storage unavailable")
		return
	}
	events, err := h.deps.Collectors.ListReplayEvents(c.Request.Context(), tenantID(c), c.Param("sessionId"))
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(c, events)
}

func (h *Handler) ListSyntheticMonitors(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.Collectors != nil {
		if list, err := h.deps.Collectors.ListSyntheticMonitors(c.Request.Context(), tid); err == nil && len(list) > 0 {
			writeSuccess(c, list)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.ListSyntheticMonitors())
}

func (h *Handler) CreateSyntheticMonitor(c *gin.Context) {
	var m SyntheticMonitor
	if err := c.ShouldBindJSON(&m); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	tid := tenantID(c)
	if h.deps.Collectors != nil {
		if saved, err := h.deps.Collectors.SaveSyntheticMonitor(c.Request.Context(), tid, m); err == nil {
			writeSuccess(c, saved)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.SaveSyntheticMonitor(m))
}

func (h *Handler) SyntheticRuns(c *gin.Context) {
	tid := tenantID(c)
	id := c.Param("id")
	if h.deps.Collectors != nil {
		if runs, err := h.deps.Collectors.ListSyntheticRuns(c.Request.Context(), tid, id); err == nil && len(runs) > 0 {
			writeSuccess(c, runs)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.SyntheticRuns(id))
}

func (h *Handler) ListWorkflows(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.PG != nil {
		if list, err := h.deps.PG.ListWorkflows(c.Request.Context(), tid); err == nil && len(list) > 0 {
			writeSuccess(c, list)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.ListWorkflows())
}

func (h *Handler) ListNotebooks(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.PG != nil {
		if list, err := h.deps.PG.ListNotebooks(c.Request.Context(), tid); err == nil && len(list) > 0 {
			writeSuccess(c, list)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.ListNotebooks())
}

func (h *Handler) CreateWorkflow(c *gin.Context) {
	var w Workflow
	if err := c.ShouldBindJSON(&w); err != nil || w.Name == "" {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if h.deps.PG != nil {
		if saved, err := h.deps.PG.SaveWorkflow(c.Request.Context(), tenantID(c), w); err == nil {
			writeSuccess(c, saved)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.SaveWorkflow(w))
}

func (h *Handler) CreateNotebook(c *gin.Context) {
	var n Notebook
	if err := c.ShouldBindJSON(&n); err != nil || n.Name == "" {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if h.deps.PG != nil {
		if saved, err := h.deps.PG.SaveNotebook(c.Request.Context(), tenantID(c), n); err == nil {
			writeSuccess(c, saved)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.SaveNotebook(n))
}

func (h *Handler) ListVulnerabilities(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListVulnerabilities())
}

func (h *Handler) ListAttacks(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListAttacks())
}

func (h *Handler) GetAttack(c *gin.Context) {
	id := c.Param("id")
	if a, ok := h.deps.Mem.GetAttack(id); ok {
		writeSuccess(c, a)
		return
	}
	writeError(c, http.StatusNotFound, "attack not found")
}

func (h *Handler) ListIntegrations(c *gin.Context) {
	tid := tenantID(c)
	repo := NewIntegrationsRepo(h.deps.Pool)
	if repo.available() {
		if list, err := repo.List(c.Request.Context(), tid); err == nil && len(list) > 0 {
			writeSuccess(c, list)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.ListIntegrations())
}

func (h *Handler) ConnectIntegration(c *gin.Context) {
	key := c.Param("id")
	var req ConnectRequest
	_ = c.ShouldBindJSON(&req)
	tid := tenantID(c)
	repo := NewIntegrationsRepo(h.deps.Pool)
	if repo.available() && len(req.Config) > 0 {
		rec, err := repo.Connect(c.Request.Context(), tid, key, req.Config)
		if err == nil {
			writeSuccess(c, rec)
			return
		}
	}
	if i, ok := h.deps.Mem.ConnectIntegration(key); ok {
		writeSuccess(c, i)
		return
	}
	writeError(c, http.StatusNotFound, "integration not found")
}

func (h *Handler) ListAdminUsers(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.Identity != nil {
		if users, err := auth.ListUsersByTenant(c.Request.Context(), h.deps.Identity, tid); err == nil && len(users) > 0 {
			writeSuccess(c, users)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.DemoAdminUsers())
}

func (h *Handler) ListAPIKeys(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.Identity != nil {
		if keys, err := auth.ListAPIKeysByTenant(c.Request.Context(), h.deps.Identity, tid); err == nil && len(keys) > 0 {
			writeSuccess(c, keys)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.DemoAPIKeys())
}

type createAPIKeyRequest struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

func (h *Handler) CreateAPIKey(c *gin.Context) {
	if h.deps.Identity == nil {
		writeError(c, http.StatusServiceUnavailable, "identity store unavailable")
		return
	}
	var req createAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Role == "" {
		req.Role = "developer"
	}
	key, raw, err := auth.CreateAPIKey(c.Request.Context(), h.deps.Identity, tenantID(c), req.Name, req.Role)
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(c, gin.H{"key": key, "secret": raw})
}

func (h *Handler) ListAudit(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.Pool != nil {
		if entries, err := ListAuditEntries(c.Request.Context(), h.deps.Pool, tid, 100); err == nil && len(entries) > 0 {
			writeSuccess(c, entries)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.DemoAuditEntries())
}

func (h *Handler) Usage(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.Usage())
}

func tenantID(c *gin.Context) string {
	if v, ok := c.Get("tenant_id"); ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	if h := c.GetHeader("X-Tenant-ID"); h != "" {
		return h
	}
	return "default"
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
		"errorCode": "OBS001",
		"message":   message,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
