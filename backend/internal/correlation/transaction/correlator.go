package transaction

import (
	"context"
	"sync"
	"time"

	"github.com/neuralops/platform/internal/correlation/repository"
	"github.com/neuralops/platform/internal/domain"
	"go.uber.org/zap"
)

type pendingTxn struct {
	txnID   string
	hops    []domain.TxnHop
	started time.Time
}

// Correlator builds banking transaction journeys grouped by txnId.
type Correlator struct {
	log   *zap.Logger
	store *repository.ClickHouseStore
	mu    sync.Mutex
	pending map[string]*pendingTxn
}

// NewCorrelator creates a transaction correlator.
func NewCorrelator(log *zap.Logger, store *repository.ClickHouseStore) *Correlator {
	return &Correlator{
		log:     log,
		store:   store,
		pending: make(map[string]*pendingTxn),
	}
}

// HandleLog appends a hop for a transaction ID.
func (c *Correlator) HandleLog(entry domain.LogEntry) {
	if entry.TxnID == "" {
		return
	}

	status := domain.HopStatusSuccess
	if entry.IsErrorSeverity() {
		status = domain.HopStatusFailed
	}

	hop := domain.TxnHop{
		ServiceName:  entry.Service,
		SpanID:       entry.TraceID,
		TraceID:      entry.TraceID,
		LatencyMs:    0,
		Status:       status,
		ErrorMessage: entry.Message,
		StartedAt:    entry.Timestamp,
	}

	c.mu.Lock()
	item, ok := c.pending[entry.TxnID]
	if !ok {
		item = &pendingTxn{txnID: entry.TxnID, started: entry.Timestamp}
		c.pending[entry.TxnID] = item
	}
	item.hops = append(item.hops, hop)
	c.mu.Unlock()
}

// FlushCompleted persists journeys that appear complete or stale.
func (c *Correlator) FlushCompleted(ctx context.Context, maxAge time.Duration) {
	now := time.Now().UTC()
	c.mu.Lock()
	defer c.mu.Unlock()

	for txnID, item := range c.pending {
		if now.Sub(item.started) < maxAge && !c.hasFailure(item.hops) {
			continue
		}
		if c.store == nil {
			delete(c.pending, txnID)
			continue
		}
		txn := c.buildTransaction(item)
		if err := c.store.SaveTransaction(ctx, txn); err != nil {
			c.log.Warn("save transaction journey failed", zap.Error(err), zap.String("txnId", txnID))
		} else {
			delete(c.pending, txnID)
		}
	}
}

func (c *Correlator) buildTransaction(item *pendingTxn) domain.Transaction {
	status := domain.TxnStatusSuccess
	failedAt := ""
	var totalLatency int64

	for _, hop := range item.hops {
		totalLatency += hop.LatencyMs
		if hop.Status == domain.HopStatusFailed || hop.Status == domain.HopStatusPartial {
			if status == domain.TxnStatusSuccess {
				status = domain.TxnStatusFailed
				failedAt = hop.ServiceName
			} else if status == domain.TxnStatusFailed {
				status = domain.TxnStatusPartial
			}
		}
	}

	return domain.Transaction{
		TxnID:          item.txnID,
		TxnType:        domain.TxnTypeUPI,
		Status:         status,
		Hops:           item.hops,
		TotalLatencyMs: totalLatency,
		FailedAt:       failedAt,
		RetryCount:     0,
		StartedAt:      item.started,
	}
}

func (c *Correlator) hasFailure(hops []domain.TxnHop) bool {
	for _, hop := range hops {
		if hop.Status == domain.HopStatusFailed || hop.Status == domain.HopStatusPartial {
			return true
		}
	}
	return false
}

// StartFlusher periodically persists transaction journeys.
func (c *Correlator) StartFlusher(ctx context.Context, interval, maxAge time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				c.FlushCompleted(context.Background(), maxAge)
				return
			case <-ticker.C:
				c.FlushCompleted(ctx, maxAge)
			}
		}
	}()
}

// GetTransaction retrieves a transaction from ClickHouse.
func (c *Correlator) GetTransaction(ctx context.Context, txnID string) (*domain.Transaction, error) {
	return c.store.GetTransaction(ctx, txnID)
}
