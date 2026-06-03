package observability

import (
	"fmt"
	"net/http"
	"strings"
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

func (h *Handler) UnifiedQuery(c *gin.Context) {
	var req UnifiedQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	recordUnifiedQuery(tenantID(c))
	writeSuccess(c, PlanUnifiedQuery(h.deps.Mem, req))
}

func (h *Handler) ValidateUnifiedQuery(c *gin.Context) {
	var req UnifiedQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	explain := BuildQueryExplain(req)
	writeSuccess(c, gin.H{
		"valid":   explain.Valid,
		"message": explain.Message,
		"explain": explain,
	})
}

func (h *Handler) ListUnifiedQueryFunctions(c *gin.Context) {
	writeSuccess(c, []string{
		"where(field, op, value)",
		"group_by(field)",
		"count()",
		"avg(field)",
		"rate(field, window)",
	})
}

func (h *Handler) ListCollectorFleet(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListCollectorFleet())
}

func (h *Handler) CreateCollectorAgent(c *gin.Context) {
	var a CollectorFleetAgent
	if err := c.ShouldBindJSON(&a); err != nil || a.Name == "" {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	writeSuccess(c, h.deps.Mem.SaveCollectorAgent(a))
}

func (h *Handler) UpdateCollectorAgent(c *gin.Context) {
	var a CollectorFleetAgent
	if err := c.ShouldBindJSON(&a); err != nil || a.Name == "" {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	updated, ok := h.deps.Mem.UpdateCollectorAgent(c.Param("id"), a)
	if !ok {
		writeError(c, http.StatusNotFound, "collector agent not found")
		return
	}
	writeSuccess(c, updated)
}

func (h *Handler) ListCollectorPipelines(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListCollectorPipelines())
}

func (h *Handler) CreateCollectorPipeline(c *gin.Context) {
	var p CollectorPipeline
	if err := c.ShouldBindJSON(&p); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	ok, msg := h.deps.Mem.ValidateCollectorPipeline(p)
	if !ok {
		writeError(c, http.StatusBadRequest, msg)
		return
	}
	writeSuccess(c, h.deps.Mem.SaveCollectorPipeline(p))
}

func (h *Handler) UpdateCollectorPipeline(c *gin.Context) {
	var p CollectorPipeline
	if err := c.ShouldBindJSON(&p); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	p.ID = c.Param("id")
	ok, msg := h.deps.Mem.ValidateCollectorPipeline(p)
	if !ok {
		writeError(c, http.StatusBadRequest, msg)
		return
	}
	writeSuccess(c, h.deps.Mem.SaveCollectorPipeline(p))
}

func (h *Handler) ValidateCollectorPipeline(c *gin.Context) {
	var p CollectorPipeline
	if err := c.ShouldBindJSON(&p); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	ok, msg := h.deps.Mem.ValidateCollectorPipeline(p)
	writeSuccess(c, gin.H{"valid": ok, "message": msg})
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
	tid := tenantID(c)
	if h.deps.Spans != nil {
		ctx, cancel := withAnalyticsTimeout(c.Request.Context())
		defer cancel()
		if list, err := h.deps.Spans.DetectAnomalies(ctx, tid); err == nil && len(list) > 0 {
			out := make([]EntityAnomaly, 0, len(list))
			for i, a := range list {
				out = append(out, EntityAnomaly{
					ID:         fmt.Sprintf("an-%d", i+1),
					EntityID:   a.Service,
					EntityType: "service",
					Service:    a.Service,
					Metric:     a.Metric,
					Score:      a.Score,
					Message:    a.Message,
					DetectedAt: a.DetectedAt,
				})
			}
			writeSuccess(c, out)
			return
		}
	}
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
	tid := tenantID(c)
	if h.deps.Spans != nil {
		ctx, cancel := withAnalyticsTimeout(c.Request.Context())
		defer cancel()
		if list, err := h.deps.Spans.ListDatabases(ctx, tid); err == nil && len(list) > 0 {
			out := make([]DatabaseInstance, 0, len(list))
			for _, d := range list {
				status := "healthy"
				if d.SlowQueries > 10 {
					status = "degraded"
				}
				out = append(out, DatabaseInstance{
					ID:          d.ID,
					Name:        d.Name,
					Engine:      d.Engine,
					Status:      status,
					QPS:         d.QPS,
					SlowQueries: d.SlowQueries,
					Connections: 0, // not available from trace telemetry
				})
			}
			writeSuccess(c, out)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.ListDatabases())
}

func (h *Handler) DBStatements(c *gin.Context) {
	tid := tenantID(c)
	id := c.Param("id")
	if h.deps.Spans != nil {
		ctx, cancel := withAnalyticsTimeout(c.Request.Context())
		defer cancel()
		if list, err := h.deps.Spans.DatabaseStatements(ctx, tid, id); err == nil && len(list) > 0 {
			out := make([]DBStatement, 0, len(list))
			for _, st := range list {
				out = append(out, DBStatement{Query: st.Query, Calls: st.Calls, AvgMs: st.AvgMs, TotalMs: st.TotalMs})
			}
			writeSuccess(c, out)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.DBStatements(id))
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

// reconcileWorkflowGraph treats the authored graph as the source of truth: when
// nodes are present the linear execution `steps` are recomputed via topological
// sort so the executor never diverges from the branching topology the user drew.
func reconcileWorkflowGraph(w *Workflow) {
	if w.Graph != nil && len(w.Graph.Nodes) > 0 {
		w.Steps = LinearizeGraph(w.Graph)
	}
}

func (h *Handler) CreateWorkflow(c *gin.Context) {
	var w Workflow
	if err := c.ShouldBindJSON(&w); err != nil || w.Name == "" {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	reconcileWorkflowGraph(&w)
	if h.deps.PG != nil {
		if saved, err := h.deps.PG.SaveWorkflow(c.Request.Context(), tenantID(c), w); err == nil {
			writeSuccess(c, saved)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.SaveWorkflow(w))
}

func (h *Handler) UpdateWorkflow(c *gin.Context) {
	var w Workflow
	if err := c.ShouldBindJSON(&w); err != nil || w.Name == "" {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	w.ID = c.Param("id")
	reconcileWorkflowGraph(&w)
	if h.deps.PG != nil {
		if saved, err := h.deps.PG.UpdateWorkflow(c.Request.Context(), tenantID(c), w); err == nil {
			writeSuccess(c, saved)
			return
		}
	}
	if saved, ok := h.deps.Mem.UpdateWorkflow(w); ok {
		writeSuccess(c, saved)
		return
	}
	writeError(c, http.StatusNotFound, "workflow not found")
}

func (h *Handler) DeleteWorkflow(c *gin.Context) {
	id := c.Param("id")
	if h.deps.PG != nil {
		if err := h.deps.PG.DeleteWorkflow(c.Request.Context(), tenantID(c), id); err == nil {
			writeSuccess(c, gin.H{"deleted": true})
			return
		}
	}
	if h.deps.Mem.DeleteWorkflow(id) {
		writeSuccess(c, gin.H{"deleted": true})
		return
	}
	writeError(c, http.StatusNotFound, "workflow not found")
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
	tid := tenantID(c)
	if h.deps.CH != nil {
		ctx, cancel := withAnalyticsTimeout(c.Request.Context())
		defer cancel()
		if list, err := QuerySecurityVulnerabilities(ctx, h.deps.CH, tid); err == nil && len(list) > 0 {
			writeSuccess(c, list)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.ListVulnerabilities())
}

type batchVulnerabilityFinding struct {
	Source      string `json:"source"`
	AssetType   string `json:"assetType"`
	AssetID     string `json:"assetId"`
	CVE         string `json:"cveId"`
	CVEAlt      string `json:"cve"`
	Severity    string `json:"severity"`
	Service     string `json:"service"`
	Package     string `json:"package"`
	Description string `json:"description"`
	DetectedAt  time.Time `json:"detectedAt"`
}

type batchImportVulnerabilitiesRequest struct {
	Source      string                      `json:"source"`
	Findings    []batchVulnerabilityFinding `json:"findings"`
	Vulnerabilities []SecurityVulnerability `json:"vulnerabilities"`
}

func (h *Handler) BatchImportVulnerabilities(c *gin.Context) {
	var req batchImportVulnerabilitiesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	items := append([]SecurityVulnerability(nil), req.Vulnerabilities...)
	for _, f := range req.Findings {
		cve := f.CVE
		if cve == "" {
			cve = f.CVEAlt
		}
		desc := f.Description
		if desc == "" && f.Package != "" {
			desc = f.Package
		}
		svc := f.Service
		if svc == "" && f.AssetID != "" {
			svc = f.AssetID
		}
		items = append(items, SecurityVulnerability{
			CVE: cve, Severity: strings.ToUpper(f.Severity), Service: svc,
			Description: desc, DetectedAt: f.DetectedAt,
		})
	}
	if len(items) == 0 {
		writeError(c, http.StatusBadRequest, "findings or vulnerabilities required")
		return
	}
	count := h.deps.Mem.IngestVulnerabilities(items)
	writeSuccess(c, gin.H{"accepted": true, "source": req.Source, "count": count})
}

func (h *Handler) ListAttacks(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.CH != nil {
		ctx, cancel := withAnalyticsTimeout(c.Request.Context())
		defer cancel()
		if list, err := QuerySecurityAttacks(ctx, h.deps.CH, tid); err == nil && len(list) > 0 {
			writeSuccess(c, list)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.ListAttacks())
}

func (h *Handler) GetAttack(c *gin.Context) {
	id := c.Param("id")
	tid := tenantID(c)
	if h.deps.CH != nil {
		ctx, cancel := withAnalyticsTimeout(c.Request.Context())
		defer cancel()
		if list, err := QuerySecurityAttacks(ctx, h.deps.CH, tid); err == nil {
			for _, a := range list {
				if a.ID == id {
					writeSuccess(c, a)
					return
				}
			}
		}
	}
	if a, ok := h.deps.Mem.GetAttack(id); ok {
		writeSuccess(c, a)
		return
	}
	writeError(c, http.StatusNotFound, "attack not found")
}

func (h *Handler) ListSecurityFindings(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.SRS != nil && h.deps.SRS.available() {
		if list, err := h.deps.SRS.ListSecurityFindings(c.Request.Context(), tid); err == nil && len(list) > 0 {
			writeSuccess(c, list)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.ListSecurityFindings())
}

type importSecurityFindingsRequest struct {
	Source   string            `json:"source"`
	Findings []SecurityFinding `json:"findings"`
}

func (h *Handler) ImportSecurityFindings(c *gin.Context) {
	var req importSecurityFindingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	count := h.deps.Mem.IngestSecurityFindings(req.Findings)
	writeSuccess(c, gin.H{"accepted": true, "source": req.Source, "count": count})
}

type cspmImportRequest struct {
	Source   string            `json:"source"`
	Findings []SecurityFinding `json:"findings"`
}

func (h *Handler) ImportCSPMFindings(c *gin.Context) {
	var req cspmImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	for i := range req.Findings {
		if req.Findings[i].Category == "" {
			req.Findings[i].Category = "cspm"
		}
	}
	count := h.deps.Mem.IngestSecurityFindings(req.Findings)
	writeSuccess(c, gin.H{"accepted": true, "source": req.Source, "category": "cspm", "count": count})
}

func (h *Handler) GetSecurityPosture(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.SecurityPosture())
}

type siemExportRequest struct {
	Provider string `json:"provider"`
	Target   string `json:"target"`
}

func (h *Handler) ExportSecurityToSIEM(c *gin.Context) {
	var req siemExportRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Provider == "" {
		writeError(c, http.StatusBadRequest, "provider is required")
		return
	}
	if h.deps.SIEM != nil {
		res, err := h.deps.SIEM.Export(c.Request.Context(), req.Provider, req.Target, nil)
		if err != nil {
			writeError(c, http.StatusBadGateway, err.Error())
			return
		}
		writeSuccess(c, res)
		return
	}
	writeSuccess(c, gin.H{
		"queued":   true,
		"provider": req.Provider,
		"target":   req.Target,
	})
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

type createUserRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (h *Handler) CreateAdminUser(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
		writeError(c, http.StatusBadRequest, "email is required")
		return
	}
	role := string(auth.ParseRole(req.Role))
	tid := tenantID(c)
	if h.deps.Identity != nil {
		if user, err := auth.CreateUser(c.Request.Context(), h.deps.Identity, tid, req.Email, role); err == nil {
			writeSuccess(c, user)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.CreateAdminUser(tid, req.Email, role))
}

type updateUserRequest struct {
	Role   *string `json:"role"`
	Active *bool   `json:"active"`
}

func (h *Handler) UpdateAdminUser(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}
	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil || (req.Role == nil && req.Active == nil) {
		writeError(c, http.StatusBadRequest, "role or active is required")
		return
	}
	id := c.Param("id")
	tid := tenantID(c)
	var normalizedRole *string
	if req.Role != nil {
		r := string(auth.ParseRole(*req.Role))
		normalizedRole = &r
	}
	if h.deps.Identity != nil {
		ok := true
		if normalizedRole != nil {
			if err := auth.UpdateUserRole(c.Request.Context(), h.deps.Identity, tid, id, *normalizedRole); err != nil {
				ok = false
			}
		}
		if ok && req.Active != nil {
			if err := auth.SetUserActive(c.Request.Context(), h.deps.Identity, tid, id, *req.Active); err != nil {
				ok = false
			}
		}
		if ok {
			writeSuccess(c, gin.H{"updated": true})
			return
		}
	}
	if _, ok := h.deps.Mem.UpdateAdminUser(id, normalizedRole, req.Active); !ok {
		writeError(c, http.StatusNotFound, "user not found")
		return
	}
	writeSuccess(c, gin.H{"updated": true})
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
	stats := h.deps.Mem.Usage()
	if h.deps.CH != nil {
		ctx, cancel := withAnalyticsTimeout(c.Request.Context())
		defer cancel()
		if live, err := QueryUsageFromClickHouse(ctx, h.deps.CH, tenantID(c), time.Now().Add(-24*time.Hour)); err == nil {
			if live.LogsIngestedGB > 0 {
				stats.LogsIngestedGB = live.LogsIngestedGB
			}
			if live.TracesIngested > 0 {
				stats.TracesIngested = live.TracesIngested
			}
			if live.MetricsIngested > 0 {
				stats.MetricsIngested = live.MetricsIngested
			}
		}
	}
	writeSuccess(c, stats)
}

func (h *Handler) ListAlertPolicies(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.SRS != nil && h.deps.SRS.available() {
		if list, err := h.deps.SRS.ListAlertPolicies(c.Request.Context(), tid); err == nil && len(list) > 0 {
			writeSuccess(c, list)
			return
		}
	}
	writeSuccess(c, h.deps.Mem.ListAlertPolicies())
}

func (h *Handler) CreateAlertPolicy(c *gin.Context) {
	var p AlertPolicy
	if err := c.ShouldBindJSON(&p); err != nil || p.Name == "" {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	p = h.deps.Mem.SaveAlertPolicy(p)
	if h.deps.SRS != nil && h.deps.SRS.available() {
		_ = h.deps.SRS.SaveAlertPolicy(c.Request.Context(), tenantID(c), p)
	}
	writeSuccess(c, p)
}

func (h *Handler) UpdateAlertPolicy(c *gin.Context) {
	var p AlertPolicy
	if err := c.ShouldBindJSON(&p); err != nil || p.Name == "" {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	p.ID = c.Param("id")
	writeSuccess(c, h.deps.Mem.SaveAlertPolicy(p))
}

func (h *Handler) DeleteAlertPolicy(c *gin.Context) {
	if !h.deps.Mem.DeleteAlertPolicy(c.Param("id")) {
		writeError(c, http.StatusNotFound, "alert policy not found")
		return
	}
	writeSuccess(c, gin.H{"deleted": true})
}

type createSuppressionRequest struct {
	ServicePattern string `json:"servicePattern"`
	Reason         string `json:"reason"`
	Duration       string `json:"duration"`
}

func (h *Handler) ListAlertSuppressions(c *gin.Context) {
	writeSuccess(c, h.deps.Mem.ListAlertSuppressions())
}

func (h *Handler) CreateAlertSuppression(c *gin.Context) {
	var req createSuppressionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	dur := 30 * time.Minute
	if req.Duration != "" {
		if parsed, err := time.ParseDuration(req.Duration); err == nil {
			dur = parsed
		}
	}
	now := time.Now().UTC()
	sup := AlertSuppression{
		ServicePattern: req.ServicePattern,
		Reason:         req.Reason,
		StartsAt:       now,
		EndsAt:         now.Add(dur),
		CreatedBy:      "ui",
	}
	writeSuccess(c, h.deps.Mem.SaveAlertSuppression(sup))
}

func (h *Handler) DeleteAlertSuppression(c *gin.Context) {
	if !h.deps.Mem.DeleteAlertSuppression(c.Param("id")) {
		writeError(c, http.StatusNotFound, "alert suppression not found")
		return
	}
	writeSuccess(c, gin.H{"deleted": true})
}

// requireAdmin enforces the ADMIN role for mutating admin endpoints. When no
// principal is present (e.g. auth disabled in local dev) the call is allowed,
// consistent with the gateway's dev-mode behaviour.
func (h *Handler) requireAdmin(c *gin.Context) bool {
	if p, ok := auth.FromContext(c.Request.Context()); ok && p.Role != auth.RoleAdmin {
		writeError(c, http.StatusForbidden, "admin role required")
		return false
	}
	return true
}
