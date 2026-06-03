package connectors

import (
	"os"
	"strconv"
	"strings"
)

// BillingMode controls how cloud billing is ingested (FIN-PROD-01..03).
const (
	BillingModeSimulated = "simulated"
	BillingModeLive      = "live"
	BillingModeHybrid    = "hybrid"
)

// BillingConfig is loaded from environment for production FinOps.
type BillingConfig struct {
	Mode                    string
	RequirePostgres         bool
	AllowSimulationFallback bool
	AWS_CURFile             string
	AWSInvoiceUSD           float64
	GCPBillingFile          string
	AzureBillingFile        string
}

// LoadBillingConfig reads FinOps billing settings from the environment.
func LoadBillingConfig() BillingConfig {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("FINOPS_BILLING_MODE")))
	if mode == "" {
		mode = BillingModeSimulated
	}
	inv, _ := strconv.ParseFloat(strings.TrimSpace(os.Getenv("FINOPS_AWS_INVOICE_USD")), 64)
	return BillingConfig{
		Mode:                    mode,
		RequirePostgres:         envBool("FINOPS_REQUIRE_POSTGRES"),
		AllowSimulationFallback: envBool("FINOPS_ALLOW_SIMULATION"),
		AWS_CURFile:             strings.TrimSpace(os.Getenv("FINOPS_AWS_CUR_FILE")),
		AWSInvoiceUSD:           inv,
		GCPBillingFile:          strings.TrimSpace(os.Getenv("FINOPS_GCP_BILLING_FILE")),
		AzureBillingFile:        strings.TrimSpace(os.Getenv("FINOPS_AZURE_BILLING_FILE")),
	}
}

func (c BillingConfig) LiveProviderEnabled(provider string) bool {
	switch strings.ToLower(provider) {
	case "aws":
		return c.AWS_CURFile != "" || c.AWSInvoiceUSD > 0
	case "gcp":
		return c.GCPBillingFile != ""
	case "azure":
		return c.AzureBillingFile != ""
	default:
		return false
	}
}

func (c BillingConfig) UseSimulation(provider string) bool {
	switch c.Mode {
	case BillingModeSimulated:
		return true
	case BillingModeLive:
		return !c.LiveProviderEnabled(provider)
	case BillingModeHybrid:
		return !c.LiveProviderEnabled(provider) && c.AllowSimulationFallback
	default:
		return true
	}
}

func envBool(key string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	return v == "1" || v == "true" || v == "yes"
}
