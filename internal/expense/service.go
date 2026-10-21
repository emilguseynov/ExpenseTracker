package expense

import (
	"context"
	"errors"
	"strings"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	service := Service{repo: repo}
	return &service
}

var ErrInvalidDateRange = errors.New("invalid date range")

func (s *Service) List(
	ctx context.Context,
	category string,
	from string,
	to string,
) ([]Expense, error) {
	filter, err := createFilter(ctx, category, from, to)
	if err != nil {
		return nil, err
	}
	return s.repo.List(ctx, filter)
}

func (s *Service) GetByID(ctx context.Context, id int64) (Expense, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Stats(
	ctx context.Context,
	category string,
	from string,
	to string,
) (Stats, error) {
	filter, err := createFilter(ctx, category, from, to)
	if err != nil {
		return Stats{}, err
	}
	return s.repo.Stats(ctx, filter)
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
	var desc string
	if description != "" {
		desc = description
	}

	return s.repo.Create(
		ctx,
		amountCents,
		category,
		&desc,
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

func createFilter(
	ctx context.Context,
	category string,
	from string,
	to string,
) (Filter, error) {
	filter := Filter{}

	if strings.TrimSpace(category) != "" {
		normalized, err := normalizeCategory(category)
		if err != nil {
			return Filter{}, err
		}
		filter.Category = normalized
	}

	if from != "" {
		fromDate, err := time.Parse("2006-01-02", from)
		if err != nil {
			return Filter{}, ErrInvalidDateRange
		}
		filter.From = fromDate
	}

	if to != "" {
		toDate, err := time.Parse("2006-01-02", to)
		if err != nil {
			return Filter{}, ErrInvalidDateRange
		}
		filter.End = toDate.AddDate(0, 0, 1)
	}

	if !filter.From.IsZero() && !filter.End.IsZero() && !filter.From.Before(filter.End) {
		return Filter{}, ErrInvalidDateRange
	}

	return filter, nil
}
