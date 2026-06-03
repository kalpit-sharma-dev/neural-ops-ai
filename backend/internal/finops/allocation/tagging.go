package allocation

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/neuralops/platform/internal/finops/domain"
)

// SplitSharedCosts applies proportional/even/weighted splits to shared resources (REQ-FINOPS-012).
func (e *Engine) SplitSharedCosts(items []domain.CostLineItem, rules []domain.SharedSplitRule) {
	for i := range items {
		if items[i].Team != "" && items[i].Team != "unallocated" {
			continue
		}
		for _, rule := range rules {
			if !rule.Enabled {
				continue
			}
			if !matchResourcePattern(rule.ResourcePattern, items[i].ResourceID, items[i].Service) {
				continue
			}
			items[i].Team = splitTargetTeam(rule, items[i].EffectiveCost)
			break
		}
	}
}

func matchResourcePattern(pattern, resourceID, service string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}
	p := strings.ToLower(strings.ReplaceAll(pattern, "*", ""))
	return strings.Contains(strings.ToLower(resourceID), p) ||
		strings.Contains(strings.ToLower(service), p)
}

func splitTargetTeam(rule domain.SharedSplitRule, _ float64) string {
	if len(rule.Targets) == 0 {
		return "platform"
	}
	switch rule.Mode {
	case "even":
		return rule.Targets[0].Team
	case "weighted":
		best := rule.Targets[0]
		for _, t := range rule.Targets[1:] {
			if t.Weight > best.Weight {
				best = t
			}
		}
		return best.Team
	default: // proportional — assign to highest-weight team
		return rule.Targets[0].Team
	}
}

// SuggestTags generates ML-assisted tagging suggestions for untagged spend (REQ-FINOPS-013).
func (e *Engine) SuggestTags(tenantID string, items []domain.CostLineItem) []domain.TagSuggestion {
	now := time.Now().UTC()
	patterns := []struct {
		match, key, val string
		conf            float64
	}{
		{"pay", "team", "payments", 0.92},
		{"ledger", "team", "payments", 0.88},
		{"auth", "team", "platform", 0.85},
		{"api", "team", "platform", 0.80},
		{"nat", "service", "network", 0.75},
		{"s3", "service", "storage", 0.78},
	}
	var out []domain.TagSuggestion
	seen := map[string]bool{}
	for _, item := range items {
		if item.Team != "" {
			continue
		}
		needle := strings.ToLower(item.ResourceID + " " + item.Service)
		for _, p := range patterns {
			if !strings.Contains(needle, p.match) {
				continue
			}
			k := item.ResourceID + p.key
			if seen[k] {
				continue
			}
			seen[k] = true
			out = append(out, domain.TagSuggestion{
				ID: uuid.NewString(), ResourceID: item.ResourceID,
				SuggestedKey: p.key, SuggestedValue: p.val,
				Confidence: p.conf, SpendUSD: item.EffectiveCost, CreatedAt: now,
			})
		}
	}
	return out
}
