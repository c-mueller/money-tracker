package service

import (
	"context"
	"fmt"
	"time"

	"icekalt.dev/money-tracker/internal/domain"
)

type TransactionService struct {
	repo      domain.TransactionRepo
	household *HouseholdService
}

func NewTransactionService(repo domain.TransactionRepo, household *HouseholdService) *TransactionService {
	return &TransactionService{repo: repo, household: household}
}

func (s *TransactionService) Create(ctx context.Context, householdID, categoryID int, amount domain.Money, description string, date time.Time) (*domain.Transaction, error) {
	if err := domain.ValidateAmount(amount); err != nil {
		return nil, err
	}
	if err := domain.ValidateDescription(description); err != nil {
		return nil, err
	}

	if _, err := s.household.GetByID(ctx, householdID); err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, &domain.Transaction{
		HouseholdID: householdID,
		CategoryID:  categoryID,
		Amount:      amount,
		Description: description,
		Date:        date,
	})
}

func (s *TransactionService) GetByID(ctx context.Context, id int) (*domain.Transaction, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *TransactionService) ListByMonth(ctx context.Context, householdID int, year int, month time.Month) ([]*domain.Transaction, error) {
	if _, err := s.household.GetByID(ctx, householdID); err != nil {
		return nil, err
	}

	return s.repo.ListByHouseholdAndMonth(ctx, householdID, year, month)
}

func (s *TransactionService) Update(ctx context.Context, householdID, id, categoryID int, amount domain.Money, description string, date time.Time) (*domain.Transaction, error) {
	if err := domain.ValidateAmount(amount); err != nil {
		return nil, err
	}
	if err := domain.ValidateDescription(description); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing.HouseholdID != householdID {
		return nil, fmt.Errorf("%w: transaction does not belong to household", domain.ErrForbidden)
	}

	existing.CategoryID = categoryID
	existing.Amount = amount
	existing.Description = description
	existing.Date = date
	return s.repo.Update(ctx, existing)
}

func (s *TransactionService) Delete(ctx context.Context, householdID, id int) error {
	if _, err := s.household.GetByID(ctx, householdID); err != nil {
		return err
	}

	tx, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if tx.HouseholdID != householdID {
		return fmt.Errorf("%w: transaction does not belong to household", domain.ErrForbidden)
	}

	return s.repo.Delete(ctx, id)
}
