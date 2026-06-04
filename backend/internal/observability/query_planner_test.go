package observability

import "testing"

func TestBuildPlannerStepsCrossSignal(t *testing.T) {
	steps := buildPlannerSteps(UnifiedQueryRequest{Query: "error", From: "all"})
	if len(steps) < 6 {
		t.Fatalf("expected cross-signal plan, got %d steps: %+v", len(steps), steps)
	}
	signals := make(map[string]bool)
	for _, s := range steps {
		if s.Signal == "metrics" || s.Signal == "logs" || s.Signal == "traces" || s.Signal == "events" {
			signals[s.Signal] = true
		}
	}
	for _, want := range []string{"metrics", "logs", "traces", "events"} {
		if !signals[want] {
			t.Fatalf("missing signal step %s in %+v", want, steps)
		}
	}
	if !hasStep(steps, "merge", "correlate_by_timestamp") {
		t.Fatal("expected merge/correlate step")
	}
	if !hasStep(steps, "guard", "cardinality_limit") {
		t.Fatal("expected cardinality guard step")
	}
}

func TestBuildPlannerStepsWithServiceAndTrace(t *testing.T) {
	steps := buildPlannerSteps(UnifiedQueryRequest{
		Query: "timeout", From: "logs", Service: "payment-service", TraceID: "trace-demo-04",
	})
	if !hasStep(steps, "filter", "service_scope:payment-service") {
		t.Fatal("expected service filter")
	}
	if !hasStep(steps, "join", "trace_id=trace-demo-04") {
		t.Fatal("expected trace join")
	}
	if len(resolvePlannerSignals("logs")) != 1 {
		t.Fatal("logs should be single signal")
	}
	if hasStep(steps, "merge", "correlate_by_timestamp") {
		t.Fatal("single-signal plan should not merge")
	}
}

func TestPlanUnifiedQueryPlannerPopulated(t *testing.T) {
	mem := NewStore()
	res := PlanUnifiedQuery(mem, UnifiedQueryRequest{Query: "error", From: "all"})
	if len(res.Planner) < 6 {
		t.Fatalf("planner on response: got %d steps", len(res.Planner))
	}
}

func hasStep(steps []QueryPlannerStep, signal, action string) bool {
	for _, s := range steps {
		if s.Signal == signal && s.Action == action {
			return true
		}
	}
	return false
}
