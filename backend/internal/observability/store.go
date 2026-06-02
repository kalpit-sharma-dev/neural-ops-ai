package observability

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Store holds observability data for UI APIs (demo + in-memory CRUD).
type Store struct {
	mu sync.RWMutex

	traces        map[string]TraceDetail
	dashboards    map[string]Dashboard
	logMetrics    map[string]LogMetricRule
	logParsing    map[string]LogParsingRule
	slos          map[string]SLO
	workflows     map[string]Workflow
	notebooks     map[string]Notebook
	synthetic     map[string]SyntheticMonitor
	syntheticRuns map[string][]SyntheticRun
	zones         map[string]ManagementZone
	adminUsers    []AdminUser
	// extensionInstalls maps extension key -> stored config (marketplace demo state).
	extensionInstalls map[string]map[string]string
}

// NewStore creates a store seeded with demo observability data.
func NewStore() *Store {
	s := &Store{
		traces:        make(map[string]TraceDetail),
		dashboards:    make(map[string]Dashboard),
		logMetrics:    make(map[string]LogMetricRule),
		logParsing:    make(map[string]LogParsingRule),
		slos:          make(map[string]SLO),
		workflows:     make(map[string]Workflow),
		notebooks:     make(map[string]Notebook),
		synthetic:     make(map[string]SyntheticMonitor),
		syntheticRuns: make(map[string][]SyntheticRun),
		zones:         make(map[string]ManagementZone),
		adminUsers: []AdminUser{
			{ID: "u1", Email: "demo@neuralops.ai", Role: "ADMIN", Active: true, TenantID: "default"},
			{ID: "u2", Email: "sre@neuralops.ai", Role: "SRE", Active: true, TenantID: "default"},
		},
		extensionInstalls: map[string]map[string]string{
			"otel-collector": {},
		},
	}
	s.seed()
	return s
}

func (s *Store) seed() {
	now := time.Now().UTC()
	services := []string{"api-gateway", "payment-service", "auth-service", "ledger-service", "notification-service"}
	ops := []string{"POST /payments", "GET /health", "POST /auth/login", "POST /ledger/debit", "POST /notify"}

	for i := 0; i < 20; i++ {
		traceID := fmt.Sprintf("trace-demo-%02d", i+1)
		root := uuid.New().String()[:16]
		spans := make([]Span, 0, len(services))
		start := now.Add(-time.Duration(i*7) * time.Minute)
		offset := int64(0)
		parent := ""
		status := "OK"
		if i%4 == 0 {
			status = "ERROR"
		}
		for j, svc := range services {
			spanID := uuid.New().String()[:16]
			dur := int64(20 + j*35 + (i % 5 * 10))
			sp := Span{
				TraceID: traceID, SpanID: spanID, ParentID: parent, Service: svc,
				Operation: ops[j], StartTime: start.Add(time.Duration(offset) * time.Millisecond),
				DurationMs: dur, Status: status,
				Tags: map[string]string{"http.status_code": "200"},
			}
			if status == "ERROR" && j == len(services)-1 {
				sp.Tags["http.status_code"] = "500"
			}
			spans = append(spans, sp)
			parent = spanID
			offset += dur
		}
		total := offset
		s.traces[traceID] = TraceDetail{
			TraceID: traceID, RootSpan: root, Spans: spans, TotalMs: total,
			Service: services[0], Status: status, SpanCount: len(spans),
		}
	}

	dashID := uuid.New().String()
	s.dashboards[dashID] = Dashboard{
		ID: dashID, TenantID: "default", Name: "Platform Overview", Shared: true,
		Tiles: []DashboardTile{
			{ID: "t1", Type: "stat", Title: "Error Rate", Metric: "error_rate", Position: map[string]int{"x": 0, "y": 0, "w": 3, "h": 2}},
			{ID: "t2", Type: "timeseries", Title: "Request Latency", Metric: "latency_p95", Position: map[string]int{"x": 3, "y": 0, "w": 6, "h": 2}},
		},
		CreatedAt: now, UpdatedAt: now,
	}

	s.slos["slo-1"] = SLO{
		ID: "slo-1", TenantID: "default", Name: "Payment Availability", Service: "payment-service",
		SLIQuery: "sum(success)/sum(total)", Target: 99.9, WindowDays: 30,
		ErrorBudget: 0.72, BurnRate: 0.3, Status: "OK", CreatedAt: now,
	}

	s.workflows["wf-1"] = Workflow{
		ID: "wf-1", Name: "P1 Incident Escalation", Trigger: "incident.p1", Enabled: true,
		Steps: []string{"notify-slack", "page-oncall", "create-jira"},
		Graph: &WorkflowGraph{
			Nodes: []WorkflowNode{
				{ID: "n1", Type: "slack", Label: "notify-slack", X: 80, Y: 120},
				{ID: "n2", Type: "pagerduty", Label: "page-oncall", X: 360, Y: 40},
				{ID: "n3", Type: "jira", Label: "create-jira", X: 360, Y: 220},
			},
			Edges: []WorkflowEdge{
				{ID: "e-n1-n2", Source: "n1", Target: "n2"},
				{ID: "e-n1-n3", Source: "n1", Target: "n3"},
			},
		},
	}

	s.notebooks["nb-1"] = Notebook{
		ID: "nb-1", Name: "Payment RCA Template",
		Cells: []NotebookCell{
			{ID: "c1", Type: "markdown", Content: "# Payment failures investigation"},
			{ID: "c2", Type: "query", Content: "service:payment-service severity:ERROR"},
		},
		UpdatedAt: now,
	}

	s.synthetic["syn-1"] = SyntheticMonitor{
		ID: "syn-1", Name: "Checkout API", Type: "http", URL: "https://demo.neuralops.ai/health",
		Interval: "5m", Locations: []string{"us-east", "eu-west"}, Enabled: true,
		LastStatus: "OK", LastRunAt: now.Add(-2 * time.Minute),
	}
	s.syntheticRuns["syn-1"] = []SyntheticRun{
		{ID: uuid.New().String(), MonitorID: "syn-1", Status: "OK", LatencyMs: 142, Location: "us-east", RanAt: now.Add(-2 * time.Minute)},
	}

	s.zones["zone-1"] = ManagementZone{ID: "zone-1", Name: "Payments", Services: []string{"payment-service", "ledger-service"}}
	s.zones["zone-2"] = ManagementZone{ID: "zone-2", Name: "Platform", Services: []string{"api-gateway", "auth-service"}}
}

// GetTrace returns trace detail by ID.
func (s *Store) GetTrace(traceID string) (TraceDetail, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.traces[traceID]
	return t, ok
}

// SearchTraces returns trace summaries matching filters.
func (s *Store) SearchTraces(req TraceSearchRequest) []TraceSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	out := make([]TraceSummary, 0, limit)
	for _, t := range s.traces {
		if req.Service != "" && !strings.EqualFold(t.Service, req.Service) {
			continue
		}
		if req.Status != "" && !strings.EqualFold(t.Status, req.Status) {
			continue
		}
		if req.MinMs > 0 && t.TotalMs < req.MinMs {
			continue
		}
		if req.MaxMs > 0 && t.TotalMs > req.MaxMs {
			continue
		}
		op := ""
		if len(t.Spans) > 0 {
			op = t.Spans[0].Operation
		}
		start := time.Now().UTC()
		if len(t.Spans) > 0 {
			start = t.Spans[0].StartTime
		}
		out = append(out, TraceSummary{
			TraceID: t.TraceID, Service: t.Service, Operation: op,
			DurationMs: t.TotalMs, Status: t.Status, StartTime: start, SpanCount: t.SpanCount,
		})
		if len(out) >= limit {
			break
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartTime.After(out[j].StartTime) })
	return out
}

// ServiceFlow returns aggregated service flow edges.
func (s *Store) ServiceFlow() []FlowEdge {
	return []FlowEdge{
		{Source: "api-gateway", Target: "auth-service", CallCount: 12000, ErrorRate: 0.2, P50Ms: 12, P95Ms: 45},
		{Source: "api-gateway", Target: "payment-service", CallCount: 8500, ErrorRate: 1.4, P50Ms: 85, P95Ms: 320},
		{Source: "payment-service", Target: "ledger-service", CallCount: 8200, ErrorRate: 0.8, P50Ms: 55, P95Ms: 180},
		{Source: "payment-service", Target: "notification-service", CallCount: 7800, ErrorRate: 0.1, P50Ms: 22, P95Ms: 65},
	}
}

// MetricCatalog returns available metrics.
func (s *Store) MetricCatalog() []MetricCatalogEntry {
	return []MetricCatalogEntry{
		{Name: "request_rate", Description: "Requests per minute", Unit: "rpm", Labels: []string{"service"}},
		{Name: "error_rate", Description: "Error percentage", Unit: "percent", Labels: []string{"service"}},
		{Name: "latency_p95", Description: "95th percentile latency", Unit: "ms", Labels: []string{"service"}},
		{Name: "cpu_usage", Description: "CPU utilization", Unit: "percent", Labels: []string{"host"}},
	}
}

// QueryMetric returns synthetic time series for a metric.
func (s *Store) QueryMetric(name, service string, start, end time.Time) MetricSeries {
	if end.IsZero() {
		end = time.Now().UTC()
	}
	if start.IsZero() {
		start = end.Add(-1 * time.Hour)
	}
	points := make([]MetricSeriesPoint, 0, 12)
	step := end.Sub(start) / 12
	base := 50.0
	switch name {
	case "error_rate":
		base = 1.2
	case "latency_p95":
		base = 120
	case "cpu_usage":
		base = 45
	}
	for i := 0; i <= 12; i++ {
		ts := start.Add(time.Duration(i) * step)
		v := base + math.Sin(float64(i)/2)*base*0.2
		points = append(points, MetricSeriesPoint{Timestamp: ts, Value: v})
	}
	return MetricSeries{
		Name: name, Labels: map[string]string{"service": service}, Unit: "auto",
		Points: points,
	}
}

// Topology returns live topology graph.
func (s *Store) Topology(zone string) TopologyGraph {
	nodes := []TopologyNode{
		{ID: "api-gateway", DisplayName: "api-gateway", Type: "service", Health: "degraded", ErrorRate: 1.1, Throughput: 2400, Zone: "Platform"},
		{ID: "payment-service", DisplayName: "payment-service", Type: "service", Health: "critical", ErrorRate: 3.2, Throughput: 980, Zone: "Payments"},
		{ID: "auth-service", DisplayName: "auth-service", Type: "service", Health: "healthy", ErrorRate: 0.2, Throughput: 1800, Zone: "Platform"},
		{ID: "ledger-service", DisplayName: "ledger-service", Type: "service", Health: "degraded", ErrorRate: 0.9, Throughput: 920, Zone: "Payments"},
		{ID: "notification-service", DisplayName: "notification-service", Type: "service", Health: "healthy", ErrorRate: 0.1, Throughput: 860, Zone: "Payments"},
	}
	if zone != "" {
		filtered := make([]TopologyNode, 0)
		for _, n := range nodes {
			if strings.EqualFold(n.Zone, zone) {
				filtered = append(filtered, n)
			}
		}
		nodes = filtered
	}
	edges := []TopologyEdge{
		{Source: "api-gateway", Target: "auth-service", CallCount: 12000, ErrorRate: 0.2, P95Ms: 45},
		{Source: "api-gateway", Target: "payment-service", CallCount: 8500, ErrorRate: 1.4, P95Ms: 320},
		{Source: "payment-service", Target: "ledger-service", CallCount: 8200, ErrorRate: 0.8, P95Ms: 180},
		{Source: "payment-service", Target: "notification-service", CallCount: 7800, ErrorRate: 0.1, P95Ms: 65},
	}
	return TopologyGraph{Nodes: nodes, Edges: edges, At: time.Now().UTC()}
}

// ListDashboards returns dashboards for tenant.
func (s *Store) ListDashboards(_ string) []Dashboard {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Dashboard, 0, len(s.dashboards))
	for _, d := range s.dashboards {
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out
}

// GetDashboard returns one dashboard.
func (s *Store) GetDashboard(id string) (Dashboard, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.dashboards[id]
	return d, ok
}

// SaveDashboard upserts a dashboard.
func (s *Store) SaveDashboard(d Dashboard) Dashboard {
	s.mu.Lock()
	defer s.mu.Unlock()
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if d.CreatedAt.IsZero() {
		d.CreatedAt = now
	}
	d.UpdatedAt = now
	s.dashboards[d.ID] = d
	return d
}

// DeleteDashboard removes a dashboard.
func (s *Store) DeleteDashboard(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.dashboards[id]; !ok {
		return false
	}
	delete(s.dashboards, id)
	return true
}

// ListLogMetricRules returns log metric rules.
func (s *Store) ListLogMetricRules(tenantID string) []LogMetricRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]LogMetricRule, 0)
	for _, r := range s.logMetrics {
		if r.TenantID == tenantID || tenantID == "" {
			out = append(out, r)
		}
	}
	return out
}

// SaveLogMetricRule upserts a log metric rule.
func (s *Store) SaveLogMetricRule(r LogMetricRule) LogMetricRule {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	if r.TenantID == "" {
		r.TenantID = "default"
	}
	s.logMetrics[r.ID] = r
	return r
}

// ListLogParsingRules returns parsing rules.
func (s *Store) ListLogParsingRules(tenantID string) []LogParsingRule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]LogParsingRule, 0)
	for _, r := range s.logParsing {
		if r.TenantID == tenantID || tenantID == "" {
			out = append(out, r)
		}
	}
	return out
}

// SaveLogParsingRule upserts a parsing rule.
func (s *Store) SaveLogParsingRule(r LogParsingRule) LogParsingRule {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	if r.TenantID == "" {
		r.TenantID = "default"
	}
	s.logParsing[r.ID] = r
	return r
}

// ListSLOs returns SLOs.
func (s *Store) ListSLOs() []SLO {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]SLO, 0, len(s.slos))
	for _, slo := range s.slos {
		out = append(out, slo)
	}
	return out
}

// SaveSLO upserts an SLO.
func (s *Store) SaveSLO(slo SLO) SLO {
	s.mu.Lock()
	defer s.mu.Unlock()
	if slo.ID == "" {
		slo.ID = uuid.New().String()
	}
	if slo.CreatedAt.IsZero() {
		slo.CreatedAt = time.Now().UTC()
	}
	s.slos[slo.ID] = slo
	return slo
}

// ListAnomalies returns entity anomalies.
func (s *Store) ListAnomalies() []EntityAnomaly {
	now := time.Now().UTC()
	return []EntityAnomaly{
		{ID: "a1", EntityID: "payment-service", EntityType: "service", Service: "payment-service", Metric: "error_rate", Score: 78, Message: "Error rate 3x baseline", DetectedAt: now.Add(-15 * time.Minute)},
		{ID: "a2", EntityID: "ledger-service", EntityType: "service", Service: "ledger-service", Metric: "latency_p95", Score: 52, Message: "Latency elevated", DetectedAt: now.Add(-40 * time.Minute)},
	}
}

// ListHosts returns infrastructure hosts.
func (s *Store) ListHosts() []HostSummary {
	return []HostSummary{
		{ID: "host-1", Name: "prod-node-01", Status: "healthy", CPU: 42, Memory: 68, Disk: 55, Zone: "us-east-1a"},
		{ID: "host-2", Name: "prod-node-02", Status: "degraded", CPU: 78, Memory: 82, Disk: 61, Zone: "us-east-1b"},
		{ID: "host-3", Name: "prod-node-03", Status: "healthy", CPU: 35, Memory: 54, Disk: 48, Zone: "us-east-1c"},
	}
}

// ListK8sClusters returns K8s clusters.
func (s *Store) ListK8sClusters() []K8sCluster {
	return []K8sCluster{{ID: "k8s-1", Name: "prod-cluster", Nodes: 12, Pods: 186, Health: "healthy", NamespaceCount: 8}}
}

// ListK8sPods returns pods.
func (s *Store) ListK8sPods(namespace string) []K8sPod {
	pods := []K8sPod{
		{ID: "p1", Name: "payment-service-7d4f9", Namespace: "payments", Node: "prod-node-02", Status: "Running", CPU: 62, Memory: 71, Restarts: 1},
		{ID: "p2", Name: "api-gateway-5c8a1", Namespace: "platform", Node: "prod-node-01", Status: "Running", CPU: 38, Memory: 45, Restarts: 0},
	}
	if namespace == "" {
		return pods
	}
	out := make([]K8sPod, 0)
	for _, p := range pods {
		if p.Namespace == namespace {
			out = append(out, p)
		}
	}
	return out
}

// ListDatabases returns DB instances.
func (s *Store) ListDatabases() []DatabaseInstance {
	return []DatabaseInstance{
		{ID: "db-1", Name: "payments-primary", Engine: "postgres", Status: "healthy", QPS: 420, SlowQueries: 3, Connections: 85},
		{ID: "db-2", Name: "ledger-replica", Engine: "postgres", Status: "degraded", QPS: 280, SlowQueries: 12, Connections: 64},
	}
}

// DBStatements returns top statements for a database.
func (s *Store) DBStatements(_ string) []DBStatement {
	return []DBStatement{
		{Query: "UPDATE accounts SET balance = $1 WHERE id = $2", Calls: 12400, AvgMs: 8.2, TotalMs: 101680},
		{Query: "SELECT * FROM transactions WHERE txn_id = $1", Calls: 8900, AvgMs: 3.1, TotalMs: 27590},
	}
}

// KafkaLag returns consumer lag entries.
func (s *Store) KafkaLag() []KafkaLag {
	return []KafkaLag{
		{Topic: "raw-logs", ConsumerGroup: "analysis", Lag: 120, Partition: 0},
		{Topic: "raw-traces", ConsumerGroup: "correlation", Lag: 45, Partition: 1},
	}
}

// ListRUMSessions returns RUM sessions.
func (s *Store) ListRUMSessions() []RUMSession {
	now := time.Now().UTC()
	return []RUMSession{
		{ID: "s1", UserID: "user-42", Page: "/checkout", Device: "Chrome/macOS", Country: "IN", DurationMs: 4200, Errors: 1, LCP: 2.1, StartedAt: now.Add(-8 * time.Minute)},
		{ID: "s2", UserID: "user-17", Page: "/dashboard", Device: "Safari/iOS", Country: "US", DurationMs: 8900, Errors: 0, LCP: 1.4, StartedAt: now.Add(-22 * time.Minute)},
	}
}

// ListSyntheticMonitors returns synthetic monitors.
func (s *Store) ListSyntheticMonitors() []SyntheticMonitor {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]SyntheticMonitor, 0, len(s.synthetic))
	for _, m := range s.synthetic {
		out = append(out, m)
	}
	return out
}

// SaveSyntheticMonitor upserts a monitor.
func (s *Store) SaveSyntheticMonitor(m SyntheticMonitor) SyntheticMonitor {
	s.mu.Lock()
	defer s.mu.Unlock()
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	s.synthetic[m.ID] = m
	return m
}

// SyntheticRuns returns runs for a monitor.
func (s *Store) SyntheticRuns(monitorID string) []SyntheticRun {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]SyntheticRun(nil), s.syntheticRuns[monitorID]...)
}

// ListWorkflows returns workflows in a stable (name) order.
func (s *Store) ListWorkflows() []Workflow {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Workflow, 0, len(s.workflows))
	for _, w := range s.workflows {
		out = append(out, w)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// SaveWorkflow upserts a workflow in memory.
func (s *Store) SaveWorkflow(w Workflow) Workflow {
	s.mu.Lock()
	defer s.mu.Unlock()
	if w.ID == "" {
		w.ID = uuid.New().String()
	}
	s.workflows[w.ID] = w
	return w
}

// UpdateWorkflow replaces an existing workflow by ID.
func (s *Store) UpdateWorkflow(w Workflow) (Workflow, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.workflows[w.ID]; !ok {
		return Workflow{}, false
	}
	s.workflows[w.ID] = w
	return w, true
}

// DeleteWorkflow removes a workflow by ID.
func (s *Store) DeleteWorkflow(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.workflows[id]; !ok {
		return false
	}
	delete(s.workflows, id)
	return true
}

// ListExtensionInstalls returns a copy of installed marketplace extensions
// (key -> config) for demo/in-memory mode.
func (s *Store) ListExtensionInstalls() map[string]map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]map[string]string, len(s.extensionInstalls))
	for k, cfg := range s.extensionInstalls {
		copyCfg := make(map[string]string, len(cfg))
		for ck, cv := range cfg {
			copyCfg[ck] = cv
		}
		out[k] = copyCfg
	}
	return out
}

// InstallExtension records an extension install (and its config) in memory.
func (s *Store) InstallExtension(key string, cfg map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cfg == nil {
		cfg = map[string]string{}
	}
	s.extensionInstalls[key] = cfg
}

// UninstallExtension removes an extension install from memory.
func (s *Store) UninstallExtension(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.extensionInstalls, key)
}

// ListNotebooks returns notebooks.
func (s *Store) ListNotebooks() []Notebook {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Notebook, 0, len(s.notebooks))
	for _, n := range s.notebooks {
		out = append(out, n)
	}
	return out
}

// SaveNotebook upserts a notebook in memory.
func (s *Store) SaveNotebook(n Notebook) Notebook {
	s.mu.Lock()
	defer s.mu.Unlock()
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	n.UpdatedAt = time.Now().UTC()
	s.notebooks[n.ID] = n
	return n
}

// ListVulnerabilities returns security vulnerabilities.
func (s *Store) ListVulnerabilities() []SecurityVulnerability {
	now := time.Now().UTC()
	return []SecurityVulnerability{
		{ID: "v1", CVE: "CVE-2025-1234", Severity: "HIGH", Service: "payment-service", Description: "Outdated dependency in payment SDK", DetectedAt: now.Add(-24 * time.Hour)},
	}
}

// ListAttacks returns attack events.
func (s *Store) ListAttacks() []SecurityAttack {
	now := time.Now().UTC()
	return []SecurityAttack{
		{ID: "atk-1", Type: "SQL_INJECTION", SourceIP: "203.0.113.42", Service: "api-gateway", Blocked: true, DetectedAt: now.Add(-2 * time.Hour)},
	}
}

// GetAttack returns one attack by ID.
func (s *Store) GetAttack(id string) (SecurityAttack, bool) {
	for _, a := range s.ListAttacks() {
		if a.ID == id {
			return a, true
		}
	}
	return SecurityAttack{}, false
}

// ListIntegrations returns integrations.
func (s *Store) ListIntegrations() []Integration {
	return []Integration{
		{ID: "int-1", Name: "Slack", Type: "slack", Status: "connected", Connected: true},
		{ID: "int-2", Name: "Jira", Type: "jira", Status: "disconnected", Connected: false},
		{ID: "int-3", Name: "PagerDuty", Type: "pagerduty", Status: "connected", Connected: true},
	}
}

// ConnectIntegration marks an integration as connected (in-memory demo).
func (s *Store) ConnectIntegration(id string) (Integration, bool) {
	integrations := s.ListIntegrations()
	for _, i := range integrations {
		if i.ID == id {
			i.Connected = true
			i.Status = "connected"
			return i, true
		}
	}
	return Integration{}, false
}

// ListZones returns management zones.
func (s *Store) ListZones() []ManagementZone {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ManagementZone, 0, len(s.zones))
	for _, z := range s.zones {
		out = append(out, z)
	}
	return out
}

// DemoAdminUsers returns the in-memory admin users.
func (s *Store) DemoAdminUsers() []AdminUser {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]AdminUser, len(s.adminUsers))
	copy(out, s.adminUsers)
	return out
}

// CreateAdminUser provisions (or re-invites) an in-memory tenant user.
func (s *Store) CreateAdminUser(tenantID, email, role string) AdminUser {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.adminUsers {
		if strings.EqualFold(s.adminUsers[i].Email, email) {
			s.adminUsers[i].Role = role
			s.adminUsers[i].Active = true
			return s.adminUsers[i]
		}
	}
	user := AdminUser{
		ID:       "u" + uuid.New().String()[:8],
		Email:    email,
		Role:     role,
		Active:   true,
		TenantID: tenantID,
	}
	s.adminUsers = append(s.adminUsers, user)
	return user
}

// UpdateAdminUser mutates an in-memory user's role and/or active flag.
func (s *Store) UpdateAdminUser(id string, role *string, active *bool) (AdminUser, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.adminUsers {
		if s.adminUsers[i].ID == id {
			if role != nil {
				s.adminUsers[i].Role = *role
			}
			if active != nil {
				s.adminUsers[i].Active = *active
			}
			return s.adminUsers[i], true
		}
	}
	return AdminUser{}, false
}

// DemoAPIKeys returns demo API keys.
func (s *Store) DemoAPIKeys() []APIKey {
	return []APIKey{
		{ID: "k1", Name: "CI Pipeline", Prefix: "no_live_", Scopes: []string{"logs.read"}, CreatedAt: time.Now().UTC().Add(-30 * 24 * time.Hour)},
	}
}

// DemoAuditEntries returns demo audit log.
func (s *Store) DemoAuditEntries() []AuditEntry {
	now := time.Now().UTC()
	return []AuditEntry{
		{ID: "aud-1", UserID: "demo@neuralops.ai", Action: "login", Resource: "auth", Detail: "OIDC login success", Timestamp: now.Add(-1 * time.Hour)},
		{ID: "aud-2", UserID: "sre@neuralops.ai", Action: "create", Resource: "alert_rule", Detail: "Created rule payment-errors", Timestamp: now.Add(-3 * time.Hour)},
	}
}

// Usage returns demo usage stats.
func (s *Store) Usage() UsageStats {
	return UsageStats{LogsIngestedGB: 42.5, TracesIngested: 1_240_000, MetricsIngested: 8_900_000, AITokensUsed: 125_000, ActiveUsers: 12}
}

// ListProfiles returns demo code hotspots for a service.
func (s *Store) ListProfiles(service string) []ProfileHotspot {
	return []ProfileHotspot{
		{Service: service, FunctionName: "ProcessPayment", FilePath: "internal/payment/handler.go", LineNo: 142, SelfTimeMs: 38.2, SampleCount: 1200},
		{Service: service, FunctionName: "ValidateLedger", FilePath: "internal/ledger/validate.go", LineNo: 88, SelfTimeMs: 22.1, SampleCount: 890},
	}
}

// ListCloudDashboards returns demo cloud dashboards.
func (s *Store) ListCloudDashboards(provider string) []CloudDashboard {
	all := []CloudDashboard{
		{ID: "aws-1", Provider: "aws", Name: "EC2 & Lambda Overview", Region: "us-east-1", Metrics: []string{"CPUUtilization", "Duration", "Errors"}},
		{ID: "azure-1", Provider: "azure", Name: "App Service Health", Region: "eastus", Metrics: []string{"HttpResponseTime", "Http5xx"}},
		{ID: "gcp-1", Provider: "gcp", Name: "Cloud Run & GCE", Region: "us-central1", Metrics: []string{"request_count", "cpu_utilization"}},
	}
	if provider == "" {
		return all
	}
	out := make([]CloudDashboard, 0)
	for _, d := range all {
		if d.Provider == provider {
			out = append(out, d)
		}
	}
	return out
}
