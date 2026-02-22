package repository

import (
	"context"
	"fmt"

	"icekalt.dev/money-tracker/ent"
	entrecurring "icekalt.dev/money-tracker/ent/recurringexpense"
	entoverride "icekalt.dev/money-tracker/ent/recurringscheduleoverride"
	"icekalt.dev/money-tracker/internal/domain"
)

type RecurringScheduleOverrideRepository struct {
	client *ent.Client
}

func NewRecurringScheduleOverrideRepository(client *ent.Client) *RecurringScheduleOverrideRepository {
	return &RecurringScheduleOverrideRepository{client: client}
}

func (r *RecurringScheduleOverrideRepository) Create(ctx context.Context, override *domain.RecurringScheduleOverride) (*domain.RecurringScheduleOverride, error) {
	o, err := r.client.RecurringScheduleOverride.Create().
		SetEffectiveDate(override.EffectiveDate).
		SetAmount(override.Amount.String()).
		SetFrequency(string(override.Frequency)).
		SetRecurringExpenseID(override.RecurringExpenseID).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	o.Edges.RecurringExpense = &ent.RecurringExpense{ID: override.RecurringExpenseID}
	return overrideToDomain(o), nil
}

func (r *RecurringScheduleOverrideRepository) GetByID(ctx context.Context, id int) (*domain.RecurringScheduleOverride, error) {
	o, err := r.client.RecurringScheduleOverride.Query().
		Where(entoverride.ID(id)).
		WithRecurringExpense().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("%w: schedule override %d", domain.ErrNotFound, id)
		}
		return nil, err
	}
	return overrideToDomain(o), nil
}

func (r *RecurringScheduleOverrideRepository) ListByRecurringExpense(ctx context.Context, recurringExpenseID int) ([]*domain.RecurringScheduleOverride, error) {
	items, err := r.client.RecurringScheduleOverride.Query().
		Where(entoverride.HasRecurringExpenseWith(entrecurring.IDEQ(recurringExpenseID))).
		WithRecurringExpense().
		Order(ent.Asc(entoverride.FieldEffectiveDate)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.RecurringScheduleOverride, 0, len(items))
	for _, o := range items {
		result = append(result, overrideToDomain(o))
	}
	return result, nil
}

func (r *RecurringScheduleOverrideRepository) Update(ctx context.Context, override *domain.RecurringScheduleOverride) (*domain.RecurringScheduleOverride, error) {
	o, err := r.client.RecurringScheduleOverride.UpdateOneID(override.ID).
		SetEffectiveDate(override.EffectiveDate).
		SetAmount(override.Amount.String()).
		SetFrequency(string(override.Frequency)).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("%w: schedule override %d", domain.ErrNotFound, override.ID)
		}
		return nil, err
	}
	o.Edges.RecurringExpense = &ent.RecurringExpense{ID: override.RecurringExpenseID}
	return overrideToDomain(o), nil
}

func (r *RecurringScheduleOverrideRepository) Delete(ctx context.Context, id int) error {
	err := r.client.RecurringScheduleOverride.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("%w: schedule override %d", domain.ErrNotFound, id)
		}
		return err
	}
	return nil
}
