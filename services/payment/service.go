package payment

import (
	"context"
	"errors"
	"time"

	"github.com/adwaiy/zap/pkg/money"
	"github.com/google/uuid"
)

var (
	ErrUnauthorized         = errors.New("payment not authorized")
	ErrAlreadyCaptured      = errors.New("payment already captured")
	ErrAlreadyRefunded      = errors.New("payment already refunded")
	ErrAlreadyCancelled     = errors.New("payment already cancelled")
	ErrCaptureMethod        = errors.New("invalid capture method")
	ErrInvalidAmount        = errors.New("invalid amount")
	ErrRefundExceedsCapture = errors.New("refund amount exceeds captured amount")
)

// State machine transitions
var validTransitions = map[PaymentStatus][]PaymentStatus{
	PaymentStatusPending:    {PaymentStatusAuthorized, PaymentStatusCancelled},
	PaymentStatusAuthorized: {PaymentStatusCaptured, PaymentStatusCancelled},
	PaymentStatusCaptured:   {PaymentStatusRefunded},
	PaymentStatusCancelled:  {},
	PaymentStatusRefunded:   {},
}

func canTransition(from, to PaymentStatus) bool {
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, status := range allowed {
		if status == to {
			return true
		}
	}
	return false
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type CreatePaymentInput struct {
	MerchantID    uuid.UUID
	Amount        int64 // in cents
	Currency      string
	CaptureMethod CaptureMethod
	Metadata      []byte
}

func (s *Service) CreatePayment(ctx context.Context, input *CreatePaymentInput) (*Payment, error) {
	if input.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	if input.CaptureMethod != CaptureMethodAutomatic && input.CaptureMethod != CaptureMethodManual {
		return nil, ErrCaptureMethod
	}

	now := time.Now().UTC()
	payment := &Payment{
		ID:            uuid.New(),
		MerchantID:    input.MerchantID,
		Amount:        input.Amount,
		Currency:      input.Currency,
		Status:        PaymentStatusPending,
		CaptureMethod: input.CaptureMethod,
		Metadata:      input.Metadata,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.CreatePayment(ctx, payment); err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *Service) GetPayment(ctx context.Context, id uuid.UUID) (*Payment, error) {
	return s.repo.GetPayment(ctx, id)
}

func (s *Service) GetPaymentsByMerchant(ctx context.Context, merchantID uuid.UUID) ([]Payment, error) {
	return s.repo.GetPaymentsByMerchant(ctx, merchantID)
}

// Authorize moves payment from PENDING to AUTHORIZED
func (s *Service) Authorize(ctx context.Context, paymentID uuid.UUID) (*Payment, *Charge, error) {
	payment, err := s.repo.GetPayment(ctx, paymentID)
	if err != nil {
		return nil, nil, err
	}

	if !canTransition(payment.Status, PaymentStatusAuthorized) {
		return nil, nil, ErrInvalidState
	}

	now := time.Now().UTC()
	charge := &Charge{
		ID:           uuid.New(),
		PaymentID:    paymentID,
		Kind:         ChargeKindAuthorization,
		Amount:       payment.Amount,
		Currency:     payment.Currency,
		Status:       ChargeStatusSucceeded, // Simplified - would call payment processor
		ProcessorRef: "auth_" + uuid.New().String()[:8],
		CreatedAt:    now,
		CompletedAt:  &now,
	}

	if err := s.repo.CreateCharge(ctx, charge); err != nil {
		return nil, nil, err
	}

	if err := s.repo.UpdatePaymentStatus(ctx, paymentID, PaymentStatusAuthorized); err != nil {
		return nil, nil, err
	}

	payment.Status = PaymentStatusAuthorized
	payment.UpdatedAt = now

	return payment, charge, nil
}

// Capture moves payment from AUTHORIZED to CAPTURED
func (s *Service) Capture(ctx context.Context, paymentID uuid.UUID) (*Payment, *Charge, error) {
	payment, err := s.repo.GetPayment(ctx, paymentID)
	if err != nil {
		return nil, nil, err
	}

	if !canTransition(payment.Status, PaymentStatusCaptured) {
		return nil, nil, ErrInvalidState
	}

	now := time.Now().UTC()
	charge := &Charge{
		ID:           uuid.New(),
		PaymentID:    paymentID,
		Kind:         ChargeKindCapture,
		Amount:       payment.Amount,
		Currency:     payment.Currency,
		Status:       ChargeStatusSucceeded,
		ProcessorRef: "cap_" + uuid.New().String()[:8],
		CreatedAt:    now,
		CompletedAt:  &now,
	}

	if err := s.repo.CreateCharge(ctx, charge); err != nil {
		return nil, nil, err
	}

	if err := s.repo.UpdatePaymentStatus(ctx, paymentID, PaymentStatusCaptured); err != nil {
		return nil, nil, err
	}

	payment.Status = PaymentStatusCaptured
	payment.UpdatedAt = now

	return payment, charge, nil
}

// Cancel moves payment from AUTHORIZED/PENDING to CANCELLED
func (s *Service) Cancel(ctx context.Context, paymentID uuid.UUID) (*Payment, error) {
	payment, err := s.repo.GetPayment(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	if !canTransition(payment.Status, PaymentStatusCancelled) {
		return nil, ErrInvalidState
	}

	now := time.Now().UTC()
	charge := &Charge{
		ID:           uuid.New(),
		PaymentID:    paymentID,
		Kind:         ChargeKindVoid,
		Amount:       payment.Amount,
		Currency:     payment.Currency,
		Status:       ChargeStatusSucceeded,
		ProcessorRef: "void_" + uuid.New().String()[:8],
		CreatedAt:    now,
		CompletedAt:  &now,
	}

	if err := s.repo.CreateCharge(ctx, charge); err != nil {
		return nil, err
	}

	if err := s.repo.UpdatePaymentStatus(ctx, paymentID, PaymentStatusCancelled); err != nil {
		return nil, err
	}

	payment.Status = PaymentStatusCancelled
	payment.UpdatedAt = now

	return payment, nil
}

// Refund moves payment from CAPTURED to REFUNDED
func (s *Service) Refund(ctx context.Context, paymentID uuid.UUID, amount int64) (*Payment, *Charge, error) {
	payment, err := s.repo.GetPayment(ctx, paymentID)
	if err != nil {
		return nil, nil, err
	}

	if payment.Status != PaymentStatusCaptured {
		return nil, nil, ErrInvalidState
	}

	if amount <= 0 || amount > payment.Amount {
		return nil, nil, ErrRefundExceedsCapture
	}

	now := time.Now().UTC()
	charge := &Charge{
		ID:           uuid.New(),
		PaymentID:    paymentID,
		Kind:         ChargeKindRefund,
		Amount:       amount,
		Currency:     payment.Currency,
		Status:       ChargeStatusSucceeded,
		ProcessorRef: "ref_" + uuid.New().String()[:8],
		CreatedAt:    now,
		CompletedAt:  &now,
	}

	if err := s.repo.CreateCharge(ctx, charge); err != nil {
		return nil, nil, err
	}

	// If full refund, update status
	if amount == payment.Amount {
		if err := s.repo.UpdatePaymentStatus(ctx, paymentID, PaymentStatusRefunded); err != nil {
			return nil, nil, err
		}
		payment.Status = PaymentStatusRefunded
	}

	payment.UpdatedAt = now

	return payment, charge, nil
}

func (s *Service) GetCharges(ctx context.Context, paymentID uuid.UUID) ([]Charge, error) {
	return s.repo.GetChargesByPayment(ctx, paymentID)
}

// Convenience function to get money from payment
func (p *Payment) GetMoney() money.Money {
	return money.New(p.Amount, p.Currency)
}
