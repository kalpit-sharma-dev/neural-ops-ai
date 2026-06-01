package normalizer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/neuralops/platform/internal/alerting/model"
	"github.com/neuralops/platform/internal/domain"
)

// Fingerprint builds deduplication fingerprint.
func Fingerprint(source domain.AlertSource, alertName, service string, labels map[string]string) string {
	payload := fmt.Sprintf("%s|%s|%s|%s", source, alertName, service, stableLabels(labels))
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:16])
}

func stableLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	raw, _ := json.Marshal(labels)
	return string(raw)
}

func mapSeverity(value string) domain.IncidentSeverity {
	switch strings.ToUpper(value) {
	case "P1", "CRITICAL", "SEV1", "SEVERE":
		return domain.IncidentSeverityP1
	case "P2", "HIGH", "SEV2", "ERROR":
		return domain.IncidentSeverityP2
	case "P3", "MEDIUM", "SEV3", "WARN", "WARNING":
		return domain.IncidentSeverityP3
	default:
		return domain.IncidentSeverityP4
	}
}

func serviceFromLabels(labels map[string]string, fallback string) string {
	for _, key := range []string{"service", "service_name", "job", "app", "application"} {
		if value := strings.TrimSpace(labels[key]); value != "" {
			return value
		}
	}
	if fallback != "" {
		return fallback
	}
	return "unknown-service"
}

// PrometheusPayload is Alertmanager webhook payload.
type PrometheusPayload struct {
	Status string `json:"status"`
	Alerts []struct {
		Status       string            `json:"status"`
		Labels       map[string]string `json:"labels"`
		Annotations  map[string]string `json:"annotations"`
		StartsAt     time.Time         `json:"startsAt"`
		EndsAt       time.Time         `json:"endsAt"`
		GeneratorURL string            `json:"generatorURL"`
	} `json:"alerts"`
}

// FromPrometheus normalizes Alertmanager webhook payload.
func FromPrometheus(payload PrometheusPayload) []model.IncomingAlert {
	out := make([]model.IncomingAlert, 0, len(payload.Alerts))
	for _, alert := range payload.Alerts {
		labels := alert.Labels
		if labels == nil {
			labels = map[string]string{}
		}
		alertName := labels["alertname"]
		if alertName == "" {
			alertName = "prometheus-alert"
		}
		service := serviceFromLabels(labels, labels["namespace"])
		title := alert.Annotations["summary"]
		if title == "" {
			title = alertName
		}
		description := alert.Annotations["description"]
		if description == "" {
			description = title
		}
		severityLabel := labels["severity"]
		if severityLabel == "" {
			severityLabel = labels["priority"]
		}
		firedAt := alert.StartsAt
		if firedAt.IsZero() {
			firedAt = time.Now().UTC()
		}
		out = append(out, model.IncomingAlert{
			Source:      domain.AlertSourcePrometheus,
			AlertName:   alertName,
			Service:     service,
			Title:       title,
			Description: description,
			Severity:    mapSeverity(severityLabel),
			FiredAt:     firedAt.UTC(),
			Labels:      labels,
			Resolved:    strings.EqualFold(alert.Status, "resolved"),
		})
	}
	return out
}

// DynatracePayload is Dynatrace problem notification payload.
type DynatracePayload struct {
	ProblemID   string `json:"problemId"`
	Title       string `json:"title"`
	ImpactLevel string `json:"impactLevel"`
	Status      string `json:"status"`
	ProblemURL  string `json:"problemUrl"`
	Tags        []struct {
		Context string `json:"context"`
		Key     string `json:"key"`
		Value   string `json:"value"`
	} `json:"tags"`
}

// FromDynatrace normalizes Dynatrace webhook payload.
func FromDynatrace(payload DynatracePayload) []model.IncomingAlert {
	labels := map[string]string{"problemId": payload.ProblemID}
	service := "dynatrace-monitored"
	for _, tag := range payload.Tags {
		key := strings.ToLower(tag.Key)
		if key == "service" || key == "application" {
			service = tag.Value
		}
		labels[tag.Key] = tag.Value
	}
	return []model.IncomingAlert{{
		Source:      domain.AlertSourceDynatrace,
		AlertName:   payload.ProblemID,
		Service:     service,
		Title:       payload.Title,
		Description: payload.ProblemURL,
		Severity:    mapSeverity(payload.ImpactLevel),
		FiredAt:     time.Now().UTC(),
		Labels:      labels,
		Resolved:    strings.EqualFold(payload.Status, "RESOLVED"),
	}}
}

// CloudWatchPayload is AWS CloudWatch alarm notification.
type CloudWatchPayload struct {
	AlarmName        string `json:"AlarmName"`
	AlarmDescription string `json:"AlarmDescription"`
	NewStateValue    string `json:"NewStateValue"`
	NewStateReason   string `json:"NewStateReason"`
	StateChangeTime  string `json:"StateChangeTime"`
	Trigger          struct {
		Namespace string `json:"Namespace"`
	} `json:"Trigger"`
}

// FromCloudWatch normalizes AWS CloudWatch alarm payload.
func FromCloudWatch(payload CloudWatchPayload) []model.IncomingAlert {
	service := payload.Trigger.Namespace
	if service == "" {
		service = "aws-cloudwatch"
	}
	firedAt := time.Now().UTC()
	if payload.StateChangeTime != "" {
		if parsed, err := time.Parse(time.RFC3339, payload.StateChangeTime); err == nil {
			firedAt = parsed.UTC()
		}
	}
	severity := domain.IncidentSeverityP2
	if strings.EqualFold(payload.NewStateValue, "OK") {
		severity = domain.IncidentSeverityP4
	}
	return []model.IncomingAlert{{
		Source:      domain.AlertSourceCloud,
		AlertName:   payload.AlarmName,
		Service:     service,
		Title:       payload.AlarmName,
		Description: firstNonEmpty(payload.NewStateReason, payload.AlarmDescription, payload.AlarmName),
		Severity:    severity,
		FiredAt:     firedAt,
		Labels: map[string]string{
			"state": payload.NewStateValue,
		},
		Resolved: strings.EqualFold(payload.NewStateValue, "OK"),
	}}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
