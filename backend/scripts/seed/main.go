package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/neuralops/platform/internal/seed"
)

func main() {
	dsn := flag.String("dsn", envOr("POSTGRES_DSN", "postgres://neuralops:neuralops@localhost:5432/neuralops?sslmode=disable"), "PostgreSQL DSN")
	clickhouseDSN := flag.String("clickhouse-dsn", envOr("CLICKHOUSE_DSN", "clickhouse://default:@localhost:9000/neuralops"), "ClickHouse DSN")
	esURL := flag.String("es-url", envOr("ELASTICSEARCH_URL", "http://localhost:9200"), "Elasticsearch URL")
	ingestionURL := flag.String("ingestion-url", envOr("INGESTION_URL", "http://localhost:8081/api/v1/logs"), "Ingestion logs endpoint")
	outputDir := flag.String("output-dir", envOr("SEED_OUTPUT_DIR", "scripts/seed/output"), "Directory for export artifacts")
	logCount := flag.Int("logs", seed.DefaultLogCount, "Number of baseline log entries to generate")
	skipPostgres := flag.Bool("skip-postgres", false, "Skip PostgreSQL seed")
	skipClickHouse := flag.Bool("skip-clickhouse", false, "Skip ClickHouse load")
	skipElasticsearch := flag.Bool("skip-elasticsearch", false, "Skip Elasticsearch bulk index")
	skipHTTP := flag.Bool("skip-http", false, "Skip ingestion HTTP posts")
	skipExport := flag.Bool("skip-export", false, "Skip writing export files")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Minute)
	defer cancel()

	summary, err := seed.Run(ctx, seed.Options{
		PostgresDSN:       *dsn,
		ClickHouseDSN:     *clickhouseDSN,
		ElasticsearchURL:  *esURL,
		IngestionURL:      *ingestionURL,
		OutputDir:         *outputDir,
		LogCount:          *logCount,
		SkipPostgres:      *skipPostgres,
		SkipClickHouse:    *skipClickHouse,
		SkipElasticsearch: *skipElasticsearch,
		SkipHTTP:          *skipHTTP,
		SkipExport:        *skipExport,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "seed failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("")
	fmt.Println("NeuralOps demo data seeded successfully.")
	fmt.Printf("  Logs:         %d (+ UPI outage simulation logs)\n", summary.LogCount)
	fmt.Printf("  Transactions: %d (includes %d failed UPI outage txns)\n", summary.TransactionCount, seed.UPIFailedTxnCount)
	fmt.Printf("  Deployments:  %d\n", summary.DeploymentCount)
	fmt.Printf("  Incidents:    %d\n", summary.IncidentCount)
	if summary.ExportDir != "" {
		fmt.Printf("  Export dir:   %s\n", summary.ExportDir)
	}
	fmt.Printf("  Email:        %s\n", seed.DemoEmail)
	fmt.Printf("  Password:     %s\n", seed.DemoPassword)
	fmt.Println("")
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
