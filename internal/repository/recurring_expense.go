package repository

import (
	"context"
	"fmt"

	"icekalt.dev/money-tracker/ent"
	enthousehold "icekalt.dev/money-tracker/ent/household"
	entrecurring "icekalt.dev/money-tracker/ent/recurringexpense"
	"icekalt.dev/money-tracker/internal/domain"
)

type RecurringExpenseRepository struct {
	client *ent.Client
}

func NewRecurringExpenseRepository(client *ent.Client) *RecurringExpenseRepository {
	return &RecurringExpenseRepository{client: client}
}

func (r *RecurringExpenseRepository) Create(ctx context.Context, expense *domain.RecurringExpense) (*domain.RecurringExpense, error) {
	q := r.client.RecurringExpense.Create().
		SetName(expense.Name).
		SetAmount(expense.Amount.String()).
		SetFrequency(string(expense.Frequency)).
		SetActive(expense.Active).
		SetStartDate(expense.StartDate).
		SetHouseholdID(expense.HouseholdID).
		SetCategoryID(expense.CategoryID)

	if expense.EndDate != nil {
		q.SetEndDate(*expense.EndDate)
	}

	re, err := q.Save(ctx)
	if err != nil {
		return nil, err
	}
	re.Edges.Household = &ent.Household{ID: expense.HouseholdID}
	re.Edges.Category = &ent.Category{ID: expense.CategoryID}
	return recurringExpenseToDomain(re), nil
}

func (r *RecurringExpenseRepository) GetByID(ctx context.Context, id int) (*domain.RecurringExpense, error) {
	re, err := r.client.RecurringExpense.Query().
		Where(entrecurring.ID(id)).
		WithHousehold().
		WithCategory().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("%w: recurring expense %d", domain.ErrNotFound, id)
		}
		return nil, err
	}
	return recurringExpenseToDomain(re), nil
}

func (r *RecurringExpenseRepository) ListByHousehold(ctx context.Context, householdID int) ([]*domain.RecurringExpense, error) {
	items, err := r.client.RecurringExpense.Query().
		Where(entrecurring.HasHouseholdWith(enthousehold.IDEQ(householdID))).
		WithHousehold().
		WithCategory().
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.RecurringExpense, 0, len(items))
	for _, re := range items {
		result = append(result, recurringExpenseToDomain(re))
	}
	return result, nil
}

func (r *RecurringExpenseRepository) ListActiveByHousehold(ctx context.Context, householdID int) ([]*domain.RecurringExpense, error) {
	items, err := r.client.RecurringExpense.Query().
		Where(
			entrecurring.HasHouseholdWith(enthousehold.IDEQ(householdID)),
			entrecurring.ActiveEQ(true),
		).
		WithHousehold().
		WithCategory().
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.RecurringExpense, 0, len(items))
	for _, re := range items {
		result = append(result, recurringExpenseToDomain(re))
	}
	return result, nil
}

func (r *RecurringExpenseRepository) Update(ctx context.Context, expense *domain.RecurringExpense) (*domain.RecurringExpense, error) {
	q := r.client.RecurringExpense.UpdateOneID(expense.ID).
		SetName(expense.Name).
		SetAmount(expense.Amount.String()).
		SetFrequency(string(expense.Frequency)).
		SetActive(expense.Active).
		SetStartDate(expense.StartDate)

	if expense.EndDate != nil {
		q.SetEndDate(*expense.EndDate)
	} else {
		q.ClearEndDate()
	}

	re, err := q.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("%w: recurring expense %d", domain.ErrNotFound, expense.ID)
		}
		return nil, err
	}
	re.Edges.Household = &ent.Household{ID: expense.HouseholdID}
	re.Edges.Category = &ent.Category{ID: expense.CategoryID}
	return recurringExpenseToDomain(re), nil
}

func (r *RecurringExpenseRepository) Delete(ctx context.Context, id int) error {
	err := r.client.RecurringExpense.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("%w: recurring expense %d", domain.ErrNotFound, id)
		}
		return err
	}
	return nil
}
