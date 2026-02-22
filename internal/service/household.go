package service

import (
	"context"
	"fmt"

	"icekalt.dev/money-tracker/internal/domain"
)

type HouseholdService struct {
	repo        domain.HouseholdRepo
	categoryR   domain.CategoryRepo
	txR         domain.TransactionRepo
	recurringR  domain.RecurringExpenseRepo
}

func NewHouseholdService(
	repo domain.HouseholdRepo,
	categoryR domain.CategoryRepo,
	txR domain.TransactionRepo,
	recurringR domain.RecurringExpenseRepo,
) *HouseholdService {
	return &HouseholdService{
		repo:       repo,
		categoryR:  categoryR,
		txR:        txR,
		recurringR: recurringR,
	}
}

func (s *HouseholdService) Create(ctx context.Context, name, currency string) (*domain.Household, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("%w: no authenticated user", domain.ErrForbidden)
	}

	if err := domain.ValidateHouseholdName(name); err != nil {
		return nil, err
	}
	if err := domain.ValidateCurrency(currency); err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, &domain.Household{
		Name:     name,
		Currency: currency,
		OwnerID:  userID,
	})
}

func (s *HouseholdService) GetByID(ctx context.Context, id int) (*domain.Household, error) {
	hh, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.authorize(ctx, hh); err != nil {
		return nil, err
	}

	return hh, nil
}

func (s *HouseholdService) List(ctx context.Context) ([]*domain.Household, error) {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("%w: no authenticated user", domain.ErrForbidden)
	}

	return s.repo.ListByOwner(ctx, userID)
}

func (s *HouseholdService) Update(ctx context.Context, id int, name, currency string) (*domain.Household, error) {
	hh, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.authorize(ctx, hh); err != nil {
		return nil, err
	}

	if err := domain.ValidateHouseholdName(name); err != nil {
		return nil, err
	}
	if err := domain.ValidateCurrency(currency); err != nil {
		return nil, err
	}

	hh.Name = name
	hh.Currency = currency
	return s.repo.Update(ctx, hh)
}

func (s *HouseholdService) Delete(ctx context.Context, id int) error {
	hh, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.authorize(ctx, hh); err != nil {
		return err
	}

	return s.repo.Delete(ctx, id)
}

func (s *HouseholdService) authorize(ctx context.Context, hh *domain.Household) error {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("%w: no authenticated user", domain.ErrForbidden)
	}
	if hh.OwnerID != userID {
		return fmt.Errorf("%w: not household owner", domain.ErrForbidden)
	}
	return nil
}
