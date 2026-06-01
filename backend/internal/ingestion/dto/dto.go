package dto

import (
	"encoding/json"
	"time"

	"github.com/neuralops/platform/internal/domain"
)

// LogIngestRequest is the API payload for log ingestion.
type LogIngestRequest struct {
	Timestamp   time.Time         `json:"timestamp"`
	Service     string            `json:"service"`
	Environment string            `json:"environment"`
	Severity    string            `json:"severity"`
	Message     string            `json:"message"`
	TraceID     string            `json:"traceId"`
	TxnID       string            `json:"txnId"`
	Thread      string            `json:"thread"`
	Host        string            `json:"host"`
	Pod         string            `json:"pod"`
	Namespace   string            `json:"namespace"`
	Labels      map[string]string `json:"labels"`
	RawJSON     json.RawMessage   `json:"rawJson"`
	TenantID    string            `json:"tenantId"`
	Plan        string            `json:"plan"`
}

// MetricIngestRequest is the API payload for metric ingestion.
type MetricIngestRequest struct {
	ServiceName string            `json:"serviceName"`
	Host        string            `json:"host"`
	Pod         string            `json:"pod"`
	MetricType  string            `json:"metricType"`
	Value       float64           `json:"value"`
	Timestamp   time.Time         `json:"timestamp"`
	Labels      map[string]string `json:"labels"`
	TenantID    string            `json:"tenantId"`
	Plan        string            `json:"plan"`
}

// EventIngestRequest represents deployment and configuration events.
type EventIngestRequest struct {
	ID          string    `json:"id"`
	Service     string    `json:"service"`
	Version     string    `json:"version"`
	DeployedAt  time.Time `json:"deployedAt"`
	DeployedBy  string    `json:"deployedBy"`
	ChangeType  string    `json:"changeType"`
	Environment string    `json:"environment"`
	TenantID    string    `json:"tenantId"`
	Plan        string    `json:"plan"`
}

// TraceSpanRequest represents a distributed trace span.
type TraceSpanRequest struct {
	TraceID     string            `json:"traceId"`
	SpanID      string            `json:"spanId"`
	ParentID    string            `json:"parentId"`
	Service     string            `json:"service"`
	Operation   string            `json:"operation"`
	StartTime   time.Time         `json:"startTime"`
	DurationMs  int64             `json:"durationMs"`
	Status      string            `json:"status"`
	Tags        map[string]string `json:"tags"`
	TenantID    string            `json:"tenantId"`
	Plan        string            `json:"plan"`
}

// AlertWebhookRequest represents third-party alert webhook payloads.
type AlertWebhookRequest struct {
	Source      string            `json:"source"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Severity    string            `json:"severity"`
	FiredAt     time.Time         `json:"firedAt"`
	Labels      map[string]string `json:"labels"`
	TenantID    string            `json:"tenantId"`
	Plan        string            `json:"plan"`
}

// IngestionMetadata is attached to enriched records before publishing.
type IngestionMetadata struct {
	ReceivedAt time.Time `json:"receivedAt"`
	IngestorID string    `json:"ingestorId"`
	Datacenter string    `json:"datacenter"`
	TenantID   string    `json:"tenantId,omitempty"`
	Source     string    `json:"source"`
	GeoCountry string    `json:"geoCountry,omitempty"`
	GeoCity    string    `json:"geoCity,omitempty"`
}

// EnrichedLog is a log entry with ingestion metadata for Kafka.
type EnrichedLog struct {
	domain.LogEntry
	IngestionMetadata IngestionMetadata `json:"ingestionMetadata"`
}

// IngestResponse is the standard ingestion API response.
type IngestResponse struct {
	Status    string `json:"status"`
	Accepted  int    `json:"accepted"`
	Rejected  int    `json:"rejected"`
	Timestamp string `json:"timestamp"`
}
