package expense

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type TestSystem struct {
	Repo    *Repository
	Handler *Handler
	Mux     *http.ServeMux
}

func newTestSystem(t *testing.T) *TestSystem {
	t.Helper()

	repo := newTestRepository(t)
	handler := NewHandler(NewService(repo))
	mux := http.NewServeMux()
	handler.Register(mux)

	return &TestSystem{
		Repo:    repo,
		Handler: handler,
		Mux:     mux,
	}
}

func (ts *TestSystem) request(method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	ts.Mux.ServeHTTP(rec, req)
	return rec
}

func TestHandlerListEmpty(t *testing.T) {
	ts := newTestSystem(t)

	rec := ts.request(http.MethodGet, "/expenses")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "[]\n" {
		t.Fatalf("body = %q, want %q", rec.Body.String(), "[]\n")
	}
}

func TestHandlerGetNotFound(t *testing.T) {
	ts := newTestSystem(t)

	rec := ts.request(http.MethodGet, "/expenses/1")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestHandlerFormatCents(t *testing.T) {
	ts := newTestSystem(t)
	ctx := t.Context()

	const insertQuery = `
	INSERT INTO expenses (amount_cents, category, description, created_at)
	VALUES
    (5,  'food',        NULL, '2026-09-03 12:00:00+03'),
    (10000, 'electronics', NULL, '2026-09-18 12:00:00+03');
	`

	if _, err := ts.Repo.Pool.Exec(ctx, insertQuery); err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	rec := ts.request(http.MethodGet, "/expenses")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got []expenseResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}

	want := []string{"0.05", "100.00"}

	for i, row := range want {
		if got[i].Amount != row {
			t.Fatalf("want: %s, got: %s", row, got[i].Amount)
		}
	}
}
