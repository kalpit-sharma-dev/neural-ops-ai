package observability

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *Store) seedAI() {
	now := time.Now().UTC()
	expID := "exp-checkout-latency"
	rca := RCAResponse{
		ExplanationID: expID,
		IncidentID:    "inc-1",
		Summary:       "Payment latency spike driven by ledger connection pool saturation after deploy v2.4.1.",
		Confidence:    0.91,
		GeneratedAt:   now,
		Root: AIExplanationNode{
			ID: "root", Label: "Checkout SLO breach", Confidence: 0.91,
			Evidence: []AIExplanationEvidence{
				{Signal: "slo", Ref: "slo-checkout-availability", Detail: "Error budget burn 3.2x baseline"},
			},
			Children: []AIExplanationNode{
				{
					ID: "n1", Label: "payment-service p95 latency +240%", Confidence: 0.88,
					Evidence: []AIExplanationEvidence{
						{Signal: "metrics", Ref: "payment.latency.p95", Detail: "p95 842ms vs 250ms SLO"},
						{Signal: "traces", Ref: "trace-demo-04", Detail: "Downstream ledger spans dominate duration"},
					},
					Children: []AIExplanationNode{
						{
							ID: "n2", Label: "ledger-service pool exhausted", Confidence: 0.86,
							Evidence: []AIExplanationEvidence{
								{Signal: "logs", Ref: "log-ledger-pool", Detail: "connection pool timeout count +420%"},
								{Signal: "deploy", Ref: "deploy-ledger-v241", Detail: "Deploy v2.4.1 correlated at T-12m"},
							},
						},
					},
				},
			},
		},
	}
	s.rcaByIncident["inc-1"] = rca
	s.explanations[expID] = rca
}

// RCAForIncident returns causal analysis for an incident.
func (s *Store) RCAForIncident(incidentID string) (RCAResponse, bool) {
	s.mu.RLock()
	if r, ok := s.rcaByIncident[incidentID]; ok {
		s.mu.RUnlock()
		return r, true
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	if r, ok := s.rcaByIncident[incidentID]; ok {
		return r, true
	}
	expID := "exp-" + uuid.New().String()[:8]
	rca := RCAResponse{
		ExplanationID: expID,
		IncidentID:    incidentID,
		Summary:       fmt.Sprintf("Probable upstream dependency degradation affecting incident %s.", incidentID),
		Confidence:    0.72,
		GeneratedAt:   time.Now().UTC(),
		Root: AIExplanationNode{
			ID: "root", Label: "Service dependency latency", Confidence: 0.72,
			Evidence: []AIExplanationEvidence{
				{Signal: "traces", Ref: "trace-demo-01", Detail: "Error spans concentrated on downstream calls"},
				{Signal: "metrics", Ref: "service.error_rate", Detail: "Error rate elevated vs 1h baseline"},
			},
		},
	}
	s.rcaByIncident[incidentID] = rca
	s.explanations[expID] = rca
	return rca, true
}

// GetAIExplanation returns a stored explanation by id.
func (s *Store) GetAIExplanation(id string) (RCAResponse, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.explanations[id]
	return r, ok
}

// Forecast returns a deterministic capacity/SLO projection.
func (s *Store) Forecast(metric, service, horizon string) AIForecast {
	if horizon == "" {
		horizon = "24h"
	}
	base := 42.0
	switch strings.ToLower(metric) {
	case "cpu.utilization", "cpu":
		base = 68.5
	case "error_rate", "errors":
		base = 2.4
	case "latency.p95", "latency":
		base = 310.0
	case "slo.burn":
		base = 1.8
	}
	if service != "" {
		base += float64(len(service)%7) * 1.3
	}
	hours := 24.0
	if strings.HasSuffix(horizon, "h") {
		fmt.Sscanf(horizon, "%fh", &hours)
	}
	trend := math.Sin(hours/12) * 5
	pred := base + trend
	return AIForecast{
		ID:             "fc-" + uuid.New().String()[:8],
		Metric:         metric,
		Service:        service,
		Horizon:        horizon,
		Prediction:     math.Round(pred*100) / 100,
		LowerBound:     math.Round((pred-8)*100) / 100,
		UpperBound:     math.Round((pred+12)*100) / 100,
		Unit:           forecastUnit(metric),
		Recommendation: forecastRecommendation(metric, pred),
		GeneratedAt:    time.Now().UTC(),
	}
}

func forecastUnit(metric string) string {
	m := strings.ToLower(metric)
	switch {
	case strings.Contains(m, "latency"), strings.Contains(m, "duration"):
		return "ms"
	case strings.Contains(m, "cpu"), strings.Contains(m, "util"):
		return "%"
	case strings.Contains(m, "error"), strings.Contains(m, "rate"):
		return "%"
	case strings.Contains(m, "burn"):
		return "x"
	default:
		return "units"
	}
}

func forecastRecommendation(metric string, pred float64) string {
	if strings.Contains(strings.ToLower(metric), "cpu") && pred > 75 {
		return "Scale payment-service replicas by +2 before peak window."
	}
	if strings.Contains(strings.ToLower(metric), "slo") && pred > 2 {
		return "Freeze non-critical deploys; enable alert suppression only after RCA sign-off."
	}
	return "Monitor trend; no immediate capacity action required."
}

// CreateAutoFixPlan builds a remediation plan with policy evaluation.
func (s *Store) CreateAutoFixPlan(incidentID string) AutoFixPlan {
	s.mu.Lock()
	defer s.mu.Unlock()
	highBlast := false
	steps := []AutoFixStep{
		{ID: "s1", Action: "scale", Description: "Scale payment-service replicas +1", BlastRadius: "low"},
		{ID: "s2", Action: "rollback", Description: "Rollback ledger-service to v2.4.0", BlastRadius: "medium"},
		{ID: "s3", Action: "restart", Description: "Rolling restart ledger connection pool", BlastRadius: "high"},
	}
	for _, st := range steps {
		if st.BlastRadius == "high" {
			highBlast = true
		}
	}
	plan := AutoFixPlan{
		ID:               "plan-" + uuid.New().String()[:8],
		IncidentID:       incidentID,
		Summary:          "Remediate checkout latency via scale + controlled ledger rollback.",
		RequiresApproval: highBlast,
		PolicyPass:       !highBlast,
		Steps:            steps,
		CreatedAt:        time.Now().UTC(),
	}
	if highBlast {
		plan.PolicyPass = false
		plan.PolicyReason = "High blast-radius step requires human approval per tenant AutoFix policy."
		plan.RequiresApproval = true
	}
	s.autofixPlans[plan.ID] = plan
	return plan
}

// ExecuteAutoFix runs an approved plan and records action logs.
func (s *Store) ExecuteAutoFix(planID string, approved bool) (AutoFixActionRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	plan, ok := s.autofixPlans[planID]
	if !ok {
		return AutoFixActionRecord{}, errors.New("plan not found")
	}
	if plan.RequiresApproval && !approved {
		return AutoFixActionRecord{}, errors.New("approval required before execution")
	}
	if !plan.PolicyPass && !approved {
		return AutoFixActionRecord{}, errors.New("policy check failed; set approved=true after review")
	}
	now := time.Now().UTC()
	action := AutoFixActionRecord{
		ID:         "afx-" + uuid.New().String()[:8],
		PlanID:     planID,
		IncidentID: plan.IncidentID,
		Status:     "completed",
		StartedAt:  now,
		FinishedAt: ptrTime(now.Add(4 * time.Second)),
		Logs: []string{
			"policy: evaluated tenant AutoFix guardrails",
			"step s1: scaled payment-service 3 -> 4 replicas",
			"step s2: skipped rollback (approval pending)",
			"verification: checkout p95 improved 18%",
		},
	}
	if plan.RequiresApproval && approved {
		action.Logs = append(action.Logs, "step s2: rolled back ledger-service to v2.4.0")
	}
	s.autofixActions[action.ID] = action
	return action, nil
}

// RollbackAutoFix reverts a completed action.
func (s *Store) RollbackAutoFix(actionID string) (AutoFixActionRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	action, ok := s.autofixActions[actionID]
	if !ok {
		return AutoFixActionRecord{}, errors.New("action not found")
	}
	if action.Status == "rolled_back" {
		return AutoFixActionRecord{}, errors.New("action already rolled back")
	}
	now := time.Now().UTC()
	action.Status = "rolled_back"
	action.FinishedAt = &now
	action.Logs = append(action.Logs,
		"rollback: restored payment-service replica count",
		"rollback: ledger-service left at pre-remediation revision",
	)
	s.autofixActions[actionID] = action
	return action, nil
}

func ptrTime(t time.Time) *time.Time { return &t }

// ListLLMWorkloads returns seeded LLM inference workloads (AI-03 preview).
func (s *Store) ListLLMWorkloads() []LLMWorkload {
	now := time.Now().UTC()
	return []LLMWorkload{
		{ID: "llm-rca", Name: "incident-rca", Model: "gpt-4o", Provider: "azure-openai",
			Requests24h: 420, Tokens24h: 890_000, P95LatencyMs: 2400, ErrorRatePct: 0.2, CostUSD24h: 18.4, UpdatedAt: now},
		{ID: "llm-chat", Name: "ops-copilot", Model: "claude-sonnet", Provider: "bedrock",
			Requests24h: 1200, Tokens24h: 2_100_000, P95LatencyMs: 1800, ErrorRatePct: 0.4, CostUSD24h: 42.1, UpdatedAt: now},
	}
}

// LLMUsageSummary returns aggregate LLM usage for a time window.
func (s *Store) LLMUsageSummary(window string) LLMUsageSummary {
	if window == "" {
		window = "24h"
	}
	var tokens int64
	var cost float64
	for _, w := range s.ListLLMWorkloads() {
		tokens += w.Tokens24h
		cost += w.CostUSD24h
	}
	return LLMUsageSummary{
		Window: window, TotalTokens: tokens, TotalCostUSD: cost,
		TopModel: "claude-sonnet", BudgetPct: math.Min(100, cost/100*100),
	}
}

// LLMTraceSpan is one inference span for LLM observability.
type LLMTraceSpan struct {
	ID           string    `json:"id"`
	WorkloadID   string    `json:"workloadId"`
	Model        string    `json:"model"`
	PromptTokens int       `json:"promptTokens"`
	OutputTokens int       `json:"outputTokens"`
	LatencyMs    float64   `json:"latencyMs"`
	Status       string    `json:"status"`
	StartedAt    time.Time `json:"startedAt"`
}

// LLMWorkloadTraces returns recent inference spans for a workload.
func (s *Store) LLMWorkloadTraces(workloadID string) []LLMTraceSpan {
	now := time.Now().UTC()
	return []LLMTraceSpan{
		{ID: "llm-span-1", WorkloadID: workloadID, Model: "gpt-4o", PromptTokens: 1200, OutputTokens: 340,
			LatencyMs: 2100, Status: "ok", StartedAt: now.Add(-5 * time.Minute)},
		{ID: "llm-span-2", WorkloadID: workloadID, Model: "gpt-4o", PromptTokens: 800, OutputTokens: 120,
			LatencyMs: 980, Status: "ok", StartedAt: now.Add(-12 * time.Minute)},
	}
}
