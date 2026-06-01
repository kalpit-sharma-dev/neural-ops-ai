package seed

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Options controls the full demo seed run.
type Options struct {
	PostgresDSN      string
	ClickHouseDSN    string
	ElasticsearchURL string
	IngestionURL     string
	OutputDir        string
	LogCount         int
	SkipPostgres     bool
	SkipClickHouse   bool
	SkipElasticsearch bool
	SkipHTTP         bool
	SkipExport       bool
}

// Run executes the Phase 20 demo seed pipeline.
func Run(ctx context.Context, opts Options) (Summary, error) {
	if opts.LogCount <= 0 {
		opts.LogCount = DefaultLogCount
	}
	now := time.Now().UTC()
	logs := GenerateAllLogs(now, opts.LogCount)
	txns := GenerateTransactions(now)
	deployments := GenerateDeployments(now)

	summary := Summary{
		LogCount:        len(logs),
		TransactionCount: len(txns),
		DeploymentCount: len(deployments),
		IncidentCount:   30,
	}

	if !opts.SkipExport {
		outputDir := opts.OutputDir
		if outputDir == "" {
			outputDir = "scripts/seed/output"
		}
		if err := ExportArtifacts(outputDir, logs, txns, deployments); err != nil {
			return summary, fmt.Errorf("export artifacts: %w", err)
		}
		summary.ExportDir = outputDir
		log.Printf("exported seed artifacts to %s", outputDir)
	}

	if !opts.SkipPostgres && opts.PostgresDSN != "" {
		if err := SeedPostgres(ctx, PostgresOptions{DSN: opts.PostgresDSN}); err != nil {
			return summary, fmt.Errorf("seed postgres: %w", err)
		}
		log.Printf("seeded postgres relational data")
	}

	if !opts.SkipClickHouse && opts.ClickHouseDSN != "" {
		if err := LoadClickHouse(ctx, opts.ClickHouseDSN, logs, txns); err != nil {
			return summary, fmt.Errorf("seed clickhouse: %w", err)
		}
		log.Printf("loaded %d logs and %d transactions into clickhouse", len(logs), len(txns))
	}

	if !opts.SkipElasticsearch && opts.ElasticsearchURL != "" {
		if err := LoadElasticsearch(ctx, opts.ElasticsearchURL, "", logs); err != nil {
			return summary, fmt.Errorf("seed elasticsearch: %w", err)
		}
		log.Printf("bulk indexed %d logs into elasticsearch", len(logs))
	}

	if !opts.SkipHTTP && opts.IngestionURL != "" {
		if err := PostLogsHTTP(ctx, opts.IngestionURL, logs, chunkSize); err != nil {
			log.Printf("warning: ingestion HTTP seed skipped: %v", err)
		} else {
			log.Printf("posted %d logs to ingestion API", len(logs))
		}
	}

	return summary, nil
}

// Summary describes generated demo data volumes.
type Summary struct {
	LogCount         int
	TransactionCount int
	DeploymentCount  int
	IncidentCount    int
	ExportDir        string
}
