package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

var (
	ErrNotFound = errors.New("outbox message not found")
)

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusPublished Status = "PUBLISHED"
	StatusFailed    Status = "FAILED"
)

type Message struct {
	ID          uuid.UUID  `json:"id"`
	Topic       string     `json:"topic"`
	Key         string     `json:"key"`
	Value       []byte     `json:"value"`
	Status      Status     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	RetryCount  int        `json:"retry_count"`
}

type Repository interface {
	Create(ctx context.Context, msg *Message) error
	GetPending(ctx context.Context, limit int) ([]Message, error)
	MarkPublished(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID) error
	DeleteOld(ctx context.Context, olderThan time.Duration) error
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, msg *Message) error {
	query := `
		INSERT INTO outbox (id, topic, key, value, status, created_at, retry_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		msg.ID,
		msg.Topic,
		msg.Key,
		msg.Value,
		msg.Status,
		msg.CreatedAt,
		msg.RetryCount,
	)
	return err
}

func (r *PostgresRepository) GetPending(ctx context.Context, limit int) ([]Message, error) {
	query := `
		SELECT id, topic, key, value, status, created_at, published_at, retry_count
		FROM outbox
		WHERE status = $1 AND retry_count < 5
		ORDER BY created_at ASC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, StatusPending, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.Topic, &msg.Key, &msg.Value, &msg.Status, &msg.CreatedAt, &msg.PublishedAt, &msg.RetryCount)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

func (r *PostgresRepository) MarkPublished(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE outbox SET status = $1, published_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, StatusPublished, time.Now().UTC(), id)
	return err
}

func (r *PostgresRepository) MarkFailed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE outbox SET status = $1, retry_count = retry_count + 1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, StatusFailed, id)
	return err
}

func (r *PostgresRepository) DeleteOld(ctx context.Context, olderThan time.Duration) error {
	query := `DELETE FROM outbox WHERE status = $1 AND created_at < $2`
	_, err := r.db.ExecContext(ctx, query, StatusPublished, time.Now().UTC().Add(-olderThan))
	return err
}

type Publisher interface {
	Publish(ctx context.Context, topic, key string, value interface{}) error
}

type KafkaPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPublisher(brokers []string) *KafkaPublisher {
	return &KafkaPublisher{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *KafkaPublisher) Publish(ctx context.Context, topic, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
	})
}

func (p *KafkaPublisher) Close() error {
	return p.writer.Close()
}

type Processor struct {
	repo      Repository
	publisher Publisher
	interval  time.Duration
}

func NewProcessor(repo Repository, publisher Publisher, interval time.Duration) *Processor {
	return &Processor{
		repo:      repo,
		publisher: publisher,
		interval:  interval,
	}
}

func (p *Processor) Start(ctx context.Context) error {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := p.process(ctx); err != nil {
				continue
			}
		}
	}
}

func (p *Processor) process(ctx context.Context) error {
	msgs, err := p.repo.GetPending(ctx, 100)
	if err != nil {
		return err
	}

	for _, msg := range msgs {
		if err := p.publisher.Publish(ctx, msg.Topic, msg.Key, msg.Value); err != nil {
			p.repo.MarkFailed(ctx, msg.ID)
			continue
		}
		p.repo.MarkPublished(ctx, msg.ID)
	}

	return nil
}

func CreateAndPublish(ctx context.Context, repo Repository, publisher Publisher, topic, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	msg := &Message{
		ID:        uuid.New(),
		Topic:     topic,
		Key:       key,
		Value:     data,
		Status:    StatusPending,
		CreatedAt: time.Now().UTC(),
	}

	return repo.Create(ctx, msg)
}
