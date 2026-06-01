package seed

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ExportArtifacts writes SQL, Elasticsearch bulk, and ClickHouse CSV files.
func ExportArtifacts(outputDir string, logs []LogRecord, txns []TransactionRecord, deployments []DeploymentRecord) error {
	if outputDir == "" {
		outputDir = "scripts/seed/output"
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	if err := exportLogsBulkNDJSON(filepath.Join(outputDir, "logs_bulk.ndjson"), logs); err != nil {
		return err
	}
	if err := exportLogsCSV(filepath.Join(outputDir, "logs.csv"), logs); err != nil {
		return err
	}
	if err := exportTransactionsCSV(filepath.Join(outputDir, "transactions.csv"), txns); err != nil {
		return err
	}
	if err := exportDeploymentsSQL(filepath.Join(outputDir, "deployments.sql"), deployments); err != nil {
		return err
	}
	return writeManifest(outputDir, len(logs), len(txns), len(deployments))
}

func exportLogsBulkNDJSON(path string, logs []LogRecord) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	index := "neuralops-logs-000001"
	enc := json.NewEncoder(file)
	for _, log := range logs {
		meta := map[string]any{"index": map[string]any{"_index": index, "_id": log.ID}}
		if err := enc.Encode(meta); err != nil {
			return err
		}
		doc := map[string]any{
			"id": log.ID, "timestamp": log.Timestamp.UTC().Format(time.RFC3339Nano),
			"tenantId": log.TenantID, "service": log.Service, "environment": log.Environment,
			"severity": log.Severity, "message": log.Message, "traceId": log.TraceID,
			"txnId": log.TxnID, "host": log.Host, "pod": log.Pod,
		}
		if log.Classification != "" {
			doc["classification"] = log.Classification
		}
		if err := enc.Encode(doc); err != nil {
			return err
		}
	}
	return nil
}

func exportLogsCSV(path string, logs []LogRecord) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	_ = writer.Write([]string{
		"tenant_id", "timestamp", "service", "environment", "severity", "message",
		"trace_id", "txn_id", "host", "pod", "classification", "explanation",
	})
	for _, log := range logs {
		_ = writer.Write([]string{
			log.TenantID,
			log.Timestamp.UTC().Format("2006-01-02 15:04:05.000"),
			log.Service,
			log.Environment,
			log.Severity,
			escapeCSV(log.Message),
			log.TraceID,
			log.TxnID,
			log.Host,
			log.Pod,
			log.Classification,
			"",
		})
	}
	return writer.Error()
}

func exportTransactionsCSV(path string, txns []TransactionRecord) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	_ = writer.Write([]string{
		"txn_id", "txn_type", "status", "failed_at", "total_latency_ms",
		"retry_count", "hops", "started_at", "completed_at", "tenant_id",
	})
	for _, txn := range txns {
		completed := ""
		if txn.CompletedAt != nil {
			completed = txn.CompletedAt.UTC().Format("2006-01-02 15:04:05.000")
		}
		_ = writer.Write([]string{
			txn.TxnID,
			txn.TxnType,
			txn.Status,
			txn.FailedAt,
			fmt.Sprintf("%d", txn.TotalLatencyMs),
			fmt.Sprintf("%d", txn.RetryCount),
			escapeCSV(txn.HopsJSON),
			txn.StartedAt.UTC().Format("2006-01-02 15:04:05.000"),
			completed,
			DemoTenantID,
		})
	}
	return writer.Error()
}

func exportDeploymentsSQL(path string, deployments []DeploymentRecord) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintln(file, "-- Generated NeuralOps deployment seed SQL")
	for _, dep := range deployments {
		meta := fmt.Sprintf(`'{"category":"%s"}'`, dep.Category)
		fmt.Fprintf(file, `
INSERT INTO deployments (id, tenant_id, service, version, environment, deployed_by, deployed_at, change_type, metadata)
VALUES ('%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', %s::jsonb)
ON CONFLICT (id) DO NOTHING;`,
			dep.ID, DemoTenantID, dep.Service, dep.Version, dep.Environment,
			dep.DeployedBy, dep.DeployedAt.UTC().Format(time.RFC3339), dep.ChangeType, meta,
		)
	}
	return nil
}

func writeManifest(dir string, logs, txns, deployments int) error {
	manifest := map[string]any{
		"generatedAt": time.Now().UTC().Format(time.RFC3339),
		"logCount":    logs,
		"txnCount":    txns,
		"deployCount": deployments,
		"files": []string{
			"logs_bulk.ndjson",
			"logs.csv",
			"transactions.csv",
			"deployments.sql",
		},
	}
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "manifest.json"), raw, 0o644)
}

func escapeCSV(value string) string {
	if strings.ContainsAny(value, ",\"\n") {
		return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
	}
	return value
}
