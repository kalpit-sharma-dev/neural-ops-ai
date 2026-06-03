package observability

import "time"

// SupportedPlatform describes an OS/kernel matrix entry (COLL-01).
type SupportedPlatform struct {
	OS            string   `json:"os"`
	Arch          string   `json:"arch"`
	KernelMin     string   `json:"kernelMin,omitempty"`
	Status        string   `json:"status"` // certified|preview|planned
	AgentVersion  string   `json:"agentVersion"`
	EBPFSupported bool     `json:"ebpfSupported"`
	Notes         string   `json:"notes,omitempty"`
	Features      []string `json:"features,omitempty"`
}

// DiscoveredService is auto-discovered from a collector agent.
type DiscoveredService struct {
	AgentID     string    `json:"agentId"`
	ServiceName string    `json:"serviceName"`
	Language    string    `json:"language,omitempty"`
	Port        int       `json:"port,omitempty"`
	DiscoveredAt time.Time `json:"discoveredAt"`
}

// CollectorSpoolStatus reports air-gap buffer state (COLL-06).
type CollectorSpoolStatus struct {
	AgentID        string    `json:"agentId"`
	PendingBatches int       `json:"pendingBatches"`
	OldestAgeSec   int       `json:"oldestAgeSec"`
	BytesOnDisk    int64     `json:"bytesOnDisk"`
	MaxBytes       int64     `json:"maxBytes"`
	ReplayLagSec   int       `json:"replayLagSec"`
	Healthy        bool      `json:"healthy"`
	CheckedAt      time.Time `json:"checkedAt"`
}

// CollectorBenchmarkResult is idle resource measurement (COLL-05).
type CollectorBenchmarkResult struct {
	AgentID       string    `json:"agentId"`
	CPUPercent    float64   `json:"cpuPercent"`
	MemoryMB      float64   `json:"memoryMB"`
	TargetCPU     float64   `json:"targetCpuPercent"`
	TargetMemory  float64   `json:"targetMemoryMb"`
	WithinTarget  bool      `json:"withinTarget"`
	MeasuredAt    time.Time `json:"measuredAt"`
}

// CollectorVersionDrift summarizes fleet version skew (COLL-09).
type CollectorVersionDrift struct {
	TargetVersion string `json:"targetVersion"`
	Drifted       []struct {
		AgentID string `json:"agentId"`
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"drifted"`
	HealthyCount int `json:"healthyCount"`
	DriftCount   int `json:"driftCount"`
}

// CollectorRolloutRequest configures canary rollout (COLL-07).
type CollectorRolloutRequest struct {
	TargetVersion string `json:"targetVersion"`
	Strategy      string `json:"strategy"` // all|canary|rolling
	CanaryPercent int    `json:"canaryPercent,omitempty"`
}

// CardinalityAlertPolicy configures NexQL cardinality breach alerts (MET-02).
type CardinalityAlertPolicy struct {
	Enabled           bool    `json:"enabled"`
	ThresholdServices int     `json:"thresholdServices"`
	NotifyChannel     string  `json:"notifyChannel,omitempty"`
	LastBreachAt      *time.Time `json:"lastBreachAt,omitempty"`
}

// MaintenanceWindow is a scheduled alert silence (ALR-02).
type MaintenanceWindow struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	ServicePattern string    `json:"servicePattern"`
	Reason         string    `json:"reason"`
	Recurrence     string    `json:"recurrence,omitempty"` // once|daily|weekly
	StartsAt       time.Time `json:"startsAt"`
	EndsAt         time.Time `json:"endsAt"`
	CreatedBy      string    `json:"createdBy,omitempty"`
}

// AlertFatigueConfig tunes fatigue scoring per policy (ALR-04).
type AlertFatigueConfig struct {
	PolicyID          string  `json:"policyId"`
	AutoSuppressScore float64 `json:"autoSuppressScore"`
	WeightDuplicate   float64 `json:"weightDuplicate"`
	WeightFrequency   float64 `json:"weightFrequency"`
}

// APMErrorGroup is a grouped error for release tracking (APM-04).
type APMErrorGroup struct {
	ID          string    `json:"id"`
	Service     string    `json:"service"`
	Fingerprint string    `json:"fingerprint"`
	Message     string    `json:"message"`
	Count       int       `json:"count"`
	Release     string    `json:"release,omitempty"`
	FirstSeen   time.Time `json:"firstSeen"`
	LastSeen    time.Time `json:"lastSeen"`
	TraceID     string    `json:"traceId,omitempty"`
}

// APMRelease tracks deployment versions (APM-04).
type APMRelease struct {
	ID        string    `json:"id"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	CommitSHA string    `json:"commitSha,omitempty"`
	DeployedAt time.Time `json:"deployedAt"`
	Environment string  `json:"environment,omitempty"`
}

// ServiceCatalogEntry maps services to owners (APM-06).
type ServiceCatalogEntry struct {
	Service      string   `json:"service"`
	DisplayName  string   `json:"displayName,omitempty"`
	OwnerTeam    string   `json:"ownerTeam"`
	OwnerEmail   string   `json:"ownerEmail,omitempty"`
	Tier         string   `json:"tier"` // P0|P1|P2
	Dependencies []string `json:"dependencies,omitempty"`
	RunbookURL   string   `json:"runbookUrl,omitempty"`
}

// BusinessTransactionView joins traces/logs for a txn (APM-07).
type BusinessTransactionView struct {
	TxnID      string              `json:"txnId"`
	Service    string              `json:"service,omitempty"`
	Status     string              `json:"status"`
	TraceIDs   []string            `json:"traceIds"`
	LogCount   int                 `json:"logCount"`
	DurationMs float64             `json:"durationMs"`
	Spans      []Span              `json:"spans,omitempty"`
	StartedAt  time.Time           `json:"startedAt"`
}

// LogIngestFormat describes supported log ingest paths (LOG-01).
type LogIngestFormat struct {
	Format      string   `json:"format"`
	Endpoint    string   `json:"endpoint"`
	ContentType string   `json:"contentType,omitempty"`
	Status      string   `json:"status"`
	Notes       string   `json:"notes,omitempty"`
	Features    []string `json:"features,omitempty"`
}

// LogCorrelation links a log entry to traces and metrics (LOG-03).
type LogCorrelation struct {
	LogID     string   `json:"logId"`
	TraceID   string   `json:"traceId,omitempty"`
	SpanID    string   `json:"spanId,omitempty"`
	Service   string   `json:"service,omitempty"`
	TxnID     string   `json:"txnId,omitempty"`
	TraceLink string   `json:"traceLink,omitempty"`
	Related   []string `json:"relatedLogIds,omitempty"`
}

// LogAnomaly is an AI-detected log pattern anomaly (LOG-06).
type LogAnomaly struct {
	ID         string    `json:"id"`
	Service    string    `json:"service"`
	Pattern    string    `json:"pattern"`
	Score      float64   `json:"score"`
	Precision  float64   `json:"precision,omitempty"`
	Recall     float64   `json:"recall,omitempty"`
	Message    string    `json:"message"`
	DetectedAt time.Time `json:"detectedAt"`
	Feedback   string    `json:"feedback,omitempty"` // useful|noise
}

// SignalPolicy defines per-signal retention and masking (ADM-02).
type SignalPolicy struct {
	Signal         string `json:"signal"` // logs|traces|metrics|events
	RetentionDays  int    `json:"retentionDays"`
	MaskingEnabled bool   `json:"maskingEnabled"`
	ResidencyRegion string `json:"residencyRegion,omitempty"`
	PIIFields      []string `json:"piiFields,omitempty"`
}

// SIEMConfig holds per-tenant SIEM export targets (SEC-02).
type SIEMConfig struct {
	Provider    string            `json:"provider"` // splunk|sentinel|qradar
	Endpoint    string            `json:"endpoint"`
	Index       string            `json:"index,omitempty"`
	WorkspaceID string            `json:"workspaceId,omitempty"`
	FailClosed  bool              `json:"failClosed"`
	Extra       map[string]string `json:"extra,omitempty"`
}

// ThreatIndicator is a threat intel record (SEC-03).
type ThreatIndicator struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // ip|domain|hash
	Value       string    `json:"value"`
	Severity    string    `json:"severity"`
	Source      string    `json:"source"`
	Description string    `json:"description,omitempty"`
	ObservedAt  time.Time `json:"observedAt"`
}

// CloudMetricCatalogEntry describes a cloud metric series.
type CloudMetricCatalogEntry struct {
	Provider    string `json:"provider"`
	Namespace   string `json:"namespace"`
	MetricName  string `json:"metricName"`
	Unit        string `json:"unit"`
	Description string `json:"description,omitempty"`
}

// DevOpsWebhookEvent is a CI/CD deploy notification (INT-01).
type DevOpsWebhookEvent struct {
	Provider  string    `json:"provider"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	CommitSHA string    `json:"commitSha,omitempty"`
	Status    string    `json:"status"`
	ReceivedAt time.Time `json:"receivedAt"`
}
