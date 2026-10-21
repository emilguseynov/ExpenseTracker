package expense

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	Pool *pgxpool.Pool
}

var ErrNotFound = errors.New("expense not found")

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{Pool: pool}
}

func (r *Repository) List(ctx context.Context, filter Filter) ([]Expense, error) {
	query := `
		SELECT id, amount_cents, category, description, created_at
		FROM expenses
	`
	filterQuery, args := filterQueryBuilder(filter)
	query += filterQuery
	query += " ORDER BY id DESC"

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	expenses := make([]Expense, 0)

	for rows.Next() {
		var expense Expense
		var description *string

		err := rows.Scan(
			&expense.ID,
			&expense.AmountCents,
			&expense.Category,
			&description,
			&expense.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if description != nil {
			expense.Description = *description
		}

		expenses = append(expenses, expense)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return expenses, nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (Expense, error) {
	const query = `
		SELECT id, amount_cents, category, description, created_at
		FROM expenses
		WHERE id = $1
		ORDER BY id;
		`
	row := r.Pool.QueryRow(ctx, query, id)

	expense := Expense{}
	var description *string

	err := row.Scan(
		&expense.ID,
		&expense.AmountCents,
		&expense.Category,
		&description,
		&expense.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Expense{}, ErrNotFound
		}
		return Expense{}, err
	}

	if description != nil {
		expense.Description = *description
	}

	return expense, nil
}

func (r *Repository) Stats(ctx context.Context, filter Filter) (Stats, error) {
	query := `
		SELECT category, SUM(amount_cents)
		FROM expenses
	`
	filterQuery, args := filterQueryBuilder(filter)
	query += filterQuery
	query += " GROUP BY category"

	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return Stats{}, err
	}
	defer rows.Close()

	stats := Stats{
		ByCategory: make(map[string]int64),
	}

	for rows.Next() {
		var category string
		var amountCents int64

		if err := rows.Scan(&category, &amountCents); err != nil {
			return Stats{}, err
		}

		stats.ByCategory[category] = amountCents
	}
	if err = rows.Err(); err != nil {
		return Stats{}, err
	}

	for _, sum := range stats.ByCategory {
		stats.TotalCents += sum
	}

	return stats, nil
}

func (r *Repository) Create(
	ctx context.Context,
	amountCents int64,
	category string,
	description *string,
) (Expense, error) {
	insertQuery := `
		INSERT INTO expenses (amount_cents, category, description)
		VALUES ($1, $2, $3)
		RETURNING id, amount_cents, category, description, created_at
	`

	row := r.Pool.QueryRow(ctx, insertQuery, amountCents, category, description)

	expense := Expense{}
	var desc *string

	err := row.Scan(
		&expense.ID,
		&expense.AmountCents,
		&expense.Category,
		&desc,
		&expense.CreatedAt,
	)
	if err != nil {
		return Expense{}, err
	}

	if desc != nil {
		expense.Description = *desc
	}

	return expense, nil
}

func (r *Repository) Delete(
	ctx context.Context,
	id int64,
) error {
	deleteQuery := `
		DELETE FROM expenses
		WHERE id = $1;
	`

	tag, err := r.Pool.Exec(ctx, deleteQuery, id)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func filterQueryBuilder(filter Filter) (query string, args []any) {
	conditions := make([]string, 0, 3)

	if filter.Category != "" {
		args = append(args, filter.Category)
		conditions = append(
			conditions,
			fmt.Sprintf("category = $%d", len(args)),
		)
	}

	if !filter.From.IsZero() {
		args = append(args, filter.From)
		conditions = append(
			conditions,
			fmt.Sprintf("created_at >= $%d", len(args)),
		)
	}

	if !filter.End.IsZero() {
		args = append(args, filter.End)
		conditions = append(
			conditions,
			fmt.Sprintf("created_at < $%d", len(args)),
		)
	}

	if len(conditions) > 0 {
		return " WHERE " + strings.Join(conditions, " AND "), args
	}

	return "", []any{}
}
