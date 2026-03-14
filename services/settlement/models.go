package settlement

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SettlementStatus string

const (
	SettlementStatusPending    SettlementStatus = "PENDING"
	SettlementStatusProcessing SettlementStatus = "PROCESSING"
	SettlementStatusCompleted  SettlementStatus = "COMPLETED"
	SettlementStatusFailed     SettlementStatus = "FAILED"
	SettlementStatusCancelled  SettlementStatus = "CANCELLED"
)

type SettlementType string

const (
	SettlementTypeACH  SettlementType = "ACH"
	SettlementTypeWire SettlementType = "WIRE"
)

type Settlement struct {
	ID             uuid.UUID        `json:"id"`
	MerchantID     uuid.UUID        `json:"merchant_id"`
	PaymentID      *uuid.UUID       `json:"payment_id,omitempty"`
	Amount         int64            `json:"amount"`
	Currency       string           `json:"currency"`
	Type           SettlementType   `json:"type"`
	Status         SettlementStatus `json:"status"`
	BankAccount    string           `json:"bank_account"`
	RoutingNumber  string           `json:"routing_number"`
	TraceNumber    string           `json:"trace_number,omitempty"`
	FailureCode    string           `json:"failure_code,omitempty"`
	FailureMessage string           `json:"failure_message,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	CompletedAt    *time.Time       `json:"completed_at,omitempty"`
}

type Repository interface {
	CreateSettlement(ctx context.Context, settlement *Settlement) error
	GetSettlement(ctx context.Context, id uuid.UUID) (*Settlement, error)
	GetSettlementsByMerchant(ctx context.Context, merchantID uuid.UUID) ([]Settlement, error)
	GetSettlementsByPayment(ctx context.Context, paymentID uuid.UUID) ([]Settlement, error)
	UpdateSettlementStatus(ctx context.Context, id uuid.UUID, status SettlementStatus, traceNumber, failureCode, failureMessage string) error
}
