package expense

import (
	"context"
	"errors"

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

func (r *Repository) List(ctx context.Context) ([]Expense, error) {
	const query = `
		SELECT id, amount_cents, category, description, created_at
		FROM expenses
		ORDER BY id;

	`
	rows, err := r.Pool.Query(ctx, query)
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
