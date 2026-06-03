package observability

import (
	"strings"
	"time"
)

// AlertContext enriches alerts with runbook and pivot links.
type AlertContext struct {
	RunbookURL     string `json:"runbookUrl,omitempty"`
	Owner          string `json:"owner,omitempty"`
	TopologyLink   string `json:"topologyLink,omitempty"`
	TracePivotLink string `json:"tracePivotLink,omitempty"`
	LogPivotLink   string `json:"logPivotLink,omitempty"`
}

// AlertEvaluationResult is the outcome of routing policy evaluation.
type AlertEvaluationResult struct {
	PolicyID    string           `json:"policyId"`
	Matched     bool             `json:"matched"`
	Suppressed  bool             `json:"suppressed"`
	Routes      []AlertPolicyRoute `json:"routes,omitempty"`
	Context     AlertContext     `json:"context"`
	DedupeKey   string           `json:"dedupeKey,omitempty"`
	Explanation string           `json:"explanation,omitempty"`
	TriggeredAt time.Time        `json:"triggeredAt"`
}

// EvaluateAlertPolicy matches service/severity and applies suppressions with AND/OR on expression.
func EvaluateAlertPolicy(mem *Store, policyID, service, severity string) AlertEvaluationResult {
	res := AlertEvaluationResult{PolicyID: policyID, TriggeredAt: time.Now().UTC()}
	policies := mem.ListAlertPolicies()
	var policy *AlertPolicy
	for i := range policies {
		if policies[i].ID == policyID {
			policy = &policies[i]
			break
		}
	}
	if policy == nil {
		res.Explanation = "policy not found"
		return res
	}
	if !policy.Enabled {
		res.Explanation = "policy disabled"
		return res
	}
	if !matchPattern(policy.ServicePattern, service) {
		res.Explanation = "service pattern mismatch"
		return res
	}
	if policy.Severity != "" && !strings.EqualFold(policy.Severity, severity) {
		res.Explanation = "severity mismatch"
		return res
	}
	if policy.Expression != "" && !evaluateAlertExpression(policy.Expression, service, severity) {
		res.Explanation = "expression not satisfied"
		return res
	}
	for _, sup := range mem.ListAlertSuppressions() {
		if time.Now().Before(sup.EndsAt) && time.Now().After(sup.StartsAt) && matchPattern(sup.ServicePattern, service) {
			res.Suppressed = true
			res.Explanation = "suppression active: " + sup.Reason
			return res
		}
	}
	res.Matched = true
	res.Routes = policy.Routes
	res.Context = policy.Context
	if res.Context.RunbookURL == "" {
		res.Context = defaultAlertContext(service)
	}
	res.DedupeKey = service + ":" + severity + ":" + policy.ID
	res.Explanation = "policy matched; notifications scheduled"
	return res
}

func defaultAlertContext(service string) AlertContext {
	return AlertContext{
		RunbookURL:     "/docs/RUNBOOK.md#" + service,
		Owner:          "oncall@" + service + ".neuralops.ai",
		TopologyLink:   "/service-map?service=" + service,
		TracePivotLink: "/traces?service=" + service,
		LogPivotLink:   "/logs?service=" + service,
	}
}

func matchPattern(pattern, value string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}
	re := strings.ReplaceAll(pattern, "*", "")
	return strings.Contains(strings.ToLower(value), strings.ToLower(re))
}

// evaluateAlertExpression supports AND/OR with service: and severity: clauses (ALR-01).
func evaluateAlertExpression(expr, service, severity string) bool {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return true
	}
	parts := splitAlertExpr(expr, " OR ")
	for _, orPart := range parts {
		andParts := splitAlertExpr(orPart, " AND ")
		ok := true
		for _, clause := range andParts {
			clause = strings.TrimSpace(clause)
			if !matchAlertClause(clause, service, severity) {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

func splitAlertExpr(expr, sep string) []string {
	return strings.Split(strings.ToUpper(expr), strings.ToUpper(sep))
}

func matchAlertClause(clause, service, severity string) bool {
	clause = strings.TrimSpace(clause)
	lower := strings.ToLower(clause)
	switch {
	case strings.HasPrefix(lower, "service:"):
		return matchPattern(strings.TrimSpace(clause[8:]), service)
	case strings.HasPrefix(lower, "severity:"):
		return strings.EqualFold(strings.TrimSpace(clause[9:]), severity)
	case strings.HasPrefix(lower, "service="):
		return matchPattern(strings.TrimSpace(clause[8:]), service)
	case strings.HasPrefix(lower, "severity="):
		return strings.EqualFold(strings.TrimSpace(clause[9:]), severity)
	default:
		return strings.Contains(strings.ToLower(service), strings.ToLower(clause))
	}
}
