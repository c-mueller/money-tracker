package service

import (
	"context"
	"fmt"
	"time"

	"icekalt.dev/money-tracker/internal/domain"
)

type RecurringExpenseService struct {
	repo      domain.RecurringExpenseRepo
	household *HouseholdService
}

func NewRecurringExpenseService(repo domain.RecurringExpenseRepo, household *HouseholdService) *RecurringExpenseService {
	return &RecurringExpenseService{repo: repo, household: household}
}

func (s *RecurringExpenseService) Create(ctx context.Context, householdID, categoryID int, name string, amount domain.Money, freq domain.Frequency, startDate time.Time, endDate *time.Time) (*domain.RecurringExpense, error) {
	if err := domain.ValidateHouseholdName(name); err != nil {
		return nil, err
	}
	if err := domain.ValidateAmount(amount); err != nil {
		return nil, err
	}
	if err := freq.Validate(); err != nil {
		return nil, err
	}
	if endDate != nil {
		if err := domain.ValidateDateRange(startDate, *endDate); err != nil {
			return nil, err
		}
	}

	if _, err := s.household.GetByID(ctx, householdID); err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, &domain.RecurringExpense{
		HouseholdID: householdID,
		CategoryID:  categoryID,
		Name:        name,
		Amount:      amount,
		Frequency:   freq,
		Active:      true,
		StartDate:   startDate,
		EndDate:     endDate,
	})
}

func (s *RecurringExpenseService) List(ctx context.Context, householdID int) ([]*domain.RecurringExpense, error) {
	if _, err := s.household.GetByID(ctx, householdID); err != nil {
		return nil, err
	}

	return s.repo.ListByHousehold(ctx, householdID)
}

func (s *RecurringExpenseService) Update(ctx context.Context, id int, name string, amount domain.Money, freq domain.Frequency, active bool, startDate time.Time, endDate *time.Time) (*domain.RecurringExpense, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if _, err := s.household.GetByID(ctx, existing.HouseholdID); err != nil {
		return nil, err
	}

	if err := domain.ValidateHouseholdName(name); err != nil {
		return nil, err
	}
	if err := domain.ValidateAmount(amount); err != nil {
		return nil, err
	}
	if err := freq.Validate(); err != nil {
		return nil, err
	}
	if endDate != nil {
		if err := domain.ValidateDateRange(startDate, *endDate); err != nil {
			return nil, err
		}
	}

	existing.Name = name
	existing.Amount = amount
	existing.Frequency = freq
	existing.Active = active
	existing.StartDate = startDate
	existing.EndDate = endDate
	return s.repo.Update(ctx, existing)
}

func (s *RecurringExpenseService) Delete(ctx context.Context, householdID, id int) error {
	if _, err := s.household.GetByID(ctx, householdID); err != nil {
		return err
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if existing.HouseholdID != householdID {
		return fmt.Errorf("%w: recurring expense does not belong to household", domain.ErrForbidden)
	}

	return s.repo.Delete(ctx, id)
}
