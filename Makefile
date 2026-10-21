.PHONY: run test

DATABASE_URL ?= postgres://localhost:5432/expense_tracker?sslmode=disable
TEST_DATABASE_URL ?= postgres://localhost:5432/expense_tracker_test?sslmode=disable

run:
	DATABASE_URL='$(DATABASE_URL)' go run ./cmd/server

test:
	TEST_DATABASE_URL='$(TEST_DATABASE_URL)' go test ./...
