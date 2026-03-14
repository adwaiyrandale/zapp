package payment

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type mockPaymentRepository struct {
	payments map[uuid.UUID]*Payment
	charges  map[uuid.UUID]*Charge
}

func newMockPaymentRepository() *mockPaymentRepository {
	return &mockPaymentRepository{
		payments: make(map[uuid.UUID]*Payment),
		charges:  make(map[uuid.UUID]*Charge),
	}
}

func (m *mockPaymentRepository) CreatePayment(ctx context.Context, p *Payment) error {
	m.payments[p.ID] = p
	return nil
}

func (m *mockPaymentRepository) GetPayment(ctx context.Context, id uuid.UUID) (*Payment, error) {
	if p, ok := m.payments[id]; ok {
		return p, nil
	}
	return nil, ErrPaymentNotFound
}

func (m *mockPaymentRepository) GetPaymentsByMerchant(ctx context.Context, merchantID uuid.UUID) ([]Payment, error) {
	var result []Payment
	for _, p := range m.payments {
		if p.MerchantID == merchantID {
			result = append(result, *p)
		}
	}
	return result, nil
}

func (m *mockPaymentRepository) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status PaymentStatus) error {
	if p, ok := m.payments[id]; ok {
		p.Status = status
		return nil
	}
	return ErrPaymentNotFound
}

func (m *mockPaymentRepository) CreateCharge(ctx context.Context, charge *Charge) error {
	m.charges[charge.ID] = charge
	return nil
}

func (m *mockPaymentRepository) GetCharge(ctx context.Context, id uuid.UUID) (*Charge, error) {
	if c, ok := m.charges[id]; ok {
		return c, nil
	}
	return nil, ErrChargeNotFound
}

func (m *mockPaymentRepository) GetChargesByPayment(ctx context.Context, paymentID uuid.UUID) ([]Charge, error) {
	var result []Charge
	for _, c := range m.charges {
		if c.PaymentID == paymentID {
			result = append(result, *c)
		}
	}
	return result, nil
}

func (m *mockPaymentRepository) UpdateChargeStatus(ctx context.Context, id uuid.UUID, status ChargeStatus, processorRef, failureCode, failureMessage string) error {
	if c, ok := m.charges[id]; ok {
		c.Status = status
		c.ProcessorRef = processorRef
		return nil
	}
	return ErrChargeNotFound
}

func TestCreatePayment(t *testing.T) {
	repo := newMockPaymentRepository()
	svc := NewService(repo)

	merchantID := uuid.New()
	input := &CreatePaymentInput{
		MerchantID:    merchantID,
		Amount:        10000, // $100.00
		Currency:      "USD",
		CaptureMethod: CaptureMethodAutomatic,
	}

	p, err := svc.CreatePayment(context.Background(), input)
	if err != nil {
		t.Fatalf("CreatePayment failed: %v", err)
	}

	if p.Status != PaymentStatusPending {
		t.Errorf("Expected PENDING status, got %s", p.Status)
	}

	if p.Amount != 10000 {
		t.Errorf("Expected amount 10000, got %d", p.Amount)
	}
}

func TestAuthorizePayment(t *testing.T) {
	repo := newMockPaymentRepository()
	svc := NewService(repo)

	merchantID := uuid.New()
	p, _ := svc.CreatePayment(context.Background(), &CreatePaymentInput{
		MerchantID:    merchantID,
		Amount:        10000,
		Currency:      "USD",
		CaptureMethod: CaptureMethodManual,
	})

	authorizedPayment, charge, err := svc.Authorize(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("Authorize failed: %v", err)
	}

	if authorizedPayment.Status != PaymentStatusAuthorized {
		t.Errorf("Expected AUTHORIZED status, got %s", authorizedPayment.Status)
	}

	if charge.Kind != ChargeKindAuthorization {
		t.Errorf("Expected AUTHORIZATION charge, got %s", charge.Kind)
	}
}

func TestCapturePayment(t *testing.T) {
	repo := newMockPaymentRepository()
	svc := NewService(repo)

	merchantID := uuid.New()
	p, _ := svc.CreatePayment(context.Background(), &CreatePaymentInput{
		MerchantID:    merchantID,
		Amount:        10000,
		Currency:      "USD",
		CaptureMethod: CaptureMethodManual,
	})

	svc.Authorize(context.Background(), p.ID)
	capturedPayment, charge, err := svc.Capture(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("Capture failed: %v", err)
	}

	if capturedPayment.Status != PaymentStatusCaptured {
		t.Errorf("Expected CAPTURED status, got %s", capturedPayment.Status)
	}

	if charge.Kind != ChargeKindCapture {
		t.Errorf("Expected CAPTURE charge, got %s", charge.Kind)
	}
}

func TestInvalidStateTransition(t *testing.T) {
	repo := newMockPaymentRepository()
	svc := NewService(repo)

	merchantID := uuid.New()
	p, _ := svc.CreatePayment(context.Background(), &CreatePaymentInput{
		MerchantID:    merchantID,
		Amount:        10000,
		Currency:      "USD",
		CaptureMethod: CaptureMethodManual,
	})

	// Try to capture without authorizing - should fail
	_, _, err := svc.Capture(context.Background(), p.ID)
	if err == nil {
		t.Error("Expected error when capturing non-authorized payment")
	}
}

func TestRefundPayment(t *testing.T) {
	repo := newMockPaymentRepository()
	svc := NewService(repo)

	merchantID := uuid.New()
	p, _ := svc.CreatePayment(context.Background(), &CreatePaymentInput{
		MerchantID:    merchantID,
		Amount:        10000,
		Currency:      "USD",
		CaptureMethod: CaptureMethodManual,
	})

	svc.Authorize(context.Background(), p.ID)
	svc.Capture(context.Background(), p.ID)

	refundedPayment, charge, err := svc.Refund(context.Background(), p.ID, 5000)
	if err != nil {
		t.Fatalf("Refund failed: %v", err)
	}

	if refundedPayment.Status != PaymentStatusCaptured {
		t.Errorf("Expected partial refund to keep CAPTURED status, got %s", refundedPayment.Status)
	}

	if charge.Kind != ChargeKindRefund {
		t.Errorf("Expected REFUND charge, got %s", charge.Kind)
	}
}
