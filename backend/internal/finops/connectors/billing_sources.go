package connectors

import (
	"os"
	"time"
)

// BillingSourceStatus describes one configured billing feed (FIN-PROD-01..03).
type BillingSourceStatus struct {
	Provider      string    `json:"provider"`
	Mode          string    `json:"mode"`
	FilePath      string    `json:"filePath,omitempty"`
	FileExists    bool      `json:"fileExists"`
	FileSizeBytes int64     `json:"fileSizeBytes,omitempty"`
	InvoiceUSD    float64   `json:"invoiceUsd,omitempty"`
	LiveEnabled   bool      `json:"liveEnabled"`
	Simulation    bool      `json:"simulationFallback"`
	CheckedAt     time.Time `json:"checkedAt"`
}

// BillingSourcesReport aggregates live billing configuration health.
type BillingSourcesReport struct {
	BillingMode string                `json:"billingMode"`
	Sources     []BillingSourceStatus `json:"sources"`
	Ready       bool                  `json:"ready"`
	Message     string                `json:"message,omitempty"`
}

// BillingSourcesStatus inspects env-configured billing files for Gate F readiness.
func BillingSourcesStatus() BillingSourcesReport {
	cfg := LoadBillingConfig()
	sources := []BillingSourceStatus{
		sourceStatus(cfg, "aws", cfg.AWS_CURFile, cfg.AWSInvoiceUSD),
		sourceStatus(cfg, "gcp", cfg.GCPBillingFile, 0),
		sourceStatus(cfg, "azure", cfg.AzureBillingFile, 0),
	}
	ready := cfg.Mode == BillingModeLive || cfg.Mode == BillingModeHybrid
	liveCount := 0
	for _, s := range sources {
		if s.LiveEnabled && s.FileExists {
			liveCount++
		}
		if cfg.Mode == BillingModeLive && s.LiveEnabled && !s.FileExists && s.FilePath != "" {
			ready = false
		}
	}
	msg := "simulated billing"
	if cfg.Mode == BillingModeLive {
		msg = "live billing"
		if liveCount == 0 {
			msg = "live mode but no billing files found on disk"
			ready = false
		}
	} else if cfg.Mode == BillingModeHybrid {
		msg = "hybrid billing"
	}
	return BillingSourcesReport{
		BillingMode: cfg.Mode,
		Sources:     sources,
		Ready:       ready,
		Message:     msg,
	}
}

func sourceStatus(cfg BillingConfig, provider, path string, invoice float64) BillingSourceStatus {
	st := BillingSourceStatus{
		Provider:    provider,
		Mode:        cfg.Mode,
		FilePath:    path,
		InvoiceUSD:  invoice,
		LiveEnabled: cfg.LiveProviderEnabled(provider),
		Simulation:  cfg.UseSimulation(provider),
		CheckedAt:   time.Now().UTC(),
	}
	if path != "" {
		if info, err := os.Stat(path); err == nil {
			st.FileExists = true
			st.FileSizeBytes = info.Size()
		}
	}
	return st
}
