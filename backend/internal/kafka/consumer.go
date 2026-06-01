package kafka

import (
	"context"
	"sync"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/neuralops/platform/pkg/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// HandlerFunc processes a Kafka message.
type HandlerFunc func(ctx context.Context, message Message) error

// Consumer reads messages from Kafka topics with a worker pool.
type Consumer struct {
	reader      *kafkago.Reader
	log         *zap.Logger
	handler     HandlerFunc
	concurrency int
	serviceName string
	groupID     string
	topic       string
}

// ConsumerConfig configures a Kafka consumer.
type ConsumerConfig struct {
	Brokers     []string
	GroupID     string
	Topics      []string
	Concurrency int
	ServiceName string
}

// NewConsumer creates a Kafka consumer.
func NewConsumer(cfg ConsumerConfig, log *zap.Logger, handler HandlerFunc) *Consumer {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 4
	}
	if cfg.GroupID == "" || len(cfg.Topics) == 0 {
		log.Warn("kafka consumer disabled: group id and at least one topic are required",
			zap.String("group_id", cfg.GroupID),
			zap.Strings("topics", cfg.Topics),
		)
		return nil
	}

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:     cfg.Brokers,
		GroupID:     cfg.GroupID,
		GroupTopics: cfg.Topics,
		MinBytes:    1,
		MaxBytes:    10e6,
	})

	topic := ""
	if len(cfg.Topics) > 0 {
		topic = cfg.Topics[0]
	}

	return &Consumer{
		reader:      reader,
		log:         log,
		handler:     handler,
		concurrency: cfg.Concurrency,
		serviceName: cfg.ServiceName,
		groupID:     cfg.GroupID,
		topic:       topic,
	}
}

// Run starts consuming until context cancellation.
func (c *Consumer) Run(ctx context.Context) error {
	if c == nil || c.reader == nil {
		<-ctx.Done()
		return nil
	}
	if c.serviceName != "" && c.groupID != "" && c.topic != "" {
		go c.reportLag(ctx)
	}

	jobs := make(chan kafkago.Message, c.concurrency*2)
	var wg sync.WaitGroup

	for i := 0; i < c.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tracer := otel.Tracer("kafka.consumer")
			for message := range jobs {
				msgCtx, span := tracer.Start(ctx, "kafka.consume",
					trace.WithSpanKind(trace.SpanKindConsumer),
					trace.WithAttributes(
						attribute.String("messaging.system", "kafka"),
						attribute.String("messaging.destination", message.Topic),
						attribute.String("messaging.kafka.consumer_group", c.groupID),
					),
				)
				err := c.handler(msgCtx, Message{Topic: message.Topic, Key: string(message.Key), Value: message.Value})
				if err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
					c.log.Warn("kafka message handling failed", zap.Error(err), zap.String("topic", message.Topic))
				}
				span.End()
			}
		}()
	}

	for {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return c.reader.Close()
		default:
			message, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					close(jobs)
					wg.Wait()
					return c.reader.Close()
				}
				c.log.Warn("kafka fetch failed", zap.Error(err))
				continue
			}
			jobs <- message
			if err := c.reader.CommitMessages(ctx, message); err != nil {
				c.log.Warn("kafka commit failed", zap.Error(err))
			}
		}
	}
}

// Close closes the consumer.
func (c *Consumer) Close() error {
	if c == nil || c.reader == nil {
		return nil
	}
	return c.reader.Close()
}

func (c *Consumer) reportLag(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := c.reader.Stats()
			metrics.SetKafkaConsumerLag(c.serviceName, c.groupID, c.topic, stats.Lag)
		}
	}
}
