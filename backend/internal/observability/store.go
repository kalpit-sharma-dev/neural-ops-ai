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

	traces             map[string]TraceDetail
	dashboards         map[string]Dashboard
	logMetrics         map[string]LogMetricRule
	logParsing         map[string]LogParsingRule
	slos               map[string]SLO
	workflows          map[string]Workflow
	notebooks          map[string]Notebook
	synthetic          map[string]SyntheticMonitor
	syntheticRuns      map[string][]SyntheticRun
	zones              map[string]ManagementZone
	alertPolicies      map[string]AlertPolicy
	suppressions       map[string]AlertSuppression
	collectorFleet     map[string]CollectorFleetAgent
	collectorPipelines map[string]CollectorPipeline
	rcaByIncident      map[string]RCAResponse
	explanations       map[string]RCAResponse
	autofixPlans       map[string]AutoFixPlan
	autofixActions     map[string]AutoFixActionRecord
	cloudAssets        []CloudAsset
	cloudAssetTopology CloudAssetTopology
	finopsAnomalies    []FinOpsCostAnomaly
	finopsCarbon       FinOpsCarbonFootprint
	networkFlows       []NetworkFlow
	networkDevices     []NetworkDevice
	networkTopology    NetworkTopologyGraph
	networkAnomalies          []NetworkAnomaly
	rumFunnels                map[string]RUMFunnel
	syntheticBrowserTests     map[string]SyntheticBrowserTest
	syntheticMobileTests      map[string]SyntheticMobileTest
	syntheticPrivateLocations map[string]SyntheticPrivateLocation
	businessKPIPacks          map[string]BusinessKPIPack
	abacPolicy                ABACPolicy
	dataResidency             DataResidencyPolicy
	branding                  BrandingTheme
	mspTenants                []MSPTenant
	exportJobs                map[string]ExportJob
	derivedMetrics            map[string]DerivedMetric
	securityFindingIncidents  map[string]string
	nfrBenchmarks             []NFRBenchmark
	nfrReliability            []NFRReliabilityDrill
	nfrA11y                   NFRA11yReport
	nfrLocales                []NFRLocale
	nfrCertification          NFRCertificationReport
	adminUsers                []AdminUser
	vulnerabilities           map[string]SecurityVulnerability
	securityFindings          map[string]SecurityFinding
	logTierPolicies           map[string]LogTierPolicy
	serverlessFunctions       map[string]ServerlessFunction
	savedQueries              map[string][]SavedQuery
	signalPolicies            map[string][]SignalPolicy
	siemConfigs               map[string]SIEMConfig
	serviceCatalog            map[string]ServiceCatalogEntry
	apmReleases               map[string]APMRelease
	logAnomalies              map[string]LogAnomaly
	cardinalityPolicies       map[string]CardinalityAlertPolicy
	maintenanceWindows        map[string]MaintenanceWindow
	fatigueConfigs            map[string]AlertFatigueConfig
	discoveredServices        []DiscoveredService
	threatFeed                []ThreatIndicator
	netflowRecords            []NetFlowRecord
	// extensionInstalls maps extension key -> stored config (marketplace demo state).
	extensionInstalls map[string]map[string]string
}

// NewStore creates a store seeded with demo observability data.
func NewStore() *Store {
	s := &Store{
		traces:             make(map[string]TraceDetail),
		dashboards:         make(map[string]Dashboard),
		logMetrics:         make(map[string]LogMetricRule),
		logParsing:         make(map[string]LogParsingRule),
		slos:               make(map[string]SLO),
		workflows:          make(map[string]Workflow),
		notebooks:          make(map[string]Notebook),
		synthetic:          make(map[string]SyntheticMonitor),
		syntheticRuns:      make(map[string][]SyntheticRun),
		zones:              make(map[string]ManagementZone),
		alertPolicies:      make(map[string]AlertPolicy),
		suppressions:       make(map[string]AlertSuppression),
		collectorFleet:     make(map[string]CollectorFleetAgent),
		collectorPipelines: make(map[string]CollectorPipeline),
		rcaByIncident:      make(map[string]RCAResponse),
		explanations:       make(map[string]RCAResponse),
		autofixPlans:       make(map[string]AutoFixPlan),
		autofixActions:            make(map[string]AutoFixActionRecord),
		rumFunnels:                make(map[string]RUMFunnel),
		syntheticBrowserTests:     make(map[string]SyntheticBrowserTest),
		syntheticMobileTests:      make(map[string]SyntheticMobileTest),
		syntheticPrivateLocations: make(map[string]SyntheticPrivateLocation),
		businessKPIPacks:          make(map[string]BusinessKPIPack),
		exportJobs:                make(map[string]ExportJob),
		derivedMetrics:           make(map[string]DerivedMetric),
		securityFindingIncidents: make(map[string]string),
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
	s.alertPolicies["ap-1"] = AlertPolicy{
		ID:             "ap-1",
		Name:           "P1 payment escalation",
		ServicePattern: "payment-*",
		Severity:       "P1",
		Enabled:        true,
		Expression:     "severity:P1 AND service:payment-*",
		DedupeKey:      "payment-p1",
		Context:        defaultAlertContext("payment-service"),
		Routes: []AlertPolicyRoute{
			{Channel: "slack", Target: "#oncall-payments", After: "0m", Priority: 1},
			{Channel: "pagerduty", Target: "payments-primary", After: "5m", Priority: 2},
		},
	}
	s.alertPolicies["finops-budget-alert"] = AlertPolicy{
		ID:             "finops-budget-alert",
		Name:           "FinOps budget threshold",
		ServicePattern: "*",
		Severity:       "P2",
		Enabled:        true,
		Expression:     "finops:budget_threshold",
		DedupeKey:      "finops-budget",
		Context: AlertContext{
			RunbookURL: "/docs/RUNBOOK.md#finops-budgets",
			Owner:      "finops@neuralops.ai",
		},
		Routes: []AlertPolicyRoute{
			{Channel: "slack", Target: "#finops-alerts", After: "0m", Priority: 1},
			{Channel: "email", Target: "finops@neuralops.ai", After: "0m", Priority: 2},
		},
	}
	s.alertPolicies["finops-ingest-stale"] = AlertPolicy{
		ID:             "finops-ingest-stale",
		Name:           "FinOps billing ingest stale",
		ServicePattern: "finops-ingest",
		Severity:       "P2",
		Enabled:        true,
		Expression:     "finops:ingest_stale",
		DedupeKey:      "finops-ingest",
		Context: AlertContext{
			RunbookURL: "/docs/RUNBOOK.md#finops-ingest",
			Owner:      "finops@neuralops.ai",
		},
		Routes: []AlertPolicyRoute{
			{Channel: "slack", Target: "#finops-alerts", After: "0m", Priority: 1},
		},
	}
	s.alertPolicies["finops-commitment-alert"] = AlertPolicy{
		ID:             "finops-commitment-alert",
		Name:           "FinOps commitment expiry/utilization",
		ServicePattern: "*",
		Severity:       "P2",
		Enabled:        true,
		Expression:     "finops:commitment_alert",
		DedupeKey:      "finops-commitment",
		Context: AlertContext{
			RunbookURL: "/docs/RUNBOOK.md#finops-commitments",
			Owner:      "finops@neuralops.ai",
		},
		Routes: []AlertPolicyRoute{
			{Channel: "slack", Target: "#finops-alerts", After: "0m", Priority: 1},
			{Channel: "email", Target: "finops@neuralops.ai", After: "1h", Priority: 2},
		},
	}
	s.alertPolicies["finops-cost-anomaly"] = AlertPolicy{
		ID:             "finops-cost-anomaly",
		Name:           "FinOps cost anomaly",
		ServicePattern: "*",
		Severity:       "P2",
		Enabled:        true,
		Expression:     "finops:cost_anomaly",
		DedupeKey:      "finops-cost",
		Context: AlertContext{
			RunbookURL: "/docs/RUNBOOK.md#finops-cost-anomaly",
			Owner:      "finops@neuralops.ai",
		},
		Routes: []AlertPolicyRoute{
			{Channel: "slack", Target: "#finops-alerts", After: "0m", Priority: 1},
			{Channel: "jira", Target: "FINOPS", After: "15m", Priority: 2},
		},
	}
	s.collectorFleet["agent-1"] = CollectorFleetAgent{
		ID: "agent-1", Name: "prod-node-01", Environment: "prod", Version: "1.2.0", Status: "healthy",
		LastHeartbeatAt: now.Add(-25 * time.Second), PolicyID: "default",
	}
	s.collectorFleet["agent-2"] = CollectorFleetAgent{
		ID: "agent-2", Name: "prod-node-02", Environment: "prod", Version: "1.1.4", Status: "degraded",
		LastHeartbeatAt: now.Add(-95 * time.Second), PolicyID: "default",
	}
	s.collectorPipelines["pipe-1"] = CollectorPipeline{
		ID: "pipe-1", Name: "default-ingest", Description: "Primary ingest parse+enrich pipeline", Enabled: true,
		Stages: []CollectorPipelineStage{
			{ID: "st-parse", Type: "parse", Config: map[string]string{"format": "json"}, Enabled: true},
			{ID: "st-mask", Type: "mask", Config: map[string]string{"fields": "password,token"}, Enabled: true},
			{ID: "st-route", Type: "route", Config: map[string]string{"target": "analytics"}, Enabled: true},
		},
		UpdatedAt: now,
	}
	s.seedAI()
	s.seedPhase5()
	s.seedPhase6()
	s.seedPhase7()
	s.seedPhase8()
	s.seedRoadmapPartial()
	s.seedDerivedMetricsCatalog()
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
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.vulnerabilities) > 0 {
		out := make([]SecurityVulnerability, 0, len(s.vulnerabilities))
		for _, v := range s.vulnerabilities {
			out = append(out, v)
		}
		sort.Slice(out, func(i, j int) bool {
			return out[i].DetectedAt.After(out[j].DetectedAt)
		})
		return out
	}
	now := time.Now().UTC()
	return []SecurityVulnerability{
		{ID: "v1", CVE: "CVE-2025-1234", Severity: "HIGH", Service: "payment-service", Description: "Outdated dependency in payment SDK", DetectedAt: now.Add(-24 * time.Hour)},
	}
}

// IngestVulnerabilities upserts scanner findings (SEC-01 batch ingest).
func (s *Store) IngestVulnerabilities(items []SecurityVulnerability) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.vulnerabilities == nil {
		s.vulnerabilities = make(map[string]SecurityVulnerability)
	}
	count := 0
	for _, item := range items {
		if item.CVE == "" && item.Description == "" {
			continue
		}
		if item.ID == "" {
			item.ID = "v-" + uuid.New().String()[:8]
		}
		if item.DetectedAt.IsZero() {
			item.DetectedAt = time.Now().UTC()
		}
		s.vulnerabilities[item.ID] = item
		count++
	}
	return count
}

// IngestSecurityFindings upserts normalized AppSec findings.
func (s *Store) IngestSecurityFindings(findings []SecurityFinding) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.securityFindings == nil {
		s.securityFindings = make(map[string]SecurityFinding)
	}
	count := 0
	for _, f := range findings {
		if f.Title == "" && f.ID == "" {
			continue
		}
		if f.ID == "" {
			f.ID = "sf-" + uuid.New().String()[:8]
		}
		if f.DetectedAt.IsZero() {
			f.DetectedAt = time.Now().UTC()
		}
		if f.Status == "" {
			f.Status = "OPEN"
		}
		s.securityFindings[f.ID] = f
		count++
	}
	return count
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

// ListSecurityFindings returns normalized security findings across AppSec and posture feeds.
func (s *Store) ListSecurityFindings() []SecurityFinding {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now().UTC()
	list := []SecurityFinding{
		{
			ID: "sf-1", Title: "Vulnerable dependency in payment-service", Category: "sca", Severity: "HIGH",
			Service: "payment-service", Asset: "container:payment-api:v1.4.2", Status: "OPEN", Exploitability: "reachable",
			IncidentID: "inc-1", DetectedAt: now.Add(-2 * time.Hour),
		},
		{
			ID: "sf-2", Title: "Privileged container in prod namespace", Category: "cspm", Severity: "MEDIUM",
			Service: "gateway", Asset: "k8s:platform/gateway", Status: "OPEN", Exploitability: "configuration", DetectedAt: now.Add(-5 * time.Hour),
		},
	}
	if len(s.securityFindings) > 0 {
		list = make([]SecurityFinding, 0, len(s.securityFindings))
		for _, f := range s.securityFindings {
			list = append(list, f)
		}
		sort.Slice(list, func(i, j int) bool {
			return list[i].DetectedAt.After(list[j].DetectedAt)
		})
	}
	for i := range list {
		if inc, ok := s.securityFindingIncidents[list[i].ID]; ok {
			list[i].IncidentID = inc
		}
	}
	return list
}

// LinkSecurityFindingIncident associates a finding with an incident id.
func (s *Store) LinkSecurityFindingIncident(findingID, incidentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.securityFindingIncidents == nil {
		s.securityFindingIncidents = make(map[string]string)
	}
	s.securityFindingIncidents[findingID] = incidentID
}

// SecurityPosture returns posture checks summary.
func (s *Store) SecurityPosture() []SecurityPostureCheck {
	return []SecurityPostureCheck{
		{ID: "pc-1", Name: "Public S3 bucket disabled", Provider: "aws", Resource: "s3://prod-audit-logs", Status: "pass", Severity: "LOW"},
		{ID: "pc-2", Name: "Kubernetes API server anonymous auth disabled", Provider: "kubernetes", Resource: "prod-cluster", Status: "fail", Severity: "HIGH"},
		{ID: "pc-3", Name: "Security group ingress restricted", Provider: "aws", Resource: "sg-0a12b", Status: "warn", Severity: "MEDIUM"},
	}
}

// ListIntegrations returns integrations.
func (s *Store) ListIntegrations() []Integration {
	return s.IntegrationsWithDevOps()
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

// UnifiedQuery runs a simple cross-signal in-memory query.
func (s *Store) UnifiedQuery(req UnifiedQueryRequest) UnifiedQueryResponse {
	q := strings.TrimSpace(strings.ToLower(req.Query))
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}
	now := time.Now().UTC()
	hits := make([]UnifiedQueryHit, 0, 8)
	add := func(h UnifiedQueryHit) {
		if len(hits) >= limit {
			return
		}
		hits = append(hits, h)
	}
	want := func(signal string) bool {
		switch strings.ToLower(strings.TrimSpace(req.From)) {
		case "", "all":
			return true
		default:
			return strings.EqualFold(signal, req.From)
		}
	}
	matches := func(values ...string) bool {
		if q == "" {
			return true
		}
		for _, v := range values {
			if strings.Contains(strings.ToLower(v), q) {
				return true
			}
		}
		return false
	}
	if want("logs") && matches("error rate spike", req.Service, "payment-service") {
		add(UnifiedQueryHit{
			ID: "log-err-1", Signal: "logs", Service: "payment-service",
			Title: "Error rate spike in payment-service", Summary: "5xx responses increased 3.2x baseline in last 15m",
			Severity: "P1", Timestamp: now.Add(-12 * time.Minute), Link: "/logs?q=payment-service+status:500",
			Fields: map[string]interface{}{"errorRate": 3.2, "window": "15m"},
		})
	}
	if want("metrics") && matches("latency p95", req.Service, "payment-service") {
		add(UnifiedQueryHit{
			ID: "metric-lat-1", Signal: "metrics", Service: "payment-service",
			Title: "P95 latency elevated", Summary: "P95 latency above 300ms threshold",
			Severity: "P2", Timestamp: now.Add(-10 * time.Minute), Link: "/metrics",
			Fields: map[string]interface{}{"p95Ms": 348},
		})
	}
	if want("traces") && matches("trace error", req.Service, "payment-service") {
		traceID := "trace-demo-04"
		if req.TraceID != "" {
			traceID = req.TraceID
		}
		add(UnifiedQueryHit{
			ID: traceID, Signal: "traces", Service: "payment-service",
			Title: "Failing checkout trace", Summary: "Trace with downstream ledger timeout",
			Severity: "P1", Timestamp: now.Add(-8 * time.Minute), Link: "/traces/" + traceID,
			Fields: map[string]interface{}{"traceId": traceID, "txnId": req.TxnID},
		})
	}
	if req.TxnID != "" && want("logs") {
		add(UnifiedQueryHit{
			ID: "log-txn-" + req.TxnID, Signal: "logs", Service: coalesce(req.Service, "payment-service"),
			Title: "Transaction log correlation", Summary: "Logs correlated by txnId=" + req.TxnID,
			Severity: "info", Timestamp: now.Add(-6 * time.Minute), Link: "/logs?q=txnId:" + req.TxnID,
			Fields: map[string]interface{}{"txnId": req.TxnID},
		})
	}
	if want("events") && matches("deployment", req.Service, "payment-service") {
		add(UnifiedQueryHit{
			ID: "evt-deploy-1", Signal: "events", Service: "payment-service",
			Title: "Recent deployment detected", Summary: "Release 2026.06.02.4 deployed 9 minutes before anomaly",
			Severity: "info", Timestamp: now.Add(-9 * time.Minute), Link: "/incidents",
		})
	}
	return UnifiedQueryResponse{Hits: hits, Count: len(hits)}
}

func coalesce(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// ListSavedQueries returns saved NexQL queries for a tenant.
func (s *Store) ListSavedQueries(tenantID string) []SavedQuery {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.savedQueries == nil {
		return []SavedQuery{}
	}
	out := append([]SavedQuery(nil), s.savedQueries[tenantID]...)
	return out
}

// SaveSavedQuery upserts a saved query.
func (s *Store) SaveSavedQuery(tenantID string, q SavedQuery) SavedQuery {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.savedQueries == nil {
		s.savedQueries = make(map[string][]SavedQuery)
	}
	q.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	list := s.savedQueries[tenantID]
	found := false
	for i, existing := range list {
		if existing.ID == q.ID {
			list[i] = q
			found = true
			break
		}
	}
	if !found {
		list = append(list, q)
	}
	s.savedQueries[tenantID] = list
	return q
}

// DeleteSavedQuery removes a saved query by id.
func (s *Store) DeleteSavedQuery(tenantID, id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.savedQueries == nil {
		return false
	}
	list := s.savedQueries[tenantID]
	out := list[:0]
	deleted := false
	for _, q := range list {
		if q.ID == id {
			deleted = true
			continue
		}
		out = append(out, q)
	}
	if deleted {
		s.savedQueries[tenantID] = out
	}
	return deleted
}

// ListAlertPolicies returns all alert policies in stable order.
func (s *Store) ListAlertPolicies() []AlertPolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]AlertPolicy, 0, len(s.alertPolicies))
	for _, p := range s.alertPolicies {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// SaveAlertPolicy upserts an alert policy.
func (s *Store) SaveAlertPolicy(p AlertPolicy) AlertPolicy {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p.ID == "" {
		p.ID = "ap-" + uuid.New().String()[:8]
	}
	s.alertPolicies[p.ID] = p
	return p
}

// DeleteAlertPolicy removes one policy by ID.
func (s *Store) DeleteAlertPolicy(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.alertPolicies[id]; !ok {
		return false
	}
	delete(s.alertPolicies, id)
	return true
}

// ListAlertSuppressions returns all suppressions by latest end time.
func (s *Store) ListAlertSuppressions() []AlertSuppression {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]AlertSuppression, 0, len(s.suppressions))
	for _, sup := range s.suppressions {
		out = append(out, sup)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].EndsAt.After(out[j].EndsAt) })
	return out
}

// SaveAlertSuppression creates a suppression.
func (s *Store) SaveAlertSuppression(sup AlertSuppression) AlertSuppression {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sup.ID == "" {
		sup.ID = "sup-" + uuid.New().String()[:8]
	}
	s.suppressions[sup.ID] = sup
	return sup
}

// DeleteAlertSuppression removes a suppression by ID.
func (s *Store) DeleteAlertSuppression(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.suppressions[id]; !ok {
		return false
	}
	delete(s.suppressions, id)
	return true
}

// ListDerivedMetrics returns configured derived metrics.
func (s *Store) ListDerivedMetrics() []DerivedMetric {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.derivedMetrics) == 0 {
		return []DerivedMetric{
			{ID: "dm-1", Name: "checkout_error_budget", Expression: "error_rate * 100", Unit: "percent", UpdatedAt: time.Now().UTC()},
		}
	}
	out := make([]DerivedMetric, 0, len(s.derivedMetrics))
	for _, m := range s.derivedMetrics {
		out = append(out, m)
	}
	return out
}

// SaveDerivedMetric stores a derived metric definition.
func (s *Store) SaveDerivedMetric(m DerivedMetric) DerivedMetric {
	s.mu.Lock()
	defer s.mu.Unlock()
	if m.ID == "" {
		m.ID = "dm-" + uuid.New().String()[:8]
	}
	m.UpdatedAt = time.Now().UTC()
	m.Value = EvaluateDerivedMetric(m.Expression)
	if s.derivedMetrics == nil {
		s.derivedMetrics = make(map[string]DerivedMetric)
	}
	s.derivedMetrics[m.ID] = m
	return m
}

// ListCollectorFleet returns collector agents in stable order.
func (s *Store) ListCollectorFleet() []CollectorFleetAgent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]CollectorFleetAgent, 0, len(s.collectorFleet))
	for _, a := range s.collectorFleet {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// SaveCollectorAgent upserts one collector agent.
func (s *Store) SaveCollectorAgent(a CollectorFleetAgent) CollectorFleetAgent {
	s.mu.Lock()
	defer s.mu.Unlock()
	if a.ID == "" {
		a.ID = "agent-" + uuid.New().String()[:8]
	}
	if a.LastHeartbeatAt.IsZero() {
		a.LastHeartbeatAt = time.Now().UTC()
	}
	s.collectorFleet[a.ID] = a
	return a
}

// UpdateCollectorAgent updates an existing collector agent.
func (s *Store) UpdateCollectorAgent(id string, a CollectorFleetAgent) (CollectorFleetAgent, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.collectorFleet[id]; !ok {
		return CollectorFleetAgent{}, false
	}
	a.ID = id
	if a.LastHeartbeatAt.IsZero() {
		a.LastHeartbeatAt = time.Now().UTC()
	}
	s.collectorFleet[id] = a
	return a, true
}

// ListCollectorPipelines returns pipelines in stable order.
func (s *Store) ListCollectorPipelines() []CollectorPipeline {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]CollectorPipeline, 0, len(s.collectorPipelines))
	for _, p := range s.collectorPipelines {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// SaveCollectorPipeline upserts one collector pipeline.
func (s *Store) SaveCollectorPipeline(p CollectorPipeline) CollectorPipeline {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p.ID == "" {
		p.ID = "pipe-" + uuid.New().String()[:8]
	}
	p.UpdatedAt = time.Now().UTC()
	s.collectorPipelines[p.ID] = p
	return p
}

// ValidateCollectorPipeline performs lightweight validation checks.
func (s *Store) ValidateCollectorPipeline(p CollectorPipeline) (bool, string) {
	if strings.TrimSpace(p.Name) == "" {
		return false, "pipeline name is required"
	}
	if len(p.Stages) == 0 {
		return false, "at least one stage is required"
	}
	for _, st := range p.Stages {
		if strings.TrimSpace(st.Type) == "" {
			return false, "stage type is required"
		}
	}
	return true, "pipeline accepted"
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

// LogTierPolicy returns retention tiers for a tenant.
func (s *Store) LogTierPolicy(tenantID string) LogTierPolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.logTierPolicies != nil {
		if p, ok := s.logTierPolicies[tenantID]; ok {
			return p
		}
	}
	return LogTierPolicy{
		TenantID: tenantID, HotRetentionDays: 7, WarmRetentionDays: 30,
		ColdRetentionDays: 90, RestoreSLAHours: 4, UpdatedAt: time.Now().UTC(),
	}
}

// SaveLogTierPolicy upserts log tier policy for a tenant.
func (s *Store) SaveLogTierPolicy(tenantID string, p LogTierPolicy) LogTierPolicy {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.logTierPolicies == nil {
		s.logTierPolicies = make(map[string]LogTierPolicy)
	}
	p.TenantID = tenantID
	p.UpdatedAt = time.Now().UTC()
	s.logTierPolicies[tenantID] = p
	return p
}

// ListServerlessFunctions returns monitored FaaS workloads.
func (s *Store) ListServerlessFunctions() []ServerlessFunction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.serverlessFunctions) > 0 {
		out := make([]ServerlessFunction, 0, len(s.serverlessFunctions))
		for _, fn := range s.serverlessFunctions {
			out = append(out, fn)
		}
		sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
		return out
	}
	now := time.Now().UTC()
	return []ServerlessFunction{
		{ID: "fn-upi-auth", Name: "upi-auth-validator", Provider: "aws", Runtime: "provided.al2023", Region: "ap-south-1",
			Invocations24h: 1_240_000, ErrorRatePct: 0.08, P95DurationMs: 42, ColdStartPct: 1.2, MemoryMB: 512, Status: "healthy", UpdatedAt: now},
		{ID: "fn-ledger-sync", Name: "ledger-nightly-sync", Provider: "azure", Runtime: "node20", Region: "centralindia",
			Invocations24h: 12_400, ErrorRatePct: 0.0, P95DurationMs: 890, ColdStartPct: 8.5, MemoryMB: 1024, Status: "healthy", UpdatedAt: now},
	}
}

// GetServerlessFunction returns one function by id.
func (s *Store) GetServerlessFunction(id string) (ServerlessFunction, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if fn, ok := s.serverlessFunctions[id]; ok {
		return fn, true
	}
	now := time.Now().UTC()
	seed := []ServerlessFunction{
		{ID: "fn-upi-auth", Name: "upi-auth-validator", Provider: "aws", Runtime: "provided.al2023", Region: "ap-south-1",
			Invocations24h: 1_240_000, ErrorRatePct: 0.08, P95DurationMs: 42, ColdStartPct: 1.2, MemoryMB: 512, Status: "healthy", UpdatedAt: now},
		{ID: "fn-ledger-sync", Name: "ledger-nightly-sync", Provider: "azure", Runtime: "node20", Region: "centralindia",
			Invocations24h: 12_400, ErrorRatePct: 0.0, P95DurationMs: 890, ColdStartPct: 8.5, MemoryMB: 1024, Status: "healthy", UpdatedAt: now},
	}
	for _, fn := range seed {
		if fn.ID == id {
			return fn, true
		}
	}
	return ServerlessFunction{}, false
}
