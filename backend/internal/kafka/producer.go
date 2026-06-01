package kafka

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Message represents a Kafka message to publish.
type Message struct {
	Topic string
	Key   string
	Value []byte
}

// Producer publishes messages to Kafka with async batching.
type Producer struct {
	writer      *kafka.Writer
	dlqWriter   *kafka.Writer
	log         *zap.Logger
	batchSize   int
	flushEvery  time.Duration
	buffer      []Message
	mu          sync.Mutex
	flushTicker *time.Ticker
	done        chan struct{}
	wg          sync.WaitGroup
}

// ProducerConfig configures the Kafka producer.
type ProducerConfig struct {
	Brokers      []string
	BatchSize    int
	FlushEvery   time.Duration
	DLQTopic     string
	RequiredAcks kafka.RequiredAcks
}

// NewProducer creates an async batched Kafka producer.
func NewProducer(cfg ProducerConfig, log *zap.Logger) *Producer {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 500
	}
	if cfg.FlushEvery <= 0 {
		cfg.FlushEvery = time.Second
	}
	if cfg.RequiredAcks == 0 {
		cfg.RequiredAcks = kafka.RequireOne
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     &kafka.Hash{},
		BatchSize:    cfg.BatchSize,
		BatchTimeout: cfg.FlushEvery,
		RequiredAcks: cfg.RequiredAcks,
		Async:        true,
	}

	dlqWriter := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.DLQTopic,
		BatchSize:    cfg.BatchSize,
		BatchTimeout: cfg.FlushEvery,
		RequiredAcks: cfg.RequiredAcks,
		Async:        true,
	}

	p := &Producer{
		writer:      writer,
		dlqWriter:   dlqWriter,
		log:         log,
		batchSize:   cfg.BatchSize,
		flushEvery:  cfg.FlushEvery,
		buffer:      make([]Message, 0, cfg.BatchSize),
		flushTicker: time.NewTicker(cfg.FlushEvery),
		done:        make(chan struct{}),
	}

	p.wg.Add(1)
	go p.flushLoop()

	return p
}

// Publish queues a message for async delivery.
func (p *Producer) Publish(_ context.Context, msg Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.buffer = append(p.buffer, msg)
	if len(p.buffer) >= p.batchSize {
		return p.flushLocked(context.Background())
	}
	return nil
}

// Close flushes pending messages and stops background workers.
func (p *Producer) Close() error {
	close(p.done)
	p.flushTicker.Stop()
	p.wg.Wait()

	p.mu.Lock()
	err := p.flushLocked(context.Background())
	p.mu.Unlock()

	closeErr := p.writer.Close()
	dlqCloseErr := p.dlqWriter.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return dlqCloseErr
}

func (p *Producer) flushLoop() {
	defer p.wg.Done()
	for {
		select {
		case <-p.done:
			return
		case <-p.flushTicker.C:
			p.mu.Lock()
			if err := p.flushLocked(context.Background()); err != nil {
				p.log.Warn("kafka periodic flush failed", zap.Error(err))
			}
			p.mu.Unlock()
		}
	}
}

func (p *Producer) flushLocked(ctx context.Context) error {
	if len(p.buffer) == 0 {
		return nil
	}

	messages := make([]kafka.Message, 0, len(p.buffer))
	for _, item := range p.buffer {
		messages = append(messages, kafka.Message{
			Topic: item.Topic,
			Key:   []byte(item.Key),
			Value: item.Value,
			Time:  time.Now().UTC(),
		})
	}

	p.buffer = p.buffer[:0]

	if err := p.writer.WriteMessages(ctx, messages...); err != nil {
		p.log.Error("kafka publish failed, sending to dlq", zap.Error(err), zap.Int("count", len(messages)))
		if dlqErr := p.sendToDLQ(ctx, messages, err); dlqErr != nil {
			return fmt.Errorf("kafka publish failed: %w; dlq failed: %v", err, dlqErr)
		}
		return err
	}

	return nil
}

func (p *Producer) sendToDLQ(ctx context.Context, messages []kafka.Message, cause error) error {
	dlqMessages := make([]kafka.Message, 0, len(messages))
	for _, msg := range messages {
		wrapped := fmt.Sprintf(`{"error":%q,"topic":%q,"key":%q,"value":%s}`,
			cause.Error(), msg.Topic, string(msg.Key), string(msg.Value))
		dlqMessages = append(dlqMessages, kafka.Message{
			Topic: p.dlqWriter.Topic,
			Key:   msg.Key,
			Value: []byte(wrapped),
			Time:  time.Now().UTC(),
		})
	}
	return p.dlqWriter.WriteMessages(ctx, dlqMessages...)
}

// PartitionKey builds a Kafka key from service and environment.
func PartitionKey(service, environment string) string {
	return service + "+" + environment
}
