package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is an HTTP client for the Money Tracker REST API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new API client.
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) do(method, path string, body any) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func decode[T any](data []byte) (T, error) {
	var result T
	if len(data) == 0 {
		return result, nil
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return result, fmt.Errorf("decoding response: %w", err)
	}
	return result, nil
}

// --- Household endpoints ---

type Household struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Currency    string `json:"currency"`
	Icon        string `json:"icon"`
	OwnerID     int    `json:"owner_id"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (c *Client) ListHouseholds() ([]Household, error) {
	data, err := c.do("GET", "/api/v1/households", nil)
	if err != nil {
		return nil, err
	}
	return decode[[]Household](data)
}

func (c *Client) CreateHousehold(req map[string]any) (*Household, error) {
	data, err := c.do("POST", "/api/v1/households", req)
	if err != nil {
		return nil, err
	}
	return decodePtr[Household](data)
}

func (c *Client) UpdateHousehold(id int, req map[string]any) (*Household, error) {
	data, err := c.do("PUT", fmt.Sprintf("/api/v1/households/%d", id), req)
	if err != nil {
		return nil, err
	}
	return decodePtr[Household](data)
}

func (c *Client) DeleteHousehold(id int) error {
	_, err := c.do("DELETE", fmt.Sprintf("/api/v1/households/%d", id), nil)
	return err
}

// --- Category endpoints ---

type Category struct {
	ID          int    `json:"id"`
	HouseholdID int    `json:"household_id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (c *Client) ListCategories(householdID int) ([]Category, error) {
	data, err := c.do("GET", fmt.Sprintf("/api/v1/households/%d/categories", householdID), nil)
	if err != nil {
		return nil, err
	}
	return decode[[]Category](data)
}

func (c *Client) CreateCategory(householdID int, req map[string]any) (*Category, error) {
	data, err := c.do("POST", fmt.Sprintf("/api/v1/households/%d/categories", householdID), req)
	if err != nil {
		return nil, err
	}
	return decodePtr[Category](data)
}

func (c *Client) UpdateCategory(householdID, categoryID int, req map[string]any) (*Category, error) {
	data, err := c.do("PUT", fmt.Sprintf("/api/v1/households/%d/categories/%d", householdID, categoryID), req)
	if err != nil {
		return nil, err
	}
	return decodePtr[Category](data)
}

func (c *Client) DeleteCategory(householdID, categoryID int) error {
	_, err := c.do("DELETE", fmt.Sprintf("/api/v1/households/%d/categories/%d", householdID, categoryID), nil)
	return err
}

// --- Transaction endpoints ---

type Transaction struct {
	ID          int    `json:"id"`
	HouseholdID int    `json:"household_id"`
	CategoryID  int    `json:"category_id"`
	Amount      string `json:"amount"`
	Description string `json:"description"`
	Date        string `json:"date"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func (c *Client) ListTransactions(householdID int, month string) ([]Transaction, error) {
	path := fmt.Sprintf("/api/v1/households/%d/transactions", householdID)
	if month != "" {
		path += "?month=" + month
	}
	data, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	return decode[[]Transaction](data)
}

func (c *Client) CreateTransaction(householdID int, req map[string]any) (*Transaction, error) {
	data, err := c.do("POST", fmt.Sprintf("/api/v1/households/%d/transactions", householdID), req)
	if err != nil {
		return nil, err
	}
	return decodePtr[Transaction](data)
}

func (c *Client) UpdateTransaction(householdID, transactionID int, req map[string]any) (*Transaction, error) {
	data, err := c.do("PUT", fmt.Sprintf("/api/v1/households/%d/transactions/%d", householdID, transactionID), req)
	if err != nil {
		return nil, err
	}
	return decodePtr[Transaction](data)
}

func (c *Client) DeleteTransaction(householdID, transactionID int) error {
	_, err := c.do("DELETE", fmt.Sprintf("/api/v1/households/%d/transactions/%d", householdID, transactionID), nil)
	return err
}

// --- Recurring Expense endpoints ---

type RecurringExpense struct {
	ID          int     `json:"id"`
	HouseholdID int     `json:"household_id"`
	CategoryID  int     `json:"category_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Amount      string  `json:"amount"`
	Frequency   string  `json:"frequency"`
	Active      bool    `json:"active"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func (c *Client) ListRecurringExpenses(householdID int) ([]RecurringExpense, error) {
	data, err := c.do("GET", fmt.Sprintf("/api/v1/households/%d/recurring-expenses", householdID), nil)
	if err != nil {
		return nil, err
	}
	return decode[[]RecurringExpense](data)
}

func (c *Client) CreateRecurringExpense(householdID int, req map[string]any) (*RecurringExpense, error) {
	data, err := c.do("POST", fmt.Sprintf("/api/v1/households/%d/recurring-expenses", householdID), req)
	if err != nil {
		return nil, err
	}
	return decodePtr[RecurringExpense](data)
}

func (c *Client) UpdateRecurringExpense(householdID, recurringID int, req map[string]any) (*RecurringExpense, error) {
	data, err := c.do("PUT", fmt.Sprintf("/api/v1/households/%d/recurring-expenses/%d", householdID, recurringID), req)
	if err != nil {
		return nil, err
	}
	return decodePtr[RecurringExpense](data)
}

func (c *Client) DeleteRecurringExpense(householdID, recurringID int) error {
	_, err := c.do("DELETE", fmt.Sprintf("/api/v1/households/%d/recurring-expenses/%d", householdID, recurringID), nil)
	return err
}

// --- Schedule Override endpoints ---

type ScheduleOverride struct {
	ID                 int    `json:"id"`
	RecurringExpenseID int    `json:"recurring_expense_id"`
	EffectiveDate      string `json:"effective_date"`
	Amount             string `json:"amount"`
	Frequency          string `json:"frequency"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

func (c *Client) ListScheduleOverrides(householdID, recurringID int) ([]ScheduleOverride, error) {
	data, err := c.do("GET", fmt.Sprintf("/api/v1/households/%d/recurring-expenses/%d/overrides", householdID, recurringID), nil)
	if err != nil {
		return nil, err
	}
	return decode[[]ScheduleOverride](data)
}

func (c *Client) CreateScheduleOverride(householdID, recurringID int, req map[string]any) (*ScheduleOverride, error) {
	data, err := c.do("POST", fmt.Sprintf("/api/v1/households/%d/recurring-expenses/%d/overrides", householdID, recurringID), req)
	if err != nil {
		return nil, err
	}
	return decodePtr[ScheduleOverride](data)
}

func (c *Client) UpdateScheduleOverride(householdID, recurringID, overrideID int, req map[string]any) (*ScheduleOverride, error) {
	data, err := c.do("PUT", fmt.Sprintf("/api/v1/households/%d/recurring-expenses/%d/overrides/%d", householdID, recurringID, overrideID), req)
	if err != nil {
		return nil, err
	}
	return decodePtr[ScheduleOverride](data)
}

func (c *Client) DeleteScheduleOverride(householdID, recurringID, overrideID int) error {
	_, err := c.do("DELETE", fmt.Sprintf("/api/v1/households/%d/recurring-expenses/%d/overrides/%d", householdID, recurringID, overrideID), nil)
	return err
}

// --- Summary endpoint ---

type Summary struct {
	Month             string            `json:"month"`
	HouseholdID       int               `json:"household_id"`
	TotalIncome       string            `json:"total_income"`
	TotalExpenses     string            `json:"total_expenses"`
	RecurringTotal    string            `json:"recurring_total"`
	RecurringIncome   string            `json:"recurring_income"`
	RecurringExpenses string            `json:"recurring_expenses"`
	OneTimeTotal      string            `json:"one_time_total"`
	OneTimeIncome     string            `json:"one_time_income"`
	OneTimeExpenses   string            `json:"one_time_expenses"`
	MonthlyTotal      string            `json:"monthly_total"`
	CategoryBreakdown []CategorySummary `json:"category_breakdown"`
}

type CategorySummary struct {
	CategoryID   int    `json:"category_id"`
	CategoryName string `json:"category_name"`
	Recurring    string `json:"recurring"`
	OneTime      string `json:"one_time"`
	Total        string `json:"total"`
}

func (c *Client) GetSummary(householdID int, month string) (*Summary, error) {
	path := fmt.Sprintf("/api/v1/households/%d/summary", householdID)
	if month != "" {
		path += "?month=" + month
	}
	data, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	return decodePtr[Summary](data)
}

func decodePtr[T any](data []byte) (*T, error) {
	var result T
	if len(data) == 0 {
		return &result, nil
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}
