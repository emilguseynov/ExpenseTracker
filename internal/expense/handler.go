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
	Description string `json:"description,omitempty"`
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

type createExpenseRequest struct {
	Amount      string `json:"amount"`
	Category    string `json:"category"`
	Description string `json:"description"`
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
	mux.HandleFunc("POST /expenses", h.post)
	mux.HandleFunc("DELETE /expenses/{id}", h.delete)
}

// GET

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	category := query.Get("category")
	from := query.Get("from")
	to := query.Get("to")

	expenses, err := h.service.List(
		r.Context(),
		category,
		from,
		to,
	)
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

// POST

func (h *Handler) post(w http.ResponseWriter, r *http.Request) {
	var req createExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	exp, err := h.service.Create(
		r.Context(),
		req.Amount,
		req.Category,
		req.Description,
	)
	if errors.Is(err, ErrInvalidAmount) || errors.Is(err, ErrInvalidCategory) {
		http.Error(w, "invalid amount", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "failed creating an expense: %v", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, newExpenseResponse(exp))
}

// DELETE

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err = h.service.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, "failed deleting an expense", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper function

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
