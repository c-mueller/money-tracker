package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"icekalt.dev/money-tracker/internal/domain"
)

func (s *Server) handleListTransactions(c echo.Context) error {
	householdID, err := parseID(c, "id")
	if err != nil {
		return respondError(c, err)
	}

	year, month, err := parseMonth(c)
	if err != nil {
		return respondError(c, err)
	}

	transactions, err := s.services.Transaction.ListByMonth(c.Request().Context(), householdID, year, month)
	if err != nil {
		return respondError(c, err)
	}

	resp := make([]TransactionResponse, len(transactions))
	for i, tx := range transactions {
		resp[i] = toTransactionResponse(tx)
	}
	return c.JSON(http.StatusOK, resp)
}

func (s *Server) handleCreateTransaction(c echo.Context) error {
	householdID, err := parseID(c, "id")
	if err != nil {
		return respondError(c, err)
	}

	var req CreateTransactionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	amount, err := domain.NewMoney(req.Amount)
	if err != nil {
		return respondError(c, fmt.Errorf("%w: invalid amount", domain.ErrValidation))
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return respondError(c, fmt.Errorf("%w: invalid date format, expected YYYY-MM-DD", domain.ErrValidation))
	}

	tx, err := s.services.Transaction.Create(c.Request().Context(), householdID, req.CategoryID, amount, req.Description, date)
	if err != nil {
		return respondError(c, err)
	}

	return c.JSON(http.StatusCreated, toTransactionResponse(tx))
}

func (s *Server) handleDeleteTransaction(c echo.Context) error {
	householdID, err := parseID(c, "id")
	if err != nil {
		return respondError(c, err)
	}

	txID, err := parseID(c, "transactionId")
	if err != nil {
		return respondError(c, err)
	}

	if err := s.services.Transaction.Delete(c.Request().Context(), householdID, txID); err != nil {
		return respondError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func toTransactionResponse(tx *domain.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:          tx.ID,
		HouseholdID: tx.HouseholdID,
		CategoryID:  tx.CategoryID,
		Amount:      tx.Amount.String(),
		Description: tx.Description,
		Date:        tx.Date.Format("2006-01-02"),
		CreatedAt:   tx.CreatedAt,
		UpdatedAt:   tx.UpdatedAt,
	}
}
