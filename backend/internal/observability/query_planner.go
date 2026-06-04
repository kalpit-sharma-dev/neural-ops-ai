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
	steps := make([]QueryPlannerStep, 0, 12)
	if svc := strings.TrimSpace(req.Service); svc != "" {
		steps = append(steps, QueryPlannerStep{
			Signal: "filter", Action: "service_scope:" + svc, EstimatedCost: 1,
		})
	}
	if tid := strings.TrimSpace(req.TraceID); tid != "" {
		steps = append(steps, QueryPlannerStep{
			Signal: "join", Action: "trace_id=" + tid, EstimatedCost: 2,
		})
	}
	if xid := strings.TrimSpace(req.TxnID); xid != "" {
		steps = append(steps, QueryPlannerStep{
			Signal: "join", Action: "txn_id=" + xid, EstimatedCost: 2,
		})
	}
	signals := resolvePlannerSignals(req.From)
	for _, sig := range signals {
		steps = append(steps, signalPlannerSteps(sig)...)
	}
	if len(signals) > 1 {
		steps = append(steps, QueryPlannerStep{
			Signal: "merge", Action: "correlate_by_timestamp", EstimatedCost: 3,
		})
	}
	steps = append(steps, QueryPlannerStep{
		Signal: "guard", Action: "cardinality_limit", EstimatedCost: 1,
	})
	return steps
}

func resolvePlannerSignals(from string) []string {
	switch strings.ToLower(strings.TrimSpace(from)) {
	case "", "all":
		return []string{"metrics", "logs", "traces", "events"}
	case "logs", "metrics", "traces", "events":
		return []string{strings.ToLower(strings.TrimSpace(from))}
	default:
		return []string{strings.ToLower(strings.TrimSpace(from))}
	}
}

func signalPlannerSteps(sig string) []QueryPlannerStep {
	switch sig {
	case "logs":
		return []QueryPlannerStep{{Signal: "logs", Action: "elasticsearch_search", EstimatedCost: 4}}
	case "metrics":
		return []QueryPlannerStep{{Signal: "metrics", Action: "promql_range_query", EstimatedCost: 3}}
	case "traces":
		return []QueryPlannerStep{{Signal: "traces", Action: "clickhouse_span_lookup", EstimatedCost: 3}}
	case "events":
		return []QueryPlannerStep{{Signal: "events", Action: "kafka_time_range", EstimatedCost: 2}}
	default:
		return []QueryPlannerStep{{Signal: sig, Action: "primary_scan", EstimatedCost: 3}}
	}
}
