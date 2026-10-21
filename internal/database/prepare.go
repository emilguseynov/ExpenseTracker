package database

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"
)

var dbNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func Prepare(ctx context.Context, databaseURL string) error {
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is empty")
	}

	cfg, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return fmt.Errorf("parse database url: %w", err)
	}
	dbName := cfg.Database
	if !dbNamePattern.MatchString(dbName) {
		return fmt.Errorf("database name %q is not allowed", dbName)
	}

	adminCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	adminCfg, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return fmt.Errorf("parse database url: %w", err)
	}
	adminCfg.Database = "postgres"

	admin, err := pgx.ConnectConfig(adminCtx, adminCfg)
	if err != nil {
		if unreachable(err) {
			return postgresUnavailable(adminCfg.Host, adminCfg.Port)
		}
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer admin.Close(adminCtx)

	var exists bool
	err = admin.QueryRow(
		adminCtx,
		`SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`,
		dbName,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check database: %w", err)
	}
	if !exists {
		_, err = admin.Exec(adminCtx, `CREATE DATABASE `+dbName)
		if err != nil {
			return fmt.Errorf("create database: %w", err)
		}
	}

	target, err := pgx.ConnectConfig(adminCtx, cfg)
	if err != nil {
		if unreachable(err) {
			return postgresUnavailable(cfg.Host, cfg.Port)
		}
		return fmt.Errorf("connect to %s: %w", dbName, err)
	}
	defer target.Close(adminCtx)

	var table *string
	err = target.QueryRow(adminCtx, `SELECT to_regclass('public.expenses')`).Scan(&table)
	if err != nil {
		return fmt.Errorf("check expenses table: %w", err)
	}
	if table != nil {
		return nil
	}

	migration, err := readMigration()
	if err != nil {
		return err
	}
	if _, err := target.Exec(adminCtx, migration); err != nil {
		return fmt.Errorf("apply migration: %w", err)
	}
	return nil
}

func postgresUnavailable(host string, port uint16) error {
	if host == "" {
		host = "localhost"
	}
	return fmt.Errorf(
		"PostgreSQL is not responding at %s:%d.\nInstall and start PostgreSQL, then run make run again",
		host,
		port,
	)
}

func unreachable(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr)
}

func readMigration() (string, error) {
	candidates := []string{
		"migrations/001_create_expenses.sql",
		"../../migrations/001_create_expenses.sql",
	}
	for _, path := range candidates {
		body, err := os.ReadFile(path)
		if err == nil {
			return string(body), nil
		}
	}
	return "", fmt.Errorf("read migration migrations/001_create_expenses.sql")
}
