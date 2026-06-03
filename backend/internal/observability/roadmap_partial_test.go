package observability

import (
	"testing"
)

func TestEvaluateAlertExpression(t *testing.T) {
	tests := []struct {
		expr     string
		service  string
		severity string
		want     bool
	}{
		{"", "payment-service", "P1", true},
		{"service:payment*", "payment-service", "P1", true},
		{"service:ledger*", "payment-service", "P1", false},
		{"severity:P1", "payment-service", "P2", false},
		{"service:payment* AND severity:P1", "payment-service", "P1", true},
		{"service:ledger* OR service:payment*", "payment-service", "P1", true},
	}
	for _, tc := range tests {
		if got := evaluateAlertExpression(tc.expr, tc.service, tc.severity); got != tc.want {
			t.Fatalf("expr %q service=%s severity=%s: got %v want %v", tc.expr, tc.service, tc.severity, got, tc.want)
		}
	}
}

func TestCollectorDiscoveryAndPlatforms(t *testing.T) {
	s := NewStore()
	if len(s.ListDiscoveredServices()) == 0 {
		t.Fatal("expected discovered services")
	}
	if len(s.SupportedPlatforms()) == 0 {
		t.Fatal("expected supported platforms")
	}
}

func TestSignalPoliciesAndSIEM(t *testing.T) {
	s := NewStore()
	policies := s.SignalPolicies("default")
	if len(policies) < 4 {
		t.Fatalf("expected 4 signal policies, got %d", len(policies))
	}
	cfg := s.SIEMConfig("default")
	if cfg.Provider == "" {
		t.Fatal("expected default SIEM provider")
	}
}

func TestBusinessTransaction(t *testing.T) {
	s := NewStore()
	view, ok := s.BusinessTransaction("txn-upi-12345")
	if !ok || view.TxnID != "txn-upi-12345" {
		t.Fatalf("unexpected business transaction: %+v ok=%v", view, ok)
	}
}
