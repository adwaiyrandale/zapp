package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/adwaiy/zap/services/ledger"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service *ledger.Service
}

func NewHandler(service *ledger.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/accounts", h.ListAccounts)
	r.Get("/accounts/{code}", h.GetAccount)
	r.Get("/accounts/{code}/balance", h.GetAccountBalance)
	r.Get("/balances", h.GetAllBalances)

	r.Get("/journals", h.ListJournals)
	r.Get("/journals/{id}", h.GetJournal)
	r.Post("/journals", h.PostJournal)
}

func (h *Handler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.service.ListAccounts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(accounts)
}

func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	account, err := h.service.GetAccount(r.Context(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(account)
}

func (h *Handler) GetAccountBalance(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	balance, err := h.service.GetAccountBalance(r.Context(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(balance)
}

func (h *Handler) GetAllBalances(w http.ResponseWriter, r *http.Request) {
	balances, err := h.service.GetAllBalances(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(balances)
}

func (h *Handler) ListJournals(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	journals, err := h.service.ListJournals(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(journals)
}

func (h *Handler) GetJournal(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid journal id", http.StatusBadRequest)
		return
	}

	journal, err := h.service.GetJournal(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(journal)
}

func (h *Handler) PostJournal(w http.ResponseWriter, r *http.Request) {
	var input ledger.PostJournalInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	journal, err := h.service.PostJournal(r.Context(), &input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(journal)
}
