package service_test

import (
	"testing"
	"time"

	"icekalt.dev/money-tracker/internal/domain"
)

func TestGetMonthlySummary(t *testing.T) {
	svc := setupTestServices(t)
	ctx, _ := createTestUser(t, svc)
	hh := createTestHousehold(t, svc, ctx)
	cat := createTestCategory(t, svc, ctx, hh.ID)

	t.Run("empty month", func(t *testing.T) {
		summary, err := svc.Summary.GetMonthlySummary(ctx, hh.ID, 2026, time.March)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if summary.Month != "2026-03" {
			t.Errorf("Month = %q, want %q", summary.Month, "2026-03")
		}
		if !summary.TotalIncome.IsZero() {
			t.Errorf("TotalIncome = %s, want 0", summary.TotalIncome.String())
		}
		if !summary.TotalExpenses.IsZero() {
			t.Errorf("TotalExpenses = %s, want 0", summary.TotalExpenses.String())
		}
	})

	t.Run("only one-time transactions", func(t *testing.T) {
		expense, _ := domain.NewMoney("-50.00")
		income, _ := domain.NewMoney("1000.00")
		svc.Transaction.Create(ctx, hh.ID, cat.ID, expense, "Groceries", time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC))
		svc.Transaction.Create(ctx, hh.ID, cat.ID, income, "Salary", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))

		summary, err := svc.Summary.GetMonthlySummary(ctx, hh.ID, 2026, time.January)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantIncome, _ := domain.NewMoney("1000")
		wantExpenses, _ := domain.NewMoney("-50")
		if !summary.TotalIncome.Equal(wantIncome) {
			t.Errorf("TotalIncome = %s, want %s", summary.TotalIncome.String(), wantIncome.String())
		}
		if !summary.TotalExpenses.Equal(wantExpenses) {
			t.Errorf("TotalExpenses = %s, want %s", summary.TotalExpenses.String(), wantExpenses.String())
		}
		if summary.RecurringTotal.IsZero() == false {
			t.Errorf("RecurringTotal = %s, want 0", summary.RecurringTotal.String())
		}
	})

	t.Run("only recurring expenses", func(t *testing.T) {
		// Use a fresh household to avoid interference
		hh2, _ := svc.Household.Create(ctx, "Summary Test 2", "", "EUR", "")
		cat2, _ := svc.Category.Create(ctx, hh2.ID, "Bills", "")

		recurAmount, _ := domain.NewMoney("-800.00")
		svc.RecurringExpense.Create(ctx, hh2.ID, cat2.ID, "Rent", "", recurAmount, domain.FrequencyMonthly, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), nil)

		summary, err := svc.Summary.GetMonthlySummary(ctx, hh2.ID, 2026, time.January)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantRecurring, _ := domain.NewMoney("-800")
		if !summary.RecurringTotal.Equal(wantRecurring) {
			t.Errorf("RecurringTotal = %s, want %s", summary.RecurringTotal.String(), wantRecurring.String())
		}
		if !summary.OneTimeTotal.IsZero() {
			t.Errorf("OneTimeTotal = %s, want 0", summary.OneTimeTotal.String())
		}
	})

	t.Run("mixed recurring and one-time", func(t *testing.T) {
		hh3, _ := svc.Household.Create(ctx, "Summary Test 3", "", "EUR", "")
		cat3, _ := svc.Category.Create(ctx, hh3.ID, "Mixed", "")

		recurAmount, _ := domain.NewMoney("-800.00")
		svc.RecurringExpense.Create(ctx, hh3.ID, cat3.ID, "Rent", "", recurAmount, domain.FrequencyMonthly, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), nil)

		oneTime, _ := domain.NewMoney("-50.00")
		svc.Transaction.Create(ctx, hh3.ID, cat3.ID, oneTime, "Extra", time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC))

		summary, err := svc.Summary.GetMonthlySummary(ctx, hh3.ID, 2026, time.January)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantRecurring, _ := domain.NewMoney("-800")
		wantOneTime, _ := domain.NewMoney("-50")
		if !summary.RecurringTotal.Equal(wantRecurring) {
			t.Errorf("RecurringTotal = %s, want %s", summary.RecurringTotal.String(), wantRecurring.String())
		}
		if !summary.OneTimeTotal.Equal(wantOneTime) {
			t.Errorf("OneTimeTotal = %s, want %s", summary.OneTimeTotal.String(), wantOneTime.String())
		}

		if len(summary.CategoryBreakdown) == 0 {
			t.Error("expected category breakdown to have entries")
		}
	})

	t.Run("recurring income and expenses split", func(t *testing.T) {
		hh5, _ := svc.Household.Create(ctx, "Summary Test 5", "", "EUR", "")
		cat5, _ := svc.Category.Create(ctx, hh5.ID, "Split", "")

		recurExpense, _ := domain.NewMoney("-800.00")
		recurIncome, _ := domain.NewMoney("3000.00")
		svc.RecurringExpense.Create(ctx, hh5.ID, cat5.ID, "Rent", "", recurExpense, domain.FrequencyMonthly, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), nil)
		svc.RecurringExpense.Create(ctx, hh5.ID, cat5.ID, "Salary", "", recurIncome, domain.FrequencyMonthly, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), nil)

		oneTimeExpense, _ := domain.NewMoney("-50.00")
		oneTimeIncome, _ := domain.NewMoney("200.00")
		svc.Transaction.Create(ctx, hh5.ID, cat5.ID, oneTimeExpense, "Groceries", time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC))
		svc.Transaction.Create(ctx, hh5.ID, cat5.ID, oneTimeIncome, "Refund", time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC))

		summary, err := svc.Summary.GetMonthlySummary(ctx, hh5.ID, 2026, time.January)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		wantRecIncome, _ := domain.NewMoney("3000")
		wantRecExpenses, _ := domain.NewMoney("-800")
		wantOneIncome, _ := domain.NewMoney("200")
		wantOneExpenses, _ := domain.NewMoney("-50")
		wantMonthlyTotal, _ := domain.NewMoney("2350") // (3000-800) + (200-50)

		if !summary.RecurringIncome.Equal(wantRecIncome) {
			t.Errorf("RecurringIncome = %s, want %s", summary.RecurringIncome.String(), wantRecIncome.String())
		}
		if !summary.RecurringExpenses.Equal(wantRecExpenses) {
			t.Errorf("RecurringExpenses = %s, want %s", summary.RecurringExpenses.String(), wantRecExpenses.String())
		}
		if !summary.OneTimeIncome.Equal(wantOneIncome) {
			t.Errorf("OneTimeIncome = %s, want %s", summary.OneTimeIncome.String(), wantOneIncome.String())
		}
		if !summary.OneTimeExpenses.Equal(wantOneExpenses) {
			t.Errorf("OneTimeExpenses = %s, want %s", summary.OneTimeExpenses.String(), wantOneExpenses.String())
		}
		if !summary.MonthlyTotal.Equal(wantMonthlyTotal) {
			t.Errorf("MonthlyTotal = %s, want %s", summary.MonthlyTotal.String(), wantMonthlyTotal.String())
		}
	})

	t.Run("recurring with schedule override", func(t *testing.T) {
		hh6, _ := svc.Household.Create(ctx, "Summary Test 6", "", "EUR", "")
		cat6, _ := svc.Category.Create(ctx, hh6.ID, "Override", "")

		recurAmount, _ := domain.NewMoney("-800.00")
		re, _ := svc.RecurringExpense.Create(ctx, hh6.ID, cat6.ID, "Rent", "", recurAmount, domain.FrequencyMonthly, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), nil)

		// Add override: from April, rent increases to -900
		overrideAmount, _ := domain.NewMoney("-900.00")
		svc.RecurringExpense.CreateOverride(ctx, re.ID, time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC), overrideAmount, domain.FrequencyMonthly)

		// January: base amount applies
		summaryJan, err := svc.Summary.GetMonthlySummary(ctx, hh6.ID, 2026, time.January)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		wantJan, _ := domain.NewMoney("-800")
		if !summaryJan.RecurringTotal.Equal(wantJan) {
			t.Errorf("Jan RecurringTotal = %s, want %s", summaryJan.RecurringTotal.String(), wantJan.String())
		}

		// May: override amount applies
		summaryMay, err := svc.Summary.GetMonthlySummary(ctx, hh6.ID, 2026, time.May)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		wantMay, _ := domain.NewMoney("-900")
		if !summaryMay.RecurringTotal.Equal(wantMay) {
			t.Errorf("May RecurringTotal = %s, want %s", summaryMay.RecurringTotal.String(), wantMay.String())
		}
	})

	t.Run("recurring not started yet excluded", func(t *testing.T) {
		hhF, _ := svc.Household.Create(ctx, "Summary Test Future", "", "EUR", "")
		catF, _ := svc.Category.Create(ctx, hhF.ID, "Future", "")

		recurAmount, _ := domain.NewMoney("-500.00")
		// Recurring starts in March 2026
		svc.RecurringExpense.Create(ctx, hhF.ID, catF.ID, "Future Rent", "", recurAmount, domain.FrequencyMonthly, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), nil)

		// January summary should not include it
		summary, err := svc.Summary.GetMonthlySummary(ctx, hhF.ID, 2026, time.January)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !summary.RecurringTotal.IsZero() {
			t.Errorf("RecurringTotal = %s, want 0 (recurring not yet started)", summary.RecurringTotal.String())
		}

		// March summary should include it
		summaryMar, err := svc.Summary.GetMonthlySummary(ctx, hhF.ID, 2026, time.March)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		wantMar, _ := domain.NewMoney("-500")
		if !summaryMar.RecurringTotal.Equal(wantMar) {
			t.Errorf("RecurringTotal = %s, want %s", summaryMar.RecurringTotal.String(), wantMar.String())
		}
	})

	t.Run("recurring already ended excluded", func(t *testing.T) {
		hhE, _ := svc.Household.Create(ctx, "Summary Test Ended", "", "EUR", "")
		catE, _ := svc.Category.Create(ctx, hhE.ID, "Ended", "")

		recurAmount, _ := domain.NewMoney("-300.00")
		endDate := time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
		svc.RecurringExpense.Create(ctx, hhE.ID, catE.ID, "Old Sub", "", recurAmount, domain.FrequencyMonthly, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), &endDate)

		// February should include it
		summaryFeb, err := svc.Summary.GetMonthlySummary(ctx, hhE.ID, 2026, time.February)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		wantFeb, _ := domain.NewMoney("-300")
		if !summaryFeb.RecurringTotal.Equal(wantFeb) {
			t.Errorf("RecurringTotal = %s, want %s", summaryFeb.RecurringTotal.String(), wantFeb.String())
		}

		// March should exclude it
		summaryMar, err := svc.Summary.GetMonthlySummary(ctx, hhE.ID, 2026, time.March)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !summaryMar.RecurringTotal.IsZero() {
			t.Errorf("RecurringTotal = %s, want 0 (recurring already ended)", summaryMar.RecurringTotal.String())
		}
	})

	t.Run("category breakdown", func(t *testing.T) {
		hh4, _ := svc.Household.Create(ctx, "Summary Test 4", "", "EUR", "")
		catA, _ := svc.Category.Create(ctx, hh4.ID, "Food", "")
		catB, _ := svc.Category.Create(ctx, hh4.ID, "Transport", "")

		amountA, _ := domain.NewMoney("-100.00")
		amountB, _ := domain.NewMoney("-200.00")
		svc.Transaction.Create(ctx, hh4.ID, catA.ID, amountA, "Groceries", time.Date(2026, 2, 5, 0, 0, 0, 0, time.UTC))
		svc.Transaction.Create(ctx, hh4.ID, catB.ID, amountB, "Gas", time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC))

		summary, err := svc.Summary.GetMonthlySummary(ctx, hh4.ID, 2026, time.February)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(summary.CategoryBreakdown) != 2 {
			t.Errorf("expected 2 category breakdowns, got %d", len(summary.CategoryBreakdown))
		}
	})
}
