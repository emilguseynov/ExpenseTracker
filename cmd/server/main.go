package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"expense-tracker/internal/database"
	"expense-tracker/internal/expense"
)

func main() {
	ctx := context.Background()

	pool, err := database.NewPool(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Error("database is not ready", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	slog.Info("database is ready")

	repo := expense.NewRepository(pool)    // data
	service := expense.NewService(repo)    // domain
	handler := expense.NewHandler(service) // presentation

	mux := http.NewServeMux()
	handler.Register(mux)

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
