package contract_test

import (
	"strings"
	"testing"
)

// bankProductionPaths are Wave 2–6 bank production APIs documented in gateway-v1.yaml.
var bankProductionPaths = []string{
	"/finops/reconciliation:",
	"/finops/audit/export:",
	"/ai/llm/workloads:",
	"/ai/llm/usage:",
}

// TestBankProductionOpenAPIPaths ensures bank wave endpoints exist in the canonical spec.
func TestBankProductionOpenAPIPaths(t *testing.T) {
	spec := loadOpenAPISpec(t)
	for _, path := range bankProductionPaths {
		if !strings.Contains(spec, path) {
			t.Fatalf("openapi missing bank path %s", path)
		}
	}
	if !strings.Contains(spec, "/admin/msp/tenants:") {
		t.Fatal("openapi missing /admin/msp/tenants")
	}
	if !strings.Contains(spec, "/security/vulnerabilities/batch:") {
		t.Fatal("openapi missing /security/vulnerabilities/batch")
	}
	// POST on MSP tenants
	if !strings.Contains(spec, "Register MSP child tenant") {
		t.Fatal("openapi missing POST summary for MSP tenants")
	}
}
