package observability

import (
	"time"

	"github.com/neuralops/platform/internal/apm"
)

// Trace APM types (shared with tracequery / correlation).
type (
	Span               = apm.Span
	TraceDetail        = apm.TraceDetail
	TraceSearchRequest = apm.TraceSearchRequest
	TraceSummary       = apm.TraceSummary
	FlowEdge           = apm.FlowEdge
)

// MetricSeriesPoint is one time-series sample.
type MetricSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// MetricSeries is a named metric time series.
type MetricSeries struct {
	Name   string              `json:"name"`
	Labels map[string]string   `json:"labels,omitempty"`
	Unit   string              `json:"unit,omitempty"`
	Points []MetricSeriesPoint `json:"points"`
}

// MetricCatalogEntry describes an available metric.
type MetricCatalogEntry struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Unit        string   `json:"unit"`
	Labels      []string `json:"labels"`
}

// TopologyNode is a node in the live service graph.
type TopologyNode struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"displayName"`
	Type        string  `json:"type"`
	Health      string  `json:"health"`
	ErrorRate   float64 `json:"errorRate"`
	Throughput  float64 `json:"throughputRpm"`
	Zone        string  `json:"zone,omitempty"`
}

// TopologyEdge connects topology nodes.
type TopologyEdge struct {
	Source    string  `json:"source"`
	Target    string  `json:"target"`
	CallCount int64   `json:"callCount"`
	ErrorRate float64 `json:"errorRate"`
	P95Ms     float64 `json:"p95Ms"`
}

// TopologyGraph is the live Smartscape graph payload.
type TopologyGraph struct {
	Nodes []TopologyNode `json:"nodes"`
	Edges []TopologyEdge `json:"edges"`
	At    time.Time      `json:"at"`
}

// DashboardTile is one tile in a custom dashboard.
type DashboardTile struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Title    string         `json:"title"`
	Metric   string         `json:"metric,omitempty"`
	Query    string         `json:"query,omitempty"`
	Position map[string]int `json:"position"`
}

// Dashboard is a user-defined dashboard.
type Dashboard struct {
	ID          string          `json:"id"`
	TenantID    string          `json:"tenantId"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Shared      bool            `json:"shared"`
	Tiles       []DashboardTile `json:"tiles"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

// LogMetricRule counts log pattern occurrences.
type LogMetricRule struct {
	ID       string `json:"id"`
	TenantID string `json:"tenantId"`
	Name     string `json:"name"`
	Pattern  string `json:"pattern"`
	Service  string `json:"service,omitempty"`
	Enabled  bool   `json:"enabled"`
}

// LogParsingRule defines log parsing configuration.
type LogParsingRule struct {
	ID       string `json:"id"`
	TenantID string `json:"tenantId"`
	Name     string `json:"name"`
	Pattern  string `json:"pattern"`
	Field    string `json:"field"`
	Enabled  bool   `json:"enabled"`
}

// SLO defines a service level objective.
type SLO struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenantId"`
	Name        string    `json:"name"`
	Service     string    `json:"service"`
	SLIQuery    string    `json:"sliQuery"`
	Target      float64   `json:"target"`
	WindowDays  int       `json:"windowDays"`
	ErrorBudget float64   `json:"errorBudget"`
	BurnRate    float64   `json:"burnRate"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

// EntityAnomaly is an entity-scoped anomaly record.
type EntityAnomaly struct {
	ID         string    `json:"id"`
	EntityID   string    `json:"entityId"`
	EntityType string    `json:"entityType"`
	Service    string    `json:"service"`
	Metric     string    `json:"metric"`
	Score      float64   `json:"score"`
	Message    string    `json:"message"`
	DetectedAt time.Time `json:"detectedAt"`
}

// HostSummary is an infrastructure host list item.
type HostSummary struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Status string  `json:"status"`
	CPU    float64 `json:"cpuPercent"`
	Memory float64 `json:"memoryPercent"`
	Disk   float64 `json:"diskPercent"`
	Zone   string  `json:"zone"`
}

// K8sCluster is a kubernetes cluster summary.
type K8sCluster struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Nodes          int    `json:"nodes"`
	Pods           int    `json:"pods"`
	Health         string `json:"health"`
	NamespaceCount int    `json:"namespaceCount"`
}

// K8sPod is a pod list item.
type K8sPod struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Namespace string  `json:"namespace"`
	Node      string  `json:"node"`
	Status    string  `json:"status"`
	CPU       float64 `json:"cpuPercent"`
	Memory    float64 `json:"memoryPercent"`
	Restarts  int     `json:"restarts"`
}

// DatabaseInstance is a monitored database.
type DatabaseInstance struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Engine      string  `json:"engine"`
	Status      string  `json:"status"`
	QPS         float64 `json:"qps"`
	SlowQueries int     `json:"slowQueries"`
	Connections int     `json:"connections"`
}

// DBStatement is a top database statement.
type DBStatement struct {
	Query   string  `json:"query"`
	Calls   int64   `json:"calls"`
	AvgMs   float64 `json:"avgMs"`
	TotalMs float64 `json:"totalMs"`
}

// KafkaLag is consumer group lag.
type KafkaLag struct {
	Topic         string `json:"topic"`
	ConsumerGroup string `json:"consumerGroup"`
	Lag           int64  `json:"lag"`
	Partition     int    `json:"partition"`
}

// RUMSession is a real-user session summary.
type RUMSession struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId"`
	Page       string    `json:"page"`
	Device     string    `json:"device"`
	Country    string    `json:"country"`
	DurationMs int64     `json:"durationMs"`
	Errors     int       `json:"errors"`
	LCP        float64   `json:"lcp"`
	StartedAt  time.Time `json:"startedAt"`
}

// SyntheticMonitor is an HTTP/browser synthetic check.
type SyntheticMonitor struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	URL        string    `json:"url"`
	Interval   string    `json:"interval"`
	Locations  []string  `json:"locations"`
	Enabled    bool      `json:"enabled"`
	LastStatus string    `json:"lastStatus"`
	LastRunAt  time.Time `json:"lastRunAt"`
}

// SyntheticRun is one synthetic execution result.
type SyntheticRun struct {
	ID        string    `json:"id"`
	MonitorID string    `json:"monitorId"`
	Status    string    `json:"status"`
	LatencyMs int64     `json:"latencyMs"`
	Location  string    `json:"location"`
	RanAt     time.Time `json:"ranAt"`
}

// WorkflowNode is a single step node in the workflow graph.
type WorkflowNode struct {
	ID    string  `json:"id"`
	Type  string  `json:"type"`
	Label string  `json:"label"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
}

// WorkflowEdge connects two workflow nodes (source -> target). Condition gates
// traversal at runtime: "" / "success" (default) follow on predecessor success,
// "failure" on predecessor failure, "always" regardless, or "key=value" to match
// the trigger context.
type WorkflowEdge struct {
	ID        string `json:"id"`
	Source    string `json:"source"`
	Target    string `json:"target"`
	Condition string `json:"condition,omitempty"`
}

// WorkflowGraph is the full branching topology authored in the visual editor.
type WorkflowGraph struct {
	Nodes []WorkflowNode `json:"nodes"`
	Edges []WorkflowEdge `json:"edges"`
}

// Workflow defines an automation workflow.
type Workflow struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Trigger   string         `json:"trigger"`
	Enabled   bool           `json:"enabled"`
	Steps     []string       `json:"steps"`
	Graph     *WorkflowGraph `json:"graph,omitempty"`
	LastRunAt *time.Time     `json:"lastRunAt,omitempty"`
}

// LinearizeGraph returns step labels in execution order via a topological sort
// (Kahn's algorithm) of the graph edges. Disconnected or cyclic nodes are
// appended in their declared order so no step is silently dropped. This keeps
// the executor's linear `steps` consistent with the authored branching graph.
func LinearizeGraph(g *WorkflowGraph) []string {
	if g == nil || len(g.Nodes) == 0 {
		return nil
	}
	byID := make(map[string]WorkflowNode, len(g.Nodes))
	indegree := make(map[string]int, len(g.Nodes))
	for _, n := range g.Nodes {
		byID[n.ID] = n
		indegree[n.ID] = 0
	}
	adjacency := make(map[string][]string)
	for _, e := range g.Edges {
		if _, ok := byID[e.Source]; !ok {
			continue
		}
		if _, ok := byID[e.Target]; !ok {
			continue
		}
		adjacency[e.Source] = append(adjacency[e.Source], e.Target)
		indegree[e.Target]++
	}
	queue := make([]string, 0)
	for _, n := range g.Nodes {
		if indegree[n.ID] == 0 {
			queue = append(queue, n.ID)
		}
	}
	visited := make(map[string]bool, len(g.Nodes))
	labels := make([]string, 0, len(g.Nodes))
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if visited[id] {
			continue
		}
		visited[id] = true
		labels = append(labels, byID[id].Label)
		for _, target := range adjacency[id] {
			indegree[target]--
			if indegree[target] <= 0 && !visited[target] {
				queue = append(queue, target)
			}
		}
	}
	for _, n := range g.Nodes {
		if !visited[n.ID] {
			labels = append(labels, n.Label)
		}
	}
	return labels
}

// Notebook is a saved analysis notebook.
type Notebook struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Cells     []NotebookCell `json:"cells"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

// NotebookCell is one notebook cell.
type NotebookCell struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

// SecurityVulnerability is a runtime vulnerability finding.
type SecurityVulnerability struct {
	ID          string    `json:"id"`
	CVE         string    `json:"cve"`
	Severity    string    `json:"severity"`
	Service     string    `json:"service"`
	Description string    `json:"description"`
	DetectedAt  time.Time `json:"detectedAt"`
}

// SecurityAttack is an attack event.
type SecurityAttack struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	SourceIP   string    `json:"sourceIp"`
	Service    string    `json:"service"`
	Blocked    bool      `json:"blocked"`
	DetectedAt time.Time `json:"detectedAt"`
}

// SecurityFinding is a normalized AppSec/SecOps finding record.
type SecurityFinding struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Category       string    `json:"category"` // rasp|sca|image|cspm
	Severity       string    `json:"severity"`
	Service        string    `json:"service,omitempty"`
	Asset          string    `json:"asset,omitempty"`
	Status         string    `json:"status"`
	Exploitability string    `json:"exploitability,omitempty"`
	IncidentID     string    `json:"incidentId,omitempty"`
	DetectedAt     time.Time `json:"detectedAt"`
}

// SecurityPostureCheck is one CSPM policy check summary.
type SecurityPostureCheck struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Resource string `json:"resource"`
	Status   string `json:"status"` // pass|fail|warn
	Severity string `json:"severity"`
}

// Integration describes a third-party integration.
type Integration struct {
	ID             string `json:"id"`
	IntegrationKey string `json:"integrationKey,omitempty"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	Status         string `json:"status"`
	Connected      bool   `json:"connected"`
}

// AdminUser is a tenant user for admin UI.
type AdminUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Active   bool   `json:"active"`
	TenantID string `json:"tenantId"`
}

// APIKey is an API key record (never returns raw secret on list).
type APIKey struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Prefix    string     `json:"prefix"`
	Scopes    []string   `json:"scopes"`
	CreatedAt time.Time  `json:"createdAt"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

// AuditEntry is an audit log row.
type AuditEntry struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Detail    string    `json:"detail"`
	Timestamp time.Time `json:"timestamp"`
}

// UsageStats is tenant usage summary.
type UsageStats struct {
	LogsIngestedGB  float64 `json:"logsIngestedGb"`
	TracesIngested  int64   `json:"tracesIngested"`
	MetricsIngested int64   `json:"metricsIngested"`
	AITokensUsed    int64   `json:"aiTokensUsed"`
	ActiveUsers     int     `json:"activeUsers"`
}

// UnifiedQueryRequest is a cross-signal query request.
type UnifiedQueryRequest struct {
	Query   string `json:"query"`
	From    string `json:"from,omitempty"` // signal source: logs|metrics|traces|events|all
	Service string `json:"service,omitempty"`
	TraceID string `json:"traceId,omitempty"`
	TxnID   string `json:"txnId,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

// UnifiedQueryHit is one item in a cross-signal result set.
type UnifiedQueryHit struct {
	ID        string                 `json:"id"`
	Signal    string                 `json:"signal"`
	Service   string                 `json:"service,omitempty"`
	Title     string                 `json:"title"`
	Summary   string                 `json:"summary"`
	Severity  string                 `json:"severity,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Link      string                 `json:"link,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// UnifiedQueryResponse bundles results and execution metadata.
type UnifiedQueryResponse struct {
	Hits        []UnifiedQueryHit  `json:"hits"`
	Count       int                `json:"count"`
	Planner     []QueryPlannerStep `json:"planner,omitempty"`
	Cardinality CardinalityGuard   `json:"cardinality,omitempty"`
}

// AlertPolicyRoute defines one route destination in an alert policy.
type AlertPolicyRoute struct {
	Channel  string `json:"channel"`
	Target   string `json:"target"`
	After    string `json:"after"`
	Priority int    `json:"priority"`
}

// AlertPolicy defines alert routing/suppression policy controls.
type AlertPolicy struct {
	ID             string             `json:"id"`
	Name           string             `json:"name"`
	ServicePattern string             `json:"servicePattern"`
	Severity       string             `json:"severity"`
	Enabled        bool               `json:"enabled"`
	Expression     string             `json:"expression,omitempty"`
	DedupeKey      string             `json:"dedupeKey,omitempty"`
	Routes         []AlertPolicyRoute `json:"routes"`
	Context        AlertContext       `json:"context,omitempty"`
}

// AlertSuppression defines a temporary suppression window.
type AlertSuppression struct {
	ID             string    `json:"id"`
	ServicePattern string    `json:"servicePattern"`
	Reason         string    `json:"reason"`
	StartsAt       time.Time `json:"startsAt"`
	EndsAt         time.Time `json:"endsAt"`
	CreatedBy      string    `json:"createdBy"`
}

// CollectorFleetAgent is an enrolled telemetry collector/agent instance.
type CollectorFleetAgent struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Environment     string    `json:"environment"`
	Version         string    `json:"version"`
	Status          string    `json:"status"`
	LastHeartbeatAt time.Time `json:"lastHeartbeatAt"`
	PolicyID        string    `json:"policyId,omitempty"`
}

// CollectorPipelineStage is one processing stage in a collector pipeline.
type CollectorPipelineStage struct {
	ID      string            `json:"id"`
	Type    string            `json:"type"` // parse|enrich|drop|mask|route
	Config  map[string]string `json:"config,omitempty"`
	Enabled bool              `json:"enabled"`
}

// CollectorPipeline defines a visual data processing pipeline.
type CollectorPipeline struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	Enabled     bool                     `json:"enabled"`
	Stages      []CollectorPipelineStage `json:"stages"`
	UpdatedAt   time.Time                `json:"updatedAt"`
}

// AIExplanationEvidence links a causal node to observability evidence.
type AIExplanationEvidence struct {
	Signal string `json:"signal"`
	Ref    string `json:"ref"`
	Detail string `json:"detail"`
}

// AIExplanationNode is one node in an explainable RCA tree.
type AIExplanationNode struct {
	ID         string                  `json:"id"`
	Label      string                  `json:"label"`
	Confidence float64                 `json:"confidence"`
	Evidence   []AIExplanationEvidence `json:"evidence,omitempty"`
	Children   []AIExplanationNode     `json:"children,omitempty"`
}

// RCAResponse is causal root-cause analysis for an incident.
type RCAResponse struct {
	ExplanationID string            `json:"explanationId"`
	IncidentID    string            `json:"incidentId"`
	Summary       string            `json:"summary"`
	Confidence    float64           `json:"confidence"`
	Root          AIExplanationNode `json:"root"`
	GeneratedAt   time.Time         `json:"generatedAt"`
}

// AIForecast is a predictive capacity/SLO/incident-risk projection.
type AIForecast struct {
	ID             string    `json:"id"`
	Metric         string    `json:"metric"`
	Service        string    `json:"service,omitempty"`
	Horizon        string    `json:"horizon"`
	Prediction     float64   `json:"prediction"`
	LowerBound     float64   `json:"lowerBound"`
	UpperBound     float64   `json:"upperBound"`
	Unit           string    `json:"unit"`
	Recommendation string    `json:"recommendation,omitempty"`
	GeneratedAt    time.Time `json:"generatedAt"`
}

// AutoFixStep is one remediation step in a plan.
type AutoFixStep struct {
	ID          string `json:"id"`
	Action      string `json:"action"`
	Description string `json:"description"`
	BlastRadius string `json:"blastRadius"`
}

// AutoFixPlan is a proposed autonomous remediation plan.
type AutoFixPlan struct {
	ID               string        `json:"id"`
	IncidentID       string        `json:"incidentId"`
	Summary          string        `json:"summary"`
	RequiresApproval bool          `json:"requiresApproval"`
	PolicyPass       bool          `json:"policyPass"`
	PolicyReason     string        `json:"policyReason,omitempty"`
	Steps            []AutoFixStep `json:"steps"`
	CreatedAt        time.Time     `json:"createdAt"`
}

// AutoFixActionRecord tracks execution/rollback lifecycle.
type AutoFixActionRecord struct {
	ID         string     `json:"id"`
	PlanID     string     `json:"planId"`
	IncidentID string     `json:"incidentId"`
	Status     string     `json:"status"`
	StartedAt  time.Time  `json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
	Logs       []string   `json:"logs"`
}

// LLMWorkload is an inference deployment tracked for LLM observability (AI-03).
type LLMWorkload struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Model        string    `json:"model"`
	Provider     string    `json:"provider"`
	Requests24h  int64     `json:"requests24h"`
	Tokens24h    int64     `json:"tokens24h"`
	P95LatencyMs float64   `json:"p95LatencyMs"`
	ErrorRatePct float64   `json:"errorRatePct"`
	CostUSD24h   float64   `json:"costUsd24h"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// LLMUsageSummary aggregates token and cost usage for finance guardrails.
type LLMUsageSummary struct {
	Window       string  `json:"window"`
	TotalTokens  int64   `json:"totalTokens"`
	TotalCostUSD float64 `json:"totalCostUsd"`
	TopModel     string  `json:"topModel"`
	BudgetPct    float64 `json:"budgetPct"`
}

// CloudAsset is an inventoried cloud resource.
type CloudAsset struct {
	ID         string            `json:"id"`
	Provider   string            `json:"provider"`
	Type       string            `json:"type"`
	Name       string            `json:"name"`
	Region     string            `json:"region"`
	AccountID  string            `json:"accountId"`
	Status     string            `json:"status"`
	Tags       map[string]string `json:"tags,omitempty"`
	MonthlyUSD float64           `json:"monthlyUsd,omitempty"`
	UpdatedAt  time.Time         `json:"updatedAt"`
}

// CloudAssetTopology is multi-cloud resource relationship graph.
type CloudAssetTopology struct {
	Nodes []CloudAssetNode `json:"nodes"`
	Edges []CloudAssetEdge `json:"edges"`
	At    time.Time        `json:"at"`
}

// CloudAssetNode is one vertex in the cloud asset graph.
type CloudAssetNode struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Provider string `json:"provider"`
	Type     string `json:"type"`
	Health   string `json:"health"`
}

// CloudAssetEdge links cloud resources.
type CloudAssetEdge struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	Relation string `json:"relation"`
}

// FinOpsCostSeries is spend over time for a scope.
type FinOpsCostSeries struct {
	Scope  string            `json:"scope"`
	Unit   string            `json:"unit"`
	Total  float64           `json:"total"`
	Budget float64           `json:"budget"`
	Points []FinOpsCostPoint `json:"points"`
}

// FinOpsCostPoint is one cost sample.
type FinOpsCostPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Amount    float64   `json:"amount"`
}

// FinOpsCostAnomaly is detected abnormal spend.
type FinOpsCostAnomaly struct {
	ID          string    `json:"id"`
	Scope       string    `json:"scope"`
	Service     string    `json:"service"`
	Provider    string    `json:"provider"`
	DeltaPct    float64   `json:"deltaPct"`
	AmountUSD   float64   `json:"amountUsd"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	DetectedAt  time.Time `json:"detectedAt"`
}

// FinOpsCarbonFootprint is estimated emissions for a tenant scope.
type FinOpsCarbonFootprint struct {
	Scope          string  `json:"scope"`
	Period         string  `json:"period"`
	Co2eKg         float64 `json:"co2eKg"`
	RenewablePct   float64 `json:"renewablePct"`
	Recommendation string  `json:"recommendation,omitempty"`
}

// NetworkFlow is a sampled L4/L7 flow record.
type NetworkFlow struct {
	ID          string    `json:"id"`
	Source      string    `json:"source"`
	Destination string    `json:"destination"`
	Protocol    string    `json:"protocol"`
	Port        int       `json:"port"`
	Bytes       int64     `json:"bytes"`
	Packets     int64     `json:"packets"`
	LatencyMs   float64   `json:"latencyMs"`
	LossPct     float64   `json:"lossPct"`
	JitterMs    float64   `json:"jitterMs"`
	Timestamp   time.Time `json:"timestamp"`
}

// NetworkDevice is an SNMP/monitored network appliance.
type NetworkDevice struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Site      string  `json:"site"`
	Status    string  `json:"status"`
	CPUUtil   float64 `json:"cpuUtil"`
	MemUtil   float64 `json:"memUtil"`
	UptimePct float64 `json:"uptimePct"`
}

// NetworkTopologyGraph is NPM topology for path analysis.
type NetworkTopologyGraph struct {
	Nodes []NetworkTopologyNode `json:"nodes"`
	Edges []NetworkTopologyEdge `json:"edges"`
	At    time.Time             `json:"at"`
}

// NetworkTopologyNode is one NPM topology vertex.
type NetworkTopologyNode struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Type   string `json:"type"`
	Health string `json:"health"`
}

// NetworkTopologyEdge is one NPM link.
type NetworkTopologyEdge struct {
	Source    string  `json:"source"`
	Target    string  `json:"target"`
	LatencyMs float64 `json:"latencyMs"`
	LossPct   float64 `json:"lossPct"`
	UtilPct   float64 `json:"utilPct"`
}

// NetworkAnomaly is detected NPM degradation.
type NetworkAnomaly struct {
	ID          string    `json:"id"`
	Link        string    `json:"link"`
	Metric      string    `json:"metric"`
	Value       float64   `json:"value"`
	Baseline    float64   `json:"baseline"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	DetectedAt  time.Time `json:"detectedAt"`
}

// RUMFunnelStep is one stage in a conversion funnel.
type RUMFunnelStep struct {
	Name          string  `json:"name"`
	Event         string  `json:"event"`
	Count         int64   `json:"count"`
	ConversionPct float64 `json:"conversionPct"`
}

// RUMFunnel is a RUM conversion funnel definition with computed metrics.
type RUMFunnel struct {
	ID                   string          `json:"id"`
	Name                 string          `json:"name"`
	Steps                []RUMFunnelStep `json:"steps"`
	OverallConversionPct float64         `json:"overallConversionPct"`
	CreatedAt            time.Time       `json:"createdAt"`
	UpdatedAt            time.Time       `json:"updatedAt"`
}

// SyntheticBrowserTest is a Playwright-style browser synthetic check.
type SyntheticBrowserTest struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	URL        string    `json:"url"`
	Script     string    `json:"script"`
	Locations  []string  `json:"locations"`
	Enabled    bool      `json:"enabled"`
	LastStatus string    `json:"lastStatus"`
	CreatedAt  time.Time `json:"createdAt"`
}

// SyntheticMobileTest is a mobile synthetic journey.
type SyntheticMobileTest struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Platform   string    `json:"platform"`
	BundleID   string    `json:"bundleId"`
	Script     string    `json:"script"`
	Enabled    bool      `json:"enabled"`
	LastStatus string    `json:"lastStatus"`
	CreatedAt  time.Time `json:"createdAt"`
}

// SyntheticPrivateLocation is a self-hosted synthetic probe site.
type SyntheticPrivateLocation struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Region          string    `json:"region"`
	AgentVersion    string    `json:"agentVersion"`
	Status          string    `json:"status"`
	LastHeartbeatAt time.Time `json:"lastHeartbeatAt"`
}

// BusinessKPIDefinition is one KPI in a business pack.
type BusinessKPIDefinition struct {
	Key    string  `json:"key"`
	Label  string  `json:"label"`
	Unit   string  `json:"unit"`
	Target float64 `json:"target"`
}

// BusinessKPIPack is an installable business observability template.
type BusinessKPIPack struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Category    string                  `json:"category"`
	Description string                  `json:"description"`
	KPIs        []BusinessKPIDefinition `json:"kpis"`
	Connectors  []string                `json:"connectors"`
	Enabled     bool                    `json:"enabled"`
}

// ABACPolicyRule is one attribute-based access rule.
type ABACPolicyRule struct {
	ID        string `json:"id"`
	Effect    string `json:"effect"`
	Action    string `json:"action"`
	Resource  string `json:"resource"`
	Condition string `json:"condition"`
}

// ABACPolicy is tenant ABAC configuration layered on RBAC.
type ABACPolicy struct {
	Enabled   bool             `json:"enabled"`
	Rules     []ABACPolicyRule `json:"rules"`
	UpdatedAt time.Time        `json:"updatedAt"`
}

// DataResidencyPolicy controls regional data placement.
type DataResidencyPolicy struct {
	PrimaryRegion     string    `json:"primaryRegion"`
	AllowedRegions    []string  `json:"allowedRegions"`
	PIIStorageRegion  string    `json:"piiStorageRegion"`
	CrossBorderDenied bool      `json:"crossBorderDenied"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// BrandingTheme is white-label MSP branding.
type BrandingTheme struct {
	ProductName  string    `json:"productName"`
	LogoURL      string    `json:"logoUrl"`
	PrimaryColor string    `json:"primaryColor"`
	AccentColor  string    `json:"accentColor"`
	SupportEmail string    `json:"supportEmail"`
	CustomDomain string    `json:"customDomain"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// MSPTenant is a child tenant in an MSP control plane.
type MSPTenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Plan      string    `json:"plan"`
	Status    string    `json:"status"`
	UserCount int       `json:"userCount"`
	Region    string    `json:"region"`
	CreatedAt time.Time `json:"createdAt"`
}

// ExportJob tracks warehouse/BI/event export runs.
type ExportJob struct {
	ID           string     `json:"id"`
	Type         string     `json:"type"`
	Destination  string     `json:"destination"`
	Status       string     `json:"status"`
	RowsExported int64      `json:"rowsExported"`
	StartedAt    time.Time  `json:"startedAt"`
	FinishedAt   *time.Time `json:"finishedAt,omitempty"`
	Message      string     `json:"message,omitempty"`
}

// NFRBenchmark is one performance/latency certification check.
type NFRBenchmark struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Category   string    `json:"category"`
	Target     string    `json:"target"`
	Actual     string    `json:"actual"`
	Unit       string    `json:"unit"`
	Pass       bool      `json:"pass"`
	MeasuredAt time.Time `json:"measuredAt"`
}

// NFRReliabilityDrill is an HA/DR/failover exercise result.
type NFRReliabilityDrill struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"`
	Region        string    `json:"region"`
	Status        string    `json:"status"`
	RTOSeconds    int       `json:"rtoSeconds"`
	RTOSLOSeconds int       `json:"rtoSloSeconds"`
	Pass          bool      `json:"pass"`
	ExecutedAt    time.Time `json:"executedAt"`
}

// NFRA11yReport summarizes WCAG accessibility certification.
type NFRA11yReport struct {
	Standard           string    `json:"standard"`
	Level              string    `json:"level"`
	PagesAudited       int       `json:"pagesAudited"`
	ViolationsCritical int       `json:"violationsCritical"`
	ViolationsSerious  int       `json:"violationsSerious"`
	Pass               bool      `json:"pass"`
	LastAuditAt        time.Time `json:"lastAuditAt"`
}

// NFRLocale describes an enabled UI locale.
type NFRLocale struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	CoveragePct float64 `json:"coveragePct"`
	Enabled     bool    `json:"enabled"`
}

// NFRCertificationReport is the signed NFR exit-gate bundle.
type NFRCertificationReport struct {
	Version           string    `json:"version"`
	SignedAt          time.Time `json:"signedAt"`
	OverallPass       bool      `json:"overallPass"`
	BenchmarksPass    bool      `json:"benchmarksPass"`
	ReliabilityPass   bool      `json:"reliabilityPass"`
	AccessibilityPass bool      `json:"accessibilityPass"`
	I18nReady         bool      `json:"i18nReady"`
	EvidenceURIs      []string  `json:"evidenceUris"`
}

// ManagementZone scopes topology visibility.
type ManagementZone struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Services []string `json:"services"`
}
