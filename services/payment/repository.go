package payment

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
	ErrChargeNotFound  = errors.New("charge not found")
	ErrInvalidState    = errors.New("invalid state transition")
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) CreatePayment(ctx context.Context, payment *Payment) error {
	query := `
		INSERT INTO payments (id, merchant_id, amount, currency, status, capture_method, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.pool.Exec(ctx, query,
		payment.ID,
		payment.MerchantID,
		payment.Amount,
		payment.Currency,
		payment.Status,
		payment.CaptureMethod,
		payment.Metadata,
		payment.CreatedAt,
		payment.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) GetPayment(ctx context.Context, id uuid.UUID) (*Payment, error) {
	query := `SELECT id, merchant_id, amount, currency, status, capture_method, metadata, created_at, updated_at FROM payments WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)

	var payment Payment
	err := row.Scan(
		&payment.ID,
		&payment.MerchantID,
		&payment.Amount,
		&payment.Currency,
		&payment.Status,
		&payment.CaptureMethod,
		&payment.Metadata,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPaymentNotFound
	}
	return &payment, err
}

func (r *PostgresRepository) GetPaymentsByMerchant(ctx context.Context, merchantID uuid.UUID) ([]Payment, error) {
	query := `SELECT id, merchant_id, amount, currency, status, capture_method, metadata, created_at, updated_at FROM payments WHERE merchant_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, merchantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []Payment
	for rows.Next() {
		var p Payment
		err := rows.Scan(&p.ID, &p.MerchantID, &p.Amount, &p.Currency, &p.Status, &p.CaptureMethod, &p.Metadata, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, nil
}

func (r *PostgresRepository) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status PaymentStatus) error {
	query := `UPDATE payments SET status = $1, updated_at = $2 WHERE id = $3`
	result, err := r.pool.Exec(ctx, query, status, time.Now().UTC(), id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrPaymentNotFound
	}
	return nil
}

func (r *PostgresRepository) CreateCharge(ctx context.Context, charge *Charge) error {
	query := `
		INSERT INTO charges (id, payment_id, kind, amount, currency, status, processor_ref, failure_code, failure_message, created_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.pool.Exec(ctx, query,
		charge.ID,
		charge.PaymentID,
		charge.Kind,
		charge.Amount,
		charge.Currency,
		charge.Status,
		charge.ProcessorRef,
		charge.FailureCode,
		charge.FailureMessage,
		charge.CreatedAt,
		charge.CompletedAt,
	)
	return err
}

func (r *PostgresRepository) GetCharge(ctx context.Context, id uuid.UUID) (*Charge, error) {
	query := `SELECT id, payment_id, kind, amount, currency, status, processor_ref, failure_code, failure_message, created_at, completed_at FROM charges WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)

	var charge Charge
	err := row.Scan(
		&charge.ID,
		&charge.PaymentID,
		&charge.Kind,
		&charge.Amount,
		&charge.Currency,
		&charge.Status,
		&charge.ProcessorRef,
		&charge.FailureCode,
		&charge.FailureMessage,
		&charge.CreatedAt,
		&charge.CompletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrChargeNotFound
	}
	return &charge, err
}

func (r *PostgresRepository) GetChargesByPayment(ctx context.Context, paymentID uuid.UUID) ([]Charge, error) {
	query := `SELECT id, payment_id, kind, amount, currency, status, processor_ref, failure_code, failure_message, created_at, completed_at FROM charges WHERE payment_id = $1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, paymentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var charges []Charge
	for rows.Next() {
		var c Charge
		err := rows.Scan(&c.ID, &c.PaymentID, &c.Kind, &c.Amount, &c.Currency, &c.Status, &c.ProcessorRef, &c.FailureCode, &c.FailureMessage, &c.CreatedAt, &c.CompletedAt)
		if err != nil {
			return nil, err
		}
		charges = append(charges, c)
	}
	return charges, nil
}

func (r *PostgresRepository) UpdateChargeStatus(ctx context.Context, id uuid.UUID, status ChargeStatus, processorRef, failureCode, failureMessage string) error {
	completedAt := time.Now().UTC()
	query := `UPDATE charges SET status = $1, processor_ref = $2, failure_code = $3, failure_message = $4, completed_at = $5 WHERE id = $6`
	result, err := r.pool.Exec(ctx, query, status, processorRef, failureCode, failureMessage, completedAt, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrChargeNotFound
	}
	return nil
}
