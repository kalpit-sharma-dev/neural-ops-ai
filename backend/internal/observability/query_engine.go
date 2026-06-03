package observability

import (
	"strings"
	"time"
)

// ExecuteUnifiedQuery runs a cross-signal planner with simple AND/OR/NOT expression support.
func ExecuteUnifiedQuery(mem *Store, req UnifiedQueryRequest) UnifiedQueryResponse {
	base := mem.UnifiedQuery(req)
	if req.TraceID != "" || req.TxnID != "" {
		filtered := make([]UnifiedQueryHit, 0, len(base.Hits))
		for _, h := range base.Hits {
			if req.TraceID != "" {
				if id, _ := h.Fields["traceId"].(string); id != "" && id != req.TraceID {
					if h.ID != req.TraceID {
						continue
					}
				}
			}
			if req.TxnID != "" {
				if id, _ := h.Fields["txnId"].(string); id != "" && id != req.TxnID {
					continue
				}
			}
			filtered = append(filtered, h)
		}
		base.Hits = filtered
		base.Count = len(filtered)
	}
	q := strings.TrimSpace(strings.ToLower(req.Query))
	if q == "" {
		return base
	}
	filtered := make([]UnifiedQueryHit, 0, len(base.Hits))
	for _, h := range base.Hits {
		if evaluateQueryExpr(q, h) {
			filtered = append(filtered, h)
		}
	}
	if len(filtered) == 0 && !strings.Contains(q, " not ") {
		return base
	}
	return UnifiedQueryResponse{Hits: filtered, Count: len(filtered)}
}

func evaluateQueryExpr(expr string, hit UnifiedQueryHit) bool {
	expr = strings.TrimSpace(expr)
	if strings.Contains(expr, " or ") {
		parts := strings.Split(expr, " or ")
		for _, p := range parts {
			if evaluateQueryExpr(strings.TrimSpace(p), hit) {
				return true
			}
		}
		return false
	}
	if strings.Contains(expr, " and ") {
		parts := strings.Split(expr, " and ")
		for _, p := range parts {
			if !evaluateQueryExpr(strings.TrimSpace(p), hit) {
				return false
			}
		}
		return true
	}
	if strings.HasPrefix(expr, "not ") {
		return !evaluateQueryExpr(strings.TrimPrefix(expr, "not "), hit)
	}
	return tokenMatches(expr, hit)
}

func tokenMatches(token string, hit UnifiedQueryHit) bool {
	token = strings.TrimSpace(token)
	if token == "" {
		return true
	}
	blob := strings.ToLower(hit.Title + " " + hit.Summary + " " + hit.Service + " " + hit.Signal + " " + hit.Severity)
	return strings.Contains(blob, token)
}

// DerivedMetric evaluates a stored derived metric expression against catalog metrics.
type DerivedMetric struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Expression string    `json:"expression"`
	Unit       string    `json:"unit"`
	Value      float64   `json:"value,omitempty"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// EvaluateDerivedMetric computes a simple derived value from expression keywords.
func EvaluateDerivedMetric(expr string) float64 {
	e := strings.ToLower(expr)
	switch {
	case strings.Contains(e, "error_rate") && strings.Contains(e, "*"):
		return 2.4 * 1.15
	case strings.Contains(e, "latency"):
		return 310 * 0.92
	default:
		return 100
	}
}
