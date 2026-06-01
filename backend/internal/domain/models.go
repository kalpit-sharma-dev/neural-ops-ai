package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	maxLogMessageLength = 65536
	maxConfidence       = 1.0
)

// LogEntry represents a normalized observability log record.
type LogEntry struct {
	ID                uuid.UUID             `json:"id" validate:"required" db:"id"`
	Timestamp         time.Time             `json:"timestamp" validate:"required" db:"timestamp"`
	Service           string                `json:"service" validate:"required,max=255" db:"service"`
	Environment       Environment           `json:"environment" validate:"required,oneof=prod staging dev" db:"environment"`
	Severity          LogSeverity           `json:"severity" validate:"required,oneof=DEBUG INFO WARN ERROR FATAL CRITICAL" db:"severity"`
	Message           string                `json:"message" validate:"required,max=65536" db:"message"`
	TraceID           string                `json:"traceId,omitempty" validate:"omitempty,max=128" db:"trace_id"`
	TxnID             string                `json:"txnId,omitempty" validate:"omitempty,max=128" db:"txn_id"`
	Thread            string                `json:"thread,omitempty" validate:"omitempty,max=128" db:"thread"`
	Host              string                `json:"host,omitempty" validate:"omitempty,max=255" db:"host"`
	Pod               string                `json:"pod,omitempty" validate:"omitempty,max=255" db:"pod"`
	Namespace         string                `json:"namespace,omitempty" validate:"omitempty,max=255" db:"namespace"`
	Labels            map[string]string     `json:"labels,omitempty" db:"labels"`
	RawJSON           json.RawMessage       `json:"rawJson,omitempty" db:"raw_json"`
	EmbeddingVector   []float32             `json:"embeddingVector,omitempty" db:"embedding_vector"`
	ParsedStackTrace  *StackTrace           `json:"parsedStackTrace,omitempty" db:"-"`
	ClassifiedError   *ErrorClassification  `json:"classifiedError,omitempty" db:"-"`
	CreatedAt         time.Time             `json:"createdAt,omitempty" db:"created_at"`
	UpdatedAt         time.Time             `json:"updatedAt,omitempty" db:"updated_at"`
}

// StackTrace represents a parsed multi-language exception stack trace.
type StackTrace struct {
	Language          StackTraceLanguage `json:"language" validate:"required,oneof=Java Golang Python NodeJS DotNet" db:"language"`
	ExceptionType     string             `json:"exceptionType" validate:"required,max=512" db:"exception_type"`
	ExceptionMessage  string             `json:"exceptionMessage" validate:"required,max=4096" db:"exception_message"`
	Frames            []StackFrame         `json:"frames" validate:"required,dive" db:"-"`
	Hotspot           string             `json:"hotspot,omitempty" validate:"omitempty,max=512" db:"hotspot"`
}

// StackFrame represents a single frame within a stack trace.
type StackFrame struct {
	ClassName     string `json:"className,omitempty" validate:"omitempty,max=512" db:"class_name"`
	Package       string `json:"package,omitempty" validate:"omitempty,max=512" db:"package"`
	MethodName    string `json:"methodName,omitempty" validate:"omitempty,max=512" db:"method_name"`
	FunctionName  string `json:"functionName,omitempty" validate:"omitempty,max=512" db:"function_name"`
	FileName      string `json:"fileName" validate:"required,max=1024" db:"file_name"`
	LineNumber    int    `json:"lineNumber" validate:"gte=0" db:"line_number"`
	IsThirdParty  bool   `json:"isThirdParty" db:"is_third_party"`
}

// ErrorClassification captures AI or rule-based error categorization.
type ErrorClassification struct {
	Category   ErrorCategory `json:"category" validate:"required" db:"category"`
	Confidence float64       `json:"confidence" validate:"required,gte=0,lte=1" db:"confidence"`
	Reasoning  string        `json:"reasoning" validate:"required,max=8192" db:"reasoning"`
}

// Incident represents a correlated operational incident.
type Incident struct {
	ID                     uuid.UUID           `json:"id" validate:"required" db:"id"`
	Title                  string              `json:"title" validate:"required,max=512" db:"title"`
	Summary                string              `json:"summary" validate:"required,max=16384" db:"summary"`
	Severity               IncidentSeverity    `json:"severity" validate:"required,oneof=P1 P2 P3 P4" db:"severity"`
	Status                 IncidentStatus      `json:"status" validate:"required,oneof=OPEN INVESTIGATING RESOLVED SUPPRESSED" db:"status"`
	AffectedServices       []string            `json:"affectedServices" validate:"required,min=1,dive,max=255" db:"-"`
	RootCauseAnalysis      *RootCause          `json:"rootCauseAnalysis,omitempty" db:"-"`
	CorrelatedLogIDs       []uuid.UUID         `json:"correlatedLogIds,omitempty" db:"-"`
	CorrelatedTraceIDs     []uuid.UUID         `json:"correlatedTraceIds,omitempty" db:"-"`
	CorrelatedDeploymentID *uuid.UUID          `json:"correlatedDeploymentId,omitempty" db:"correlated_deployment_id"`
	BlastRadius            []string            `json:"blastRadius,omitempty" validate:"dive,max=255" db:"-"`
	StartTime              time.Time           `json:"startTime" validate:"required" db:"start_time"`
	ResolvedTime           *time.Time          `json:"resolvedTime,omitempty" db:"resolved_time"`
	AcknowledgedAt         *time.Time          `json:"acknowledgedAt,omitempty" db:"acknowledged_at"`
	MTTR                   time.Duration       `json:"mttr,omitempty" db:"mttr_ns"`
	Recommendations        []Recommendation    `json:"recommendations,omitempty" validate:"omitempty,dive" db:"-"`
	Timeline               []IncidentEvent     `json:"timeline,omitempty" validate:"omitempty,dive" db:"-"`
	CreatedAt              time.Time           `json:"createdAt,omitempty" db:"created_at"`
	UpdatedAt              time.Time           `json:"updatedAt,omitempty" db:"updated_at"`
}

// RootCause captures AI-generated root cause analysis for an incident.
type RootCause struct {
	FirstFailingService    string                  `json:"firstFailingService" validate:"required,max=255" db:"first_failing_service"`
	RootCauseDescription   string                  `json:"rootCauseDescription" validate:"required,max=16384" db:"root_cause_description"`
	Evidence               []Evidence              `json:"evidence,omitempty" validate:"omitempty,dive" db:"-"`
	DeploymentCorrelation  *DeploymentCorrelation  `json:"deploymentCorrelation,omitempty" db:"-"`
	InfraCorrelation       *InfraCorrelation       `json:"infraCorrelation,omitempty" db:"-"`
	Confidence             float64                 `json:"confidence" validate:"required,gte=0,lte=1" db:"confidence"`
}

// Evidence supports root cause analysis with correlated artifacts.
type Evidence struct {
	ID          uuid.UUID    `json:"id" validate:"required" db:"id"`
	Type        EvidenceType `json:"type" validate:"required" db:"type"`
	Description string       `json:"description" validate:"required,max=4096" db:"description"`
	Source      string       `json:"source,omitempty" validate:"omitempty,max=255" db:"source"`
	ReferenceID *uuid.UUID   `json:"referenceId,omitempty" db:"reference_id"`
	Snippet     string       `json:"snippet,omitempty" validate:"omitempty,max=16384" db:"snippet"`
	Timestamp   time.Time    `json:"timestamp" validate:"required" db:"timestamp"`
}

// DeploymentCorrelation links an incident to a deployment change.
type DeploymentCorrelation struct {
	DeploymentID      uuid.UUID `json:"deploymentId" validate:"required" db:"deployment_id"`
	Service           string    `json:"service" validate:"required,max=255" db:"service"`
	Version           string    `json:"version" validate:"required,max=128" db:"version"`
	DeployedAt        time.Time `json:"deployedAt" validate:"required" db:"deployed_at"`
	CorrelationScore  float64   `json:"correlationScore" validate:"required,gte=0,lte=1" db:"correlation_score"`
}

// InfraCorrelation links an incident to infrastructure anomalies.
type InfraCorrelation struct {
	ResourceType     string    `json:"resourceType" validate:"required,max=128" db:"resource_type"`
	ResourceID       string    `json:"resourceId" validate:"required,max=255" db:"resource_id"`
	Description      string    `json:"description" validate:"required,max=4096" db:"description"`
	CorrelationScore float64   `json:"correlationScore" validate:"required,gte=0,lte=1" db:"correlation_score"`
	DetectedAt       time.Time `json:"detectedAt" validate:"required" db:"detected_at"`
}

// Recommendation represents an actionable remediation suggestion.
type Recommendation struct {
	Type        RecommendationType     `json:"type" validate:"required" db:"type"`
	Description string                 `json:"description" validate:"required,max=8192" db:"description"`
	CodeSnippet string                 `json:"codeSnippet,omitempty" validate:"omitempty,max=16384" db:"code_snippet"`
	Priority    RecommendationPriority `json:"priority" validate:"required,oneof=HIGH MEDIUM LOW" db:"priority"`
}

// Deployment represents a service deployment or configuration change event.
type Deployment struct {
	ID          uuid.UUID   `json:"id" validate:"required" db:"id"`
	Service     string      `json:"service" validate:"required,max=255" db:"service"`
	Version     string      `json:"version" validate:"required,max=128" db:"version"`
	DeployedAt  time.Time   `json:"deployedAt" validate:"required" db:"deployed_at"`
	DeployedBy  string      `json:"deployedBy" validate:"required,max=255" db:"deployed_by"`
	ChangeType  ChangeType  `json:"changeType" validate:"required" db:"change_type"`
	Environment Environment `json:"environment" validate:"required,oneof=prod staging dev" db:"environment"`
	CreatedAt   time.Time   `json:"createdAt,omitempty" db:"created_at"`
}

// Alert represents an external or internal alert signal.
type Alert struct {
	ID               uuid.UUID        `json:"id" validate:"required" db:"id"`
	Source           AlertSource      `json:"source" validate:"required" db:"source"`
	Title            string           `json:"title" validate:"required,max=512" db:"title"`
	Description      string           `json:"description" validate:"required,max=8192" db:"description"`
	Severity         IncidentSeverity `json:"severity" validate:"required,oneof=P1 P2 P3 P4" db:"severity"`
	FiredAt          time.Time        `json:"firedAt" validate:"required" db:"fired_at"`
	ResolvedAt       *time.Time       `json:"resolvedAt,omitempty" db:"resolved_at"`
	Labels           map[string]string `json:"labels,omitempty" db:"labels"`
	LinkedIncidentID *uuid.UUID       `json:"linkedIncidentId,omitempty" db:"linked_incident_id"`
	Deduplicated     bool             `json:"deduplicated" db:"deduplicated"`
	CreatedAt        time.Time        `json:"createdAt,omitempty" db:"created_at"`
	UpdatedAt        time.Time        `json:"updatedAt,omitempty" db:"updated_at"`
}

// Transaction represents a banking transaction journey across microservices.
type Transaction struct {
	TxnID          string    `json:"txnId" validate:"required,max=128" db:"txn_id"`
	TxnType        TxnType   `json:"txnType" validate:"required,oneof=UPI NEFT RTGS IMPS" db:"txn_type"`
	Status         TxnStatus `json:"status" validate:"required,oneof=SUCCESS FAILED PARTIAL PENDING" db:"status"`
	Hops           []TxnHop  `json:"hops" validate:"required,min=1,dive" db:"-"`
	TotalLatencyMs int64     `json:"totalLatencyMs" validate:"gte=0" db:"total_latency_ms"`
	FailedAt       string    `json:"failedAt,omitempty" validate:"omitempty,max=255" db:"failed_at"`
	RetryCount     int       `json:"retryCount" validate:"gte=0" db:"retry_count"`
	StartedAt      time.Time `json:"startedAt,omitempty" db:"started_at"`
	CompletedAt    *time.Time `json:"completedAt,omitempty" db:"completed_at"`
}

// TxnHop represents one service hop within a banking transaction.
type TxnHop struct {
	ServiceName  string    `json:"serviceName" validate:"required,max=255" db:"service_name"`
	SpanID       string    `json:"spanId" validate:"required,max=128" db:"span_id"`
	TraceID      string    `json:"traceId" validate:"required,max=128" db:"trace_id"`
	LatencyMs    int64     `json:"latencyMs" validate:"gte=0" db:"latency_ms"`
	Status       HopStatus `json:"status" validate:"required,oneof=SUCCESS FAILED PARTIAL PENDING" db:"status"`
	ErrorMessage string    `json:"errorMessage,omitempty" validate:"omitempty,max=4096" db:"error_message"`
	StartedAt    time.Time `json:"startedAt,omitempty" db:"started_at"`
	CompletedAt  *time.Time `json:"completedAt,omitempty" db:"completed_at"`
}

// Metric represents a time-series metric sample.
type Metric struct {
	ServiceName string     `json:"serviceName" validate:"required,max=255" db:"service_name"`
	Host        string     `json:"host,omitempty" validate:"omitempty,max=255" db:"host"`
	Pod         string     `json:"pod,omitempty" validate:"omitempty,max=255" db:"pod"`
	MetricType  MetricType `json:"metricType" validate:"required" db:"metric_type"`
	Value       float64    `json:"value" validate:"required" db:"value"`
	Timestamp   time.Time  `json:"timestamp" validate:"required" db:"timestamp"`
	Labels      map[string]string `json:"labels,omitempty" db:"labels"`
}

// AnomalyDetection represents a detected metric anomaly.
type AnomalyDetection struct {
	ServiceName string     `json:"serviceName" validate:"required,max=255" db:"service_name"`
	MetricType  MetricType `json:"metricType" validate:"required" db:"metric_type"`
	Score       float64    `json:"score" validate:"required,gte=0,lte=1" db:"score"`
	Detected    bool       `json:"detected" db:"detected"`
	Baseline    float64    `json:"baseline" validate:"required" db:"baseline"`
	Actual      float64    `json:"actual" validate:"required" db:"actual"`
	DetectedAt  time.Time  `json:"detectedAt" validate:"required" db:"detected_at"`
	AnomalyType AnomalyType `json:"anomalyType" validate:"required,oneof=SPIKE DROP TREND SEASONAL" db:"anomaly_type"`
}

// IncidentEvent represents a single event on an incident timeline.
type IncidentEvent struct {
	ID          uuid.UUID         `json:"id" validate:"required" db:"id"`
	Timestamp   time.Time         `json:"timestamp" validate:"required" db:"timestamp"`
	EventType   IncidentEventType `json:"eventType" validate:"required" db:"event_type"`
	Title       string            `json:"title" validate:"required,max=512" db:"title"`
	Description string            `json:"description,omitempty" validate:"omitempty,max=8192" db:"description"`
	Actor       string            `json:"actor,omitempty" validate:"omitempty,max=255" db:"actor"`
	ReferenceID *uuid.UUID        `json:"referenceId,omitempty" db:"reference_id"`
	Metadata    map[string]string `json:"metadata,omitempty" db:"metadata"`
}

// ComputeMTTR calculates mean time to resolution when the incident is resolved.
func (i *Incident) ComputeMTTR() {
	if i.ResolvedTime == nil || i.StartTime.IsZero() {
		i.MTTR = 0
		return
	}
	resolved := *i.ResolvedTime
	if resolved.Before(i.StartTime) {
		i.MTTR = 0
		return
	}
	i.MTTR = resolved.Sub(i.StartTime)
}

// IsResolved reports whether the incident has been resolved.
func (i *Incident) IsResolved() bool {
	return i.Status == IncidentStatusResolved
}

// IsOpen reports whether the incident is actively open or under investigation.
func (i *Incident) IsOpen() bool {
	return i.Status == IncidentStatusOpen || i.Status == IncidentStatusInvestigating
}

// HasStackTrace reports whether the log entry contains a parsed stack trace.
func (l *LogEntry) HasStackTrace() bool {
	return l.ParsedStackTrace != nil && len(l.ParsedStackTrace.Frames) > 0
}

// IsErrorSeverity reports whether the log severity indicates an error condition.
func (l *LogEntry) IsErrorSeverity() bool {
	switch l.Severity {
	case LogSeverityError, LogSeverityFatal, LogSeverityCritical:
		return true
	default:
		return false
	}
}

// FailedHop returns the first failed hop in the transaction, if any.
func (t *Transaction) FailedHop() *TxnHop {
	for i := range t.Hops {
		if t.Hops[i].Status == HopStatusFailed || t.Hops[i].Status == HopStatusPartial {
			return &t.Hops[i]
		}
	}
	return nil
}

// ValidateMessageLength ensures log message size stays within platform limits.
func ValidateMessageLength(message string) error {
	if len(message) > maxLogMessageLength {
		return NewValidationError("message", "exceeds maximum length of 65536 bytes")
	}
	return nil
}

// ValidateConfidence ensures confidence scores remain within [0, 1].
func ValidateConfidence(confidence float64) error {
	if confidence < 0 || confidence > maxConfidence {
		return NewValidationError("confidence", "must be between 0 and 1")
	}
	return nil
}
