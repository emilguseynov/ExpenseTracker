package expense

import (
	"context"
	"errors"
	"strings"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	service := Service{repo: repo}
	return &service
}

func (s *Service) List(ctx context.Context) ([]Expense, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id int64) (Expense, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(
	ctx context.Context,
	amount, category, description string,
) (Expense, error) {
	amountCents, err := ParseCents(amount)
	if err != nil {
		return Expense{}, err
	}

	category, err = normalizeCategory(category)
	if err != nil {
		return Expense{}, err
	}

	return s.repo.Create(
		ctx,
		amountCents,
		category,
		&description,
	)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

var ErrInvalidCategory = errors.New("invalid category")

func normalizeCategory(category string) (string, error) {
	category = strings.ToLower(strings.TrimSpace(category))
	if category == "" {
		return "", ErrInvalidCategory
	}
	return category, nil
}
