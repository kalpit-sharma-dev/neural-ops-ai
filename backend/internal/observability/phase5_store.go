package observability

import (
	"strings"
	"time"
)

func (s *Store) seedPhase5() {
	now := time.Now().UTC()
	s.cloudAssets = []CloudAsset{
		{ID: "aws-ec2-pay-01", Provider: "aws", Type: "ec2", Name: "payment-asg-node-01", Region: "us-east-1", AccountID: "111122223333", Status: "running", Tags: map[string]string{"env": "prod", "team": "payments"}, MonthlyUSD: 420.5, UpdatedAt: now},
		{ID: "aws-rds-ledger", Provider: "aws", Type: "rds", Name: "ledger-primary", Region: "us-east-1", AccountID: "111122223333", Status: "available", Tags: map[string]string{"tier": "data"}, MonthlyUSD: 890.0, UpdatedAt: now},
		{ID: "aws-lambda-notify", Provider: "aws", Type: "lambda", Name: "notification-dispatch", Region: "us-east-1", AccountID: "111122223333", Status: "active", MonthlyUSD: 45.2, UpdatedAt: now},
		{ID: "gcp-run-api", Provider: "gcp", Type: "cloud_run", Name: "api-gateway-run", Region: "us-central1", AccountID: "neuralops-prod", Status: "ready", MonthlyUSD: 210.0, UpdatedAt: now},
		{ID: "azure-app-auth", Provider: "azure", Type: "app_service", Name: "auth-service-plan", Region: "eastus", AccountID: "sub-neuralops", Status: "running", MonthlyUSD: 380.0, UpdatedAt: now},
	}
	s.cloudAssetTopology = CloudAssetTopology{
		At: now,
		Nodes: []CloudAssetNode{
			{ID: "aws-ec2-pay-01", Label: "payment-asg", Provider: "aws", Type: "ec2", Health: "healthy"},
			{ID: "aws-rds-ledger", Label: "ledger-db", Provider: "aws", Type: "rds", Health: "degraded"},
			{ID: "aws-lambda-notify", Label: "notify-fn", Provider: "aws", Type: "lambda", Health: "healthy"},
			{ID: "gcp-run-api", Label: "api-gateway", Provider: "gcp", Type: "cloud_run", Health: "healthy"},
			{ID: "azure-app-auth", Label: "auth-app", Provider: "azure", Type: "app_service", Health: "healthy"},
		},
		Edges: []CloudAssetEdge{
			{Source: "gcp-run-api", Target: "aws-ec2-pay-01", Relation: "routes_to"},
			{Source: "aws-ec2-pay-01", Target: "aws-rds-ledger", Relation: "depends_on"},
			{Source: "aws-ec2-pay-01", Target: "aws-lambda-notify", Relation: "invokes"},
			{Source: "gcp-run-api", Target: "azure-app-auth", Relation: "authenticates_via"},
		},
	}
	s.finopsAnomalies = []FinOpsCostAnomaly{
		{ID: "cost-anom-1", Scope: "payments", Service: "ledger-service", Provider: "aws", DeltaPct: 34.2, AmountUSD: 1280, Severity: "high", Description: "RDS storage IOPS spike vs 7d baseline", DetectedAt: now.Add(-3 * time.Hour)},
		{ID: "cost-anom-2", Scope: "platform", Service: "api-gateway", Provider: "gcp", DeltaPct: 18.5, AmountUSD: 420, Severity: "medium", Description: "Cloud Run request volume surge", DetectedAt: now.Add(-11 * time.Hour)},
	}
	s.finopsCarbon = FinOpsCarbonFootprint{
		Scope: "tenant-default", Period: "30d", Co2eKg: 1240.5, RenewablePct: 62.0,
		Recommendation: "Shift batch workloads to us-central1 (higher renewable mix).",
	}
	s.networkDevices = []NetworkDevice{
		{ID: "sw-core-1", Name: "core-switch-01", Type: "switch", Site: "us-east-dc1", Status: "up", CPUUtil: 42, MemUtil: 58, UptimePct: 99.98},
		{ID: "fw-edge-1", Name: "edge-fw-01", Type: "firewall", Site: "us-east-dc1", Status: "up", CPUUtil: 28, MemUtil: 44, UptimePct: 99.99},
		{ID: "rt-wan-1", Name: "wan-router-01", Type: "router", Site: "us-east-dc1", Status: "degraded", CPUUtil: 71, MemUtil: 62, UptimePct: 99.2},
	}
	s.networkFlows = []NetworkFlow{
		{ID: "flow-1", Source: "10.0.1.12", Destination: "10.0.2.45", Protocol: "TCP", Port: 5432, Bytes: 1_240_000, Packets: 8200, LatencyMs: 2.1, LossPct: 0.01, JitterMs: 0.4, Timestamp: now.Add(-2 * time.Minute)},
		{ID: "flow-2", Source: "10.0.1.12", Destination: "10.0.3.10", Protocol: "TCP", Port: 443, Bytes: 4_800_000, Packets: 32000, LatencyMs: 18.4, LossPct: 0.8, JitterMs: 3.2, Timestamp: now.Add(-2 * time.Minute)},
		{ID: "flow-3", Source: "10.0.3.10", Destination: "52.94.12.8", Protocol: "TCP", Port: 443, Bytes: 9_100_000, Packets: 61000, LatencyMs: 42.0, LossPct: 2.4, JitterMs: 8.1, Timestamp: now.Add(-1 * time.Minute)},
	}
	s.networkTopology = NetworkTopologyGraph{
		At: now,
		Nodes: []NetworkTopologyNode{
			{ID: "sw-core-1", Label: "core-switch", Type: "switch", Health: "healthy"},
			{ID: "fw-edge-1", Label: "edge-fw", Type: "firewall", Health: "healthy"},
			{ID: "rt-wan-1", Label: "wan-router", Type: "router", Health: "degraded"},
			{ID: "pay-subnet", Label: "payments-subnet", Type: "subnet", Health: "healthy"},
		},
		Edges: []NetworkTopologyEdge{
			{Source: "pay-subnet", Target: "sw-core-1", LatencyMs: 1.2, LossPct: 0, UtilPct: 45},
			{Source: "sw-core-1", Target: "fw-edge-1", LatencyMs: 0.8, LossPct: 0, UtilPct: 38},
			{Source: "fw-edge-1", Target: "rt-wan-1", LatencyMs: 12.4, LossPct: 1.8, UtilPct: 82},
		},
	}
	s.networkAnomalies = []NetworkAnomaly{
		{ID: "npm-anom-1", Link: "fw-edge-1 → rt-wan-1", Metric: "loss_pct", Value: 2.4, Baseline: 0.3, Severity: "high", Description: "WAN link packet loss elevated", DetectedAt: now.Add(-25 * time.Minute)},
		{ID: "npm-anom-2", Link: "pay-subnet → sw-core-1", Metric: "jitter_ms", Value: 3.2, Baseline: 0.6, Severity: "medium", Description: "Jitter spike on payments path", DetectedAt: now.Add(-50 * time.Minute)},
	}
}

// ListCloudAssets returns cloud inventory, optionally filtered by provider.
func (s *Store) ListCloudAssets(provider string) []CloudAsset {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if provider == "" {
		out := make([]CloudAsset, len(s.cloudAssets))
		copy(out, s.cloudAssets)
		return out
	}
	out := make([]CloudAsset, 0)
	for _, a := range s.cloudAssets {
		if strings.EqualFold(a.Provider, provider) {
			out = append(out, a)
		}
	}
	return out
}

// CloudAssetTopology returns the multi-cloud relationship graph.
func (s *Store) CloudAssetTopology() CloudAssetTopology {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cloudAssetTopology
}

// FinOpsCosts returns spend series for optional scope/provider.
func (s *Store) FinOpsCosts(scope, provider string) FinOpsCostSeries {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now().UTC()
	points := make([]FinOpsCostPoint, 0, 14)
	total := 0.0
	for i := 13; i >= 0; i-- {
		amt := 8200.0 + float64(i%5)*120 - float64(i)*15
		if scope == "payments" {
			amt = 3100 + float64(i%4)*80
		}
		if provider == "aws" {
			amt *= 0.55
		} else if provider == "gcp" {
			amt *= 0.25
		} else if provider == "azure" {
			amt *= 0.20
		}
		points = append(points, FinOpsCostPoint{Timestamp: now.AddDate(0, 0, -i), Amount: amt})
		total += amt
	}
	if scope == "" {
		scope = "all"
	}
	return FinOpsCostSeries{
		Scope: scope, Unit: "USD", Total: total, Budget: total * 1.05, Points: points,
	}
}

// FinOpsCostAnomalies returns detected spend anomalies.
func (s *Store) FinOpsCostAnomalies() []FinOpsCostAnomaly {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]FinOpsCostAnomaly, len(s.finopsAnomalies))
	copy(out, s.finopsAnomalies)
	return out
}

// FinOpsCarbon returns carbon footprint summary.
func (s *Store) FinOpsCarbon(scope string) FinOpsCarbonFootprint {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c := s.finopsCarbon
	if scope != "" {
		c.Scope = scope
	}
	return c
}

// ListNetworkFlows returns recent flow records.
func (s *Store) ListNetworkFlows(limit int) []NetworkFlow {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 || limit > len(s.networkFlows) {
		limit = len(s.networkFlows)
	}
	out := make([]NetworkFlow, limit)
	copy(out, s.networkFlows[:limit])
	return out
}

// ListNetworkDevices returns SNMP/monitored devices.
func (s *Store) ListNetworkDevices() []NetworkDevice {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]NetworkDevice, len(s.networkDevices))
	copy(out, s.networkDevices)
	return out
}

// NetworkTopology returns NPM topology graph.
func (s *Store) NetworkTopology() NetworkTopologyGraph {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.networkTopology
}

// NetworkAnomalies returns NPM degradation alerts.
func (s *Store) NetworkAnomalies() []NetworkAnomaly {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]NetworkAnomaly, len(s.networkAnomalies))
	copy(out, s.networkAnomalies)
	return out
}
