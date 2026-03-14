package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/adwaiy/zap/services/payment"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service *payment.Service
}

func NewHandler(service *payment.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/payments", h.CreatePayment)
	r.Get("/payments", h.ListPayments)
	r.Get("/payments/{id}", h.GetPayment)
	r.Post("/payments/{id}/authorize", h.Authorize)
	r.Post("/payments/{id}/capture", h.Capture)
	r.Post("/payments/{id}/cancel", h.Cancel)
	r.Post("/payments/{id}/refund", h.Refund)
	r.Get("/payments/{id}/charges", h.GetCharges)
}

type CreatePaymentRequest struct {
	MerchantID    string `json:"merchant_id"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	CaptureMethod string `json:"capture_method"`
	Metadata      string `json:"metadata,omitempty"`
}

func (h *Handler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	merchantID, err := uuid.Parse(req.MerchantID)
	if err != nil {
		http.Error(w, "invalid merchant_id", http.StatusBadRequest)
		return
	}

	captureMethod := payment.CaptureMethodManual
	if req.CaptureMethod == "AUTOMATIC" {
		captureMethod = payment.CaptureMethodAutomatic
	}

	var metadata []byte
	if req.Metadata != "" {
		metadata = []byte(req.Metadata)
	}

	input := &payment.CreatePaymentInput{
		MerchantID:    merchantID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		CaptureMethod: captureMethod,
		Metadata:      metadata,
	}

	p, err := h.service.CreatePayment(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (h *Handler) GetPayment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid payment id", http.StatusBadRequest)
		return
	}

	p, err := h.service.GetPayment(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(p)
}

func (h *Handler) ListPayments(w http.ResponseWriter, r *http.Request) {
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

	payments, err := h.service.GetPaymentsByMerchant(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(payments)
}

func (h *Handler) Authorize(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid payment id", http.StatusBadRequest)
		return
	}

	p, charge, err := h.service.Authorize(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"payment": p,
		"charge":  charge,
	})
}

func (h *Handler) Capture(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid payment id", http.StatusBadRequest)
		return
	}

	p, charge, err := h.service.Capture(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"payment": p,
		"charge":  charge,
	})
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid payment id", http.StatusBadRequest)
		return
	}

	p, err := h.service.Cancel(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(p)
}

func (h *Handler) Refund(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid payment id", http.StatusBadRequest)
		return
	}

	amountStr := r.URL.Query().Get("amount")
	var amount int64
	if amountStr != "" {
		amount, err = strconv.ParseInt(amountStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid amount", http.StatusBadRequest)
			return
		}
	}

	p, charge, err := h.service.Refund(r.Context(), id, amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"payment": p,
		"charge":  charge,
	})
}

func (h *Handler) GetCharges(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid payment id", http.StatusBadRequest)
		return
	}

	charges, err := h.service.GetCharges(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(charges)
}
