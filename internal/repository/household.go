package repository

import (
	"context"
	"fmt"

	"icekalt.dev/money-tracker/ent"
	enthousehold "icekalt.dev/money-tracker/ent/household"
	entrecurring "icekalt.dev/money-tracker/ent/recurringexpense"
	enttransaction "icekalt.dev/money-tracker/ent/transaction"
	entcategory "icekalt.dev/money-tracker/ent/category"
	entuser "icekalt.dev/money-tracker/ent/user"
	"icekalt.dev/money-tracker/internal/domain"
)

type HouseholdRepository struct {
	client *ent.Client
}

func NewHouseholdRepository(client *ent.Client) *HouseholdRepository {
	return &HouseholdRepository{client: client}
}

func (r *HouseholdRepository) Create(ctx context.Context, household *domain.Household) (*domain.Household, error) {
	h, err := r.client.Household.Create().
		SetName(household.Name).
		SetCurrency(household.Currency).
		SetIcon(household.Icon).
		SetOwnerID(household.OwnerID).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	h.Edges.Owner = &ent.User{ID: household.OwnerID}
	return householdToDomain(h), nil
}

func (r *HouseholdRepository) GetByID(ctx context.Context, id int) (*domain.Household, error) {
	h, err := r.client.Household.Query().
		Where(enthousehold.ID(id)).
		WithOwner().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("%w: household %d", domain.ErrNotFound, id)
		}
		return nil, err
	}
	return householdToDomain(h), nil
}

func (r *HouseholdRepository) ListByOwner(ctx context.Context, ownerID int) ([]*domain.Household, error) {
	items, err := r.client.Household.Query().
		Where(enthousehold.HasOwnerWith(entuser.IDEQ(ownerID))).
		WithOwner().
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Household, 0, len(items))
	for _, h := range items {
		result = append(result, householdToDomain(h))
	}
	return result, nil
}

func (r *HouseholdRepository) Update(ctx context.Context, household *domain.Household) (*domain.Household, error) {
	h, err := r.client.Household.UpdateOneID(household.ID).
		SetName(household.Name).
		SetCurrency(household.Currency).
		SetIcon(household.Icon).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("%w: household %d", domain.ErrNotFound, household.ID)
		}
		return nil, err
	}
	h.Edges.Owner = &ent.User{ID: household.OwnerID}
	return householdToDomain(h), nil
}

func (r *HouseholdRepository) Delete(ctx context.Context, id int) error {
	// Cascade delete children
	_, err := r.client.RecurringExpense.Delete().
		Where(entrecurring.HasHouseholdWith(enthousehold.IDEQ(id))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("deleting recurring expenses: %w", err)
	}

	_, err = r.client.Transaction.Delete().
		Where(enttransaction.HasHouseholdWith(enthousehold.IDEQ(id))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("deleting transactions: %w", err)
	}

	_, err = r.client.Category.Delete().
		Where(entcategory.HasHouseholdWith(enthousehold.IDEQ(id))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("deleting categories: %w", err)
	}

	err = r.client.Household.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("%w: household %d", domain.ErrNotFound, id)
		}
		return err
	}
	return nil
}
