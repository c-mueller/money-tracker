package service

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"icekalt.dev/money-tracker/internal/domain"
)

type SummaryService struct {
	txRepo        domain.TransactionRepo
	recurringRepo domain.RecurringExpenseRepo
	categoryRepo  domain.CategoryRepo
	household     *HouseholdService
}

func NewSummaryService(
	txRepo domain.TransactionRepo,
	recurringRepo domain.RecurringExpenseRepo,
	categoryRepo domain.CategoryRepo,
	household *HouseholdService,
) *SummaryService {
	return &SummaryService{
		txRepo:        txRepo,
		recurringRepo: recurringRepo,
		categoryRepo:  categoryRepo,
		household:     household,
	}
}

func (s *SummaryService) GetMonthlySummary(ctx context.Context, householdID int, year int, month time.Month) (*domain.MonthlySummary, error) {
	if _, err := s.household.GetByID(ctx, householdID); err != nil {
		return nil, err
	}

	refMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

	// Get active recurring expenses
	recurring, err := s.recurringRepo.ListActiveByHousehold(ctx, householdID)
	if err != nil {
		return nil, err
	}

	// Get one-time transactions for the month
	transactions, err := s.txRepo.ListByHouseholdAndMonth(ctx, householdID, year, month)
	if err != nil {
		return nil, err
	}

	// Get all categories for the household
	categories, err := s.categoryRepo.ListByHousehold(ctx, householdID)
	if err != nil {
		return nil, err
	}

	catMap := make(map[int]string)
	for _, c := range categories {
		catMap[c.ID] = c.Name
	}

	// Build category breakdown
	catRecurring := make(map[int]decimal.Decimal)
	catOneTime := make(map[int]decimal.Decimal)

	totalRecurring := decimal.Zero
	for _, re := range recurring {
		monthly, err := domain.NormalizeToMonthly(re.Amount, re.Frequency, refMonth)
		if err != nil {
			continue
		}
		catRecurring[re.CategoryID] = catRecurring[re.CategoryID].Add(monthly)
		totalRecurring = totalRecurring.Add(monthly)
	}

	totalOneTime := decimal.Zero
	totalIncome := decimal.Zero
	totalExpenses := decimal.Zero

	for _, tx := range transactions {
		catOneTime[tx.CategoryID] = catOneTime[tx.CategoryID].Add(tx.Amount)
		totalOneTime = totalOneTime.Add(tx.Amount)

		if tx.Amount.IsPositive() {
			totalIncome = totalIncome.Add(tx.Amount)
		} else {
			totalExpenses = totalExpenses.Add(tx.Amount)
		}
	}

	// Build breakdown
	catIDs := make(map[int]bool)
	for id := range catRecurring {
		catIDs[id] = true
	}
	for id := range catOneTime {
		catIDs[id] = true
	}

	breakdown := make([]domain.CategorySummary, 0, len(catIDs))
	for id := range catIDs {
		rec := catRecurring[id]
		one := catOneTime[id]
		breakdown = append(breakdown, domain.CategorySummary{
			CategoryID:   id,
			CategoryName: catMap[id],
			Recurring:    rec,
			OneTime:      one,
			Total:        rec.Add(one),
		})
	}

	return &domain.MonthlySummary{
		Month:             refMonth.Format("2006-01"),
		HouseholdID:       householdID,
		TotalIncome:       totalIncome,
		TotalExpenses:     totalExpenses,
		RecurringTotal:    totalRecurring,
		OneTimeTotal:      totalOneTime,
		CategoryBreakdown: breakdown,
	}, nil
}
