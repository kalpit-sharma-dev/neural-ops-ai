package logtail

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/neuralops/platform/internal/gateway/websocket"
	"github.com/neuralops/platform/internal/ingestion/dto"
	"github.com/neuralops/platform/internal/kafka"
	"go.uber.org/zap"
)

// Start wires Kafka log consumption into the log tail hub, or demo feed when brokers unset.
func Start(ctx context.Context, log *zap.Logger, hub *websocket.LogTailHub) {
	brokers := strings.TrimSpace(os.Getenv("KAFKA_BROKERS"))
	if brokers == "" || hub == nil {
		log.Info("log tail: kafka unavailable, starting demo feed")
		hub.StartDemoFeed(ctx, 3*time.Second)
		return
	}

	topic := strings.TrimSpace(os.Getenv("LOG_TAIL_KAFKA_TOPIC"))
	if topic == "" {
		topic = "raw-logs"
	}
	if prefix := strings.TrimSpace(os.Getenv("KAFKA_TOPIC_PREFIX")); prefix != "" {
		topic = prefix + topic
	}

	group := strings.TrimSpace(os.Getenv("LOG_TAIL_CONSUMER_GROUP"))
	if group == "" {
		group = "gateway-log-tail"
	}

	consumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: splitBrokers(brokers),
		GroupID: group,
		Topics:  []string{topic},
	}, log, func(_ context.Context, msg kafka.Message) error {
		var enriched dto.EnrichedLog
		if err := json.Unmarshal(msg.Value, &enriched); err != nil {
			return nil
		}
		entry := enriched.LogEntry
		hub.Publish(websocket.LogTailLine{
			Timestamp: entry.Timestamp,
			Service:   entry.Service,
			Severity:  string(entry.Severity),
			Message:   entry.Message,
			TraceID:   entry.TraceID,
			Host:      entry.Host,
			Pod:       entry.Pod,
			TenantID:  enriched.IngestionMetadata.TenantID,
		})
		return nil
	})
	if consumer == nil {
		hub.StartDemoFeed(ctx, 3*time.Second)
		return
	}
	go func() {
		log.Info("log tail kafka consumer started", zap.String("topic", topic), zap.String("group", group))
		if err := consumer.Run(ctx); err != nil && ctx.Err() == nil {
			log.Warn("log tail consumer stopped", zap.Error(err))
		}
	}()
}

func splitBrokers(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
