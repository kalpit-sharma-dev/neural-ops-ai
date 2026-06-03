package observability

import "strings"

// QueryPlannerStep is one step in a cross-signal execution plan.
type QueryPlannerStep struct {
	Signal        string `json:"signal"`
	Action        string `json:"action"`
	EstimatedCost int    `json:"estimatedCost"`
}

// PlanUnifiedQuery builds an execution plan then runs the query.
func PlanUnifiedQuery(mem *Store, req UnifiedQueryRequest) UnifiedQueryResponse {
	steps := buildPlannerSteps(req)
	result := ExecuteUnifiedQuery(mem, req)
	result.Planner = steps
	result.Hits, result.Cardinality = applyCardinalityGuard(result.Hits, req)
	result.Count = len(result.Hits)
	return result
}

func buildPlannerSteps(req UnifiedQueryRequest) []QueryPlannerStep {
	from := strings.ToLower(strings.TrimSpace(req.From))
	if from == "" || from == "all" {
		return []QueryPlannerStep{
			{Signal: "metrics", Action: "scan_time_series", EstimatedCost: 2},
			{Signal: "logs", Action: "index_search", EstimatedCost: 4},
			{Signal: "traces", Action: "span_lookup", EstimatedCost: 3},
			{Signal: "events", Action: "event_bus_query", EstimatedCost: 1},
		}
	}
	return []QueryPlannerStep{
		{Signal: from, Action: "primary_scan", EstimatedCost: 3},
	}
}
