package saga

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/adwaiy/zap/pkg/config"
	"github.com/google/uuid"
)

type Coordinator struct {
	service *Service
	cfg     *config.Config
	http    *http.Client
}

func NewCoordinator(service *Service, cfg *config.Config) *Coordinator {
	return &Coordinator{
		service: service,
		cfg:     cfg,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type PaymentFlowInput struct {
	MerchantID    uuid.UUID `json:"merchant_id"`
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
	CaptureMethod string    `json:"capture_method"`
	BankAccount   string    `json:"bank_account,omitempty"`
	RoutingNumber string    `json:"routing_number,omitempty"`
}

func (c *Coordinator) RegisterPaymentFlows() {
	c.service.RegisterSaga(SagaDefinition{
		Name: "payment_capture",
		Steps: []StepDefinition{
			{
				Name: "create_payment",
				Action: func(ctx context.Context, input []byte) ([]byte, error) {
					var req PaymentFlowInput
					if err := json.Unmarshal(input, &req); err != nil {
						return nil, err
					}

					payload := map[string]interface{}{
						"merchant_id":    req.MerchantID.String(),
						"amount":         req.Amount,
						"currency":       req.Currency,
						"capture_method": req.CaptureMethod,
					}

					resp, err := c.http.Post(c.cfg.PaymentURL()+"/api/v1/payments", "application/json", bytes.NewBuffer(mustMarshal(payload)))
					if err != nil {
						return nil, err
					}
					defer resp.Body.Close()

					var result map[string]interface{}
					if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
						return nil, err
					}

					if resp.StatusCode >= 400 {
						return nil, fmt.Errorf("payment creation failed: %v", result)
					}

					return mustMarshal(result), nil
				},
				Compensate: func(ctx context.Context, output []byte) error {
					var result map[string]interface{}
					json.Unmarshal(output, &result)
					if id, ok := result["id"].(string); ok {
						paymentID, _ := uuid.Parse(id)
						req, _ := http.NewRequest("POST", c.cfg.PaymentURL()+"/api/v1/payments/"+paymentID.String()+"/cancel", nil)
						c.http.Do(req)
					}
					return nil
				},
			},
			{
				Name: "authorize_payment",
				Action: func(ctx context.Context, input []byte) ([]byte, error) {
					var result map[string]interface{}
					json.Unmarshal(input, &result)
					paymentID, _ := uuid.Parse(result["id"].(string))

					resp, err := http.Post(c.cfg.PaymentURL()+"/api/v1/payments/"+paymentID.String()+"/authorize", "application/json", nil)
					if err != nil {
						return nil, err
					}
					defer resp.Body.Close()

					var authResult map[string]interface{}
					json.NewDecoder(resp.Body).Decode(&authResult)

					if resp.StatusCode >= 400 {
						return nil, fmt.Errorf("authorization failed: %v", authResult)
					}

					if payment, ok := authResult["payment"].(map[string]interface{}); ok {
						return mustMarshal(payment), nil
					}
					return mustMarshal(authResult), nil
				},
				Compensate: func(ctx context.Context, output []byte) error {
					var result map[string]interface{}
					json.Unmarshal(output, &result)
					if id, ok := result["id"].(string); ok {
						paymentID, _ := uuid.Parse(id)
						req, _ := http.NewRequest("POST", c.cfg.PaymentURL()+"/api/v1/payments/"+paymentID.String()+"/cancel", nil)
						c.http.Do(req)
					}
					return nil
				},
			},
			{
				Name: "capture_payment",
				Action: func(ctx context.Context, input []byte) ([]byte, error) {
					var result map[string]interface{}
					json.Unmarshal(input, &result)
					paymentID, _ := uuid.Parse(result["id"].(string))

					resp, err := http.Post(c.cfg.PaymentURL()+"/api/v1/payments/"+paymentID.String()+"/capture", "application/json", nil)
					if err != nil {
						return nil, err
					}
					defer resp.Body.Close()

					var captureResult map[string]interface{}
					json.NewDecoder(resp.Body).Decode(&captureResult)

					if resp.StatusCode >= 400 {
						return nil, fmt.Errorf("capture failed: %v", captureResult)
					}

					if payment, ok := captureResult["payment"].(map[string]interface{}); ok {
						return mustMarshal(payment), nil
					}
					return mustMarshal(captureResult), nil
				},
				Compensate: func(ctx context.Context, output []byte) error {
					var result map[string]interface{}
					json.Unmarshal(output, &result)
					if id, ok := result["id"].(string); ok {
						paymentID, _ := uuid.Parse(id)
						req, _ := http.NewRequest("POST", c.cfg.PaymentURL()+"/api/v1/payments/"+paymentID.String()+"/cancel", nil)
						c.http.Do(req)
					}
					return nil
				},
			},
			{
				Name: "create_ledger_entry",
				Action: func(ctx context.Context, input []byte) ([]byte, error) {
					return input, nil
				},
				Compensate: func(ctx context.Context, output []byte) error {
					return nil
				},
			},
		},
		Compensations: []StepDefinition{
			{}, // create_payment - compensated by cancel
			{}, // authorize_payment - already cancelled
			{}, // capture_payment - already cancelled
			{}, // create_ledger_entry
		},
	})

	c.service.RegisterSaga(SagaDefinition{
		Name: "payout",
		Steps: []StepDefinition{
			{
				Name: "create_settlement",
				Action: func(ctx context.Context, input []byte) ([]byte, error) {
					var req PaymentFlowInput
					json.Unmarshal(input, &req)

					payload := map[string]interface{}{
						"merchant_id":    req.MerchantID.String(),
						"amount":         req.Amount,
						"currency":       req.Currency,
						"type":           "ACH",
						"bank_account":   req.BankAccount,
						"routing_number": req.RoutingNumber,
					}

					resp, err := http.Post(c.cfg.SettlementURL()+"/api/v1/settlements", "application/json", bytes.NewBuffer(mustMarshal(payload)))
					if err != nil {
						return nil, err
					}
					defer resp.Body.Close()

					var result map[string]interface{}
					json.NewDecoder(resp.Body).Decode(&result)

					if resp.StatusCode >= 400 {
						return nil, fmt.Errorf("settlement creation failed: %v", result)
					}

					return mustMarshal(result), nil
				},
				Compensate: func(ctx context.Context, output []byte) error {
					var result map[string]interface{}
					json.Unmarshal(output, &result)
					if id, ok := result["id"].(string); ok {
						settlementID, _ := uuid.Parse(id)
						req, _ := http.NewRequest("POST", c.cfg.SettlementURL()+"/api/v1/settlements/"+settlementID.String()+"/cancel", nil)
						c.http.Do(req)
					}
					return nil
				},
			},
			{
				Name: "process_settlement",
				Action: func(ctx context.Context, input []byte) ([]byte, error) {
					var result map[string]interface{}
					json.Unmarshal(input, &result)
					settlementID, _ := uuid.Parse(result["id"].(string))

					resp, err := http.Post(c.cfg.SettlementURL()+"/api/v1/settlements/"+settlementID.String()+"/process", "application/json", nil)
					if err != nil {
						return nil, err
					}
					defer resp.Body.Close()

					var processResult map[string]interface{}
					json.NewDecoder(resp.Body).Decode(&processResult)

					if resp.StatusCode >= 400 {
						return nil, fmt.Errorf("settlement processing failed: %v", processResult)
					}

					return mustMarshal(processResult), nil
				},
				Compensate: func(ctx context.Context, output []byte) error {
					return nil
				},
			},
			{
				Name: "complete_settlement",
				Action: func(ctx context.Context, input []byte) ([]byte, error) {
					var result map[string]interface{}
					json.Unmarshal(input, &result)
					settlementID, _ := uuid.Parse(result["id"].(string))

					resp, err := http.Post(c.cfg.SettlementURL()+"/api/v1/settlements/"+settlementID.String()+"/complete", "application/json", nil)
					if err != nil {
						return nil, err
					}
					defer resp.Body.Close()

					var completeResult map[string]interface{}
					json.NewDecoder(resp.Body).Decode(&completeResult)

					if resp.StatusCode >= 400 {
						return nil, fmt.Errorf("settlement completion failed: %v", completeResult)
					}

					return mustMarshal(completeResult), nil
				},
				Compensate: func(ctx context.Context, output []byte) error {
					return nil
				},
			},
		},
		Compensations: []StepDefinition{
			{}, // create_settlement - compensated by cancel
			{}, // process_settlement
			{}, // complete_settlement
		},
	})
}

func mustMarshal(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}
