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
	ID           string    `json:"id"`
	TenantID     string    `json:"tenantId"`
	Name         string    `json:"name"`
	Service      string    `json:"service"`
	SLIQuery     string    `json:"sliQuery"`
	Target       float64   `json:"target"`
	WindowDays   int       `json:"windowDays"`
	ErrorBudget  float64   `json:"errorBudget"`
	BurnRate     float64   `json:"burnRate"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
}

// EntityAnomaly is an entity-scoped anomaly record.
type EntityAnomaly struct {
	ID        string    `json:"id"`
	EntityID  string    `json:"entityId"`
	EntityType string   `json:"entityType"`
	Service   string    `json:"service"`
	Metric    string    `json:"metric"`
	Score     float64   `json:"score"`
	Message   string    `json:"message"`
	DetectedAt time.Time `json:"detectedAt"`
}

// HostSummary is an infrastructure host list item.
type HostSummary struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Status    string  `json:"status"`
	CPU       float64 `json:"cpuPercent"`
	Memory    float64 `json:"memoryPercent"`
	Disk      float64 `json:"diskPercent"`
	Zone      string  `json:"zone"`
}

// K8sCluster is a kubernetes cluster summary.
type K8sCluster struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Nodes        int    `json:"nodes"`
	Pods         int    `json:"pods"`
	Health       string `json:"health"`
	NamespaceCount int  `json:"namespaceCount"`
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
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Engine     string  `json:"engine"`
	Status     string  `json:"status"`
	QPS        float64 `json:"qps"`
	SlowQueries int    `json:"slowQueries"`
	Connections int    `json:"connections"`
}

// DBStatement is a top database statement.
type DBStatement struct {
	Query      string  `json:"query"`
	Calls      int64   `json:"calls"`
	AvgMs      float64 `json:"avgMs"`
	TotalMs    float64 `json:"totalMs"`
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
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Page      string    `json:"page"`
	Device    string    `json:"device"`
	Country   string    `json:"country"`
	DurationMs int64    `json:"durationMs"`
	Errors    int       `json:"errors"`
	LCP       float64   `json:"lcp"`
	StartedAt time.Time `json:"startedAt"`
}

// SyntheticMonitor is an HTTP/browser synthetic check.
type SyntheticMonitor struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	URL       string    `json:"url"`
	Interval  string    `json:"interval"`
	Locations []string  `json:"locations"`
	Enabled   bool      `json:"enabled"`
	LastStatus string   `json:"lastStatus"`
	LastRunAt time.Time `json:"lastRunAt"`
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

// Workflow defines an automation workflow.
type Workflow struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Trigger     string    `json:"trigger"`
	Enabled     bool      `json:"enabled"`
	Steps       []string  `json:"steps"`
	LastRunAt   *time.Time `json:"lastRunAt,omitempty"`
}

// Notebook is a saved analysis notebook.
type Notebook struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Cells     []NotebookCell `json:"cells"`
	UpdatedAt time.Time `json:"updatedAt"`
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
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	SourceIP  string    `json:"sourceIp"`
	Service   string    `json:"service"`
	Blocked   bool      `json:"blocked"`
	DetectedAt time.Time `json:"detectedAt"`
}

// Integration describes a third-party integration.
type Integration struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	Connected bool  `json:"connected"`
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
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Prefix    string    `json:"prefix"`
	Scopes    []string  `json:"scopes"`
	CreatedAt time.Time `json:"createdAt"`
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
	LogsIngestedGB   float64 `json:"logsIngestedGb"`
	TracesIngested   int64   `json:"tracesIngested"`
	MetricsIngested  int64   `json:"metricsIngested"`
	AITokensUsed     int64   `json:"aiTokensUsed"`
	ActiveUsers      int     `json:"activeUsers"`
}

// ManagementZone scopes topology visibility.
type ManagementZone struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Services []string `json:"services"`
}
