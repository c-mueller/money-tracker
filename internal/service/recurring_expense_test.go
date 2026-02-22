package service_test

import (
	"errors"
	"testing"
	"time"

	"icekalt.dev/money-tracker/internal/domain"
)

func TestRecurringExpenseCreate(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)
	hh := createTestHousehold(t, svc, ctx)
	cat := createTestCategory(t, svc, ctx, hh.ID)

	t.Run("success", func(t *testing.T) {
		amount, _ := domain.NewMoney("-800.00")
		re, err := svc.RecurringExpense.Create(ctx, hh.ID, cat.ID, "Rent", "Monthly rent", amount, domain.FrequencyMonthly, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if re.Name != "Rent" {
			t.Errorf("Name = %q, want %q", re.Name, "Rent")
		}
		if re.Frequency != domain.FrequencyMonthly {
			t.Errorf("Frequency = %q, want %q", re.Frequency, domain.FrequencyMonthly)
		}
		if !re.Active {
			t.Error("expected Active = true by default")
		}
	})

	t.Run("with end date", func(t *testing.T) {
		amount, _ := domain.NewMoney("-100.00")
		end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
		re, err := svc.RecurringExpense.Create(ctx, hh.ID, cat.ID, "Subscription", "", amount, domain.FrequencyMonthly, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), &end)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if re.EndDate == nil {
			t.Error("expected EndDate to be set")
		}
	})

	t.Run("zero amount", func(t *testing.T) {
		_, err := svc.RecurringExpense.Create(ctx, hh.ID, cat.ID, "Test", "", domain.ZeroMoney(), domain.FrequencyMonthly, time.Now(), nil)
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("invalid frequency", func(t *testing.T) {
		amount, _ := domain.NewMoney("-50.00")
		_, err := svc.RecurringExpense.Create(ctx, hh.ID, cat.ID, "Test", "", amount, domain.Frequency("invalid"), time.Now(), nil)
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		amount, _ := domain.NewMoney("-50.00")
		_, err := svc.RecurringExpense.Create(ctx, hh.ID, cat.ID, "", "", amount, domain.FrequencyMonthly, time.Now(), nil)
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})

	t.Run("end before start", func(t *testing.T) {
		amount, _ := domain.NewMoney("-50.00")
		start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		_, err := svc.RecurringExpense.Create(ctx, hh.ID, cat.ID, "Test", "", amount, domain.FrequencyMonthly, start, &end)
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("expected ErrValidation, got %v", err)
		}
	})
}

func TestRecurringExpenseList(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)
	hh := createTestHousehold(t, svc, ctx)
	cat := createTestCategory(t, svc, ctx, hh.ID)

	amount, _ := domain.NewMoney("-100.00")
	svc.RecurringExpense.Create(ctx, hh.ID, cat.ID, "Expense 1", "", amount, domain.FrequencyMonthly, time.Now(), nil)
	svc.RecurringExpense.Create(ctx, hh.ID, cat.ID, "Expense 2", "", amount, domain.FrequencyWeekly, time.Now(), nil)

	t.Run("filter by household", func(t *testing.T) {
		list, err := svc.RecurringExpense.List(ctx, hh.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("expected 2 recurring expenses, got %d", len(list))
		}
	})
}

func TestRecurringExpenseUpdate(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)
	hh := createTestHousehold(t, svc, ctx)
	cat := createTestCategory(t, svc, ctx, hh.ID)

	amount, _ := domain.NewMoney("-800.00")
	re, _ := svc.RecurringExpense.Create(ctx, hh.ID, cat.ID, "Rent", "", amount, domain.FrequencyMonthly, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), nil)

	t.Run("success", func(t *testing.T) {
		newAmount, _ := domain.NewMoney("-900.00")
		updated, err := svc.RecurringExpense.Update(ctx, re.ID, cat.ID, "New Rent", "updated", newAmount, domain.FrequencyMonthly, true, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Name != "New Rent" {
			t.Errorf("Name = %q, want %q", updated.Name, "New Rent")
		}
		if !updated.Amount.Equal(newAmount) {
			t.Errorf("Amount = %s, want %s", updated.Amount.String(), newAmount.String())
		}
	})

	t.Run("change frequency", func(t *testing.T) {
		updated, err := svc.RecurringExpense.Update(ctx, re.ID, cat.ID, "Rent", "", amount, domain.FrequencyYearly, true, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Frequency != domain.FrequencyYearly {
			t.Errorf("Frequency = %q, want %q", updated.Frequency, domain.FrequencyYearly)
		}
	})

	t.Run("deactivate", func(t *testing.T) {
		updated, err := svc.RecurringExpense.Update(ctx, re.ID, cat.ID, "Rent", "", amount, domain.FrequencyYearly, false, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Active {
			t.Error("expected Active = false")
		}
	})

	t.Run("not found", func(t *testing.T) {
		newAmount, _ := domain.NewMoney("10")
		_, err := svc.RecurringExpense.Update(ctx, 99999, cat.ID, "Test", "", newAmount, domain.FrequencyMonthly, true, time.Now(), nil)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestRecurringExpenseDelete(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)
	hh := createTestHousehold(t, svc, ctx)
	cat := createTestCategory(t, svc, ctx, hh.ID)

	t.Run("success", func(t *testing.T) {
		amount, _ := domain.NewMoney("-100.00")
		re, _ := svc.RecurringExpense.Create(ctx, hh.ID, cat.ID, "ToDelete", "", amount, domain.FrequencyMonthly, time.Now(), nil)

		if err := svc.RecurringExpense.Delete(ctx, hh.ID, re.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err := svc.RecurringExpense.GetByID(ctx, re.ID)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})

	t.Run("wrong household", func(t *testing.T) {
		amount, _ := domain.NewMoney("-100.00")
		re, _ := svc.RecurringExpense.Create(ctx, hh.ID, cat.ID, "Test", "", amount, domain.FrequencyMonthly, time.Now(), nil)
		hh2, _ := svc.Household.Create(ctx, "Other", "", "EUR", "")

		err := svc.RecurringExpense.Delete(ctx, hh2.ID, re.ID)
		if !errors.Is(err, domain.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		err := svc.RecurringExpense.Delete(ctx, hh.ID, 99999)
		if err == nil {
			t.Error("expected error for non-existent recurring expense")
		}
	})
}
