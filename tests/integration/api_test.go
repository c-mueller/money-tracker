//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestFullFlow(t *testing.T) {
	env := setupTestEnv(t)

	// 1. Health check (no auth needed)
	resp := doRequest(t, env, "GET", "/api/v1/health", "")
	assertStatus(t, resp, http.StatusOK)

	// 2. Create household
	resp = doRequest(t, env, "POST", "/api/v1/households", `{"name":"Test Haushalt","currency":"EUR"}`)
	assertStatus(t, resp, http.StatusCreated)
	var household map[string]interface{}
	decodeJSON(t, resp, &household)
	householdID := int(household["id"].(float64))
	if household["name"] != "Test Haushalt" {
		t.Errorf("expected name 'Test Haushalt', got %v", household["name"])
	}

	// 3. List households
	resp = doRequest(t, env, "GET", "/api/v1/households", "")
	assertStatus(t, resp, http.StatusOK)
	var households []map[string]interface{}
	decodeJSON(t, resp, &households)
	if len(households) != 1 {
		t.Errorf("expected 1 household, got %d", len(households))
	}

	// 4. Create category
	resp = doRequest(t, env, "POST", "/api/v1/households/"+itoa(householdID)+"/categories", `{"name":"Miete"}`)
	assertStatus(t, resp, http.StatusCreated)
	var category map[string]interface{}
	decodeJSON(t, resp, &category)
	categoryID := int(category["id"].(float64))

	// 5. Create recurring expense
	resp = doRequest(t, env, "POST", "/api/v1/households/"+itoa(householdID)+"/recurring-expenses",
		`{"category_id":`+itoa(categoryID)+`,"name":"Kaltmiete","amount":"800.00","frequency":"monthly","start_date":"2026-01-01"}`)
	assertStatus(t, resp, http.StatusCreated)

	// 6. Create transaction
	resp = doRequest(t, env, "POST", "/api/v1/households/"+itoa(householdID)+"/transactions",
		`{"category_id":`+itoa(categoryID)+`,"amount":"-50.00","description":"Nebenkosten","date":"2026-01-15"}`)
	assertStatus(t, resp, http.StatusCreated)

	// 7. Get summary
	resp = doRequest(t, env, "GET", "/api/v1/households/"+itoa(householdID)+"/summary?month=2026-01", "")
	assertStatus(t, resp, http.StatusOK)
	var summary map[string]interface{}
	decodeJSON(t, resp, &summary)
	if summary["month"] != "2026-01" {
		t.Errorf("expected month '2026-01', got %v", summary["month"])
	}
	if summary["recurring_total"] != "800" {
		t.Errorf("expected recurring_total '800', got %v", summary["recurring_total"])
	}

	// 8. List transactions
	resp = doRequest(t, env, "GET", "/api/v1/households/"+itoa(householdID)+"/transactions?month=2026-01", "")
	assertStatus(t, resp, http.StatusOK)
	var transactions []map[string]interface{}
	decodeJSON(t, resp, &transactions)
	if len(transactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(transactions))
	}

	// 9. Delete household
	resp = doRequest(t, env, "DELETE", "/api/v1/households/"+itoa(householdID), "")
	assertStatus(t, resp, http.StatusNoContent)

	// 10. Verify deleted
	resp = doRequest(t, env, "GET", "/api/v1/households", "")
	assertStatus(t, resp, http.StatusOK)
	decodeJSON(t, resp, &households)
	if len(households) != 0 {
		t.Errorf("expected 0 households after delete, got %d", len(households))
	}
}

func TestValidation(t *testing.T) {
	env := setupTestEnv(t)

	// Empty household name
	resp := doRequest(t, env, "POST", "/api/v1/households", `{"name":"","currency":"EUR"}`)
	assertStatus(t, resp, http.StatusUnprocessableEntity)

	// Invalid currency
	resp = doRequest(t, env, "POST", "/api/v1/households", `{"name":"Test","currency":"invalid"}`)
	assertStatus(t, resp, http.StatusUnprocessableEntity)

	// Invalid ID
	resp = doRequest(t, env, "GET", "/api/v1/households/abc/categories", "")
	assertStatus(t, resp, http.StatusBadRequest)
}

func TestHouseholdCRUD(t *testing.T) {
	env := setupTestEnv(t)

	// Create
	resp := doRequest(t, env, "POST", "/api/v1/households", `{"name":"CRUD Test","currency":"USD"}`)
	assertStatus(t, resp, http.StatusCreated)
	var hh map[string]interface{}
	decodeJSON(t, resp, &hh)
	hhID := int(hh["id"].(float64))

	// Update
	resp = doRequest(t, env, "PUT", "/api/v1/households/"+itoa(hhID), `{"name":"Updated","currency":"EUR"}`)
	assertStatus(t, resp, http.StatusOK)
	decodeJSON(t, resp, &hh)
	if hh["name"] != "Updated" {
		t.Errorf("expected name 'Updated', got %v", hh["name"])
	}
	if hh["currency"] != "EUR" {
		t.Errorf("expected currency 'EUR', got %v", hh["currency"])
	}

	// Delete
	resp = doRequest(t, env, "DELETE", "/api/v1/households/"+itoa(hhID), "")
	assertStatus(t, resp, http.StatusNoContent)

	// Verify deleted
	resp = doRequest(t, env, "GET", "/api/v1/households", "")
	assertStatus(t, resp, http.StatusOK)
	var list []map[string]interface{}
	decodeJSON(t, resp, &list)
	if len(list) != 0 {
		t.Errorf("expected 0 households, got %d", len(list))
	}
}

func TestCategoryCRUD(t *testing.T) {
	env := setupTestEnv(t)

	// Setup household
	resp := doRequest(t, env, "POST", "/api/v1/households", `{"name":"Cat Test","currency":"EUR"}`)
	assertStatus(t, resp, http.StatusCreated)
	var hh map[string]interface{}
	decodeJSON(t, resp, &hh)
	hhID := itoa(int(hh["id"].(float64)))

	// Create category
	resp = doRequest(t, env, "POST", "/api/v1/households/"+hhID+"/categories", `{"name":"Food"}`)
	assertStatus(t, resp, http.StatusCreated)
	var cat map[string]interface{}
	decodeJSON(t, resp, &cat)
	catID := itoa(int(cat["id"].(float64)))
	if cat["name"] != "Food" {
		t.Errorf("expected name 'Food', got %v", cat["name"])
	}

	// List categories
	resp = doRequest(t, env, "GET", "/api/v1/households/"+hhID+"/categories", "")
	assertStatus(t, resp, http.StatusOK)
	var cats []map[string]interface{}
	decodeJSON(t, resp, &cats)
	if len(cats) != 1 {
		t.Errorf("expected 1 category, got %d", len(cats))
	}

	// Update category
	resp = doRequest(t, env, "PUT", "/api/v1/households/"+hhID+"/categories/"+catID, `{"name":"Groceries"}`)
	assertStatus(t, resp, http.StatusOK)
	decodeJSON(t, resp, &cat)
	if cat["name"] != "Groceries" {
		t.Errorf("expected name 'Groceries', got %v", cat["name"])
	}

	// Delete category
	resp = doRequest(t, env, "DELETE", "/api/v1/households/"+hhID+"/categories/"+catID, "")
	assertStatus(t, resp, http.StatusNoContent)

	// Verify deleted
	resp = doRequest(t, env, "GET", "/api/v1/households/"+hhID+"/categories", "")
	assertStatus(t, resp, http.StatusOK)
	decodeJSON(t, resp, &cats)
	if len(cats) != 0 {
		t.Errorf("expected 0 categories, got %d", len(cats))
	}
}

func TestTransactionCRUD(t *testing.T) {
	env := setupTestEnv(t)

	// Setup
	resp := doRequest(t, env, "POST", "/api/v1/households", `{"name":"Tx Test","currency":"EUR"}`)
	assertStatus(t, resp, http.StatusCreated)
	var hh map[string]interface{}
	decodeJSON(t, resp, &hh)
	hhID := itoa(int(hh["id"].(float64)))

	resp = doRequest(t, env, "POST", "/api/v1/households/"+hhID+"/categories", `{"name":"Bills"}`)
	assertStatus(t, resp, http.StatusCreated)
	var cat map[string]interface{}
	decodeJSON(t, resp, &cat)
	catID := itoa(int(cat["id"].(float64)))

	// Create transaction
	resp = doRequest(t, env, "POST", "/api/v1/households/"+hhID+"/transactions",
		`{"category_id":`+catID+`,"amount":"-25.50","description":"Electric","date":"2026-02-10"}`)
	assertStatus(t, resp, http.StatusCreated)
	var tx map[string]interface{}
	decodeJSON(t, resp, &tx)
	txID := itoa(int(tx["id"].(float64)))
	if tx["amount"] != "-25.5" {
		t.Errorf("expected amount '-25.5', got %v", tx["amount"])
	}

	// List by month
	resp = doRequest(t, env, "GET", "/api/v1/households/"+hhID+"/transactions?month=2026-02", "")
	assertStatus(t, resp, http.StatusOK)
	var txList []map[string]interface{}
	decodeJSON(t, resp, &txList)
	if len(txList) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(txList))
	}

	// Wrong month returns empty
	resp = doRequest(t, env, "GET", "/api/v1/households/"+hhID+"/transactions?month=2026-03", "")
	assertStatus(t, resp, http.StatusOK)
	decodeJSON(t, resp, &txList)
	if len(txList) != 0 {
		t.Errorf("expected 0 transactions for wrong month, got %d", len(txList))
	}

	// Delete transaction
	resp = doRequest(t, env, "DELETE", "/api/v1/households/"+hhID+"/transactions/"+txID, "")
	assertStatus(t, resp, http.StatusNoContent)
}

func TestRecurringExpenseCRUD(t *testing.T) {
	env := setupTestEnv(t)

	// Setup
	resp := doRequest(t, env, "POST", "/api/v1/households", `{"name":"Recurring Test","currency":"EUR"}`)
	assertStatus(t, resp, http.StatusCreated)
	var hh map[string]interface{}
	decodeJSON(t, resp, &hh)
	hhID := itoa(int(hh["id"].(float64)))

	resp = doRequest(t, env, "POST", "/api/v1/households/"+hhID+"/categories", `{"name":"Housing"}`)
	assertStatus(t, resp, http.StatusCreated)
	var cat map[string]interface{}
	decodeJSON(t, resp, &cat)
	catID := itoa(int(cat["id"].(float64)))

	// Create
	resp = doRequest(t, env, "POST", "/api/v1/households/"+hhID+"/recurring-expenses",
		`{"category_id":`+catID+`,"name":"Internet","amount":"-39.99","frequency":"monthly","start_date":"2026-01-01"}`)
	assertStatus(t, resp, http.StatusCreated)
	var re map[string]interface{}
	decodeJSON(t, resp, &re)
	reID := itoa(int(re["id"].(float64)))
	if re["name"] != "Internet" {
		t.Errorf("expected name 'Internet', got %v", re["name"])
	}

	// List
	resp = doRequest(t, env, "GET", "/api/v1/households/"+hhID+"/recurring-expenses", "")
	assertStatus(t, resp, http.StatusOK)
	var reList []map[string]interface{}
	decodeJSON(t, resp, &reList)
	if len(reList) != 1 {
		t.Errorf("expected 1 recurring expense, got %d", len(reList))
	}

	// Update
	resp = doRequest(t, env, "PUT", "/api/v1/households/"+hhID+"/recurring-expenses/"+reID,
		`{"category_id":`+catID+`,"name":"Fiber Internet","amount":"-49.99","frequency":"monthly","active":true,"start_date":"2026-01-01"}`)
	assertStatus(t, resp, http.StatusOK)
	decodeJSON(t, resp, &re)
	if re["name"] != "Fiber Internet" {
		t.Errorf("expected name 'Fiber Internet', got %v", re["name"])
	}

	// Delete
	resp = doRequest(t, env, "DELETE", "/api/v1/households/"+hhID+"/recurring-expenses/"+reID, "")
	assertStatus(t, resp, http.StatusNoContent)

	// Verify deleted
	resp = doRequest(t, env, "GET", "/api/v1/households/"+hhID+"/recurring-expenses", "")
	assertStatus(t, resp, http.StatusOK)
	decodeJSON(t, resp, &reList)
	if len(reList) != 0 {
		t.Errorf("expected 0 recurring expenses, got %d", len(reList))
	}
}

func TestSummaryEndpoint(t *testing.T) {
	env := setupTestEnv(t)

	// Setup
	resp := doRequest(t, env, "POST", "/api/v1/households", `{"name":"Summary Test","currency":"EUR"}`)
	assertStatus(t, resp, http.StatusCreated)
	var hh map[string]interface{}
	decodeJSON(t, resp, &hh)
	hhID := itoa(int(hh["id"].(float64)))

	resp = doRequest(t, env, "POST", "/api/v1/households/"+hhID+"/categories", `{"name":"General"}`)
	assertStatus(t, resp, http.StatusCreated)
	var cat map[string]interface{}
	decodeJSON(t, resp, &cat)
	catID := itoa(int(cat["id"].(float64)))

	// Add recurring expense
	doRequest(t, env, "POST", "/api/v1/households/"+hhID+"/recurring-expenses",
		`{"category_id":`+catID+`,"name":"Rent","amount":"-1000","frequency":"monthly","start_date":"2026-01-01"}`)

	// Add one-time income
	doRequest(t, env, "POST", "/api/v1/households/"+hhID+"/transactions",
		`{"category_id":`+catID+`,"amount":"3000","description":"Salary","date":"2026-01-05"}`)

	// Add one-time expense
	doRequest(t, env, "POST", "/api/v1/households/"+hhID+"/transactions",
		`{"category_id":`+catID+`,"amount":"-75","description":"Groceries","date":"2026-01-10"}`)

	// Get summary
	resp = doRequest(t, env, "GET", "/api/v1/households/"+hhID+"/summary?month=2026-01", "")
	assertStatus(t, resp, http.StatusOK)
	var summary map[string]interface{}
	decodeJSON(t, resp, &summary)

	if summary["month"] != "2026-01" {
		t.Errorf("month = %v, want '2026-01'", summary["month"])
	}
	if summary["recurring_total"] != "-1000" {
		t.Errorf("recurring_total = %v, want '-1000'", summary["recurring_total"])
	}
	if summary["total_income"] != "3000" {
		t.Errorf("total_income = %v, want '3000'", summary["total_income"])
	}
	if summary["total_expenses"] != "-75" {
		t.Errorf("total_expenses = %v, want '-75'", summary["total_expenses"])
	}
}

func TestTokenManagement(t *testing.T) {
	env := setupTestEnv(t)

	// List tokens (should include the test token)
	resp := doRequest(t, env, "GET", "/api/v1/tokens", "")
	assertStatus(t, resp, http.StatusOK)
	var tokens []map[string]interface{}
	decodeJSON(t, resp, &tokens)
	initialCount := len(tokens)

	// Create new token
	resp = doRequest(t, env, "POST", "/api/v1/tokens", `{"name":"New API Token"}`)
	assertStatus(t, resp, http.StatusCreated)

	// List again — should have one more
	resp = doRequest(t, env, "GET", "/api/v1/tokens", "")
	assertStatus(t, resp, http.StatusOK)
	decodeJSON(t, resp, &tokens)
	if len(tokens) != initialCount+1 {
		t.Errorf("expected %d tokens, got %d", initialCount+1, len(tokens))
	}
}

func TestUnauthorized(t *testing.T) {
	env := setupTestEnv(t)

	// Request without Bearer token
	req, _ := http.NewRequest("GET", env.server.URL+"/api/v1/households", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("executing request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func doRequest(t *testing.T, env *testEnv, method, path, body string) *http.Response {
	t.Helper()
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, env.server.URL+path, bodyReader)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if env.token != "" {
		req.Header.Set("Authorization", "Bearer "+env.token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("executing request: %v", err)
	}
	return resp
}

func assertStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("expected status %d, got %d: %s", expected, resp.StatusCode, string(body))
	}
}

func decodeJSON(t *testing.T, resp *http.Response, v interface{}) {
	t.Helper()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
}

func TestTransactionUpdate(t *testing.T) {
	env := setupTestEnv(t)

	// Setup
	resp := doRequest(t, env, "POST", "/api/v1/households", `{"name":"TxUpdate Test","currency":"EUR"}`)
	assertStatus(t, resp, http.StatusCreated)
	var hh map[string]interface{}
	decodeJSON(t, resp, &hh)
	hhID := itoa(int(hh["id"].(float64)))

	resp = doRequest(t, env, "POST", "/api/v1/households/"+hhID+"/categories", `{"name":"Bills"}`)
	assertStatus(t, resp, http.StatusCreated)
	var cat map[string]interface{}
	decodeJSON(t, resp, &cat)
	catID := itoa(int(cat["id"].(float64)))

	// Create transaction
	resp = doRequest(t, env, "POST", "/api/v1/households/"+hhID+"/transactions",
		`{"category_id":`+catID+`,"amount":"-25.50","description":"Electric","date":"2026-02-10"}`)
	assertStatus(t, resp, http.StatusCreated)
	var tx map[string]interface{}
	decodeJSON(t, resp, &tx)
	txID := itoa(int(tx["id"].(float64)))

	// Update transaction
	resp = doRequest(t, env, "PUT", "/api/v1/households/"+hhID+"/transactions/"+txID,
		`{"category_id":`+catID+`,"amount":"-30.00","description":"Electric Updated","date":"2026-02-15"}`)
	assertStatus(t, resp, http.StatusOK)
	decodeJSON(t, resp, &tx)
	if tx["amount"] != "-30" {
		t.Errorf("expected amount '-30', got %v", tx["amount"])
	}
	if tx["description"] != "Electric Updated" {
		t.Errorf("expected description 'Electric Updated', got %v", tx["description"])
	}
	if tx["date"] != "2026-02-15" {
		t.Errorf("expected date '2026-02-15', got %v", tx["date"])
	}
}

func TestHouseholdFullFields(t *testing.T) {
	env := setupTestEnv(t)

	// Create with description and icon
	resp := doRequest(t, env, "POST", "/api/v1/households",
		`{"name":"Full Fields","description":"Test description","currency":"EUR","icon":"wallet"}`)
	assertStatus(t, resp, http.StatusCreated)
	var hh map[string]interface{}
	decodeJSON(t, resp, &hh)
	hhID := itoa(int(hh["id"].(float64)))

	if hh["description"] != "Test description" {
		t.Errorf("expected description 'Test description', got %v", hh["description"])
	}
	if hh["icon"] != "wallet" {
		t.Errorf("expected icon 'wallet', got %v", hh["icon"])
	}

	// Update with new description and icon
	resp = doRequest(t, env, "PUT", "/api/v1/households/"+hhID,
		`{"name":"Full Fields Updated","description":"Updated desc","currency":"USD","icon":"piggybank"}`)
	assertStatus(t, resp, http.StatusOK)
	decodeJSON(t, resp, &hh)
	if hh["description"] != "Updated desc" {
		t.Errorf("expected description 'Updated desc', got %v", hh["description"])
	}
	if hh["icon"] != "piggybank" {
		t.Errorf("expected icon 'piggybank', got %v", hh["icon"])
	}
}

func TestCategoryWithIcon(t *testing.T) {
	env := setupTestEnv(t)

	// Setup household
	resp := doRequest(t, env, "POST", "/api/v1/households", `{"name":"Icon Test","currency":"EUR"}`)
	assertStatus(t, resp, http.StatusCreated)
	var hh map[string]interface{}
	decodeJSON(t, resp, &hh)
	hhID := itoa(int(hh["id"].(float64)))

	// Create category with icon
	resp = doRequest(t, env, "POST", "/api/v1/households/"+hhID+"/categories", `{"name":"Food","icon":"restaurant"}`)
	assertStatus(t, resp, http.StatusCreated)
	var cat map[string]interface{}
	decodeJSON(t, resp, &cat)
	catID := itoa(int(cat["id"].(float64)))
	if cat["icon"] != "restaurant" {
		t.Errorf("expected icon 'restaurant', got %v", cat["icon"])
	}

	// Update category with new icon
	resp = doRequest(t, env, "PUT", "/api/v1/households/"+hhID+"/categories/"+catID, `{"name":"Dining","icon":"dining"}`)
	assertStatus(t, resp, http.StatusOK)
	decodeJSON(t, resp, &cat)
	if cat["name"] != "Dining" {
		t.Errorf("expected name 'Dining', got %v", cat["name"])
	}
	if cat["icon"] != "dining" {
		t.Errorf("expected icon 'dining', got %v", cat["icon"])
	}
}

func TestRecurringExpenseDescription(t *testing.T) {
	env := setupTestEnv(t)

	// Setup
	resp := doRequest(t, env, "POST", "/api/v1/households", `{"name":"RE Desc Test","currency":"EUR"}`)
	assertStatus(t, resp, http.StatusCreated)
	var hh map[string]interface{}
	decodeJSON(t, resp, &hh)
	hhID := itoa(int(hh["id"].(float64)))

	resp = doRequest(t, env, "POST", "/api/v1/households/"+hhID+"/categories", `{"name":"Bills"}`)
	assertStatus(t, resp, http.StatusCreated)
	var cat map[string]interface{}
	decodeJSON(t, resp, &cat)
	catID := itoa(int(cat["id"].(float64)))

	// Create with description
	resp = doRequest(t, env, "POST", "/api/v1/households/"+hhID+"/recurring-expenses",
		`{"category_id":`+catID+`,"name":"Internet","description":"Fiber 100Mbit","amount":"-39.99","frequency":"monthly","start_date":"2026-01-01"}`)
	assertStatus(t, resp, http.StatusCreated)
	var re map[string]interface{}
	decodeJSON(t, resp, &re)
	reID := itoa(int(re["id"].(float64)))
	if re["description"] != "Fiber 100Mbit" {
		t.Errorf("expected description 'Fiber 100Mbit', got %v", re["description"])
	}

	// Update with new description
	resp = doRequest(t, env, "PUT", "/api/v1/households/"+hhID+"/recurring-expenses/"+reID,
		`{"category_id":`+catID+`,"name":"Internet","description":"Fiber 1Gbit","amount":"-49.99","frequency":"monthly","active":true,"start_date":"2026-01-01"}`)
	assertStatus(t, resp, http.StatusOK)
	decodeJSON(t, resp, &re)
	if re["description"] != "Fiber 1Gbit" {
		t.Errorf("expected description 'Fiber 1Gbit', got %v", re["description"])
	}
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
