package settlement

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidAmount         = errors.New("invalid amount")
	ErrInvalidSettlementType = errors.New("invalid settlement type")
	ErrInvalidState          = errors.New("invalid state transition")
	ErrSettlementNotPending  = errors.New("settlement is not in pending state")
)

var validTransitions = map[SettlementStatus][]SettlementStatus{
	SettlementStatusPending:    {SettlementStatusProcessing, SettlementStatusCancelled},
	SettlementStatusProcessing: {SettlementStatusCompleted, SettlementStatusFailed},
	SettlementStatusCompleted:  {},
	SettlementStatusFailed:     {},
	SettlementStatusCancelled:  {},
}

func canTransition(from, to SettlementStatus) bool {
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

type CreateSettlementInput struct {
	MerchantID    uuid.UUID
	PaymentID     *uuid.UUID
	Amount        int64
	Currency      string
	Type          SettlementType
	BankAccount   string
	RoutingNumber string
}

func (s *Service) CreateSettlement(ctx context.Context, input *CreateSettlementInput) (*Settlement, error) {
	if input.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	if input.Type != SettlementTypeACH && input.Type != SettlementTypeWire {
		return nil, ErrInvalidSettlementType
	}

	now := time.Now().UTC()
	settlement := &Settlement{
		ID:            uuid.New(),
		MerchantID:    input.MerchantID,
		PaymentID:     input.PaymentID,
		Amount:        input.Amount,
		Currency:      input.Currency,
		Type:          input.Type,
		Status:        SettlementStatusPending,
		BankAccount:   input.BankAccount,
		RoutingNumber: input.RoutingNumber,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.CreateSettlement(ctx, settlement); err != nil {
		return nil, err
	}

	return settlement, nil
}

func (s *Service) GetSettlement(ctx context.Context, id uuid.UUID) (*Settlement, error) {
	return s.repo.GetSettlement(ctx, id)
}

func (s *Service) GetSettlementsByMerchant(ctx context.Context, merchantID uuid.UUID) ([]Settlement, error) {
	return s.repo.GetSettlementsByMerchant(ctx, merchantID)
}

func (s *Service) GetSettlementsByPayment(ctx context.Context, paymentID uuid.UUID) ([]Settlement, error) {
	return s.repo.GetSettlementsByPayment(ctx, paymentID)
}

func (s *Service) Process(ctx context.Context, settlementID uuid.UUID) (*Settlement, error) {
	settlement, err := s.repo.GetSettlement(ctx, settlementID)
	if err != nil {
		return nil, err
	}

	if !canTransition(settlement.Status, SettlementStatusProcessing) {
		return nil, ErrInvalidState
	}

	traceNumber := "trace_" + uuid.New().String()[:8]

	if err := s.repo.UpdateSettlementStatus(ctx, settlementID, SettlementStatusProcessing, traceNumber, "", ""); err != nil {
		return nil, err
	}

	settlement.Status = SettlementStatusProcessing
	settlement.TraceNumber = traceNumber
	settlement.UpdatedAt = time.Now().UTC()

	return settlement, nil
}

func (s *Service) Complete(ctx context.Context, settlementID uuid.UUID) (*Settlement, error) {
	settlement, err := s.repo.GetSettlement(ctx, settlementID)
	if err != nil {
		return nil, err
	}

	if settlement.Status != SettlementStatusProcessing {
		return nil, ErrSettlementNotPending
	}

	if err := s.repo.UpdateSettlementStatus(ctx, settlementID, SettlementStatusCompleted, settlement.TraceNumber, "", ""); err != nil {
		return nil, err
	}

	settlement.Status = SettlementStatusCompleted
	now := time.Now().UTC()
	settlement.CompletedAt = &now
	settlement.UpdatedAt = now

	return settlement, nil
}

func (s *Service) Fail(ctx context.Context, settlementID uuid.UUID, failureCode, failureMessage string) (*Settlement, error) {
	settlement, err := s.repo.GetSettlement(ctx, settlementID)
	if err != nil {
		return nil, err
	}

	if settlement.Status != SettlementStatusProcessing {
		return nil, ErrSettlementNotPending
	}

	if err := s.repo.UpdateSettlementStatus(ctx, settlementID, SettlementStatusFailed, settlement.TraceNumber, failureCode, failureMessage); err != nil {
		return nil, err
	}

	settlement.Status = SettlementStatusFailed
	settlement.FailureCode = failureCode
	settlement.FailureMessage = failureMessage
	now := time.Now().UTC()
	settlement.CompletedAt = &now
	settlement.UpdatedAt = now

	return settlement, nil
}

func (s *Service) Cancel(ctx context.Context, settlementID uuid.UUID) (*Settlement, error) {
	settlement, err := s.repo.GetSettlement(ctx, settlementID)
	if err != nil {
		return nil, err
	}

	if !canTransition(settlement.Status, SettlementStatusCancelled) {
		return nil, ErrInvalidState
	}

	if err := s.repo.UpdateSettlementStatus(ctx, settlementID, SettlementStatusCancelled, "", "", ""); err != nil {
		return nil, err
	}

	settlement.Status = SettlementStatusCancelled
	settlement.UpdatedAt = time.Now().UTC()

	return settlement, nil
}
