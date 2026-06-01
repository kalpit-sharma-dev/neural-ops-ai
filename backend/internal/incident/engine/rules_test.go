package engine

import (
	"testing"
	"time"

	"github.com/neuralops/platform/internal/domain"
)

func TestEvaluateSeverityP1HighErrorRate(t *testing.T) {
	severity := EvaluateSeverity("upi-service", ServiceMetrics{TotalLogs: 100, ErrorLogs: 60}, nil)
	if severity != domain.IncidentSeverityP1 {
		t.Fatalf("expected P1, got %s", severity)
	}
}

func TestEvaluateSeverityP4Anomaly(t *testing.T) {
	severity := EvaluateSeverity("ledger-service", ServiceMetrics{AnomalyScore: 0.5}, nil)
	if severity != domain.IncidentSeverityP4 {
		t.Fatalf("expected P4, got %s", severity)
	}
}

func TestBuildFingerprintStable(t *testing.T) {
	at := mustParseTime("2026-05-31T10:07:00Z")
	a := BuildFingerprint("upi-service", domain.ErrorCategoryTimeout, at, 5*time.Minute)
	b := BuildFingerprint("upi-service", domain.ErrorCategoryTimeout, at.Add(2*time.Minute), 5*time.Minute)
	if a != b {
		t.Fatalf("expected same fingerprint in window")
	}
}

func mustParseTime(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return t
}
