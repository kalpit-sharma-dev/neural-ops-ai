package streaming_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/neuralops/platform/internal/kafka"
	"github.com/neuralops/platform/internal/streaming"
	"go.uber.org/zap"
)

func TestMaterializerHandlerAggregatesAlertSignal(t *testing.T) {
	mat := streaming.NewMaterializer(streaming.NewStore(nil), zap.NewNop())
	payload, _ := json.Marshal(streaming.AlertSignalPayload{
		PolicyID: "ap-1", Service: "payments", Count: 2, Timestamp: time.Now().UTC(),
	})
	raw, _ := streaming.PublishEvent(streaming.StreamEvent{
		Type: streaming.EventAlertSignal, TenantID: "default", Payload: payload,
	})
	if err := mat.Handler()(context.Background(), kafka.Message{Value: raw}); err != nil {
		t.Fatal(err)
	}
}
