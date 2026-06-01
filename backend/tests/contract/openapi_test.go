package contract_test

import (
	"os"
	"strings"
	"testing"
)

// TestOpenAPIPathsRegistered ensures depth-parity routes exist in OpenAPI spec (Pact-style contract).
func TestOpenAPIPathsRegistered(t *testing.T) {
	raw, err := os.ReadFile("../../../docs/openapi/observability-v1.yaml")
	if err != nil {
		raw, err = os.ReadFile("../../docs/openapi/observability-v1.yaml")
	}
	if err != nil {
		t.Skip("openapi spec not found:", err)
	}
	spec := string(raw)
	required := []string{
		"/api/v1/notebooks/{id}/execute:",
		"/api/v1/integrations/{id}/oauth/start:",
		"/api/v1/integrations/{id}/disconnect:",
		"/api/v1/admin/oncall/{id}/sync-pagerduty:",
		"/api/v1/mobile/push/register:",
	}
	for _, path := range required {
		if !strings.Contains(spec, path) {
			t.Fatalf("openapi missing path %s", path)
		}
	}
}
