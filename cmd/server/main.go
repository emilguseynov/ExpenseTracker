package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"expense-tracker/internal/database"
	"expense-tracker/internal/expense"
)

func main() {
	ctx := context.Background()
	databaseURL := os.Getenv("DATABASE_URL")

	if err := database.Prepare(ctx, databaseURL); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	pool, err := database.NewPool(ctx, databaseURL)
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
	mux.Handle("GET /", http.FileServer(http.Dir("web")))

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
