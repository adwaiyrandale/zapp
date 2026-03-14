package http

import (
	"encoding/json"
	"net/http"

	"github.com/adwaiy/zap/services/settlement"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service *settlement.Service
}

func NewHandler(service *settlement.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/settlements", h.CreateSettlement)
	r.Get("/settlements", h.ListSettlements)
	r.Get("/settlements/{id}", h.GetSettlement)
	r.Post("/settlements/{id}/process", h.Process)
	r.Post("/settlements/{id}/complete", h.Complete)
	r.Post("/settlements/{id}/fail", h.Fail)
	r.Post("/settlements/{id}/cancel", h.Cancel)
	r.Get("/payments/{payment_id}/settlements", h.GetSettlementsByPayment)
}

type CreateSettlementRequest struct {
	MerchantID    string `json:"merchant_id"`
	PaymentID     string `json:"payment_id,omitempty"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	Type          string `json:"type"`
	BankAccount   string `json:"bank_account"`
	RoutingNumber string `json:"routing_number"`
}

func (h *Handler) CreateSettlement(w http.ResponseWriter, r *http.Request) {
	var req CreateSettlementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	merchantID, err := uuid.Parse(req.MerchantID)
	if err != nil {
		http.Error(w, "invalid merchant_id", http.StatusBadRequest)
		return
	}

	var paymentID *uuid.UUID
	if req.PaymentID != "" {
		id, err := uuid.Parse(req.PaymentID)
		if err != nil {
			http.Error(w, "invalid payment_id", http.StatusBadRequest)
			return
		}
		paymentID = &id
	}

	settlementType := settlement.SettlementTypeACH
	if req.Type == "WIRE" {
		settlementType = settlement.SettlementTypeWire
	}

	input := &settlement.CreateSettlementInput{
		MerchantID:    merchantID,
		PaymentID:     paymentID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Type:          settlementType,
		BankAccount:   req.BankAccount,
		RoutingNumber: req.RoutingNumber,
	}

	s, err := h.service.CreateSettlement(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s)
}

func (h *Handler) GetSettlement(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid settlement id", http.StatusBadRequest)
		return
	}

	s, err := h.service.GetSettlement(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(s)
}

func (h *Handler) ListSettlements(w http.ResponseWriter, r *http.Request) {
	merchantID := r.URL.Query().Get("merchant_id")
	if merchantID == "" {
		http.Error(w, "merchant_id required", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(merchantID)
	if err != nil {
		http.Error(w, "invalid merchant_id", http.StatusBadRequest)
		return
	}

	settlements, err := h.service.GetSettlementsByMerchant(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(settlements)
}

func (h *Handler) Process(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid settlement id", http.StatusBadRequest)
		return
	}

	s, err := h.service.Process(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(s)
}

func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid settlement id", http.StatusBadRequest)
		return
	}

	s, err := h.service.Complete(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(s)
}

type FailRequest struct {
	FailureCode    string `json:"failure_code"`
	FailureMessage string `json:"failure_message"`
}

func (h *Handler) Fail(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid settlement id", http.StatusBadRequest)
		return
	}

	var req FailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s, err := h.service.Fail(r.Context(), id, req.FailureCode, req.FailureMessage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(s)
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid settlement id", http.StatusBadRequest)
		return
	}

	s, err := h.service.Cancel(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(s)
}

func (h *Handler) GetSettlementsByPayment(w http.ResponseWriter, r *http.Request) {
	paymentIDStr := chi.URLParam(r, "payment_id")
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		http.Error(w, "invalid payment id", http.StatusBadRequest)
		return
	}

	settlements, err := h.service.GetSettlementsByPayment(r.Context(), paymentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(settlements)
}
