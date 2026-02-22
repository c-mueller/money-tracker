package repository

import (
	"context"
	"fmt"
	"time"

	"icekalt.dev/money-tracker/ent"
	enthousehold "icekalt.dev/money-tracker/ent/household"
	enttransaction "icekalt.dev/money-tracker/ent/transaction"
	"icekalt.dev/money-tracker/internal/domain"
)

type TransactionRepository struct {
	client *ent.Client
}

func NewTransactionRepository(client *ent.Client) *TransactionRepository {
	return &TransactionRepository{client: client}
}

func (r *TransactionRepository) Create(ctx context.Context, tx *domain.Transaction) (*domain.Transaction, error) {
	t, err := r.client.Transaction.Create().
		SetAmount(tx.Amount.String()).
		SetDescription(tx.Description).
		SetDate(tx.Date).
		SetHouseholdID(tx.HouseholdID).
		SetCategoryID(tx.CategoryID).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	t.Edges.Household = &ent.Household{ID: tx.HouseholdID}
	t.Edges.Category = &ent.Category{ID: tx.CategoryID}
	return transactionToDomain(t), nil
}

func (r *TransactionRepository) GetByID(ctx context.Context, id int) (*domain.Transaction, error) {
	t, err := r.client.Transaction.Query().
		Where(enttransaction.ID(id)).
		WithHousehold().
		WithCategory().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("%w: transaction %d", domain.ErrNotFound, id)
		}
		return nil, err
	}
	return transactionToDomain(t), nil
}

func (r *TransactionRepository) ListByHouseholdAndMonth(ctx context.Context, householdID int, year int, month time.Month) ([]*domain.Transaction, error) {
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	items, err := r.client.Transaction.Query().
		Where(
			enttransaction.HasHouseholdWith(enthousehold.IDEQ(householdID)),
			enttransaction.DateGTE(start),
			enttransaction.DateLT(end),
		).
		WithHousehold().
		WithCategory().
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Transaction, 0, len(items))
	for _, t := range items {
		result = append(result, transactionToDomain(t))
	}
	return result, nil
}

func (r *TransactionRepository) Update(ctx context.Context, tx *domain.Transaction) (*domain.Transaction, error) {
	t, err := r.client.Transaction.UpdateOneID(tx.ID).
		SetAmount(tx.Amount.String()).
		SetDescription(tx.Description).
		SetDate(tx.Date).
		SetCategoryID(tx.CategoryID).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("%w: transaction %d", domain.ErrNotFound, tx.ID)
		}
		return nil, err
	}
	t.Edges.Household = &ent.Household{ID: tx.HouseholdID}
	t.Edges.Category = &ent.Category{ID: tx.CategoryID}
	return transactionToDomain(t), nil
}

func (r *TransactionRepository) Delete(ctx context.Context, id int) error {
	err := r.client.Transaction.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("%w: transaction %d", domain.ErrNotFound, id)
		}
		return err
	}
	return nil
}
