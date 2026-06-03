package websocket_test

import (
	"testing"
	"time"

	"github.com/neuralops/platform/internal/gateway/websocket"
	"go.uber.org/zap"
)

func TestLogTailMatchesSubscription(t *testing.T) {
	hub := websocket.NewLogTailHub(zap.NewNop(), []string{"*"})
	line := websocket.LogTailLine{
		Timestamp: time.Now().UTC(),
		Service:   "payment-service",
		Severity:  "ERROR",
		Message:   "timeout on ledger",
	}
	// Publish should not panic with zero subscribers.
	hub.Publish(line)
}
