package service_test

import (
	"errors"
	"testing"

	"icekalt.dev/money-tracker/internal/domain"
)

func TestCategoryCreate(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)
	hh := createTestHousehold(t, svc, ctx)

	t.Run("success", func(t *testing.T) {
		cat, err := svc.Category.Create(ctx, hh.ID, "Groceries", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cat.Name != "Groceries" {
			t.Errorf("Name = %q, want %q", cat.Name, "Groceries")
		}
		if cat.Icon != "category" {
			t.Errorf("Icon = %q, want %q (default)", cat.Icon, "category")
		}
		if cat.HouseholdID != hh.ID {
			t.Errorf("HouseholdID = %d, want %d", cat.HouseholdID, hh.ID)
		}
	})

	t.Run("custom icon", func(t *testing.T) {
		cat, err := svc.Category.Create(ctx, hh.ID, "Transport", "car")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cat.Icon != "car" {
			t.Errorf("Icon = %q, want %q", cat.Icon, "car")
		}
	})

	t.Run("empty name", func(t *testing.T) {
		_, err := svc.Category.Create(ctx, hh.ID, "", "")
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("household not found", func(t *testing.T) {
		_, err := svc.Category.Create(ctx, 99999, "Test", "")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestCategoryList(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)
	hh := createTestHousehold(t, svc, ctx)

	t.Run("empty list", func(t *testing.T) {
		list, err := svc.Category.List(ctx, hh.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 0 {
			t.Errorf("expected empty list, got %d", len(list))
		}
	})

	t.Run("filter by household", func(t *testing.T) {
		svc.Category.Create(ctx, hh.ID, "Cat A", "")
		svc.Category.Create(ctx, hh.ID, "Cat B", "")

		hh2, _ := svc.Household.Create(ctx, "Other Home", "", "USD", "")
		svc.Category.Create(ctx, hh2.ID, "Cat C", "")

		list, err := svc.Category.List(ctx, hh.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("expected 2 categories for hh, got %d", len(list))
		}
	})
}

func TestCategoryGetByID(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)
	hh := createTestHousehold(t, svc, ctx)

	t.Run("success", func(t *testing.T) {
		cat := createTestCategory(t, svc, ctx, hh.ID)
		got, err := svc.Category.GetByID(ctx, cat.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != cat.ID {
			t.Errorf("ID = %d, want %d", got.ID, cat.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.Category.GetByID(ctx, 99999)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestCategoryUpdate(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)
	hh := createTestHousehold(t, svc, ctx)
	cat := createTestCategory(t, svc, ctx, hh.ID)

	t.Run("success", func(t *testing.T) {
		updated, err := svc.Category.Update(ctx, cat.ID, "New Name", "star")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Name != "New Name" {
			t.Errorf("Name = %q, want %q", updated.Name, "New Name")
		}
		if updated.Icon != "star" {
			t.Errorf("Icon = %q, want %q", updated.Icon, "star")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.Category.Update(ctx, 99999, "Name", "icon")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		_, err := svc.Category.Update(ctx, cat.ID, "", "icon")
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})
}

func TestCategoryDelete(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)
	hh := createTestHousehold(t, svc, ctx)

	t.Run("success", func(t *testing.T) {
		cat := createTestCategory(t, svc, ctx, hh.ID)
		if err := svc.Category.Delete(ctx, hh.ID, cat.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err := svc.Category.GetByID(ctx, cat.ID)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})

	t.Run("wrong household", func(t *testing.T) {
		cat := createTestCategory(t, svc, ctx, hh.ID)
		hh2, _ := svc.Household.Create(ctx, "Other", "", "EUR", "")

		err := svc.Category.Delete(ctx, hh2.ID, cat.ID)
		if !errors.Is(err, domain.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		err := svc.Category.Delete(ctx, hh.ID, 99999)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}
