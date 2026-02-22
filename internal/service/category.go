package service

import (
	"context"
	"fmt"

	"icekalt.dev/money-tracker/internal/domain"
)

type CategoryService struct {
	repo      domain.CategoryRepo
	household *HouseholdService
}

func NewCategoryService(repo domain.CategoryRepo, household *HouseholdService) *CategoryService {
	return &CategoryService{repo: repo, household: household}
}

func (s *CategoryService) Create(ctx context.Context, householdID int, name string) (*domain.Category, error) {
	if err := domain.ValidateCategoryName(name); err != nil {
		return nil, err
	}

	if _, err := s.household.GetByID(ctx, householdID); err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, &domain.Category{
		HouseholdID: householdID,
		Name:        name,
	})
}

func (s *CategoryService) List(ctx context.Context, householdID int) ([]*domain.Category, error) {
	if _, err := s.household.GetByID(ctx, householdID); err != nil {
		return nil, err
	}

	return s.repo.ListByHousehold(ctx, householdID)
}

func (s *CategoryService) Update(ctx context.Context, id int, name string) (*domain.Category, error) {
	cat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if _, err := s.household.GetByID(ctx, cat.HouseholdID); err != nil {
		return nil, err
	}

	if err := domain.ValidateCategoryName(name); err != nil {
		return nil, err
	}

	cat.Name = name
	return s.repo.Update(ctx, cat)
}

func (s *CategoryService) Delete(ctx context.Context, householdID, id int) error {
	if _, err := s.household.GetByID(ctx, householdID); err != nil {
		return err
	}

	cat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if cat.HouseholdID != householdID {
		return fmt.Errorf("%w: category does not belong to household", domain.ErrForbidden)
	}

	return s.repo.Delete(ctx, id)
}
