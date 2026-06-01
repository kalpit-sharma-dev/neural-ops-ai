package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/neuralops/platform/internal/domain"
	"github.com/neuralops/platform/internal/platform/db"
)

// ClickHouseStore stores high-volume transaction journeys.
type ClickHouseStore struct {
	conn  driver.Conn
	table string
}

// NewClickHouseStore creates a ClickHouse repository.
func NewClickHouseStore(ctx context.Context, dsn, table string) (*ClickHouseStore, error) {
	if table == "" {
		table = "transaction_journeys"
	}
	opts, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse clickhouse dsn: %w", err)
	}
	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}

	store := &ClickHouseStore{conn: conn, table: table}
	if err := db.RunClickHouseMigrations(ctx, conn); err != nil {
		return nil, fmt.Errorf("clickhouse migrations: %w", err)
	}
	if err := store.ensureSchema(ctx); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *ClickHouseStore) ensureSchema(ctx context.Context) error {
	query := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
    txn_id String,
    txn_type String,
    status String,
    failed_at String,
    total_latency_ms Int64,
    retry_count Int32,
    hops String,
    started_at DateTime64(3, 'UTC'),
    completed_at Nullable(DateTime64(3, 'UTC'))
) ENGINE = MergeTree()
ORDER BY (txn_id, started_at)`, s.table)
	return s.conn.Exec(ctx, query)
}

// SaveTransaction stores a transaction journey.
func (s *ClickHouseStore) SaveTransaction(ctx context.Context, txn domain.Transaction) error {
	hopsJSON := "[]"
	if len(txn.Hops) > 0 {
		raw, err := json.Marshal(txn.Hops)
		if err != nil {
			return err
		}
		hopsJSON = string(raw)
	}

	var completedAt any
	if txn.CompletedAt != nil {
		completedAt = txn.CompletedAt.UTC()
	}

	return s.conn.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s (txn_id, txn_type, status, failed_at, total_latency_ms, retry_count, hops, started_at, completed_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, s.table),
		txn.TxnID, txn.TxnType, txn.Status, txn.FailedAt, txn.TotalLatencyMs, txn.RetryCount, hopsJSON, txn.StartedAt.UTC(), completedAt,
	)
}

// GetTransaction retrieves a transaction journey by ID.
func (s *ClickHouseStore) GetTransaction(ctx context.Context, txnID string) (*domain.Transaction, error) {
	row := s.conn.QueryRow(ctx, fmt.Sprintf(`
SELECT txn_id, txn_type, status, failed_at, total_latency_ms, retry_count, hops, started_at, completed_at
FROM %s WHERE txn_id = ? ORDER BY started_at DESC LIMIT 1`, s.table), txnID)

	var txn domain.Transaction
	var hopsJSON string
	var completedAt *time.Time
	if err := row.Scan(&txn.TxnID, &txn.TxnType, &txn.Status, &txn.FailedAt, &txn.TotalLatencyMs, &txn.RetryCount, &hopsJSON, &txn.StartedAt, &completedAt); err != nil {
		return nil, err
	}
	if completedAt != nil {
		txn.CompletedAt = completedAt
	}
	if hopsJSON != "" {
		if err := json.Unmarshal([]byte(hopsJSON), &txn.Hops); err != nil {
			return nil, err
		}
	}
	return &txn, nil
}

// TransactionFilter filters transaction journey searches.
type TransactionFilter struct {
	TxnID         string
	TxnType       string
	Status        string
	FailedService string
	StartTime     *time.Time
	EndTime       *time.Time
	Limit         int
}

// SearchTransactions queries transaction journeys with optional filters.
func (s *ClickHouseStore) SearchTransactions(ctx context.Context, filter TransactionFilter) ([]domain.Transaction, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	query := fmt.Sprintf(`
SELECT txn_id, txn_type, status, failed_at, total_latency_ms, retry_count, hops, started_at, completed_at
FROM %s WHERE 1=1`, s.table)
	args := make([]any, 0, 8)

	if filter.TxnID != "" {
		query += " AND txn_id = ?"
		args = append(args, filter.TxnID)
	}
	if filter.TxnType != "" {
		query += " AND txn_type = ?"
		args = append(args, filter.TxnType)
	}
	if filter.Status != "" {
		query += " AND status = ?"
		args = append(args, filter.Status)
	}
	if filter.FailedService != "" {
		query += " AND failed_at = ?"
		args = append(args, filter.FailedService)
	}
	if filter.StartTime != nil {
		query += " AND started_at >= ?"
		args = append(args, filter.StartTime.UTC())
	}
	if filter.EndTime != nil {
		query += " AND started_at <= ?"
		args = append(args, filter.EndTime.UTC())
	}
	query += " ORDER BY started_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]domain.Transaction, 0)
	for rows.Next() {
		var txn domain.Transaction
		var hopsJSON string
		var completedAt *time.Time
		if err := rows.Scan(&txn.TxnID, &txn.TxnType, &txn.Status, &txn.FailedAt, &txn.TotalLatencyMs, &txn.RetryCount, &hopsJSON, &txn.StartedAt, &completedAt); err != nil {
			return nil, err
		}
		if completedAt != nil {
			txn.CompletedAt = completedAt
		}
		if hopsJSON != "" {
			if err := json.Unmarshal([]byte(hopsJSON), &txn.Hops); err != nil {
				return nil, err
			}
		}
		results = append(results, txn)
	}
	return results, rows.Err()
}
