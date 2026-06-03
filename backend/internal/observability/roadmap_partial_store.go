package observability

import (
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *Store) seedRoadmapPartial() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.signalPolicies == nil {
		s.signalPolicies = map[string][]SignalPolicy{}
	}
	tid := "default"
	if _, ok := s.signalPolicies[tid]; !ok {
		s.signalPolicies[tid] = defaultSignalPolicies()
	}
	if s.serviceCatalog == nil {
		s.serviceCatalog = map[string]ServiceCatalogEntry{}
	}
	for _, e := range defaultServiceCatalog() {
		s.serviceCatalog[e.Service] = e
	}
	if s.logAnomalies == nil {
		s.logAnomalies = map[string]LogAnomaly{}
	}
	for _, a := range defaultLogAnomalies() {
		s.logAnomalies[a.ID] = a
	}
	if s.apmReleases == nil {
		s.apmReleases = map[string]APMRelease{}
	}
	if s.maintenanceWindows == nil {
		s.maintenanceWindows = map[string]MaintenanceWindow{}
	}
	if s.cardinalityPolicies == nil {
		s.cardinalityPolicies = map[string]CardinalityAlertPolicy{}
	}
	if s.fatigueConfigs == nil {
		s.fatigueConfigs = map[string]AlertFatigueConfig{}
	}
	if s.siemConfigs == nil {
		s.siemConfigs = map[string]SIEMConfig{}
	}
	if s.discoveredServices == nil {
		s.discoveredServices = defaultDiscoveredServices()
	}
	if s.threatFeed == nil {
		s.threatFeed = defaultThreatFeed()
	}
}

func defaultSignalPolicies() []SignalPolicy {
	return []SignalPolicy{
		{Signal: "logs", RetentionDays: 90, MaskingEnabled: true, ResidencyRegion: "in-mumbai", PIIFields: []string{"pan", "accountNumber", "email"}},
		{Signal: "traces", RetentionDays: 14, MaskingEnabled: true, ResidencyRegion: "in-mumbai"},
		{Signal: "metrics", RetentionDays: 400, MaskingEnabled: false, ResidencyRegion: "in-mumbai"},
		{Signal: "events", RetentionDays: 30, MaskingEnabled: true, ResidencyRegion: "in-mumbai"},
	}
}

func defaultServiceCatalog() []ServiceCatalogEntry {
	return []ServiceCatalogEntry{
		{Service: "payment-service", DisplayName: "Payments API", OwnerTeam: "payments-platform", OwnerEmail: "payments-oncall@bank.example", Tier: "P0", RunbookURL: "/docs/RUNBOOK.md#payment-service"},
		{Service: "ledger-service", DisplayName: "Core Ledger", OwnerTeam: "ledger", OwnerEmail: "ledger-oncall@bank.example", Tier: "P0", Dependencies: []string{"payment-service"}},
		{Service: "api-gateway", DisplayName: "API Gateway", OwnerTeam: "platform", OwnerEmail: "platform-oncall@bank.example", Tier: "P1"},
	}
}

func defaultLogAnomalies() []LogAnomaly {
	now := time.Now().UTC()
	return []LogAnomaly{
		{ID: "la-1", Service: "ledger-service", Pattern: `ERROR.*timeout`, Score: 0.91, Precision: 0.87, Recall: 0.82, Message: "Spike in ledger timeout errors", DetectedAt: now.Add(-2 * time.Hour)},
		{ID: "la-2", Service: "payment-service", Pattern: `WARN.*retry`, Score: 0.76, Precision: 0.80, Recall: 0.75, Message: "Elevated payment retry warnings", DetectedAt: now.Add(-45 * time.Minute)},
	}
}

func defaultDiscoveredServices() []DiscoveredService {
	now := time.Now().UTC()
	return []DiscoveredService{
		{AgentID: "agent-1", ServiceName: "payment-service", Language: "Java", Port: 8080, DiscoveredAt: now.Add(-24 * time.Hour)},
		{AgentID: "agent-1", ServiceName: "ledger-service", Language: "Go", Port: 8081, DiscoveredAt: now.Add(-24 * time.Hour)},
		{AgentID: "agent-2", ServiceName: "api-gateway", Language: "Go", Port: 8443, DiscoveredAt: now.Add(-12 * time.Hour)},
	}
}

func defaultThreatFeed() []ThreatIndicator {
	now := time.Now().UTC()
	return []ThreatIndicator{
		{ID: "th-1", Type: "ip", Value: "203.0.113.44", Severity: "high", Source: "internal-correlation", Description: "Brute-force auth attempts", ObservedAt: now.Add(-3 * time.Hour)},
		{ID: "th-2", Type: "domain", Value: "evil-phish.example", Severity: "medium", Source: "threat-intel-feed", ObservedAt: now.Add(-6 * time.Hour)},
	}
}

func defaultSupportedPlatforms() []SupportedPlatform {
	return []SupportedPlatform{
		{OS: "Linux", Arch: "amd64", KernelMin: "5.4", Status: "certified", AgentVersion: "1.4.0", EBPFSupported: true, Features: []string{"host", "ebpf-rtt", "otlp"}},
		{OS: "Linux", Arch: "arm64", KernelMin: "5.10", Status: "certified", AgentVersion: "1.4.0", EBPFSupported: true, Features: []string{"host", "otlp"}},
		{OS: "Windows Server", Arch: "amd64", Status: "preview", AgentVersion: "1.4.0", EBPFSupported: false, Notes: "eBPF not available", Features: []string{"host", "otlp"}},
		{OS: "RHEL", Arch: "amd64", KernelMin: "4.18", Status: "certified", AgentVersion: "1.4.0", EBPFSupported: true, Features: []string{"host", "ebpf-rtt"}},
	}
}

func defaultLogIngestFormats() []LogIngestFormat {
	return []LogIngestFormat{
		{Format: "json", Endpoint: "POST /api/v1/logs", ContentType: "application/json", Status: "ga", Features: []string{"structured", "traceId", "txnId"}},
		{Format: "syslog", Endpoint: "TCP/UDP :514", Status: "ga", Notes: "RFC5424 via ingestion connector", Features: []string{"cef", "structured-data"}},
		{Format: "otel-logs", Endpoint: "POST /api/v1/otlp/v1/logs", ContentType: "application/x-protobuf", Status: "ga", Features: []string{"resource-attrs", "trace-context"}},
		{Format: "ndjson", Endpoint: "POST /api/v1/logs/bulk", ContentType: "application/x-ndjson", Status: "preview"},
	}
}

func defaultCloudMetricCatalog() []CloudMetricCatalogEntry {
	return []CloudMetricCatalogEntry{
		{Provider: "aws", Namespace: "AWS/EC2", MetricName: "CPUUtilization", Unit: "Percent", Description: "EC2 instance CPU"},
		{Provider: "aws", Namespace: "AWS/RDS", MetricName: "DatabaseConnections", Unit: "Count"},
		{Provider: "gcp", Namespace: "compute.googleapis.com", MetricName: "instance/cpu/utilization", Unit: "Percent"},
		{Provider: "azure", Namespace: "Microsoft.Compute/virtualMachines", MetricName: "Percentage CPU", Unit: "Percent"},
	}
}

func (s *Store) ListDiscoveredServices() []DiscoveredService {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.discoveredServices == nil {
		return defaultDiscoveredServices()
	}
	out := make([]DiscoveredService, len(s.discoveredServices))
	copy(out, s.discoveredServices)
	return out
}

func (s *Store) SupportedPlatforms() []SupportedPlatform {
	return defaultSupportedPlatforms()
}

func (s *Store) CollectorSpoolStatus(agentID string) (CollectorSpoolStatus, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.collectorFleet[agentID]
	if !ok {
		return CollectorSpoolStatus{}, false
	}
	_ = a
	return CollectorSpoolStatus{
		AgentID: agentID, PendingBatches: 2, OldestAgeSec: 45, BytesOnDisk: 12_582_912,
		MaxBytes: 1_073_741_824, ReplayLagSec: 3, Healthy: true, CheckedAt: time.Now().UTC(),
	}, true
}

func (s *Store) CollectorBenchmark(agentID string) (CollectorBenchmarkResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.collectorFleet[agentID]; !ok {
		return CollectorBenchmarkResult{}, false
	}
	cpu, mem := 0.8, 198.0
	return CollectorBenchmarkResult{
		AgentID: agentID, CPUPercent: cpu, MemoryMB: mem,
		TargetCPU: 1.0, TargetMemory: 256, WithinTarget: cpu <= 1.0 && mem <= 256,
		MeasuredAt: time.Now().UTC(),
	}, true
}

func (s *Store) CollectorVersionDrift(target string) CollectorVersionDrift {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if target == "" {
		target = "1.4.0"
	}
	out := CollectorVersionDrift{TargetVersion: target}
	for _, a := range s.collectorFleet {
		if a.Status == "healthy" {
			out.HealthyCount++
		}
		if a.Version != "" && a.Version != target {
			out.Drifted = append(out.Drifted, struct {
				AgentID string `json:"agentId"`
				Name    string `json:"name"`
				Version string `json:"version"`
			}{AgentID: a.ID, Name: a.Name, Version: a.Version})
			out.DriftCount++
		}
	}
	return out
}

func (s *Store) ApplyCollectorRollout(agentID string, req CollectorRolloutRequest) (CollectorFleetAgent, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.collectorFleet[agentID]
	if !ok {
		return CollectorFleetAgent{}, false
	}
	if req.TargetVersion != "" {
		if req.Strategy == "canary" && req.CanaryPercent < 100 {
			a.Status = "canary"
		} else {
			a.Version = req.TargetVersion
			a.Status = "upgrading"
		}
	}
	a.LastHeartbeatAt = time.Now().UTC()
	s.collectorFleet[agentID] = a
	return a, true
}

func (s *Store) GetCardinalityAlertPolicy(tenantID string) CardinalityAlertPolicy {
	s.mu.RLock()
	defer s.mu.Unlock()
	if p, ok := s.cardinalityPolicies[tenantID]; ok {
		return p
	}
	return CardinalityAlertPolicy{Enabled: true, ThresholdServices: 256, NotifyChannel: "slack:#platform-alerts"}
}

func (s *Store) SaveCardinalityAlertPolicy(tenantID string, p CardinalityAlertPolicy) CardinalityAlertPolicy {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cardinalityPolicies == nil {
		s.cardinalityPolicies = map[string]CardinalityAlertPolicy{}
	}
	s.cardinalityPolicies[tenantID] = p
	return p
}

func (s *Store) ListMaintenanceWindows() []MaintenanceWindow {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]MaintenanceWindow, 0, len(s.maintenanceWindows))
	for _, w := range s.maintenanceWindows {
		out = append(out, w)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartsAt.After(out[j].StartsAt) })
	return out
}

func (s *Store) SaveMaintenanceWindow(w MaintenanceWindow) MaintenanceWindow {
	s.mu.Lock()
	defer s.mu.Unlock()
	if w.ID == "" {
		w.ID = "mw-" + uuid.New().String()[:8]
	}
	if s.maintenanceWindows == nil {
		s.maintenanceWindows = map[string]MaintenanceWindow{}
	}
	s.maintenanceWindows[w.ID] = w
	return w
}

func (s *Store) GetFatigueConfig(policyID string) AlertFatigueConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if c, ok := s.fatigueConfigs[policyID]; ok {
		return c
	}
	return AlertFatigueConfig{PolicyID: policyID, AutoSuppressScore: 0.85, WeightDuplicate: 0.4, WeightFrequency: 0.6}
}

func (s *Store) SaveFatigueConfig(c AlertFatigueConfig) AlertFatigueConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.fatigueConfigs == nil {
		s.fatigueConfigs = map[string]AlertFatigueConfig{}
	}
	s.fatigueConfigs[c.PolicyID] = c
	return c
}

func (s *Store) ListAPMErrorGroups(service string) []APMErrorGroup {
	now := time.Now().UTC()
	all := []APMErrorGroup{
		{ID: "err-1", Service: "payment-service", Fingerprint: "NullPointerException:PaymentValidator", Message: "NullPointerException in PaymentValidator", Count: 42, Release: "v2.14.1", FirstSeen: now.Add(-6 * time.Hour), LastSeen: now.Add(-5 * time.Minute), TraceID: "trace-demo-03"},
		{ID: "err-2", Service: "ledger-service", Fingerprint: "timeout:PostgresQuery", Message: "context deadline exceeded on ledger query", Count: 18, Release: "v1.8.0", FirstSeen: now.Add(-2 * time.Hour), LastSeen: now.Add(-12 * time.Minute), TraceID: "trace-demo-07"},
	}
	if service == "" {
		return all
	}
	out := make([]APMErrorGroup, 0)
	for _, e := range all {
		if strings.EqualFold(e.Service, service) {
			out = append(out, e)
		}
	}
	return out
}

func (s *Store) ListAPMReleases(service string) []APMRelease {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]APMRelease, 0, len(s.apmReleases))
	for _, r := range s.apmReleases {
		if service == "" || strings.EqualFold(r.Service, service) {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DeployedAt.After(out[j].DeployedAt) })
	if len(out) == 0 {
		now := time.Now().UTC()
		return []APMRelease{
			{ID: "rel-1", Service: "payment-service", Version: "v2.14.1", CommitSHA: "abc1234", DeployedAt: now.Add(-24 * time.Hour), Environment: "production"},
			{ID: "rel-2", Service: "ledger-service", Version: "v1.8.0", CommitSHA: "def5678", DeployedAt: now.Add(-48 * time.Hour), Environment: "production"},
		}
	}
	return out
}

func (s *Store) RecordAPMDeployment(r APMRelease) APMRelease {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.apmReleases == nil {
		s.apmReleases = map[string]APMRelease{}
	}
	if r.ID == "" {
		r.ID = "rel-" + uuid.New().String()[:8]
	}
	if r.DeployedAt.IsZero() {
		r.DeployedAt = time.Now().UTC()
	}
	s.apmReleases[r.ID] = r
	return r
}

func (s *Store) ListServiceCatalog() []ServiceCatalogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ServiceCatalogEntry, 0, len(s.serviceCatalog))
	for _, e := range s.serviceCatalog {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Service < out[j].Service })
	return out
}

func (s *Store) SaveServiceCatalogEntry(e ServiceCatalogEntry) ServiceCatalogEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.serviceCatalog == nil {
		s.serviceCatalog = map[string]ServiceCatalogEntry{}
	}
	s.serviceCatalog[e.Service] = e
	return e
}

func (s *Store) BusinessTransaction(txnID string) (BusinessTransactionView, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	traces := make([]string, 0)
	var spans []Span
	for id, t := range s.traces {
		for _, sp := range t.Spans {
			if strings.Contains(strings.ToLower(sp.Operation), "payment") || id == "trace-demo-03" {
				traces = append(traces, id)
				spans = append(spans, sp)
				break
			}
		}
	}
	if txnID == "" {
		return BusinessTransactionView{}, false
	}
	return BusinessTransactionView{
		TxnID: txnID, Service: "payment-service", Status: "completed",
		TraceIDs: traces, LogCount: 12, DurationMs: 245.5,
		Spans: spans, StartedAt: time.Now().UTC().Add(-5 * time.Minute),
	}, true
}

func (s *Store) LogIngestFormats() []LogIngestFormat {
	return defaultLogIngestFormats()
}

func (s *Store) LogCorrelation(logID string) LogCorrelation {
	traceID := "trace-demo-03"
	if logID == "" {
		logID = "log-" + uuid.New().String()[:8]
	}
	return LogCorrelation{
		LogID: logID, TraceID: traceID, SpanID: "span-root", Service: "payment-service",
		TxnID: "txn-upi-" + logID[len(logID)-4:], TraceLink: "/traces/" + traceID,
		Related: []string{"log-related-1", "log-related-2"},
	}
}

func (s *Store) ListLogAnomalies() []LogAnomaly {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]LogAnomaly, 0, len(s.logAnomalies))
	for _, a := range s.logAnomalies {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DetectedAt.After(out[j].DetectedAt) })
	return out
}

func (s *Store) FeedbackLogAnomaly(id, feedback string) (LogAnomaly, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.logAnomalies[id]
	if !ok {
		return LogAnomaly{}, false
	}
	a.Feedback = feedback
	if feedback == "useful" {
		a.Precision = min(1.0, a.Precision+0.02)
	}
	s.logAnomalies[id] = a
	return a, true
}

func (s *Store) SignalPolicies(tenantID string) []SignalPolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if p, ok := s.signalPolicies[tenantID]; ok {
		out := make([]SignalPolicy, len(p))
		copy(out, p)
		return out
	}
	return defaultSignalPolicies()
}

func (s *Store) SaveSignalPolicies(tenantID string, policies []SignalPolicy) []SignalPolicy {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.signalPolicies == nil {
		s.signalPolicies = map[string][]SignalPolicy{}
	}
	s.signalPolicies[tenantID] = policies
	return policies
}

func (s *Store) SIEMConfig(tenantID string) SIEMConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if c, ok := s.siemConfigs[tenantID]; ok {
		return c
	}
	return SIEMConfig{Provider: "splunk", Endpoint: "https://splunk.example:8088/services/collector", Index: "neuralops", FailClosed: true}
}

func (s *Store) SaveSIEMConfig(tenantID string, c SIEMConfig) SIEMConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.siemConfigs == nil {
		s.siemConfigs = map[string]SIEMConfig{}
	}
	s.siemConfigs[tenantID] = c
	return c
}

func (s *Store) ThreatFeed() []ThreatIndicator {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.threatFeed == nil {
		return defaultThreatFeed()
	}
	out := make([]ThreatIndicator, len(s.threatFeed))
	copy(out, s.threatFeed)
	return out
}

func (s *Store) CloudMetricCatalog(provider string) []CloudMetricCatalogEntry {
	all := defaultCloudMetricCatalog()
	if provider == "" {
		return all
	}
	out := make([]CloudMetricCatalogEntry, 0)
	for _, e := range all {
		if strings.EqualFold(e.Provider, provider) {
			out = append(out, e)
		}
	}
	return out
}

func (s *Store) GetExportJob(id string) (ExportJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, ok := s.exportJobs[id]
	return j, ok
}

func (s *Store) RecordDevOpsDeploy(provider, service, version, commit string) DevOpsWebhookEvent {
	ev := DevOpsWebhookEvent{
		Provider: provider, Service: service, Version: version, CommitSHA: commit,
		Status: "success", ReceivedAt: time.Now().UTC(),
	}
	s.RecordAPMDeployment(APMRelease{
		Service: service, Version: version, CommitSHA: commit,
		DeployedAt: ev.ReceivedAt, Environment: "production",
	})
	return ev
}

func (s *Store) IntegrationsWithDevOps() []Integration {
	return []Integration{
		{ID: "int-1", Name: "Slack", Type: "slack", Status: "connected", Connected: true},
		{ID: "int-2", Name: "Jira", Type: "jira", Status: "disconnected", Connected: false},
		{ID: "int-3", Name: "PagerDuty", Type: "pagerduty", Status: "connected", Connected: true},
		{ID: "int-4", Name: "GitHub", Type: "github", Status: "disconnected", Connected: false},
		{ID: "int-5", Name: "GitLab", Type: "gitlab", Status: "disconnected", Connected: false},
		{ID: "int-6", Name: "Jenkins", Type: "jenkins", Status: "disconnected", Connected: false},
		{ID: "int-7", Name: "ServiceNow", Type: "servicenow", Status: "disconnected", Connected: false},
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func (s *Store) seedDerivedMetricsCatalog() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.derivedMetrics) > 0 {
		return
	}
	now := time.Now().UTC()
	catalog := []DerivedMetric{
		{ID: "dm-error-rate", Name: "service.error_rate", Expression: "errors/total", Unit: "ratio", UpdatedAt: now},
		{ID: "dm-p95-latency", Name: "service.latency_p95", Expression: "histogram_quantile(0.95, latency)", Unit: "ms", UpdatedAt: now},
		{ID: "dm-ingest-lag", Name: "pipeline.ingest_lag_ms", Expression: "max(ingest_lag)", Unit: "ms", UpdatedAt: now},
		{ID: "dm-slo-burn", Name: "slo.burn_rate", Expression: "error_budget_burn", Unit: "ratio", UpdatedAt: now},
		{ID: "dm-cardinality", Name: "nexql.cardinality", Expression: "count(distinct service)", Unit: "count", UpdatedAt: now},
	}
	for _, m := range catalog {
		m.Value = EvaluateDerivedMetric(m.Expression)
		s.derivedMetrics[m.ID] = m
	}
}

func (s *Store) IngestThreatIndicators(items []ThreatIndicator) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.threatFeed == nil {
		s.threatFeed = defaultThreatFeed()
	}
	accepted := 0
	for _, ind := range items {
		if ind.ID == "" {
			ind.ID = "th-" + uuid.New().String()[:8]
		}
		if ind.ObservedAt.IsZero() {
			ind.ObservedAt = time.Now().UTC()
		}
		s.threatFeed = append(s.threatFeed, ind)
		accepted++
	}
	return map[string]any{"accepted": accepted, "total": len(s.threatFeed)}
}

func (s *Store) IngestNetFlowRecords(records []NetFlowRecord) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	for i := range records {
		if records[i].ID == "" {
			records[i].ID = "nf-" + uuid.New().String()[:8]
		}
		if records[i].Timestamp.IsZero() {
			records[i].Timestamp = now
		}
		s.netflowRecords = append(s.netflowRecords, records[i])
	}
	return map[string]any{"accepted": len(records), "total": len(s.netflowRecords)}
}

func (s *Store) ListNetFlowRecordsStored(limit int) []NetFlowRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.netflowRecords) == 0 {
		return nil
	}
	if limit <= 0 || limit > len(s.netflowRecords) {
		limit = len(s.netflowRecords)
	}
	out := make([]NetFlowRecord, limit)
	copy(out, s.netflowRecords[len(s.netflowRecords)-limit:])
	return out
}

func (s *Store) ListExportJobs() []ExportJob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ExportJob, 0, len(s.exportJobs))
	for _, j := range s.exportJobs {
		out = append(out, j)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartedAt.After(out[j].StartedAt) })
	return out
}

func (s *Store) DBMStatements(dbID string) []DBMStatement {
	_ = dbID
	return []DBMStatement{
		{
			Query: "UPDATE accounts SET balance = $1 WHERE id = $2", Calls: 12400, AvgMs: 8.2, P95Ms: 22.4,
			TotalMs: 101680, TraceID: "trace-demo-07", Service: "ledger-service", Database: "payments-primary",
		},
		{
			Query: "SELECT * FROM transactions WHERE txn_id = $1", Calls: 8900, AvgMs: 3.1, P95Ms: 9.8,
			TotalMs: 27590, TraceID: "trace-demo-03", Service: "payment-service", Database: "payments-primary",
		},
		{
			Query: "INSERT INTO audit_log (event, payload) VALUES ($1, $2)", Calls: 4200, AvgMs: 5.6, P95Ms: 14.1,
			TotalMs: 23520, Service: "ledger-service", Database: "ledger-replica",
		},
	}
}

func (s *Store) EBPFFlowSamples(host string) []EBPFFlowSample {
	now := time.Now().UTC()
	all := []EBPFFlowSample{
		{ID: "ebpf-1", SrcIP: "10.0.1.5", DstIP: "10.0.2.8", SrcPort: 44321, DstPort: 5432, Protocol: "TCP", RTTMs: 1.2, Bytes: 4096, Host: "nexagent-1", Timestamp: now},
		{ID: "ebpf-2", SrcIP: "10.0.1.5", DstIP: "10.0.3.2", SrcPort: 51002, DstPort: 8080, Protocol: "TCP", RTTMs: 0.8, Bytes: 8192, Host: "nexagent-1", Timestamp: now.Add(-30 * time.Second)},
		{ID: "ebpf-3", SrcIP: "10.0.2.8", DstIP: "10.0.4.1", SrcPort: 33001, DstPort: 9092, Protocol: "TCP", RTTMs: 2.1, Bytes: 2048, Host: "nexagent-2", Timestamp: now.Add(-1 * time.Minute)},
	}
	if host == "" {
		return all
	}
	out := make([]EBPFFlowSample, 0)
	for _, f := range all {
		if f.Host == host {
			out = append(out, f)
		}
	}
	return out
}

