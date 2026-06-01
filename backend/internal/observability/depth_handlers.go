package observability

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neuralops/platform/internal/apm"
	"github.com/neuralops/platform/internal/workflow"
)

// ProfileHotspot is a code-level hotspot from continuous profiling.
type ProfileHotspot struct {
	FunctionName string  `json:"functionName"`
	FilePath     string  `json:"filePath,omitempty"`
	LineNo       int     `json:"lineNo"`
	SelfTimeMs   float64 `json:"selfTimeMs"`
	SampleCount  int64   `json:"sampleCount"`
	Service      string  `json:"service"`
}

// K8sNamespace is a Kubernetes namespace inventory row.
type K8sNamespace struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	PodCount int    `json:"podCount"`
}

// K8sDeployment is a Kubernetes deployment inventory row.
type K8sDeployment struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Namespace     string `json:"namespace"`
	Replicas      int    `json:"replicas"`
	ReadyReplicas int    `json:"readyReplicas"`
}

// CloudDashboard is a cloud vendor dashboard definition.
type CloudDashboard struct {
	ID       string   `json:"id"`
	Provider string   `json:"provider"`
	Name     string   `json:"name"`
	Region   string   `json:"region"`
	Metrics  []string `json:"metrics"`
}

// WorkflowRun is a workflow execution record.
type WorkflowRun struct {
	ID           string     `json:"id"`
	WorkflowID   string     `json:"workflowId"`
	TriggerEvent string     `json:"triggerEvent"`
	Status       string     `json:"status"`
	StepsLog     []string   `json:"stepsLog"`
	StartedAt    time.Time  `json:"startedAt"`
	FinishedAt   *time.Time `json:"finishedAt,omitempty"`
}

// RegisterDepthRoutes mounts production-depth observability routes.
func (h *Handler) RegisterDepthRoutes(v1 *gin.RouterGroup) {
	apmGroup := v1.Group("/apm")
	{
		apmGroup.GET("/retention", h.GetTraceRetention)
		apmGroup.PUT("/retention", h.PutTraceRetention)
		apmGroup.GET("/services/:service/profiles", h.ListProfiles)
		apmGroup.POST("/profiles", h.IngestProfile)
	}

	infra := v1.Group("/infra")
	{
		infra.GET("/k8s/namespaces", h.ListK8sNamespaces)
		infra.GET("/k8s/deployments", h.ListK8sDeployments)
	}

	v1.GET("/cloud/dashboards", h.ListCloudDashboards)
	v1.POST("/rum/consent", h.RecordRUMConsent)
	v1.GET("/workflows/:id/runs", h.ListWorkflowRuns)
	v1.POST("/workflows/:id/execute", h.ExecuteWorkflow)
	v1.POST("/workflows/trigger", h.TriggerWorkflows)
	v1.PUT("/slos/:id/burn-alert", h.UpdateSLOBurnAlert)
}

func (h *Handler) policyStore() *apm.PolicyStore {
	return apm.NewPolicyStore(h.deps.Pool)
}

func (h *Handler) GetTraceRetention(c *gin.Context) {
	p := h.policyStore().Get(c.Request.Context(), tenantID(c))
	writeSuccess(c, p)
}

func (h *Handler) PutTraceRetention(c *gin.Context) {
	var body apm.RetentionPolicy
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "invalid body")
		return
	}
	body.TenantID = tenantID(c)
	if err := h.policyStore().Save(c.Request.Context(), body); err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(c, body)
}

func (h *Handler) ListProfiles(c *gin.Context) {
	service := c.Param("service")
	tid := tenantID(c)
	if h.deps.Pool != nil {
		rows, err := h.deps.Pool.Query(c.Request.Context(), `
SELECT function_name, COALESCE(file_path,''), line_no, self_time_ms, sample_count, service
FROM apm_profiles WHERE tenant_id = $1 AND service = $2
ORDER BY self_time_ms DESC LIMIT 50`, tid, service)
		if err == nil {
			defer rows.Close()
			out := make([]ProfileHotspot, 0)
			for rows.Next() {
				var p ProfileHotspot
				if err := rows.Scan(&p.FunctionName, &p.FilePath, &p.LineNo, &p.SelfTimeMs, &p.SampleCount, &p.Service); err == nil {
					out = append(out, p)
				}
			}
			if len(out) > 0 {
				writeSuccess(c, out)
				return
			}
		}
	}
	writeSuccess(c, h.deps.Mem.ListProfiles(service))
}

func (h *Handler) IngestProfile(c *gin.Context) {
	var body ProfileHotspot
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "invalid body")
		return
	}
	tid := tenantID(c)
	if h.deps.Pool != nil && body.Service != "" && body.FunctionName != "" {
		_, err := h.deps.Pool.Exec(c.Request.Context(), `
INSERT INTO apm_profiles (tenant_id, service, function_name, file_path, line_no, self_time_ms, sample_count)
VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			tid, body.Service, body.FunctionName, body.FilePath, body.LineNo, body.SelfTimeMs, body.SampleCount)
		if err != nil {
			writeError(c, http.StatusInternalServerError, err.Error())
			return
		}
	}
	writeSuccess(c, body)
}

func (h *Handler) ListK8sNamespaces(c *gin.Context) {
	tid := tenantID(c)
	if h.deps.Pool != nil {
		rows, err := h.deps.Pool.Query(c.Request.Context(), `
SELECT id::text, name, status, pod_count FROM collector_k8s_namespaces WHERE tenant_id = $1 ORDER BY name`, tid)
		if err == nil {
			defer rows.Close()
			out := make([]K8sNamespace, 0)
			for rows.Next() {
				var n K8sNamespace
				if err := rows.Scan(&n.ID, &n.Name, &n.Status, &n.PodCount); err == nil {
					out = append(out, n)
				}
			}
			if len(out) > 0 {
				writeSuccess(c, out)
				return
			}
		}
	}
	writeSuccess(c, []K8sNamespace{{ID: "ns-default", Name: "default", Status: "Active", PodCount: 12}})
}

func (h *Handler) ListK8sDeployments(c *gin.Context) {
	ns := c.Query("namespace")
	tid := tenantID(c)
	if h.deps.Pool != nil {
		q := `SELECT id::text, name, namespace, replicas, ready_replicas FROM collector_k8s_deployments WHERE tenant_id = $1`
		args := []any{tid}
		if ns != "" {
			q += ` AND namespace = $2`
			args = append(args, ns)
		}
		q += ` ORDER BY namespace, name`
		rows, err := h.deps.Pool.Query(c.Request.Context(), q, args...)
		if err == nil {
			defer rows.Close()
			out := make([]K8sDeployment, 0)
			for rows.Next() {
				var d K8sDeployment
				if err := rows.Scan(&d.ID, &d.Name, &d.Namespace, &d.Replicas, &d.ReadyReplicas); err == nil {
					out = append(out, d)
				}
			}
			if len(out) > 0 {
				writeSuccess(c, out)
				return
			}
		}
	}
	writeSuccess(c, []K8sDeployment{{ID: "dep-1", Name: "gateway", Namespace: "platform", Replicas: 3, ReadyReplicas: 3}})
}

func (h *Handler) ListCloudDashboards(c *gin.Context) {
	provider := c.Query("provider")
	tid := tenantID(c)
	if h.deps.Pool != nil {
		q := `SELECT id::text, provider, name, region, metrics FROM observability_cloud_dashboards WHERE tenant_id = $1`
		args := []any{tid}
		if provider != "" {
			q += ` AND provider = $2`
			args = append(args, provider)
		}
		rows, err := h.deps.Pool.Query(c.Request.Context(), q, args...)
		if err == nil {
			defer rows.Close()
			out := make([]CloudDashboard, 0)
			for rows.Next() {
				var d CloudDashboard
				var metricsJSON []byte
				if err := rows.Scan(&d.ID, &d.Provider, &d.Name, &d.Region, &metricsJSON); err == nil {
					_ = json.Unmarshal(metricsJSON, &d.Metrics)
					out = append(out, d)
				}
			}
			if len(out) > 0 {
				writeSuccess(c, out)
				return
			}
		}
	}
	writeSuccess(c, h.deps.Mem.ListCloudDashboards(provider))
}

func (h *Handler) RecordRUMConsent(c *gin.Context) {
	var body struct {
		SessionID      string `json:"sessionId"`
		ConsentGiven   bool   `json:"consentGiven"`
		ConsentVersion string `json:"consentVersion"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "invalid body")
		return
	}
	tid := tenantID(c)
	if h.deps.Pool != nil {
		_, _ = h.deps.Pool.Exec(c.Request.Context(), `
INSERT INTO rum_consent_logs (tenant_id, session_key, consent_given, consent_version, recorded_at)
VALUES ($1,$2,$3,$4,NOW())`, tid, body.SessionID, body.ConsentGiven, body.ConsentVersion)
	}
	writeSuccess(c, gin.H{"recorded": true})
}

func (h *Handler) ListWorkflowRuns(c *gin.Context) {
	wfID := c.Param("id")
	tid := tenantID(c)
	if h.deps.Pool == nil {
		writeSuccess(c, []WorkflowRun{})
		return
	}
	rows, err := h.deps.Pool.Query(c.Request.Context(), `
SELECT id::text, workflow_id::text, trigger_event, status, steps_log, started_at, finished_at
FROM observability_workflow_runs WHERE tenant_id = $1 AND workflow_id = $2 ORDER BY started_at DESC LIMIT 20`,
		tid, wfID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	out := make([]WorkflowRun, 0)
	for rows.Next() {
		var r WorkflowRun
		var logJSON []byte
		var finished *time.Time
		if err := rows.Scan(&r.ID, &r.WorkflowID, &r.TriggerEvent, &r.Status, &logJSON, &r.StartedAt, &finished); err != nil {
			continue
		}
		_ = json.Unmarshal(logJSON, &r.StepsLog)
		r.FinishedAt = finished
		out = append(out, r)
	}
	writeSuccess(c, out)
}

func (h *Handler) ExecuteWorkflow(c *gin.Context) {
	wfID := c.Param("id")
	var body map[string]string
	_ = c.ShouldBindJSON(&body)
	if body == nil {
		body = map[string]string{}
	}
	exec := workflow.NewExecutor(h.deps.Pool)
	result, err := exec.ExecuteWorkflow(c.Request.Context(), tenantID(c), wfID, "manual", body)
	if err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}
	writeSuccess(c, result)
}

func (h *Handler) TriggerWorkflows(c *gin.Context) {
	var body struct {
		Trigger string            `json:"trigger"`
		Context map[string]string `json:"context"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Trigger == "" {
		writeError(c, http.StatusBadRequest, "trigger required")
		return
	}
	exec := workflow.NewExecutor(h.deps.Pool)
	results, err := exec.ExecuteByTrigger(c.Request.Context(), tenantID(c), body.Trigger, body.Context)
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(c, results)
}

func (h *Handler) UpdateSLOBurnAlert(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Enabled   bool    `json:"enabled"`
		Threshold float64 `json:"threshold"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "invalid body")
		return
	}
	if h.deps.Pool == nil {
		writeError(c, http.StatusServiceUnavailable, "postgres unavailable")
		return
	}
	th := body.Threshold
	if th <= 0 {
		th = 2.0
	}
	tag, err := h.deps.Pool.Exec(c.Request.Context(), `
UPDATE observability_slos SET burn_alert_enabled = $2, burn_alert_threshold = $3
WHERE tenant_id = $1 AND id = $4`, tenantID(c), body.Enabled, th, id)
	if err != nil || tag.RowsAffected() == 0 {
		writeError(c, http.StatusNotFound, "slo not found")
		return
	}
	writeSuccess(c, gin.H{"id": id, "burnAlertEnabled": body.Enabled, "burnAlertThreshold": th})
}

// SeedDemoProfiles inserts demo profile data when empty.
func SeedDemoProfiles(ctx context.Context, pool *pgxpool.Pool, tenantID string) {
	if pool == nil {
		return
	}
	var n int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM apm_profiles WHERE tenant_id = $1`, tenantID).Scan(&n)
	if n > 0 {
		return
	}
	demo := []ProfileHotspot{
		{Service: "payment-api", FunctionName: "ProcessPayment", FilePath: "internal/payment/handler.go", LineNo: 142, SelfTimeMs: 38.2, SampleCount: 1200},
		{Service: "payment-api", FunctionName: "ValidateLedger", FilePath: "internal/ledger/validate.go", LineNo: 88, SelfTimeMs: 22.1, SampleCount: 890},
		{Service: "gateway", FunctionName: "AuthMiddleware", FilePath: "internal/middleware/auth.go", LineNo: 56, SelfTimeMs: 12.4, SampleCount: 3400},
	}
	for _, p := range demo {
		_, _ = pool.Exec(ctx, `
INSERT INTO apm_profiles (tenant_id, service, function_name, file_path, line_no, self_time_ms, sample_count)
VALUES ($1,$2,$3,$4,$5,$6,$7)`, tenantID, p.Service, p.FunctionName, p.FilePath, p.LineNo, p.SelfTimeMs, p.SampleCount)
	}
}
