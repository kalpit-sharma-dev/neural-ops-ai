package normalizer

import (
	"testing"
	"time"

	"github.com/neuralops/platform/internal/domain"
)

func TestFromPrometheusMapsSeverityAndService(t *testing.T) {
	payload := PrometheusPayload{
		Alerts: []struct {
			Status       string            `json:"status"`
			Labels       map[string]string `json:"labels"`
			Annotations  map[string]string `json:"annotations"`
			StartsAt     time.Time         `json:"startsAt"`
			EndsAt       time.Time         `json:"endsAt"`
			GeneratorURL string            `json:"generatorURL"`
		}{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighErrorRate",
					"severity":  "critical",
					"service":   "upi-service",
				},
				Annotations: map[string]string{
					"summary":     "High error rate",
					"description": "Errors above threshold",
				},
			},
		},
	}

	alerts := FromPrometheus(payload)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].Severity != domain.IncidentSeverityP1 {
		t.Fatalf("expected P1 severity, got %s", alerts[0].Severity)
	}
	if alerts[0].Service != "upi-service" {
		t.Fatalf("expected upi-service, got %s", alerts[0].Service)
	}
}

func TestFingerprintStable(t *testing.T) {
	a := Fingerprint(domain.AlertSourcePrometheus, "HighErrorRate", "upi-service", map[string]string{"env": "prod"})
	b := Fingerprint(domain.AlertSourcePrometheus, "HighErrorRate", "upi-service", map[string]string{"env": "prod"})
	if a != b {
		t.Fatalf("fingerprints should match")
	}
}
