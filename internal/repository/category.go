package repository

import (
	"context"
	"fmt"

	"icekalt.dev/money-tracker/ent"
	entcategory "icekalt.dev/money-tracker/ent/category"
	enthousehold "icekalt.dev/money-tracker/ent/household"
	"icekalt.dev/money-tracker/internal/domain"
)

type CategoryRepository struct {
	client *ent.Client
}

func NewCategoryRepository(client *ent.Client) *CategoryRepository {
	return &CategoryRepository{client: client}
}

func (r *CategoryRepository) Create(ctx context.Context, category *domain.Category) (*domain.Category, error) {
	c, err := r.client.Category.Create().
		SetName(category.Name).
		SetHouseholdID(category.HouseholdID).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, fmt.Errorf("%w: category name already exists in household", domain.ErrConflict)
		}
		return nil, err
	}
	c.Edges.Household = &ent.Household{ID: category.HouseholdID}
	return categoryToDomain(c), nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id int) (*domain.Category, error) {
	c, err := r.client.Category.Query().
		Where(entcategory.ID(id)).
		WithHousehold().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("%w: category %d", domain.ErrNotFound, id)
		}
		return nil, err
	}
	return categoryToDomain(c), nil
}

func (r *CategoryRepository) ListByHousehold(ctx context.Context, householdID int) ([]*domain.Category, error) {
	items, err := r.client.Category.Query().
		Where(entcategory.HasHouseholdWith(enthousehold.IDEQ(householdID))).
		WithHousehold().
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Category, 0, len(items))
	for _, c := range items {
		result = append(result, categoryToDomain(c))
	}
	return result, nil
}

func (r *CategoryRepository) Update(ctx context.Context, category *domain.Category) (*domain.Category, error) {
	c, err := r.client.Category.UpdateOneID(category.ID).
		SetName(category.Name).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("%w: category %d", domain.ErrNotFound, category.ID)
		}
		if ent.IsConstraintError(err) {
			return nil, fmt.Errorf("%w: category name already exists in household", domain.ErrConflict)
		}
		return nil, err
	}
	c.Edges.Household = &ent.Household{ID: category.HouseholdID}
	return categoryToDomain(c), nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id int) error {
	err := r.client.Category.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("%w: category %d", domain.ErrNotFound, id)
		}
		return err
	}
	return nil
}
