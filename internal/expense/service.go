package expense

import (
	"context"
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
