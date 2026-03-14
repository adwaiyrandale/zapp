package saga

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrSagaNotFound = errors.New("saga not found")
	ErrStepNotFound = errors.New("step not found")
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) CreateSaga(ctx context.Context, saga *Saga) error {
	query := `
		INSERT INTO sagas (id, kind, status, current_step, input, output, compensation_state, created_at, updated_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(ctx, query,
		saga.ID,
		saga.Kind,
		saga.Status,
		saga.CurrentStep,
		saga.Input,
		saga.Output,
		saga.CompensationState,
		saga.CreatedAt,
		saga.UpdatedAt,
		saga.CompletedAt,
	)
	return err
}

func (r *PostgresRepository) GetSaga(ctx context.Context, id uuid.UUID) (*Saga, error) {
	query := `SELECT id, kind, status, current_step, input, output, compensation_state, created_at, updated_at, completed_at FROM sagas WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)

	var s Saga
	err := row.Scan(
		&s.ID,
		&s.Kind,
		&s.Status,
		&s.CurrentStep,
		&s.Input,
		&s.Output,
		&s.CompensationState,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.CompletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSagaNotFound
	}
	return &s, err
}

func (r *PostgresRepository) UpdateSaga(ctx context.Context, saga *Saga) error {
	query := `UPDATE sagas SET status = $1, current_step = $2, output = $3, compensation_state = $4, updated_at = $5, completed_at = $6 WHERE id = $7`
	_, err := r.pool.Exec(ctx, query,
		saga.Status,
		saga.CurrentStep,
		saga.Output,
		saga.CompensationState,
		time.Now().UTC(),
		saga.CompletedAt,
		saga.ID,
	)
	return err
}

func (r *PostgresRepository) GetSagasByStatus(ctx context.Context, status SagaStatus) ([]Saga, error) {
	query := `SELECT id, kind, status, current_step, input, output, compensation_state, created_at, updated_at, completed_at FROM sagas WHERE status = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sagas []Saga
	for rows.Next() {
		var s Saga
		err := rows.Scan(
			&s.ID,
			&s.Kind,
			&s.Status,
			&s.CurrentStep,
			&s.Input,
			&s.Output,
			&s.CompensationState,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		sagas = append(sagas, s)
	}
	return sagas, nil
}

func (r *PostgresRepository) CreateSagaStep(ctx context.Context, step *SagaStep) error {
	query := `
		INSERT INTO saga_steps (id, saga_id, name, seq, status, input, output, error, started_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(ctx, query,
		step.ID,
		step.SagaID,
		step.Name,
		step.Seq,
		step.Status,
		step.Input,
		step.Output,
		step.Error,
		step.StartedAt,
		step.CompletedAt,
	)
	return err
}

func (r *PostgresRepository) GetSagaSteps(ctx context.Context, sagaID uuid.UUID) ([]SagaStep, error) {
	query := `SELECT id, saga_id, name, seq, status, input, output, error, started_at, completed_at FROM saga_steps WHERE saga_id = $1 ORDER BY seq ASC`
	rows, err := r.pool.Query(ctx, query, sagaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []SagaStep
	for rows.Next() {
		var step SagaStep
		err := rows.Scan(
			&step.ID,
			&step.SagaID,
			&step.Name,
			&step.Seq,
			&step.Status,
			&step.Input,
			&step.Output,
			&step.Error,
			&step.StartedAt,
			&step.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	return steps, nil
}

func (r *PostgresRepository) UpdateSagaStep(ctx context.Context, step *SagaStep) error {
	query := `UPDATE saga_steps SET status = $1, output = $2, error = $3, started_at = $4, completed_at = $5 WHERE id = $6`
	_, err := r.pool.Exec(ctx, query,
		step.Status,
		step.Output,
		step.Error,
		step.StartedAt,
		step.CompletedAt,
		step.ID,
	)
	return err
}
