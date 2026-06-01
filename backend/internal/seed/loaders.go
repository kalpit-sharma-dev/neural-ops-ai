package seed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/neuralops/platform/internal/platform/db"
	"github.com/neuralops/platform/internal/platform/tenant"
)

// LoadClickHouse inserts logs and transactions into ClickHouse.
func LoadClickHouse(ctx context.Context, dsn string, logs []LogRecord, txns []TransactionRecord) error {
	opts, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		return fmt.Errorf("parse clickhouse dsn: %w", err)
	}
	conn, err := clickhouse.Open(opts)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := conn.Ping(ctx); err != nil {
		return err
	}
	if err := db.RunClickHouseMigrations(ctx, conn); err != nil {
		return err
	}

	if err := insertLogs(ctx, conn, logs); err != nil {
		return err
	}
	return insertTransactions(ctx, conn, txns)
}

func insertLogs(ctx context.Context, conn driver.Conn, logs []LogRecord) error {
	batch, err := conn.PrepareBatch(ctx, `
INSERT INTO logs (
  tenant_id, timestamp, service, environment, severity, message,
  trace_id, txn_id, host, pod, classification, explanation
)`)
	if err != nil {
		return err
	}
	for _, log := range logs {
		if err := batch.Append(
			log.TenantID,
			log.Timestamp.UTC(),
			log.Service,
			log.Environment,
			log.Severity,
			log.Message,
			log.TraceID,
			log.TxnID,
			log.Host,
			log.Pod,
			log.Classification,
			"",
		); err != nil {
			return err
		}
	}
	return batch.Send()
}

func insertTransactions(ctx context.Context, conn driver.Conn, txns []TransactionRecord) error {
	batch, err := conn.PrepareBatch(ctx, `
INSERT INTO transaction_journeys (
  txn_id, txn_type, status, failed_at, total_latency_ms, retry_count, hops, started_at, completed_at
)`)
	if err != nil {
		return err
	}
	for _, txn := range txns {
		var completed any
		if txn.CompletedAt != nil {
			completed = txn.CompletedAt.UTC()
		}
		if err := batch.Append(
			txn.TxnID,
			txn.TxnType,
			txn.Status,
			txn.FailedAt,
			txn.TotalLatencyMs,
			txn.RetryCount,
			txn.HopsJSON,
			txn.StartedAt.UTC(),
			completed,
		); err != nil {
			return err
		}
	}
	return batch.Send()
}

// LoadElasticsearch bulk-indexes logs into Elasticsearch for the demo tenant.
func LoadElasticsearch(ctx context.Context, url, index string, logs []LogRecord) error {
	if index == "" && len(logs) > 0 {
		index = tenant.LogsBootstrapIndex(logs[0].TenantID)
	}
	if index == "" {
		index = tenant.LogsBootstrapIndex(DemoTenantID)
	}
	return LoadElasticsearchAuth(ctx, url, index, "", "", logs)
}

// LoadElasticsearchAuth bulk-indexes logs with optional basic auth.
func LoadElasticsearchAuth(ctx context.Context, url, index, username, password string, logs []LogRecord) error {
	tenantID := DemoTenantID
	if len(logs) > 0 && strings.TrimSpace(logs[0].TenantID) != "" {
		tenantID = logs[0].TenantID
	}
	if index == "" {
		index = tenant.LogsBootstrapIndex(tenantID)
	}
	alias := tenant.LogsAlias(tenantID)
	cfg := elasticsearch.Config{Addresses: []string{url}}
	if username != "" {
		cfg.Username = username
		cfg.Password = password
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return err
	}
	_ = db.ApplyElasticsearchLogTemplate(ctx, es, "")

	if err := ensureIndex(ctx, es, index); err != nil {
		return err
	}
	if alias != "" {
		if err := ensureWriteAlias(ctx, es, index, alias); err != nil {
			return err
		}
	}

	indexer, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client: es,
		Index:  index,
	})
	if err != nil {
		return err
	}
	defer indexer.Close(ctx)

	for _, log := range logs {
		doc := map[string]any{
			"id": log.ID, "timestamp": log.Timestamp.UTC().Format(time.RFC3339Nano),
			"tenantId": log.TenantID, "service": log.Service, "environment": log.Environment,
			"severity": log.Severity, "message": log.Message, "traceId": log.TraceID,
			"txnId": log.TxnID, "host": log.Host, "pod": log.Pod,
		}
		if log.Classification != "" {
			doc["classification"] = log.Classification
		}
		body, _ := json.Marshal(doc)
		record := log
		if err := indexer.Add(ctx, esutil.BulkIndexerItem{
			Action:     "index",
			DocumentID: record.ID,
			Body:       bytes.NewReader(body),
		}); err != nil {
			return err
		}
	}
	return indexer.Close(ctx)
}

func ensureIndex(ctx context.Context, es *elasticsearch.Client, index string) error {
	res, err := es.Indices.Exists([]string{index}, es.Indices.Exists.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == 200 {
		return nil
	}
	createRes, err := es.Indices.Create(index, es.Indices.Create.WithContext(ctx))
	if err != nil {
		return err
	}
	defer createRes.Body.Close()
	if createRes.IsError() {
		raw, _ := io.ReadAll(createRes.Body)
		return fmt.Errorf("create index: %s", string(raw))
	}
	return nil
}

func ensureWriteAlias(ctx context.Context, es *elasticsearch.Client, index, alias string) error {
	res, err := es.Indices.GetAlias(es.Indices.GetAlias.WithName(alias), es.Indices.GetAlias.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == 200 {
		return nil
	}

	body := map[string]any{
		"actions": []map[string]any{
			{
				"add": map[string]any{
					"index":          index,
					"alias":          alias,
					"is_write_index": true,
				},
			},
		},
	}
	payload, _ := json.Marshal(body)
	updateRes, err := es.Indices.UpdateAliases(bytes.NewReader(payload), es.Indices.UpdateAliases.WithContext(ctx))
	if err != nil {
		return err
	}
	defer updateRes.Body.Close()
	if updateRes.IsError() {
		raw, _ := io.ReadAll(updateRes.Body)
		return fmt.Errorf("attach write alias: %s", string(raw))
	}
	return nil
}

// PostLogsHTTP sends logs to the ingestion API in batches.
func PostLogsHTTP(ctx context.Context, url string, logs []LogRecord, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 500
	}
	client := &http.Client{Timeout: 60 * time.Second}
	for start := 0; start < len(logs); start += batchSize {
		end := start + batchSize
		if end > len(logs) {
			end = len(logs)
		}
		payload := make([]map[string]any, 0, end-start)
		for _, log := range logs[start:end] {
			payload = append(payload, map[string]any{
				"timestamp":   log.Timestamp,
				"service":     log.Service,
				"environment": log.Environment,
				"severity":    log.Severity,
				"message":     log.Message,
				"traceId":     log.TraceID,
				"txnId":       log.TxnID,
				"tenantId":    log.TenantID,
				"host":        log.Host,
				"pod":         log.Pod,
			})
		}
		body, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		resp.Body.Close()
		if resp.StatusCode >= 300 {
			return fmt.Errorf("ingestion batch %d-%d failed with status %d", start, end, resp.StatusCode)
		}
	}
	return nil
}
