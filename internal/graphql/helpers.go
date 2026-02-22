package graphql

import (
	"icekalt.dev/money-tracker/internal/domain"
	"icekalt.dev/money-tracker/internal/graphql/model"
)

func toGQLHousehold(h *domain.Household) *model.Household {
	return &model.Household{
		ID:          h.ID,
		Name:        h.Name,
		Description: h.Description,
		Currency:    h.Currency,
		Icon:        h.Icon,
		OwnerID:     h.OwnerID,
		CreatedAt:   h.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   h.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toGQLCategory(c *domain.Category) *model.Category {
	return &model.Category{
		ID:          c.ID,
		HouseholdID: c.HouseholdID,
		Name:        c.Name,
		Icon:        c.Icon,
		CreatedAt:   c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toGQLTransaction(tx *domain.Transaction) *model.Transaction {
	return &model.Transaction{
		ID:          tx.ID,
		HouseholdID: tx.HouseholdID,
		CategoryID:  tx.CategoryID,
		Amount:      tx.Amount.String(),
		Description: tx.Description,
		Date:        tx.Date.Format("2006-01-02"),
		CreatedAt:   tx.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   tx.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toGQLRecurringExpense(re *domain.RecurringExpense) *model.RecurringExpense {
	resp := &model.RecurringExpense{
		ID:          re.ID,
		HouseholdID: re.HouseholdID,
		CategoryID:  re.CategoryID,
		Name:        re.Name,
		Description: re.Description,
		Amount:      re.Amount.String(),
		Frequency:   string(re.Frequency),
		Active:      re.Active,
		StartDate:   re.StartDate.Format("2006-01-02"),
		CreatedAt:   re.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   re.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if re.EndDate != nil {
		s := re.EndDate.Format("2006-01-02")
		resp.EndDate = &s
	}
	return resp
}

func toGQLMonthlySummary(s *domain.MonthlySummary) *model.MonthlySummary {
	breakdown := make([]model.CategorySummary, len(s.CategoryBreakdown))
	for i, cb := range s.CategoryBreakdown {
		breakdown[i] = model.CategorySummary{
			CategoryID:   cb.CategoryID,
			CategoryName: cb.CategoryName,
			Recurring:    cb.Recurring.String(),
			OneTime:      cb.OneTime.String(),
			Total:        cb.Total.String(),
		}
	}
	return &model.MonthlySummary{
		Month:             s.Month,
		HouseholdID:       s.HouseholdID,
		TotalIncome:       s.TotalIncome.String(),
		TotalExpenses:     s.TotalExpenses.String(),
		RecurringTotal:    s.RecurringTotal.String(),
		RecurringIncome:   s.RecurringIncome.String(),
		RecurringExpenses: s.RecurringExpenses.String(),
		OneTimeTotal:      s.OneTimeTotal.String(),
		OneTimeIncome:     s.OneTimeIncome.String(),
		OneTimeExpenses:   s.OneTimeExpenses.String(),
		MonthlyTotal:      s.MonthlyTotal.String(),
		CategoryBreakdown: breakdown,
	}
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
