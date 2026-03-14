package settlement

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrSettlementNotFound = errors.New("settlement not found")
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) CreateSettlement(ctx context.Context, settlement *Settlement) error {
	query := `
		INSERT INTO settlements (id, merchant_id, payment_id, amount, currency, type, status, bank_account, routing_number, trace_number, failure_code, failure_message, created_at, updated_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	_, err := r.pool.Exec(ctx, query,
		settlement.ID,
		settlement.MerchantID,
		settlement.PaymentID,
		settlement.Amount,
		settlement.Currency,
		settlement.Type,
		settlement.Status,
		settlement.BankAccount,
		settlement.RoutingNumber,
		settlement.TraceNumber,
		settlement.FailureCode,
		settlement.FailureMessage,
		settlement.CreatedAt,
		settlement.UpdatedAt,
		settlement.CompletedAt,
	)
	return err
}

func (r *PostgresRepository) GetSettlement(ctx context.Context, id uuid.UUID) (*Settlement, error) {
	query := `SELECT id, merchant_id, payment_id, amount, currency, type, status, bank_account, routing_number, trace_number, failure_code, failure_message, created_at, updated_at, completed_at FROM settlements WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)

	var s Settlement
	err := row.Scan(
		&s.ID,
		&s.MerchantID,
		&s.PaymentID,
		&s.Amount,
		&s.Currency,
		&s.Type,
		&s.Status,
		&s.BankAccount,
		&s.RoutingNumber,
		&s.TraceNumber,
		&s.FailureCode,
		&s.FailureMessage,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.CompletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSettlementNotFound
	}
	return &s, err
}

func (r *PostgresRepository) GetSettlementsByMerchant(ctx context.Context, merchantID uuid.UUID) ([]Settlement, error) {
	query := `SELECT id, merchant_id, payment_id, amount, currency, type, status, bank_account, routing_number, trace_number, failure_code, failure_message, created_at, updated_at, completed_at FROM settlements WHERE merchant_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, merchantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settlements []Settlement
	for rows.Next() {
		var s Settlement
		err := rows.Scan(
			&s.ID,
			&s.MerchantID,
			&s.PaymentID,
			&s.Amount,
			&s.Currency,
			&s.Type,
			&s.Status,
			&s.BankAccount,
			&s.RoutingNumber,
			&s.TraceNumber,
			&s.FailureCode,
			&s.FailureMessage,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		settlements = append(settlements, s)
	}
	return settlements, nil
}

func (r *PostgresRepository) GetSettlementsByPayment(ctx context.Context, paymentID uuid.UUID) ([]Settlement, error) {
	query := `SELECT id, merchant_id, payment_id, amount, currency, type, status, bank_account, routing_number, trace_number, failure_code, failure_message, created_at, updated_at, completed_at FROM settlements WHERE payment_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, paymentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settlements []Settlement
	for rows.Next() {
		var s Settlement
		err := rows.Scan(
			&s.ID,
			&s.MerchantID,
			&s.PaymentID,
			&s.Amount,
			&s.Currency,
			&s.Type,
			&s.Status,
			&s.BankAccount,
			&s.RoutingNumber,
			&s.TraceNumber,
			&s.FailureCode,
			&s.FailureMessage,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		settlements = append(settlements, s)
	}
	return settlements, nil
}

func (r *PostgresRepository) UpdateSettlementStatus(ctx context.Context, id uuid.UUID, status SettlementStatus, traceNumber, failureCode, failureMessage string) error {
	now := time.Now().UTC()
	var completedAt interface{}
	if status == SettlementStatusCompleted || status == SettlementStatusFailed {
		completedAt = now
	}

	query := `UPDATE settlements SET status = $1, trace_number = $2, failure_code = $3, failure_message = $4, completed_at = $5, updated_at = $6 WHERE id = $7`
	result, err := r.pool.Exec(ctx, query, status, traceNumber, failureCode, failureMessage, completedAt, now, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrSettlementNotFound
	}
	return nil
}
