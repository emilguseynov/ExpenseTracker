package expense

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

type expenseResponse struct {
	ID          int64  `json:"id"`
	Amount      string `json:"amount"`
	Category    string `json:"category"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

func newExpenseResponse(expense Expense) expenseResponse {
	return expenseResponse{
		ID:          expense.ID,
		Amount:      FormatCents(expense.AmountCents),
		Category:    expense.Category,
		Description: expense.Description,
		CreatedAt:   expense.CreatedAt.Format(time.RFC3339),
	}
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /expenses", h.list)
	mux.HandleFunc("GET /expenses/{id}", h.get)
}

// GET

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	expenses, err := h.service.List(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response := make([]expenseResponse, 0, len(expenses))
	for _, expense := range expenses {
		response = append(response, newExpenseResponse(expense))
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	expense, err := h.service.GetByID(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		http.Error(w, "expense not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, newExpenseResponse(expense))
}

// Helper function

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
