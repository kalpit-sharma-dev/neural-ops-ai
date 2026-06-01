package seed

import "time"

// LogRecord is a normalized log row for export and bulk loading.
type LogRecord struct {
	ID             string
	Timestamp      time.Time
	TenantID       string
	Service        string
	Environment    string
	Severity       string
	Message        string
	TraceID        string
	TxnID          string
	Host           string
	Pod            string
	Classification string
}

// TransactionRecord represents a banking transaction journey.
type TransactionRecord struct {
	TxnID          string
	TxnType        string
	Status         string
	FailedAt       string
	TotalLatencyMs int64
	RetryCount     int32
	HopsJSON       string
	StartedAt      time.Time
	CompletedAt    *time.Time
}

// DeploymentRecord represents a service deployment event.
type DeploymentRecord struct {
	ID          string
	Service     string
	Version     string
	DeployedAt  time.Time
	DeployedBy  string
	ChangeType  string
	Environment string
	Category    string // clean | minor_issue | caused_incident
}

// ScenarioTimeline anchors the UPI outage simulation.
type ScenarioTimeline struct {
	OutageStart   time.Time
	DeployAt      time.Time
	ErrorsStart   time.Time
	ThreadPoolAt  time.Time
	CascadeAt     time.Time
	ResolvedAt    time.Time
	RollbackAt    time.Time
	WindowEnd     time.Time
}

// NewScenarioTimeline builds a reproducible outage window ending recently.
func NewScenarioTimeline(now time.Time) ScenarioTimeline {
	outageStart := now.Add(-45 * time.Minute)
	return ScenarioTimeline{
		OutageStart:  outageStart,
		DeployAt:     outageStart,
		ErrorsStart:  outageStart.Add(3 * time.Minute),
		ThreadPoolAt: outageStart.Add(5 * time.Minute),
		CascadeAt:    outageStart.Add(8 * time.Minute),
	ResolvedAt:   outageStart.Add(45 * time.Minute),
		RollbackAt:   outageStart.Add(45 * time.Minute),
		WindowEnd:    now,
	}
}
