package main

import (
	"context"
	"log/slog"
	"os"

	"expense-tracker/internal/database"
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
}
