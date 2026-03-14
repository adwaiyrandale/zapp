package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/segmentio/kafka-go"
)

type Message struct {
	Key   string
	Topic string
	Value interface{}
}

type Publisher interface {
	Publish(ctx context.Context, msg Message) error
	PublishBatch(ctx context.Context, msgs []Message) error
	Close() error
}

type kafkaPublisher struct {
	writer *kafka.Writer
	mu     sync.RWMutex
}

func NewPublisher(brokers []string) Publisher {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Balancer: &kafka.LeastBytes{},
		Async:    false,
	}

	return &kafkaPublisher{writer: writer}
}

func (p *kafkaPublisher) Publish(ctx context.Context, msg Message) error {
	value, err := json.Marshal(msg.Value)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(msg.Key),
		Topic: msg.Topic,
		Value: value,
	})
}

func (p *kafkaPublisher) PublishBatch(ctx context.Context, msgs []Message) error {
	kafkaMsgs := make([]kafka.Message, len(msgs))
	for i, msg := range msgs {
		value, err := json.Marshal(msg.Value)
		if err != nil {
			return fmt.Errorf("failed to marshal message %d: %w", i, err)
		}
		kafkaMsgs[i] = kafka.Message{
			Key:   []byte(msg.Key),
			Topic: msg.Topic,
			Value: value,
		}
	}

	return p.writer.WriteMessages(ctx, kafkaMsgs...)
}

func (p *kafkaPublisher) Close() error {
	return p.writer.Close()
}

type Consumer interface {
	Subscribe(ctx context.Context, topic string, handler func(msg Message) error) error
	Close() error
}

type kafkaConsumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, groupID, topic string) Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	return &kafkaConsumer{reader: reader}
}

func (c *kafkaConsumer) Subscribe(ctx context.Context, topic string, handler func(msg Message) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				continue
			}

			var value interface{}
			if err := json.Unmarshal(msg.Value, &value); err != nil {
				continue
			}

			message := Message{
				Key:   string(msg.Key),
				Topic: msg.Topic,
				Value: value,
			}

			if err := handler(message); err != nil {
				continue
			}
		}
	}
}

func (c *kafkaConsumer) Close() error {
	return c.reader.Close()
}

type Event struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp int64       `json:"timestamp"`
}

func NewEvent(eventType string, payload interface{}) Event {
	return Event{
		Type:      eventType,
		Payload:   payload,
		Timestamp: nowMillis(),
	}
}

func nowMillis() int64 {
	return currentTimeFunc()
}

var currentTimeFunc = func() int64 {
	return 0
}

func SetTimeFunc(fn func() int64) {
	currentTimeFunc = fn
}

const (
	TopicPaymentCreated      = "payment.created"
	TopicPaymentAuthorized   = "payment.authorized"
	TopicPaymentCaptured     = "payment.captured"
	TopicPaymentFailed       = "payment.failed"
	TopicSettlementCreated   = "settlement.created"
	TopicSettlementCompleted = "settlement.completed"
	TopicSettlementFailed    = "settlement.failed"
)
