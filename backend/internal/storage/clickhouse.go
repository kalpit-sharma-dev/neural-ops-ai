package storage

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/neuralops/platform/internal/domain"
	"github.com/neuralops/platform/internal/platform/db"
	"go.uber.org/zap"
)

// MetricRow represents a row in the ClickHouse metrics table.
type MetricRow struct {
	TenantID    string
	ServiceName string
	Host        string
	Pod         string
	MetricType  domain.MetricType
	Value       float64
	Timestamp   time.Time
	Labels      map[string]string
}

// ClickHouseWriter performs async batched inserts into ClickHouse.
type ClickHouseWriter struct {
	conn         driver.Conn
	log          *zap.Logger
	table        string
	maxBatchSize int
	flushEvery   time.Duration
	buffer       []MetricRow
	mu           sync.Mutex
	flushTicker  *time.Ticker
	done         chan struct{}
	wg           sync.WaitGroup
}

// ClickHouseConfig configures the ClickHouse writer.
type ClickHouseConfig struct {
	DSN           string
	Table         string
	MaxBatchSize  int
	FlushInterval time.Duration
}

// NewClickHouseWriter creates a ClickHouse metrics writer.
func NewClickHouseWriter(ctx context.Context, cfg ClickHouseConfig, log *zap.Logger) (*ClickHouseWriter, error) {
	if cfg.MaxBatchSize <= 0 {
		cfg.MaxBatchSize = 50000
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = time.Second
	}
	if cfg.Table == "" {
		cfg.Table = "metrics"
	}

	opts, err := clickhouse.ParseDSN(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse clickhouse dsn: %w", err)
	}

	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("open clickhouse: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping clickhouse: %w", err)
	}
	if err := db.RunClickHouseMigrations(ctx, conn); err != nil {
		return nil, fmt.Errorf("clickhouse migrations: %w", err)
	}

	writer := &ClickHouseWriter{
		conn:         conn,
		log:          log,
		table:        cfg.Table,
		maxBatchSize: cfg.MaxBatchSize,
		flushEvery:   cfg.FlushInterval,
		buffer:       make([]MetricRow, 0, cfg.MaxBatchSize),
		flushTicker:  time.NewTicker(cfg.FlushInterval),
		done:         make(chan struct{}),
	}

	if err := writer.ensureSchema(ctx); err != nil {
		_ = conn.Close()
		return nil, err
	}

	writer.wg.Add(1)
	go writer.flushLoop()

	return writer, nil
}

func (w *ClickHouseWriter) ensureSchema(ctx context.Context) error {
	query := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
    tenant_id String DEFAULT 'default',
    service_name String,
    host String,
    pod String,
    metric_type String,
    value Float64,
    timestamp DateTime64(3, 'UTC'),
    labels Map(String, String)
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, service_name, metric_type, timestamp)`, w.table)

	if err := w.conn.Exec(ctx, query); err != nil {
		return err
	}
	return w.conn.Exec(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS tenant_id String DEFAULT 'default'", w.table))
}

// Write queues a metric row for async insert.
func (w *ClickHouseWriter) Write(row MetricRow) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buffer = append(w.buffer, row)
	if len(w.buffer) >= w.maxBatchSize {
		return w.flushLocked(context.Background())
	}
	return nil
}

// Close flushes pending rows and closes the connection.
func (w *ClickHouseWriter) Close() error {
	close(w.done)
	w.flushTicker.Stop()
	w.wg.Wait()

	w.mu.Lock()
	err := w.flushLocked(context.Background())
	w.mu.Unlock()

	closeErr := w.conn.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func (w *ClickHouseWriter) flushLoop() {
	defer w.wg.Done()
	for {
		select {
		case <-w.done:
			return
		case <-w.flushTicker.C:
			w.mu.Lock()
			if err := w.flushLocked(context.Background()); err != nil {
				w.log.Warn("clickhouse periodic flush failed", zap.Error(err))
			}
			w.mu.Unlock()
		}
	}
}

func (w *ClickHouseWriter) flushLocked(ctx context.Context) error {
	if len(w.buffer) == 0 {
		return nil
	}

	batch, err := w.conn.PrepareBatch(ctx, fmt.Sprintf(
		"INSERT INTO %s (tenant_id, service_name, host, pod, metric_type, value, timestamp, labels)",
		w.table,
	))
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	rows := w.buffer
	w.buffer = w.buffer[:0]

	for _, row := range rows {
		tenantID := row.TenantID
		if tenantID == "" {
			tenantID = "default"
		}
		if err := batch.Append(
			tenantID,
			row.ServiceName,
			row.Host,
			row.Pod,
			string(row.MetricType),
			row.Value,
			row.Timestamp.UTC(),
			row.Labels,
		); err != nil {
			return fmt.Errorf("append row: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("send batch: %w", err)
	}

	w.log.Debug("clickhouse batch inserted", zap.Int("rows", len(rows)))
	return nil
}

// ParseDSNHost extracts host from a clickhouse DSN for logging.
func ParseDSNHost(dsn string) string {
	if idx := strings.Index(dsn, "@"); idx >= 0 {
		rest := dsn[idx+1:]
		if end := strings.Index(rest, "/"); end >= 0 {
			return rest[:end]
		}
		return rest
	}
	return dsn
}
