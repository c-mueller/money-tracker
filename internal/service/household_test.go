package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"icekalt.dev/money-tracker/internal/domain"
	"icekalt.dev/money-tracker/internal/service"
)

func TestHouseholdCreate(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)

	t.Run("success", func(t *testing.T) {
		hh, err := svc.Household.Create(ctx, "My Home", "A nice place", "EUR", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if hh.Name != "My Home" {
			t.Errorf("Name = %q, want %q", hh.Name, "My Home")
		}
		if hh.Currency != "EUR" {
			t.Errorf("Currency = %q, want %q", hh.Currency, "EUR")
		}
		if hh.Icon != "home" {
			t.Errorf("Icon = %q, want %q (default)", hh.Icon, "home")
		}
		if hh.Description != "A nice place" {
			t.Errorf("Description = %q, want %q", hh.Description, "A nice place")
		}
	})

	t.Run("custom icon", func(t *testing.T) {
		hh, err := svc.Household.Create(ctx, "Office", "", "USD", "building")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if hh.Icon != "building" {
			t.Errorf("Icon = %q, want %q", hh.Icon, "building")
		}
	})

	t.Run("empty name", func(t *testing.T) {
		_, err := svc.Household.Create(ctx, "", "", "EUR", "")
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("invalid currency", func(t *testing.T) {
		_, err := svc.Household.Create(ctx, "Test", "", "invalid", "")
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("no auth context", func(t *testing.T) {
		_, err := svc.Household.Create(context.Background(), "Test", "", "EUR", "")
		if !errors.Is(err, domain.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})
}

func TestHouseholdList(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)

	t.Run("empty list", func(t *testing.T) {
		list, err := svc.Household.List(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 0 {
			t.Errorf("expected empty list, got %d", len(list))
		}
	})

	t.Run("multiple households", func(t *testing.T) {
		svc.Household.Create(ctx, "Home 1", "", "EUR", "")
		svc.Household.Create(ctx, "Home 2", "", "USD", "")

		list, err := svc.Household.List(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("expected 2 households, got %d", len(list))
		}
	})
}

func TestHouseholdUpdate(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)
	hh := createTestHousehold(t, svc, ctx)

	t.Run("success", func(t *testing.T) {
		updated, err := svc.Household.Update(ctx, hh.ID, "New Name", "Updated desc", "USD", "star")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Name != "New Name" {
			t.Errorf("Name = %q, want %q", updated.Name, "New Name")
		}
		if updated.Currency != "USD" {
			t.Errorf("Currency = %q, want %q", updated.Currency, "USD")
		}
		if updated.Icon != "star" {
			t.Errorf("Icon = %q, want %q", updated.Icon, "star")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.Household.Update(ctx, 99999, "Name", "", "EUR", "")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("wrong owner", func(t *testing.T) {
		otherUser, _ := svc.User.GetOrCreate(context.Background(), "other-sub", "other@example.com", "Other")
		otherCtx := service.WithUserID(context.Background(), otherUser.ID)

		_, err := svc.Household.Update(otherCtx, hh.ID, "Hack", "", "EUR", "")
		if !errors.Is(err, domain.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})
}

func TestHouseholdDelete(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)

	t.Run("success", func(t *testing.T) {
		hh := createTestHousehold(t, svc, ctx)
		if err := svc.Household.Delete(ctx, hh.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err := svc.Household.GetByID(ctx, hh.ID)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})

	t.Run("cascade deletes children", func(t *testing.T) {
		hh := createTestHousehold(t, svc, ctx)
		cat := createTestCategory(t, svc, ctx, hh.ID)

		amount, _ := domain.NewMoney("-50.00")
		svc.Transaction.Create(ctx, hh.ID, cat.ID, amount, "test", time.Now())
		svc.RecurringExpense.Create(ctx, hh.ID, cat.ID, "Rent", "", amount, domain.FrequencyMonthly, time.Now(), nil)

		if err := svc.Household.Delete(ctx, hh.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err := svc.Household.GetByID(ctx, hh.ID)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected household to be deleted")
		}
	})

	t.Run("not found", func(t *testing.T) {
		err := svc.Household.Delete(ctx, 99999)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}
