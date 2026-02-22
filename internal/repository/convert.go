package repository

import (
	"github.com/shopspring/decimal"
	"icekalt.dev/money-tracker/ent"
	"icekalt.dev/money-tracker/internal/domain"
)

func userToDomain(u *ent.User) *domain.User {
	return &domain.User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Subject:   u.Subject,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func householdToDomain(h *ent.Household) *domain.Household {
	hh := &domain.Household{
		ID:        h.ID,
		Name:      h.Name,
		Currency:  h.Currency,
		CreatedAt: h.CreatedAt,
		UpdatedAt: h.UpdatedAt,
	}
	if owner := h.Edges.Owner; owner != nil {
		hh.OwnerID = owner.ID
	}
	return hh
}

func categoryToDomain(c *ent.Category) *domain.Category {
	cat := &domain.Category{
		ID:        c.ID,
		Name:      c.Name,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
	if hh := c.Edges.Household; hh != nil {
		cat.HouseholdID = hh.ID
	}
	return cat
}

func transactionToDomain(t *ent.Transaction) *domain.Transaction {
	amount, _ := decimal.NewFromString(t.Amount)
	tx := &domain.Transaction{
		ID:          t.ID,
		Amount:      amount,
		Description: t.Description,
		Date:        t.Date,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
	if hh := t.Edges.Household; hh != nil {
		tx.HouseholdID = hh.ID
	}
	if cat := t.Edges.Category; cat != nil {
		tx.CategoryID = cat.ID
	}
	return tx
}

func recurringExpenseToDomain(r *ent.RecurringExpense) *domain.RecurringExpense {
	amount, _ := decimal.NewFromString(r.Amount)
	re := &domain.RecurringExpense{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Amount:      amount,
		Frequency: domain.Frequency(r.Frequency),
		Active:    r.Active,
		StartDate: r.StartDate,
		EndDate:   r.EndDate,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
	if hh := r.Edges.Household; hh != nil {
		re.HouseholdID = hh.ID
	}
	if cat := r.Edges.Category; cat != nil {
		re.CategoryID = cat.ID
	}
	return re
}

func apiTokenToDomain(t *ent.APIToken) *domain.APIToken {
	tok := &domain.APIToken{
		ID:        t.ID,
		Name:      t.Name,
		TokenHash: t.TokenHash,
		ExpiresAt: t.ExpiresAt,
		LastUsed:  t.LastUsed,
		CreatedAt: t.CreatedAt,
	}
	if u := t.Edges.User; u != nil {
		tok.UserID = u.ID
	}
	return tok
}
