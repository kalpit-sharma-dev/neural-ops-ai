package contract_test

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// gatewayPaths are registered first-class gateway routes (Pact consumer contract).
var gatewayPaths = []string{
	"/api/v1/info",
	"/api/v1/auth/config",
	"/api/v1/auth/oidc/start",
	"/api/v1/dashboard/overview",
	"/api/v1/chat/query",
	"/api/v1/search/logs",
	"/api/v1/metrics/catalog",
	"/api/v1/metrics/promql",
	"/api/v1/apm/retention",
	"/api/v1/apm/profiles",
	"/api/v1/integrations",
	"/api/v1/integrations/{id}/oauth/start",
	"/api/v1/cloud/metrics",
	"/api/v1/notebooks/{id}/execute",
	"/api/v1/admin/sso",
	"/api/v1/admin/oncall/{id}/sync-pagerduty",
	"/api/v1/mobile/push/register",
}

func loadOpenAPISpec(t *testing.T) string {
	t.Helper()
	for _, p := range []string{
		"../../../docs/openapi/gateway-v1.yaml",
		"../../docs/openapi/gateway-v1.yaml",
		"../docs/openapi/gateway-v1.yaml",
	} {
		raw, err := os.ReadFile(p)
		if err == nil {
			return string(raw)
		}
	}
	t.Fatal("gateway openapi spec not found")
	return ""
}

func normalizeOpenAPIPath(p string) string {
	p = strings.TrimPrefix(p, "/api/v1")
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	re := regexp.MustCompile(`\{[^}]+\}`)
	return re.ReplaceAllString(p, "{id}")
}

// TestGatewayOpenAPIPactDiff ensures consumer paths exist in OpenAPI spec (Pact-style).
func TestGatewayOpenAPIPactDiff(t *testing.T) {
	spec := loadOpenAPISpec(t)
	for _, path := range gatewayPaths {
		openAPIPath := normalizeOpenAPIPath(path)
		if !strings.Contains(spec, openAPIPath+":") {
			t.Fatalf("gateway pact: path %s (openapi %s) missing from spec", path, openAPIPath)
		}
	}
}

// TestGatewayOpenAPIPathCount guards against spec drift.
func TestGatewayOpenAPIPathCount(t *testing.T) {
	spec := loadOpenAPISpec(t)
	count := strings.Count(spec, "\n  /")
	if count < 40 {
		t.Fatalf("expected >= 40 gateway paths in openapi, got %d", count)
	}
}
