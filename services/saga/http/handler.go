package http

import (
	"encoding/json"
	"net/http"

	"github.com/adwaiy/zap/services/saga"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service *saga.Service
}

func NewHandler(service *saga.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/sagas", h.CreateSaga)
	r.Get("/sagas", h.ListSagas)
	r.Get("/sagas/{id}", h.GetSaga)
	r.Get("/sagas/{id}/steps", h.GetSagaSteps)
	r.Post("/sagas/{id}/execute", h.ExecuteStep)
	r.Post("/sagas/{id}/compensate", h.Compensate)
}

type CreateSagaRequest struct {
	Kind  string      `json:"kind"`
	Input interface{} `json:"input"`
}

func (h *Handler) CreateSaga(w http.ResponseWriter, r *http.Request) {
	var req CreateSagaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input := &saga.StartSagaInput{
		Kind:  req.Kind,
		Input: req.Input,
	}

	s, err := h.service.StartSaga(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s)
}

func (h *Handler) GetSaga(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid saga id", http.StatusBadRequest)
		return
	}

	s, err := h.service.GetSaga(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(s)
}

func (h *Handler) ListSagas(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	var sagas []saga.Saga
	var err error

	switch status {
	case "RUNNING":
		sagas, err = h.service.GetRunningSagas(r.Context())
	case "COMPENSATING":
		sagas, err = h.service.GetCompensatingSagas(r.Context())
	default:
		http.Error(w, "invalid status (use RUNNING or COMPENSATING)", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(sagas)
}

func (h *Handler) GetSagaSteps(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid saga id", http.StatusBadRequest)
		return
	}

	steps, err := h.service.GetSagaSteps(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(steps)
}

func (h *Handler) ExecuteStep(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid saga id", http.StatusBadRequest)
		return
	}

	s, step, err := h.service.ExecuteNextStep(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"saga": s,
		"step": step,
	})
}

func (h *Handler) Compensate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid saga id", http.StatusBadRequest)
		return
	}

	s, err := h.service.Compensate(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(s)
}
