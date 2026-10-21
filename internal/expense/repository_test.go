package expense

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"expense-tracker/internal/database"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestMain(m *testing.M) {
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		if err := database.Prepare(context.Background(), url); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}
	os.Exit(m.Run())
}

func TestListAndGetByID(t *testing.T) {
	repo := newTestRepository(t)
	ctx := t.Context()
	const insertQuery = `
	INSERT INTO expenses (amount_cents, category, description, created_at)
	VALUES
    (5000,  'food',        NULL, '2026-09-03 12:00:00+03'),
    (2000,  'transport',   NULL, '2026-09-12 12:00:00+03'),
    (10000, 'electronics', NULL, '2026-09-18 12:00:00+03'),
    (3000,  'food',        NULL, '2026-09-27 15:00:00+03');
	`

	if _, err := repo.Pool.Exec(ctx, insertQuery); err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	got, err := repo.List(ctx, Filter{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(got) != 4 {
		t.Fatalf("list length = %d, want 4", len(got))
	}

	want := []struct {
		id       int64
		cents    int64
		category string
	}{
		{4, 3000, "food"},
		{3, 10000, "electronics"},
		{2, 2000, "transport"},
		{1, 5000, "food"},
	}
	for i, row := range want {
		if got[i].ID != row.id || got[i].AmountCents != row.cents || got[i].Category != row.category {
			t.Fatalf("row %d = %+v, want id=%d cents=%d category=%s", i, got[i], row.id, row.cents, row.category)
		}
		if got[i].Description != "" {
			t.Fatalf("row %d description = %q, want empty", i, got[i].Description)
		}
	}

	_, err = repo.GetByID(ctx, 999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("wrong err type: %v, want ErrNotFound", err)
	}
}

func newTestRepository(t *testing.T) *Repository {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	ctx := t.Context()

	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL is empty")
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to open pool: %v", err)
	}
	t.Cleanup(func() {
		pool.Close()
	})

	repo := Repository{Pool: pool}

	const dropTableQuery = `
		DROP TABLE IF EXISTS expenses;
	`
	if _, err := repo.Pool.Exec(ctx, dropTableQuery); err != nil {
		t.Fatalf("drop table: %v", err)
	}

	migration, err := os.ReadFile("../../migrations/001_create_expenses.sql")
	if err != nil {
		t.Fatalf("read migration err: %v", err)
	}

	if _, err := repo.Pool.Exec(ctx, string(migration)); err != nil {
		t.Fatalf("apply migration: %v", err)
	}

	return &repo
}
