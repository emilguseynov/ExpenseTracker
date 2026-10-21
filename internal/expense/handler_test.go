package expense

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

const standardInsertQuery = `
	INSERT INTO expenses (amount_cents, category, description, created_at)
	VALUES
	(5000,  'food',        NULL, '2026-09-03 12:00:00+03'),
	(2000,  'transport',   NULL, '2026-09-12 12:00:00+03'),
	(10000, 'electronics', NULL, '2026-09-18 12:00:00+03'),
	(3000,  'food',        NULL, '2026-09-27 15:00:00+03'),
	(2000,  'food',        NULL, '2026-10-18 12:00:00+03');
`

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

func TestHandlerCreateTrimsCategory(t *testing.T) {
	ts := newTestSystem(t)

	rec := ts.request(
		http.MethodPost,
		"/expenses",
		strings.NewReader(`{"amount":"50.00","category":" Food ","description":"lunch"}`),
	)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var created expenseResponse
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if created.Category != "food" {
		t.Fatalf("category = %q, want %q", created.Category, "food")
	}
}

func TestHandlerCreateEmptyDescription(t *testing.T) {
	ts := newTestSystem(t)

	rec := ts.request(
		http.MethodPost,
		"/expenses",
		strings.NewReader(`{"amount":"50.00","category":"food","description":""}`),
	)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if value, ok := body["description"]; ok {
		t.Fatalf("description = %v, want field absent", value)
	}
}

func TestHandlerCreateInvalidAmount(t *testing.T) {
	ts := newTestSystem(t)

	rec := ts.request(
		http.MethodPost,
		"/expenses",
		strings.NewReader(`{"amount":"0.00","category":"food","description":"lunch"}`),
	)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("response code = %v, want 400 BadRequest", rec.Code)
	}
}

func TestHandlerCreateInvalidCategory(t *testing.T) {
	ts := newTestSystem(t)

	rec := ts.request(
		http.MethodPost,
		"/expenses",
		strings.NewReader(`{"amount":"50.00","category":"      ","description":"lunch"}`),
	)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("response code = %v, want 400 BadRequest", rec.Code)
	}
}

func TestHandlerDelete(t *testing.T) {
	ts := newTestSystem(t)

	const insertQuery = `
	INSERT INTO expenses (amount_cents, category, description, created_at)
	VALUES
    (5,  'food',        NULL, '2026-09-03 12:00:00+03');
	`

	if _, err := ts.Repo.Pool.Exec(
		t.Context(),
		insertQuery,
	); err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	rec := ts.request(
		http.MethodDelete,
		"/expenses/1",
	)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("response code = %v, want 204 StatusNoContent", rec.Code)
	}

	rec = ts.request(
		http.MethodGet,
		"/expenses/1",
	)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("response code = %v, want 404 NotFound", rec.Code)
	}
}

func TestListWithCategory(t *testing.T) {
	ts := newTestSystem(t)

	if _, err := ts.Repo.Pool.Exec(t.Context(), standardInsertQuery); err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	rec := ts.request(
		http.MethodGet,
		"/expenses?category=food",
	)

	response := []expenseResponse{}

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed parsing json: %v", err)
	}

	for _, exp := range response {
		if exp.Category != "food" {
			t.Fatalf("wrong category in response: %v, want food", exp.Category)
		}
	}
}

func TestListWithDateFilter(t *testing.T) {
	ts := newTestSystem(t)

	if _, err := ts.Repo.Pool.Exec(t.Context(), standardInsertQuery); err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	rec := ts.request(
		http.MethodGet,
		"/expenses?from=2026-09-01&to=2026-09-30",
	)

	response := []expenseResponse{}

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed parsing json: %v", err)
	}

	for _, exp := range response {
		createdAt, err := time.Parse(time.RFC3339, exp.CreatedAt)
		if err != nil {
			t.Fatalf("created_at: %v", err)
		}
		if createdAt.Month() != time.September {
			t.Fatalf("created_at month = %s, want September", createdAt.Month())
		}
	}
}

func TestListWithCategoryAndDateFilter(t *testing.T) {
	ts := newTestSystem(t)

	if _, err := ts.Repo.Pool.Exec(t.Context(), standardInsertQuery); err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	rec := ts.request(
		http.MethodGet,
		"/expenses?category=food&from=2026-09-01&to=2026-09-30",
	)

	response := []expenseResponse{}

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed parsing json: %v", err)
	}

	for _, exp := range response {
		if exp.Category != "food" {
			t.Fatalf("wrong category in response: %v, want food", exp.Category)
		}

		createdAt, err := time.Parse(time.RFC3339, exp.CreatedAt)
		if err != nil {
			t.Fatalf("created_at: %v", err)
		}
		if createdAt.Month() != time.September {
			t.Fatalf("created_at month = %s, want September", createdAt.Month())
		}
	}
}

func TestListWithInvalidDateFilter(t *testing.T) {
	ts := newTestSystem(t)

	if _, err := ts.Repo.Pool.Exec(t.Context(), standardInsertQuery); err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	rec := ts.request(
		http.MethodGet,
		"/expenses?from=2026-09-03&to=2026-09-02",
	)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("wrong status code: %v, want 400 BadRequest", rec.Code)
	}
}

func TestStatsNoParameters(t *testing.T) {
	ts := newTestSystem(t)

	if _, err := ts.Repo.Pool.Exec(t.Context(), standardInsertQuery); err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	rec := ts.request(
		http.MethodGet,
		"/stats",
	)

	if rec.Code != http.StatusOK {
		t.Fatalf("status not ok: %v", rec.Code)
	}

	var response statsResponse

	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed parsing json: %v", err)
	}

	want := map[string]string{
		"electronics": "100.00",
		"food":        "100.00",
		"transport":   "20.00",
	}

	if response.Total != "220.00" {
		t.Fatalf("wrong total: %v, want 220.00", response.Total)
	}
	// if response.ByCategory["electronics"] != want[]

	for key, value := range want {
		if value != response.ByCategory[key] {
			t.Fatalf(
				"wrong sum for %s category: %v, want %v",
				key, value, want[key],
			)
		}
	}

	if len(response.ByCategory) != 3 {
		t.Fatalf("wrong amount of categories: %v, want 3", len(response.ByCategory))
	}
}

func TestStatsSeptember(t *testing.T) {
	response := requestStats(t, "/stats?from=2026-09-01&to=2026-09-30")

	if response.Total != "200.00" {
		t.Fatalf("total = %s, want 200.00", response.Total)
	}

	want := map[string]string{
		"electronics": "100.00",
		"food":        "80.00",
		"transport":   "20.00",
	}
	assertStatsCategories(t, response.ByCategory, want)
}

func TestStatsFood(t *testing.T) {
	want := map[string]string{
		"food": "100.00",
	}

	for _, path := range []string{
		"/stats?category=food",
		"/stats?category=Food",
	} {
		response := requestStats(t, path)
		if response.Total != "100.00" {
			t.Fatalf("%s total = %s, want 100.00", path, response.Total)
		}
		assertStatsCategories(t, response.ByCategory, want)
	}
}

func TestStatsFoodSeptember(t *testing.T) {
	response := requestStats(t, "/stats?category=food&from=2026-09-01&to=2026-09-30")

	if response.Total != "80.00" {
		t.Fatalf("total = %s, want 80.00", response.Total)
	}

	assertStatsCategories(t, response.ByCategory, map[string]string{
		"food": "80.00",
	})
}

func TestStatsMissingCategory(t *testing.T) {
	response := requestStats(t, "/stats?category=rent")

	if response.Total != "0.00" {
		t.Fatalf("total = %s, want 0.00", response.Total)
	}
	if len(response.ByCategory) != 0 {
		t.Fatalf("by_category = %v, want empty", response.ByCategory)
	}
}

func TestStatsInvalidDateFilter(t *testing.T) {
	ts := newTestSystem(t)

	if _, err := ts.Repo.Pool.Exec(t.Context(), standardInsertQuery); err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	for _, path := range []string{
		"/stats?from=2026-13-01",
		"/stats?from=2026-10-01&to=2026-09-01",
	} {
		rec := ts.request(http.MethodGet, path)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s status = %d, want 400", path, rec.Code)
		}
	}
}

func TestPostNotJSON(t *testing.T) {
	ts := newTestSystem(t)

	rec := ts.request(
		http.MethodPost,
		"/expenses",
		strings.NewReader(`not-json`),
	)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("wrong response code: %v, want 400", rec.Code)
	}
}

func TestIrregularFieldInJSON(t *testing.T) {
	ts := newTestSystem(t)

	rec := ts.request(
		http.MethodPost,
		"/expenses",
		strings.NewReader(`{"amount":"50.00","category":"food","note":"x"}`),
	)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("wrong response code: %v, want 400", rec.Code)
	}
}

func requestStats(t *testing.T, path string) statsResponse {
	t.Helper()

	ts := newTestSystem(t)
	if _, err := ts.Repo.Pool.Exec(t.Context(), standardInsertQuery); err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	rec := ts.request(http.MethodGet, path)
	if rec.Code != http.StatusOK {
		t.Fatalf("%s status = %d, want 200", path, rec.Code)
	}

	var response statsResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed parsing json: %v", err)
	}
	return response
}

func assertStatsCategories(t *testing.T, got, want map[string]string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("by_category = %v, want %v", got, want)
	}
	for key, value := range want {
		if got[key] != value {
			t.Fatalf("by_category[%s] = %s, want %s", key, got[key], value)
		}
	}
}
