package allocation

import (
	"sort"
	"strings"

	"github.com/neuralops/platform/internal/finops/domain"
)

// Engine applies tag/label mapping rules to line items.
type Engine struct{}

// NewEngine creates an allocation engine.
func NewEngine() *Engine { return &Engine{} }

// Apply mutates items in-place with team/environment/cost-center from rules.
func (e *Engine) Apply(tenantID string, items []domain.CostLineItem, rules []domain.AllocationRule) {
	sort.Slice(rules, func(i, j int) bool { return rules[i].Priority < rules[j].Priority })
	for i := range items {
		for _, r := range rules {
			if !r.Enabled {
				continue
			}
			tagVal, ok := items[i].Tags[r.TagKey]
			if !ok {
				continue
			}
			if r.TagValue != "" && tagVal != r.TagValue {
				continue
			}
			switch r.Dimension {
			case "team":
				items[i].Team = tagVal
			case "environment":
				items[i].Environment = tagVal
			case "cost_center":
				items[i].CostCenter = tagVal
			case "service":
				items[i].Service = tagVal
			}
		}
		if items[i].Team == "" {
			items[i].Team = inferTeam(items[i])
		}
	}
}

func inferTeam(item domain.CostLineItem) string {
	if strings.Contains(item.ResourceID, "pay") {
		return "payments"
	}
	return "platform"
}

// KubernetesCost splits cluster spend by namespace/workload with idle cost.
func (e *Engine) KubernetesCost(items []domain.CostLineItem, cluster, namespace string) []domain.K8sCostRow {
	type key struct{ ns, wl string }
	totals := map[key]domain.K8sCostRow{}
	var clusterTotal float64
	for _, item := range items {
		if item.Service != "ec2" && item.Service != "cloud_run" && item.Service != "app_service" {
			continue
		}
		clusterTotal += item.EffectiveCost
		ns := "platform"
		if item.Team != "" {
			ns = item.Team
		}
		if namespace != "" && ns != namespace {
			continue
		}
		wl := item.ResourceID
		k := key{ns: ns, wl: wl}
		row := totals[k]
		row.Cluster = cluster
		if row.Cluster == "" {
			row.Cluster = "prod-cluster"
		}
		row.Namespace = ns
		row.Workload = wl
		row.CPUCost += item.EffectiveCost * 0.6
		row.MemCost += item.EffectiveCost * 0.3
		row.TotalUSD += item.EffectiveCost
		totals[k] = row
	}
	idle := clusterTotal * 0.12
	out := make([]domain.K8sCostRow, 0, len(totals)+1)
	for _, row := range totals {
		out = append(out, row)
	}
	if len(out) > 0 {
		out[0].IdleCost = idle
		out[0].TotalUSD += idle
	}
	sort.Slice(out, func(i, j int) bool { return out[i].TotalUSD > out[j].TotalUSD })
	return out
}
