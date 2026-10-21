# Personal Expense Tracker

A small expense ledger with a web page and a JSON API. Amounts are stored as integer cents.

## Stack

- Go 1.27+ and `net/http`
- PostgreSQL
- `pgx` for SQL, without an ORM
- Static page served by the same process

## What it does

- Add an expense with an amount (`50.00`), a category, and an optional description
- List expenses, newest first
- Filter the list and the statistics by category and date range
- Delete an expense
- Show the total and the sum per category

## Run

Install and start a PostgreSQL server. The `psql` client is not required.

```bash
make run
```

Open http://localhost:8080

On startup the server creates the `expense_tracker` database and the `expenses` table. If PostgreSQL is not responding, the process exits and prints the address it tried.

Use another database:

```bash
DATABASE_URL='postgres://localhost:5432/expense_tracker?sslmode=disable' make run
```

Run tests:

```bash
make test
```
