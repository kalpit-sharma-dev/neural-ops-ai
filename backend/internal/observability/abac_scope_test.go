package observability_test

import (
	"context"
	"testing"

	"github.com/neuralops/platform/internal/observability"
)

func TestABACPaymentsVsPlatformDepartment(t *testing.T) {
	mem := observability.NewStore()
	mem.PutABACPolicy(observability.ABACPolicy{
		Enabled: true,
		Rules: []observability.ABACPolicyRule{
			{ID: "pay-read", Effect: "allow", Action: "read", Resource: "/api/v1/logs/*", Condition: "user.department == payments"},
			{ID: "plat-read", Effect: "allow", Action: "read", Resource: "/api/v1/logs/*", Condition: "user.department == platform"},
			{ID: "deny-pii-export", Effect: "deny", Action: "post", Resource: "/api/v1/exports/*", Condition: "user.clearance < 3"},
		},
	})
	gs := observability.NewGovernanceService(nil, mem)
	ctx := context.Background()

	ok, _ := gs.EvaluateABAC(ctx, "default", "DEVELOPER", "ops@payments.acme.com", "GET", "/api/v1/logs/search")
	if !ok {
		t.Fatal("payments department should read logs")
	}

	ok, _ = gs.EvaluateABAC(ctx, "default", "DEVELOPER", "dev@neuralops.ai", "GET", "/api/v1/logs/search")
	if !ok {
		t.Fatal("platform department (neuralops email) should read logs")
	}

	ok, reason := gs.EvaluateABAC(ctx, "default", "VIEWER", "guest@external.io", "POST", "/api/v1/exports/logs")
	if ok {
		t.Fatal("low clearance export should be denied")
	}
	if reason == "" {
		t.Fatal("expected deny reason")
	}
}

func TestABACFinOpsPaymentsScopeRead(t *testing.T) {
	mem := observability.NewStore()
	mem.PutABACPolicy(observability.ABACPolicy{
		Enabled: true,
		Rules: []observability.ABACPolicyRule{
			{ID: "finops-payments", Effect: "allow", Action: "read", Resource: "/api/v1/finops/*"},
		},
	})
	gs := observability.NewGovernanceService(nil, mem)
	ctx := context.Background()

	ok, _ := gs.EvaluateABAC(ctx, "default", "DEVELOPER", "analyst@payments.acme.com", "GET", "/api/v1/finops/costs/breakdown")
	if !ok {
		t.Fatal("expected finops read allowed")
	}
}
