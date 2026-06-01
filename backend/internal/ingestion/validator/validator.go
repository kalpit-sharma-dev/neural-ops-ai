package validator

import (
	"fmt"
	"strings"
	"time"

	"github.com/neuralops/platform/internal/domain"
	"github.com/neuralops/platform/internal/ingestion/dto"
)

const maxMessageLength = 65536

// ValidateLog validates a single log ingest request.
func ValidateLog(req *dto.LogIngestRequest, maxSkew time.Duration) error {
	if req == nil {
		return domain.NewValidationError("body", "must not be nil")
	}
	if req.Timestamp.IsZero() {
		return domain.NewValidationError("timestamp", "is required")
	}
	if req.Service == "" {
		return domain.NewValidationError("service", "is required")
	}
	if req.Message == "" {
		return domain.NewValidationError("message", "is required")
	}
	if len(req.Message) > maxMessageLength {
		return domain.NewValidationError("message", "exceeds maximum length of 65536 bytes")
	}
	if err := validateTimestampSkew(req.Timestamp, maxSkew); err != nil {
		return err
	}
	if req.Environment != "" {
		env := domain.Environment(req.Environment)
		if !env.IsValid() {
			return domain.NewValidationError("environment", "must be prod, staging, or dev")
		}
	}
	if req.Severity != "" {
		sev := domain.LogSeverity(strings.ToUpper(req.Severity))
		if !sev.IsValid() {
			return domain.NewValidationError("severity", "is invalid")
		}
	}
	return nil
}

// ValidateMetric validates a metric ingest request.
func ValidateMetric(req *dto.MetricIngestRequest) error {
	if req == nil {
		return domain.NewValidationError("body", "must not be nil")
	}
	if req.ServiceName == "" {
		return domain.NewValidationError("serviceName", "is required")
	}
	if req.MetricType == "" {
		return domain.NewValidationError("metricType", "is required")
	}
	metricType := domain.MetricType(strings.ToUpper(req.MetricType))
	if !metricType.IsValid() {
		return domain.NewValidationError("metricType", "is invalid")
	}
	if req.Timestamp.IsZero() {
		return domain.NewValidationError("timestamp", "is required")
	}
	return nil
}

// ValidateEvent validates a deployment/config event request.
func ValidateEvent(req *dto.EventIngestRequest) error {
	if req == nil {
		return domain.NewValidationError("body", "must not be nil")
	}
	if req.Service == "" {
		return domain.NewValidationError("service", "is required")
	}
	if req.Version == "" {
		return domain.NewValidationError("version", "is required")
	}
	if req.DeployedAt.IsZero() {
		return domain.NewValidationError("deployedAt", "is required")
	}
	if req.ChangeType != "" {
		changeType := domain.ChangeType(strings.ToUpper(req.ChangeType))
		if !changeType.IsValid() {
			return domain.NewValidationError("changeType", "is invalid")
		}
	}
	return nil
}

// ValidateTrace validates a trace span request.
func ValidateTrace(req *dto.TraceSpanRequest) error {
	if req == nil {
		return domain.NewValidationError("body", "must not be nil")
	}
	if req.TraceID == "" {
		return domain.NewValidationError("traceId", "is required")
	}
	if req.SpanID == "" {
		return domain.NewValidationError("spanId", "is required")
	}
	if req.Service == "" {
		return domain.NewValidationError("service", "is required")
	}
	if req.StartTime.IsZero() {
		return domain.NewValidationError("startTime", "is required")
	}
	return nil
}

func validateTimestampSkew(ts time.Time, maxSkew time.Duration) error {
	now := time.Now().UTC()
	diff := now.Sub(ts.UTC())
	if diff < 0 {
		diff = -diff
	}
	if diff > maxSkew {
		return domain.NewValidationError("timestamp", fmt.Sprintf("must be within ±%s of now", maxSkew))
	}
	return nil
}
