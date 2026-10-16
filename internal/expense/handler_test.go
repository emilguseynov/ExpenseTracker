package expense

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
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

func (ts *TestSystem) request(
	method, path string,
	body ...io.Reader,
) *httptest.ResponseRecorder {
	var r io.Reader
	if len(body) > 0 {
		r = body[0]
	}

	req := httptest.NewRequest(method, path, r)
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

// плохой тест, лучше отдельно протестить FormatCents (в service_test.go)
// и отдельно добавление и чтение из базы
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

func TestHandlerCreate(t *testing.T) {
	ts := newTestSystem(t)

	rec := ts.request(
		http.MethodPost,
		"/expenses",
		strings.NewReader(`{"amount":"50.00","category":"food","description":"lunch"}`),
	)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var created expenseResponse
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if created.Amount != "50.00" {
		t.Fatalf("amount = %q, want %q", created.Amount, "50.00")
	}
	if created.Category != "food" {
		t.Fatalf("category = %q, want %q", created.Category, "food")
	}
	if created.Description != "lunch" {
		t.Fatalf("description = %q, want %q", created.Description, "lunch")
	}

	got := ts.request(http.MethodGet, "/expenses/"+strconv.FormatInt(created.ID, 10))
	if got.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", got.Code, http.StatusOK)
	}

	var stored expenseResponse
	if err := json.NewDecoder(got.Body).Decode(&stored); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if stored != created {
		t.Fatalf("stored = %+v, want %+v", stored, created)
	}
}

// TODO: TestHandlerCreateTrimsCategory — POST с категорией " Food " даёт 201 и category "food".
// TODO: TestHandlerCreateEmptyDescription — пустое описание не попадает в JSON ответа.
// TODO: TestHandlerCreateInvalidAmount — сумма "0.00" даёт 400.
// TODO: TestHandlerCreateInvalidCategory — категория из пробелов даёт 400.
// TODO: TestHandlerDelete — DELETE существующего id даёт 204, следующий GET этого id даёт 404.
