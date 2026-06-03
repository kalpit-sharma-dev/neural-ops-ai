package repository

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/finops/domain"
	"github.com/neuralops/platform/internal/finops/optimize"
)

// MemoryStore is an in-memory FinOps repository with seeded defaults.
type MemoryStore struct {
	mu          sync.RWMutex
	lineItems   map[string][]domain.CostLineItem
	rules       map[string][]domain.AllocationRule
	budgets     map[string][]domain.Budget
	anomalies   map[string][]domain.Anomaly
	recs        map[string][]domain.Recommendation
	snapshots   map[string][]domain.IngestSnapshot
	imports     map[string][]domain.ImportBatch
	splits      map[string][]domain.SharedSplitRule
	tagSuggest  map[string][]domain.TagSuggestion
	sensitivity map[string]map[string]domain.AnomalySensitivity
	statements  map[string][]domain.ChargebackStatement
	scenarios   map[string][]domain.ScenarioResult
}

// NewMemoryStore creates a seeded in-memory store.
func NewMemoryStore() *MemoryStore {
	m := &MemoryStore{
		lineItems: map[string][]domain.CostLineItem{},
		rules:     map[string][]domain.AllocationRule{},
		budgets:     map[string][]domain.Budget{},
		anomalies:   map[string][]domain.Anomaly{},
		recs:        map[string][]domain.Recommendation{},
		snapshots:   map[string][]domain.IngestSnapshot{},
		imports:     map[string][]domain.ImportBatch{},
		splits:      map[string][]domain.SharedSplitRule{},
		tagSuggest:  map[string][]domain.TagSuggestion{},
		sensitivity: map[string]map[string]domain.AnomalySensitivity{},
		statements:  map[string][]domain.ChargebackStatement{},
		scenarios:   map[string][]domain.ScenarioResult{},
	}
	m.seedDefaults("default")
	return m
}

func (m *MemoryStore) seedDefaults(tenantID string) {
	now := time.Now().UTC()
	m.rules[tenantID] = []domain.AllocationRule{
		{ID: "ar-1", Name: "Team from tag", Dimension: "team", TagKey: "team", Priority: 10, Enabled: true, CreatedAt: now, UpdatedAt: now},
		{ID: "ar-2", Name: "Env from tag", Dimension: "environment", TagKey: "env", Priority: 20, Enabled: true, CreatedAt: now, UpdatedAt: now},
	}
	m.budgets[tenantID] = []domain.Budget{
		{ID: "bud-1", Name: "Payments monthly", ScopeType: "team", ScopeValue: "payments", Period: "monthly", AmountUSD: 12000, Thresholds: []int{50, 80, 100}, NotifyPolicyID: "ap-finops", CreatedAt: now, UpdatedAt: now},
		{ID: "bud-2", Name: "Platform monthly", ScopeType: "team", ScopeValue: "platform", Period: "monthly", AmountUSD: 25000, Thresholds: []int{50, 80, 100}, CreatedAt: now, UpdatedAt: now},
	}
	m.splits[tenantID] = []domain.SharedSplitRule{
		{ID: "split-1", Name: "NAT gateway shared", ResourcePattern: "nat", Mode: "proportional", Enabled: true, CreatedAt: now,
			Targets: []domain.SplitTarget{{Team: "payments", Weight: 0.6}, {Team: "platform", Weight: 0.4}}},
	}
}

func (m *MemoryStore) SaveLineItems(_ context.Context, tenantID string, items []domain.CostLineItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing := m.lineItems[tenantID]
	m.lineItems[tenantID] = dedupeItems(append(existing, items...))
	return nil
}

func (m *MemoryStore) ReplaceLineItems(_ context.Context, tenantID string, items []domain.CostLineItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lineItems[tenantID] = items
	return nil
}

func dedupeItems(items []domain.CostLineItem) []domain.CostLineItem {
	key := func(i domain.CostLineItem) string {
		return i.Provider + "|" + i.ResourceID + "|" + i.BillingPeriod.Format("2006-01-02") + "|" + i.UsageType
	}
	seen := map[string]domain.CostLineItem{}
	for _, i := range items {
		seen[key(i)] = i
	}
	out := make([]domain.CostLineItem, 0, len(seen))
	for _, v := range seen {
		out = append(out, v)
	}
	return out
}

func (m *MemoryStore) ListLineItems(_ context.Context, tenantID string, from, to time.Time, provider, scope string) ([]domain.CostLineItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]domain.CostLineItem, 0)
	for _, item := range m.lineItems[tenantID] {
		if !from.IsZero() && item.IngestedAt.Before(from) {
			continue
		}
		if !to.IsZero() && item.IngestedAt.After(to) {
			continue
		}
		if provider != "" && item.Provider != provider {
			continue
		}
		if scope != "" && scope != "all" && item.Team != scope {
			continue
		}
		out = append(out, item)
	}
	return out, nil
}

func (m *MemoryStore) SaveIngestSnapshot(_ context.Context, tenantID string, s domain.IngestSnapshot) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.snapshots[tenantID] = append(m.snapshots[tenantID], s)
	return nil
}

func (m *MemoryStore) ListAllocationRules(_ context.Context, tenantID string) ([]domain.AllocationRule, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]domain.AllocationRule(nil), m.rules[tenantID]...), nil
}

func (m *MemoryStore) CreateAllocationRule(_ context.Context, tenantID string, r domain.AllocationRule) (domain.AllocationRule, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r.ID == "" {
		r.ID = "ar-" + uuid.NewString()[:8]
	}
	now := time.Now().UTC()
	r.CreatedAt, r.UpdatedAt = now, now
	m.rules[tenantID] = append(m.rules[tenantID], r)
	return r, nil
}

func (m *MemoryStore) UpdateAllocationRule(_ context.Context, tenantID string, r domain.AllocationRule) (domain.AllocationRule, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, x := range m.rules[tenantID] {
		if x.ID == r.ID {
			r.UpdatedAt = time.Now().UTC()
			m.rules[tenantID][i] = r
			return r, nil
		}
	}
	return domain.AllocationRule{}, errors.New("rule not found")
}

func (m *MemoryStore) DeleteAllocationRule(_ context.Context, tenantID, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	rules := m.rules[tenantID]
	for i, x := range rules {
		if x.ID == id {
			m.rules[tenantID] = append(rules[:i], rules[i+1:]...)
			return nil
		}
	}
	return errors.New("rule not found")
}

func (m *MemoryStore) ListBudgets(_ context.Context, tenantID string) ([]domain.Budget, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]domain.Budget(nil), m.budgets[tenantID]...), nil
}

func (m *MemoryStore) CreateBudget(_ context.Context, tenantID string, b domain.Budget) (domain.Budget, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if b.ID == "" {
		b.ID = "bud-" + uuid.NewString()[:8]
	}
	now := time.Now().UTC()
	b.CreatedAt, b.UpdatedAt = now, now
	m.budgets[tenantID] = append(m.budgets[tenantID], b)
	return b, nil
}

func (m *MemoryStore) UpdateBudget(_ context.Context, tenantID string, b domain.Budget) (domain.Budget, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, x := range m.budgets[tenantID] {
		if x.ID == b.ID {
			b.UpdatedAt = time.Now().UTC()
			m.budgets[tenantID][i] = b
			return b, nil
		}
	}
	return domain.Budget{}, errors.New("budget not found")
}

func (m *MemoryStore) DeleteBudget(_ context.Context, tenantID, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, x := range m.budgets[tenantID] {
		if x.ID == id {
			m.budgets[tenantID] = append(m.budgets[tenantID][:i], m.budgets[tenantID][i+1:]...)
			return nil
		}
	}
	return errors.New("budget not found")
}

func (m *MemoryStore) ListAnomalies(_ context.Context, tenantID, scope, severity, status string) ([]domain.Anomaly, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := filterAnomalies(m.anomalies[tenantID], scope, severity, status)
	sort.Slice(out, func(i, j int) bool { return out[i].DetectedAt.After(out[j].DetectedAt) })
	return out, nil
}

func filterAnomalies(list []domain.Anomaly, scope, severity, status string) []domain.Anomaly {
	out := make([]domain.Anomaly, 0, len(list))
	for _, a := range list {
		if scope != "" && a.Scope != scope {
			continue
		}
		if severity != "" && !strings.EqualFold(a.Severity, severity) {
			continue
		}
		if status != "" && a.Status != status {
			continue
		}
		out = append(out, a)
	}
	return out
}

func (m *MemoryStore) GetAnomaly(_ context.Context, tenantID, id string) (domain.Anomaly, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, a := range m.anomalies[tenantID] {
		if a.ID == id {
			return a, nil
		}
	}
	return domain.Anomaly{}, errors.New("anomaly not found")
}

func (m *MemoryStore) SaveAnomaly(_ context.Context, tenantID string, a domain.Anomaly) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.anomalies[tenantID] = append(m.anomalies[tenantID], a)
	return nil
}

func (m *MemoryStore) UpdateAnomalyFeedback(_ context.Context, tenantID, id, feedback string) (domain.Anomaly, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, a := range m.anomalies[tenantID] {
		if a.ID == id {
			a.Feedback = feedback
			if feedback == "false_positive" {
				a.Status = "dismissed"
			} else if feedback == "confirm" {
				a.Status = "confirmed"
			} else if feedback == "expected" {
				a.Status = "expected"
			}
			m.anomalies[tenantID][i] = a
			return a, nil
		}
	}
	return domain.Anomaly{}, errors.New("anomaly not found")
}

func (m *MemoryStore) ListRecommendations(_ context.Context, tenantID, recType, scope, status string) ([]domain.Recommendation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]domain.Recommendation, 0)
	for _, r := range m.recs[tenantID] {
		if recType != "" && r.Type != recType {
			continue
		}
		if scope != "" && r.Scope != scope {
			continue
		}
		if status != "" && r.Status != status {
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

func (m *MemoryStore) GetRecommendation(_ context.Context, tenantID, id string) (domain.Recommendation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, r := range m.recs[tenantID] {
		if r.ID == id {
			return r, nil
		}
	}
	return domain.Recommendation{}, errors.New("recommendation not found")
}

func (m *MemoryStore) SaveRecommendation(_ context.Context, tenantID string, r domain.Recommendation) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recs[tenantID] = append(m.recs[tenantID], r)
	return nil
}

func (m *MemoryStore) UpdateRecommendationStatus(_ context.Context, tenantID, id, action string) (domain.Recommendation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, r := range m.recs[tenantID] {
		if r.ID == id {
			r = optimize.ApplyLifecycleAction(r, action)
			m.recs[tenantID][i] = r
			return r, nil
		}
	}
	return domain.Recommendation{}, errors.New("recommendation not found")
}

func (m *MemoryStore) SaveImportBatch(_ context.Context, tenantID string, b domain.ImportBatch) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.imports[tenantID] = append(m.imports[tenantID], b)
	return nil
}

func (m *MemoryStore) GetImportVersion(_ context.Context, tenantID, dedupeKey string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	max := 0
	for _, b := range m.imports[tenantID] {
		if b.DedupeKey == dedupeKey && b.Version > max {
			max = b.Version
		}
	}
	return max, nil
}

func (m *MemoryStore) ListSharedSplits(_ context.Context, tenantID string) ([]domain.SharedSplitRule, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]domain.SharedSplitRule(nil), m.splits[tenantID]...), nil
}

func (m *MemoryStore) SaveSharedSplits(_ context.Context, tenantID string, rules []domain.SharedSplitRule) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.splits[tenantID] = rules
	return nil
}

func (m *MemoryStore) ListTagSuggestions(_ context.Context, tenantID string) ([]domain.TagSuggestion, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]domain.TagSuggestion(nil), m.tagSuggest[tenantID]...), nil
}

func (m *MemoryStore) SaveTagSuggestions(_ context.Context, tenantID string, s []domain.TagSuggestion) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tagSuggest[tenantID] = s
	return nil
}

func (m *MemoryStore) GetAnomalySensitivity(_ context.Context, tenantID, scope string) (domain.AnomalySensitivity, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.sensitivity[tenantID] != nil {
		if s, ok := m.sensitivity[tenantID][scope]; ok {
			return s, nil
		}
	}
	return domain.AnomalySensitivity{Scope: scope, ZThreshold: 2.5}, nil
}

func (m *MemoryStore) SaveAnomalySensitivity(_ context.Context, tenantID string, s domain.AnomalySensitivity) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sensitivity[tenantID] == nil {
		m.sensitivity[tenantID] = map[string]domain.AnomalySensitivity{}
	}
	m.sensitivity[tenantID][s.Scope] = s
	return nil
}

func (m *MemoryStore) ListIngestSnapshots(_ context.Context, tenantID string) ([]domain.IngestSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]domain.IngestSnapshot(nil), m.snapshots[tenantID]...), nil
}

func (m *MemoryStore) SaveChargebackStatement(_ context.Context, tenantID string, st domain.ChargebackStatement) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statements[tenantID] = append(m.statements[tenantID], st)
	return nil
}

func (m *MemoryStore) GetChargebackStatement(_ context.Context, tenantID, id string) (domain.ChargebackStatement, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, st := range m.statements[tenantID] {
		if st.ID == id {
			return st, nil
		}
	}
	return domain.ChargebackStatement{}, errors.New("statement not found")
}

func (m *MemoryStore) ListChargebackStatements(_ context.Context, tenantID, costCenter string) ([]domain.ChargebackStatement, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]domain.ChargebackStatement, 0)
	for _, st := range m.statements[tenantID] {
		if costCenter != "" && costCenter != "all" && st.CostCenter != costCenter {
			continue
		}
		out = append(out, st)
	}
	return out, nil
}

func (m *MemoryStore) SaveScenarioRun(_ context.Context, tenantID string, r domain.ScenarioResult) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.scenarios[tenantID] = append(m.scenarios[tenantID], r)
	return nil
}

func (m *MemoryStore) ListScenarioRuns(_ context.Context, tenantID string, limit int) ([]domain.ScenarioResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := m.scenarios[tenantID]
	if limit <= 0 || limit > len(list) {
		limit = len(list)
	}
	start := len(list) - limit
	if start < 0 {
		start = 0
	}
	out := make([]domain.ScenarioResult, limit)
	copy(out, list[start:])
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}
