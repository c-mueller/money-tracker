package service_test

import (
	"errors"
	"testing"
	"time"

	"icekalt.dev/money-tracker/internal/domain"
	"icekalt.dev/money-tracker/internal/service"
)

func TestTransactionCreate(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)
	hh := createTestHousehold(t, svc, ctx)
	cat := createTestCategory(t, svc, ctx, hh.ID)

	t.Run("success", func(t *testing.T) {
		amount, _ := domain.NewMoney("-50.00")
		tx, err := svc.Transaction.Create(ctx, hh.ID, cat.ID, amount, "Groceries", time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tx.HouseholdID != hh.ID {
			t.Errorf("HouseholdID = %d, want %d", tx.HouseholdID, hh.ID)
		}
		if tx.Description != "Groceries" {
			t.Errorf("Description = %q, want %q", tx.Description, "Groceries")
		}
		if !tx.Amount.Equal(amount) {
			t.Errorf("Amount = %s, want %s", tx.Amount.String(), amount.String())
		}
	})

	t.Run("zero amount", func(t *testing.T) {
		_, err := svc.Transaction.Create(ctx, hh.ID, cat.ID, domain.ZeroMoney(), "test", time.Now())
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("household not found", func(t *testing.T) {
		amount, _ := domain.NewMoney("10")
		_, err := svc.Transaction.Create(ctx, 99999, cat.ID, amount, "test", time.Now())
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestTransactionListByMonth(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)
	hh := createTestHousehold(t, svc, ctx)
	cat := createTestCategory(t, svc, ctx, hh.ID)

	amount, _ := domain.NewMoney("-50.00")
	svc.Transaction.Create(ctx, hh.ID, cat.ID, amount, "Jan tx", time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC))
	svc.Transaction.Create(ctx, hh.ID, cat.ID, amount, "Feb tx", time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC))

	t.Run("filter by month", func(t *testing.T) {
		list, err := svc.Transaction.ListByMonth(ctx, hh.ID, 2026, time.January)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 transaction in Jan, got %d", len(list))
		}
		if len(list) == 1 && list[0].Description != "Jan tx" {
			t.Errorf("Description = %q, want %q", list[0].Description, "Jan tx")
		}
	})

	t.Run("empty month", func(t *testing.T) {
		list, err := svc.Transaction.ListByMonth(ctx, hh.ID, 2026, time.March)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 0 {
			t.Errorf("expected 0 transactions in March, got %d", len(list))
		}
	})
}

func TestTransactionUpdate(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)
	hh := createTestHousehold(t, svc, ctx)
	cat := createTestCategory(t, svc, ctx, hh.ID)

	amount, _ := domain.NewMoney("-50.00")
	tx, _ := svc.Transaction.Create(ctx, hh.ID, cat.ID, amount, "original", time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC))

	t.Run("success", func(t *testing.T) {
		newAmount, _ := domain.NewMoney("-75.00")
		updated, err := svc.Transaction.Update(ctx, hh.ID, tx.ID, cat.ID, newAmount, "updated", time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Description != "updated" {
			t.Errorf("Description = %q, want %q", updated.Description, "updated")
		}
		if !updated.Amount.Equal(newAmount) {
			t.Errorf("Amount = %s, want %s", updated.Amount.String(), newAmount.String())
		}
	})

	t.Run("wrong household", func(t *testing.T) {
		hh2, _ := svc.Household.Create(ctx, "Other", "", "EUR", "")
		newAmount, _ := domain.NewMoney("10")
		_, err := svc.Transaction.Update(ctx, hh2.ID, tx.ID, cat.ID, newAmount, "hack", time.Now())
		if !errors.Is(err, domain.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		newAmount, _ := domain.NewMoney("10")
		_, err := svc.Transaction.Update(ctx, hh.ID, 99999, cat.ID, newAmount, "test", time.Now())
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestTransactionDelete(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)
	hh := createTestHousehold(t, svc, ctx)
	cat := createTestCategory(t, svc, ctx, hh.ID)

	t.Run("success", func(t *testing.T) {
		amount, _ := domain.NewMoney("-50.00")
		tx, _ := svc.Transaction.Create(ctx, hh.ID, cat.ID, amount, "to delete", time.Now())

		if err := svc.Transaction.Delete(ctx, hh.ID, tx.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err := svc.Transaction.GetByID(ctx, tx.ID)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})

	t.Run("wrong household", func(t *testing.T) {
		amount, _ := domain.NewMoney("-50.00")
		tx, _ := svc.Transaction.Create(ctx, hh.ID, cat.ID, amount, "test", time.Now())
		hh2, _ := svc.Household.Create(ctx, "Other", "", "EUR", "")

		err := svc.Transaction.Delete(ctx, hh2.ID, tx.ID)
		if !errors.Is(err, domain.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		err := svc.Transaction.Delete(ctx, hh.ID, 99999)
		if err == nil {
			t.Error("expected error for non-existent transaction")
		}
	})
}

func TestTransactionNoAuth(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)
	hh := createTestHousehold(t, svc, ctx)
	cat := createTestCategory(t, svc, ctx, hh.ID)

	// Use a different user who doesn't own the household
	otherUser, _ := svc.User.GetOrCreate(ctx, "other-sub", "other@example.com", "Other")
	otherCtx := service.WithUserID(ctx, otherUser.ID)

	amount, _ := domain.NewMoney("-50.00")
	_, err := svc.Transaction.Create(otherCtx, hh.ID, cat.ID, amount, "test", time.Now())
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden for wrong owner, got %v", err)
	}
}
