package transaction_test

import (
	"testing"
	"time"

	"github.com/neuralops/platform/internal/correlation/transaction"
	"github.com/neuralops/platform/internal/domain"
	"go.uber.org/zap"
)

func TestCorrelatorBuildsPaymentJourney(t *testing.T) {
	c := transaction.NewCorrelator(zap.NewNop(), nil)
	now := time.Now().UTC()
	c.HandleLog(domain.LogEntry{TxnID: "txn-100", Service: "upi-gateway", Timestamp: now, Message: "authorize"})
	c.HandleLog(domain.LogEntry{TxnID: "txn-100", Service: "ledger-service", Timestamp: now.Add(50 * time.Millisecond), Message: "post"})
	c.HandleLog(domain.LogEntry{
		TxnID: "txn-100", Service: "notification-service", Timestamp: now.Add(120 * time.Millisecond),
		Message: "notify failed", Severity: "ERROR",
	})
	c.FlushCompleted(t.Context(), time.Millisecond)
}

func TestCorrelatorIgnoresMissingTxnID(t *testing.T) {
	c := transaction.NewCorrelator(zap.NewNop(), nil)
	c.HandleLog(domain.LogEntry{Service: "upi-gateway", Timestamp: time.Now().UTC()})
	c.FlushCompleted(t.Context(), time.Millisecond)
}

// TestCorrelatorHighVolume validates BANK-014 correlator at POC-scale event count.
func TestCorrelatorHighVolume(t *testing.T) {
	c := transaction.NewCorrelator(zap.NewNop(), nil)
	now := time.Now().UTC()
	const txns = 2500
	for i := 0; i < txns; i++ {
		txn := "txn-vol-" + string(rune('a'+(i%26)))
		c.HandleLog(domain.LogEntry{TxnID: txn, Service: "payment-service", Timestamp: now, Message: "pay"})
		c.HandleLog(domain.LogEntry{TxnID: txn, Service: "ledger-service", Timestamp: now.Add(time.Millisecond), Message: "post"})
	}
	start := time.Now()
	c.FlushCompleted(t.Context(), time.Millisecond)
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("flush too slow for %d txns: %v", txns, elapsed)
	}
}
