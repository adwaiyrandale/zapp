package payment

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "PENDING"
	PaymentStatusAuthorized PaymentStatus = "AUTHORIZED"
	PaymentStatusCaptured   PaymentStatus = "CAPTURED"
	PaymentStatusCancelled  PaymentStatus = "CANCELLED"
	PaymentStatusRefunded   PaymentStatus = "REFUNDED"
)

type CaptureMethod string

const (
	CaptureMethodAutomatic CaptureMethod = "AUTOMATIC"
	CaptureMethodManual    CaptureMethod = "MANUAL"
)

type ChargeKind string

const (
	ChargeKindAuthorization ChargeKind = "AUTHORIZATION"
	ChargeKindCapture       ChargeKind = "CAPTURE"
	ChargeKindRefund        ChargeKind = "REFUND"
	ChargeKindVoid          ChargeKind = "VOID"
)

type ChargeStatus string

const (
	ChargeStatusPending   ChargeStatus = "PENDING"
	ChargeStatusSucceeded ChargeStatus = "SUCCEEDED"
	ChargeStatusFailed    ChargeStatus = "FAILED"
)

type Payment struct {
	ID            uuid.UUID     `json:"id"`
	MerchantID    uuid.UUID     `json:"merchant_id"`
	Amount        int64         `json:"amount"` // in cents
	Currency      string        `json:"currency"`
	Status        PaymentStatus `json:"status"`
	CaptureMethod CaptureMethod `json:"capture_method"`
	Metadata      []byte        `json:"metadata,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type Charge struct {
	ID             uuid.UUID    `json:"id"`
	PaymentID      uuid.UUID    `json:"payment_id"`
	Kind           ChargeKind   `json:"kind"`
	Amount         int64        `json:"amount"`
	Currency       string       `json:"currency"`
	Status         ChargeStatus `json:"status"`
	ProcessorRef   string       `json:"processor_ref,omitempty"`
	FailureCode    string       `json:"failure_code,omitempty"`
	FailureMessage string       `json:"failure_message,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	CompletedAt    *time.Time   `json:"completed_at,omitempty"`
}

type Repository interface {
	// Payment operations
	CreatePayment(ctx context.Context, payment *Payment) error
	GetPayment(ctx context.Context, id uuid.UUID) (*Payment, error)
	GetPaymentsByMerchant(ctx context.Context, merchantID uuid.UUID) ([]Payment, error)
	UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status PaymentStatus) error

	// Charge operations
	CreateCharge(ctx context.Context, charge *Charge) error
	GetCharge(ctx context.Context, id uuid.UUID) (*Charge, error)
	GetChargesByPayment(ctx context.Context, paymentID uuid.UUID) ([]Charge, error)
	UpdateChargeStatus(ctx context.Context, id uuid.UUID, status ChargeStatus, processorRef, failureCode, failureMessage string) error
}
