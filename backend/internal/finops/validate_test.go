package finops

import (
	"testing"
)

func TestValidateFilter(t *testing.T) {
	if err := ValidateFilter("payments", "scope"); err != nil {
		t.Fatalf("expected valid scope: %v", err)
	}
	if err := ValidateFilter("all", "scope"); err != nil {
		t.Fatalf("all should be valid: %v", err)
	}
	if err := ValidateFilter("'; DROP TABLE--", "scope"); err == nil {
		t.Fatal("expected rejection for injection attempt")
	}
}

func TestMaskAccountID(t *testing.T) {
	if got := MaskAccountID("123456789012"); got == "123456789012" {
		t.Fatalf("expected masked account id, got %s", got)
	}
	if MaskAccountID("12") != "****" {
		t.Fatal("short ids should fully mask")
	}
}
