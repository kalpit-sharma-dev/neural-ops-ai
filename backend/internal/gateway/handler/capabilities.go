package handler

import "os"

// PlatformCapabilities advertises open-standards and unified-platform features for clients and integrators.
type PlatformCapabilities struct {
	Product            string   `json:"product"`
	OpenTelemetry      bool     `json:"openTelemetry"`
	Prometheus         bool     `json:"prometheus"`
	Ebpf               bool     `json:"ebpf"`
	OpenAPI            bool     `json:"openApi"`
	StreamingPipeline  bool     `json:"streamingPipeline"`
	UnifiedQuery       bool     `json:"unifiedQuery"`
	AIFeatures         []string `json:"aiFeatures"`
	GovernanceFeatures []string `json:"governanceFeatures"`
	DeploymentModes    []string `json:"deploymentModes"`
	Standards          []string `json:"standards"`
	Differentiators    []string `json:"differentiators"`
}

// DefaultCapabilities returns the platform capability matrix exposed on GET /api/v1/info.
func DefaultCapabilities(demoMode bool) PlatformCapabilities {
	_ = demoMode
	streaming := os.Getenv("KAFKA_BROKERS") != ""
	return PlatformCapabilities{
		Product:           "NeuralOps",
		OpenTelemetry:     true,
		Prometheus:        true,
		Ebpf:              true,
		OpenAPI:           true,
		StreamingPipeline: streaming,
		UnifiedQuery:      true,
		AIFeatures: []string{
			"natural-language-query",
			"incident-rca",
			"causal-explanations",
			"capacity-forecast",
			"autofix-plan-execute",
		},
		GovernanceFeatures: []string{
			"rbac",
			"abac",
			"data-residency",
			"msp-white-label",
			"multi-region-status",
			"audit-log",
		},
		DeploymentModes: []string{"docker-compose", "kubernetes", "hybrid-collectors"},
		Standards: []string{
			"OpenTelemetry OTLP",
			"Prometheus PromQL",
			"OpenAPI 3",
			"OAuth2/OIDC",
			"SAML 2.0",
			"Terraform provider",
		},
		Differentiators: []string{
			"unified-melt-plus-incident-workflow",
			"ai-native-rca-and-autofix",
			"security-finding-to-incident-correlation",
			"open-standards-first-ingestion",
			"self-host-and-data-residency",
			"streaming-derived-metrics-and-alert-fatigue",
		},
	}
}
