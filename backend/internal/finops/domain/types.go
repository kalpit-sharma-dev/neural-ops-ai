package domain

import "time"

// CostLineItem is a FOCUS-aligned canonical billing record.
type CostLineItem struct {
	ID            string            `json:"id"`
	TenantID      string            `json:"tenantId"`
	BillingPeriod time.Time         `json:"billingPeriod"`
	Provider      string            `json:"provider"`
	AccountID     string            `json:"accountId"`
	Service       string            `json:"service"`
	Region        string            `json:"region"`
	ResourceID    string            `json:"resourceId"`
	UsageType     string            `json:"usageType"`
	Quantity      float64           `json:"quantity"`
	Unit          string            `json:"unit"`
	AmortizedCost float64           `json:"amortizedCost"`
	ListCost      float64           `json:"listCost"`
	EffectiveCost float64           `json:"effectiveCost"`
	CostView      string            `json:"costView"`
	Tags          map[string]string `json:"tags"`
	Team          string            `json:"team"`
	Environment   string            `json:"environment"`
	CostCenter    string            `json:"costCenter"`
	IngestedAt    time.Time         `json:"ingestedAt"`
}

// IngestSnapshot tracks reconciliation per provider/period.
type IngestSnapshot struct {
	ID             string    `json:"id"`
	Provider       string    `json:"provider"`
	BillingPeriod  time.Time `json:"billingPeriod"`
	LineCount      int64     `json:"lineCount"`
	TotalAmortized float64   `json:"totalAmortized"`
	InvoiceTotal   float64   `json:"invoiceTotal"`
	DriftPct       float64   `json:"driftPct"`
	CostView       string    `json:"costView"`
	IngestedAt     time.Time `json:"ingestedAt"`
	DriftAlert     bool      `json:"driftAlert"`
	Version        int       `json:"version"`
}

// AllocationRule maps tags to organizational dimensions.
type AllocationRule struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Dimension string    `json:"dimension"`
	TagKey    string    `json:"tagKey"`
	TagValue  string    `json:"tagValue"`
	Priority  int       `json:"priority"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Budget scoped to allocation dimension.
type Budget struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	ScopeType      string    `json:"scopeType"`
	ScopeValue     string    `json:"scopeValue"`
	Period         string    `json:"period"`
	AmountUSD      float64   `json:"amountUsd"`
	Thresholds     []int     `json:"thresholds"`
	NotifyPolicyID string    `json:"notifyPolicyId"`
	SpendUSD       float64   `json:"spendUsd,omitempty"`
	BurnPct        float64   `json:"burnPct,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// Anomaly is a detected cost spike.
type Anomaly struct {
	ID                string    `json:"id"`
	Scope             string    `json:"scope"`
	Service           string    `json:"service"`
	Provider          string    `json:"provider"`
	DeltaPct          float64   `json:"deltaPct"`
	AmountUSD         float64   `json:"amountUsd"`
	Severity          string    `json:"severity"`
	Status            string    `json:"status"`
	Description       string    `json:"description"`
	AlertPolicyID     string    `json:"alertPolicyId,omitempty"`
	DetectedAt        time.Time `json:"detectedAt"`
	Feedback          string    `json:"feedback,omitempty"`
	ProbableCause     string    `json:"probableCause,omitempty"`
	SensitivityFactor float64   `json:"sensitivityFactor,omitempty"`
}

// Recommendation is an optimization action.
type Recommendation struct {
	ID                  string    `json:"id"`
	Type                string    `json:"type"`
	ResourceID          string    `json:"resourceId"`
	Scope               string    `json:"scope"`
	Title               string    `json:"title"`
	Description         string    `json:"description"`
	ProjectedSavingsUSD float64   `json:"projectedSavingsUsd"`
	RealizedSavingsUSD  float64   `json:"realizedSavingsUsd,omitempty"`
	RiskScore           float64   `json:"riskScore"`
	Status              string    `json:"status"`
	TicketID            string    `json:"ticketId,omitempty"`
	TicketURL           string    `json:"ticketUrl,omitempty"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

// CostSeries is enhanced spend over time (backward compatible fields included).
type CostSeries struct {
	Scope        string       `json:"scope"`
	Unit         string       `json:"unit"`
	Total        float64      `json:"total"`
	Budget       float64      `json:"budget"`
	BudgetID     string       `json:"budgetId,omitempty"`
	Points       []CostPoint  `json:"points"`
	Forecast     Forecast     `json:"forecast"`
	TagCoverage  float64      `json:"tagCoveragePct"`
	CostView     string       `json:"costView"`
	Reconciliation []IngestSnapshot `json:"reconciliation,omitempty"`
}

// CostPoint is one cost sample.
type CostPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Amount    float64   `json:"amount"`
}

// Forecast with confidence band.
type Forecast struct {
	Horizon   string  `json:"horizon"`
	P50       float64 `json:"p50"`
	P95       float64 `json:"p95"`
	Lower     float64 `json:"lower"`
	Upper     float64 `json:"upper"`
	Method    string  `json:"method"`
}

// BreakdownNode is a cost allocation tree node.
type BreakdownNode struct {
	Dimension string          `json:"dimension"`
	Key       string          `json:"key"`
	AmountUSD float64         `json:"amountUsd"`
	Pct       float64         `json:"pct"`
	Children  []BreakdownNode `json:"children,omitempty"`
}

// K8sCostRow is namespace/workload cost allocation.
type K8sCostRow struct {
	Cluster   string  `json:"cluster"`
	Namespace string  `json:"namespace"`
	Workload  string  `json:"workload"`
	CPUCost   float64 `json:"cpuCostUsd"`
	MemCost   float64 `json:"memCostUsd"`
	IdleCost  float64 `json:"idleCostUsd"`
	TotalUSD  float64 `json:"totalUsd"`
}

// CloudAsset is a minimal inventory record for billing ingest.
type CloudAsset struct {
	ID         string
	Provider   string
	Type       string
	Name       string
	Region     string
	AccountID  string
	Status     string
	Tags       map[string]string
	MonthlyUSD float64
}

// CloudAssets lists inventory for connector ingest.
type CloudAssets interface {
	ListCloudAssets(provider string) []CloudAsset
}

// CostsQuery parameters for enhanced cost series.
type CostsQuery struct {
	Scope       string
	Provider    string
	Granularity string
	From        time.Time
	To          time.Time
	GroupBy     string
	CostView    string
}

// CarbonFootprint legacy carbon summary.
type CarbonFootprint struct {
	Scope          string             `json:"scope"`
	Period         string             `json:"period"`
	Co2eKg         float64            `json:"co2eKg"`
	RenewablePct   float64            `json:"renewablePct"`
	Recommendation string             `json:"recommendation,omitempty"`
	Methodology    string             `json:"methodology,omitempty"`
	FactorVersion  string             `json:"factorVersion,omitempty"`
	ByDimension    []CarbonBreakdown  `json:"byDimension,omitempty"`
	SCI            SoftwareCarbonIntensity `json:"sci,omitempty"`
}

// CarbonBreakdown is attributed CO2e by dimension value.
type CarbonBreakdown struct {
	Dimension string  `json:"dimension"`
	Key       string  `json:"key"`
	Co2eKg    float64 `json:"co2eKg"`
	Pct       float64 `json:"pct"`
}

// SoftwareCarbonIntensity per SCI spec.
type SoftwareCarbonIntensity struct {
	Co2ePerRequest    float64 `json:"co2ePerRequest"`
	Co2ePerTransaction float64 `json:"co2ePerTransaction"`
	RequestsPerMonth  int64   `json:"requestsPerMonth"`
}

// ImportBatch tracks custom cost feed ingestion.
type ImportBatch struct {
	ID         string    `json:"id"`
	Source     string    `json:"source"`
	Format     string    `json:"format"`
	LineCount  int64     `json:"lineCount"`
	TotalUSD   float64   `json:"totalUsd"`
	Version    int       `json:"version"`
	DedupeKey  string    `json:"dedupeKey"`
	ImportedAt time.Time `json:"importedAt"`
}

// ImportRequest is the body for POST /finops/imports.
type ImportRequest struct {
	Source  string           `json:"source"`
	Format  string           `json:"format"`
	Items   []ImportLineItem `json:"items"`
	DedupeKey string         `json:"dedupeKey"`
}

// ImportLineItem is one row from a custom cost feed.
type ImportLineItem struct {
	ResourceID    string            `json:"resourceId"`
	Service       string            `json:"service"`
	Provider      string            `json:"provider"`
	Region        string            `json:"region"`
	EffectiveCost float64           `json:"effectiveCost"`
	UsageType     string            `json:"usageType"`
	Tags          map[string]string `json:"tags"`
}

// SharedSplitRule splits shared/untagged resources across teams.
type SharedSplitRule struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	ResourcePattern string         `json:"resourcePattern"`
	Mode            string         `json:"mode"` // proportional|even|weighted
	Targets         []SplitTarget  `json:"targets"`
	Enabled         bool           `json:"enabled"`
	CreatedAt       time.Time      `json:"createdAt"`
}

// SplitTarget is one allocation target for shared cost splitting.
type SplitTarget struct {
	Team   string  `json:"team"`
	Weight float64 `json:"weight"`
}

// TagSuggestion is an ML-assisted tagging recommendation.
type TagSuggestion struct {
	ID             string    `json:"id"`
	ResourceID     string    `json:"resourceId"`
	SuggestedKey   string    `json:"suggestedKey"`
	SuggestedValue string    `json:"suggestedValue"`
	Confidence     float64   `json:"confidence"`
	SpendUSD       float64   `json:"spendUsd"`
	CreatedAt      time.Time `json:"createdAt"`
}

// Commitment is RI/SP/CUD coverage record.
type Commitment struct {
	ID               string    `json:"id"`
	Provider         string    `json:"provider"`
	CommitmentType   string    `json:"commitmentType"`
	Region           string    `json:"region"`
	CoveragePct      float64   `json:"coveragePct"`
	UtilizationPct   float64   `json:"utilizationPct"`
	MonthlyCommitUSD float64   `json:"monthlyCommitUsd"`
	ExpiresAt        time.Time `json:"expiresAt"`
	Status           string    `json:"status"`
}

// CommitmentRecommendation suggests a commitment purchase.
type CommitmentRecommendation struct {
	ID              string  `json:"id"`
	Provider        string  `json:"provider"`
	CommitmentType  string  `json:"commitmentType"`
	TermMonths      int     `json:"termMonths"`
	BreakEvenMonths float64 `json:"breakEvenMonths"`
	MonthlySavings  float64 `json:"monthlySavingsUsd"`
	RiskScore       float64 `json:"riskScore"`
	Description     string  `json:"description"`
}

// UnitEconomics joins cost with usage metrics.
type UnitEconomics struct {
	Metric       string  `json:"metric"`
	Scope        string  `json:"scope"`
	TotalCostUSD float64 `json:"totalCostUsd"`
	TotalUnits   int64   `json:"totalUnits"`
	CostPerUnit  float64 `json:"costPerUnit"`
	Period       string  `json:"period"`
}

// CarbonRecommendation suggests carbon-reducing action.
type CarbonRecommendation struct {
	ID              string  `json:"id"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	Co2eReductionKg float64 `json:"co2eReductionKg"`
	CostDeltaUSD    float64 `json:"costDeltaUsd"`
	Region          string  `json:"region"`
}

// ReportSchedule is a scheduled FinOps report.
type ReportSchedule struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Scope           string     `json:"scope"`
	Format          string     `json:"format"`
	Cadence         string     `json:"cadence"`
	DeliveryChannel string     `json:"deliveryChannel"`
	DeliveryTarget  string     `json:"deliveryTarget"`
	Enabled         bool       `json:"enabled"`
	LastRunAt       *time.Time `json:"lastRunAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
}

// FinOpsReport is a generated report summary.
type FinOpsReport struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Scope     string    `json:"scope"`
	Format    string    `json:"format"`
	URL       string    `json:"url"`
	Generated time.Time `json:"generatedAt"`
}

// AuditEntry is a FinOps mutation audit log row.
type AuditEntry struct {
	ID         string            `json:"id"`
	Actor      string            `json:"actor"`
	Action     string            `json:"action"`
	EntityType string            `json:"entityType"`
	EntityID   string            `json:"entityId"`
	Detail     map[string]string `json:"detail,omitempty"`
	CreatedAt  time.Time         `json:"createdAt"`
}

// AnomalySensitivity tunes detection threshold per scope.
type AnomalySensitivity struct {
	Scope              string    `json:"scope"`
	ZThreshold         float64   `json:"zThreshold"`
	FalsePositiveCount int       `json:"falsePositiveCount"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// ChargebackStatement is a showback/chargeback export per cost-center (REQ-FINOPS-014).
type ChargebackStatement struct {
	ID          string              `json:"id"`
	CostCenter  string              `json:"costCenter"`
	Mode        string              `json:"mode"` // showback|chargeback
	Period      string              `json:"period"`
	TotalUSD    float64             `json:"totalUsd"`
	Lines       []ChargebackLine    `json:"lines"`
	ExportURL   string              `json:"exportUrl,omitempty"`
	GeneratedAt time.Time           `json:"generatedAt"`
}

// ChargebackLine is one attributed spend line on a statement.
type ChargebackLine struct {
	Team        string  `json:"team"`
	Service     string  `json:"service"`
	AmountUSD   float64 `json:"amountUsd"`
	AllocatedPct float64 `json:"allocatedPct"`
}

// ScenarioRequest models a what-if change (REQ-FINOPS-053).
type ScenarioRequest struct {
	Name         string            `json:"name"`
	ScenarioType string            `json:"scenarioType"` // region_migration|instance_family|commitment_purchase
	Scope        string            `json:"scope"`
	Params       map[string]string `json:"params"`
}

// ScenarioResult is projected cost and carbon impact.
type ScenarioResult struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	ScenarioType     string    `json:"scenarioType"`
	BaselineCostUSD  float64   `json:"baselineCostUsd"`
	ProjectedCostUSD float64   `json:"projectedCostUsd"`
	CostDeltaUSD     float64   `json:"costDeltaUsd"`
	BaselineCo2eKg  float64   `json:"baselineCo2eKg"`
	ProjectedCo2eKg float64   `json:"projectedCo2eKg"`
	Co2eDeltaKg     float64   `json:"co2eDeltaKg"`
	Summary          string    `json:"summary"`
	CreatedAt        time.Time `json:"createdAt"`
}

// CommitmentAlert is an expiring or under-utilized commitment (REQ-FINOPS-042).
type CommitmentAlert struct {
	ID             string    `json:"id"`
	CommitmentID   string    `json:"commitmentId"`
	Provider       string    `json:"provider"`
	AlertType      string    `json:"alertType"` // expiring|under_utilized
	Severity       string    `json:"severity"`
	Message        string    `json:"message"`
	UtilizationPct float64   `json:"utilizationPct,omitempty"`
	ExpiresAt      time.Time `json:"expiresAt,omitempty"`
	AlertPolicyID  string    `json:"alertPolicyId,omitempty"`
}

// CarbonActionRequest applies a carbon reduction recommendation (REQ-FINOPS-062).
type CarbonActionRequest struct {
	RecommendationID string `json:"recommendationId"`
	Action           string `json:"action"` // simulate|apply|dismiss
}

// CarbonActionResult is the outcome of a carbon action.
type CarbonActionResult struct {
	ID              string    `json:"id"`
	RecommendationID string   `json:"recommendationId"`
	Status          string    `json:"status"`
	Co2eReductionKg float64   `json:"co2eReductionKg"`
	CostDeltaUSD    float64   `json:"costDeltaUsd"`
	AppliedAt       time.Time `json:"appliedAt,omitempty"`
}

// GovernancePolicy is advanced FinOps governance config.
type GovernancePolicy struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	ResidencyRegion string    `json:"residencyRegion"`
	AllowedScopes   []string  `json:"allowedScopes"`
	ChargebackMode  string    `json:"chargebackMode"`
	Enabled         bool      `json:"enabled"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
