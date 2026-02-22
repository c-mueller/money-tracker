//go:build integration

package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"icekalt.dev/money-tracker/internal/auth"
	"icekalt.dev/money-tracker/internal/devmode"
)

// gqlRequest sends a GraphQL request with Bearer token auth.
func gqlRequest(t *testing.T, env *testEnv, query string) map[string]interface{} {
	t.Helper()
	body := map[string]string{"query": query}
	b, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", env.server.URL+"/graphql", strings.NewReader(string(b)))
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+env.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("executing request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decoding GraphQL response: %v", err)
	}
	return result
}

// gqlData extracts the "data" field from a GraphQL response, failing if there are errors.
func gqlData(t *testing.T, result map[string]interface{}) map[string]interface{} {
	t.Helper()
	if errs, ok := result["errors"]; ok {
		t.Fatalf("GraphQL errors: %v", errs)
	}
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("no data in GraphQL response: %v", result)
	}
	return data
}

func TestGraphQLHouseholds(t *testing.T) {
	env := setupTestEnv(t)

	// Create household
	result := gqlRequest(t, env, `mutation {
		createHousehold(input: {name: "GQL Household", currency: "EUR", description: "Test desc", icon: "wallet"}) {
			id name description currency icon ownerID
		}
	}`)
	data := gqlData(t, result)
	hh := data["createHousehold"].(map[string]interface{})
	if hh["name"] != "GQL Household" {
		t.Errorf("expected name 'GQL Household', got %v", hh["name"])
	}
	if hh["description"] != "Test desc" {
		t.Errorf("expected description 'Test desc', got %v", hh["description"])
	}
	if hh["icon"] != "wallet" {
		t.Errorf("expected icon 'wallet', got %v", hh["icon"])
	}
	hhID := int(hh["id"].(float64))

	// List households
	result = gqlRequest(t, env, `{ households { id name description } }`)
	data = gqlData(t, result)
	households := data["households"].([]interface{})
	if len(households) != 1 {
		t.Errorf("expected 1 household, got %d", len(households))
	}

	// Get single household
	result = gqlRequest(t, env, `{ household(id: `+itoa(hhID)+`) { id name description currency icon } }`)
	data = gqlData(t, result)
	single := data["household"].(map[string]interface{})
	if single["name"] != "GQL Household" {
		t.Errorf("expected name 'GQL Household', got %v", single["name"])
	}

	// Update household
	result = gqlRequest(t, env, `mutation {
		updateHousehold(input: {id: `+itoa(hhID)+`, name: "Updated GQL", currency: "USD", description: "New desc", icon: "home"}) {
			id name description currency icon
		}
	}`)
	data = gqlData(t, result)
	updated := data["updateHousehold"].(map[string]interface{})
	if updated["name"] != "Updated GQL" {
		t.Errorf("expected name 'Updated GQL', got %v", updated["name"])
	}
	if updated["currency"] != "USD" {
		t.Errorf("expected currency 'USD', got %v", updated["currency"])
	}
}

func TestGraphQLCategories(t *testing.T) {
	env := setupTestEnv(t)

	// Setup household
	result := gqlRequest(t, env, `mutation {
		createHousehold(input: {name: "Cat GQL", currency: "EUR"}) { id }
	}`)
	hhID := int(gqlData(t, result)["createHousehold"].(map[string]interface{})["id"].(float64))

	// Create category
	result = gqlRequest(t, env, `mutation {
		createCategory(input: {householdID: `+itoa(hhID)+`, name: "Food", icon: "restaurant"}) {
			id name icon householdID
		}
	}`)
	data := gqlData(t, result)
	cat := data["createCategory"].(map[string]interface{})
	catID := int(cat["id"].(float64))
	if cat["name"] != "Food" {
		t.Errorf("expected name 'Food', got %v", cat["name"])
	}
	if cat["icon"] != "restaurant" {
		t.Errorf("expected icon 'restaurant', got %v", cat["icon"])
	}

	// List categories
	result = gqlRequest(t, env, `{ categories(householdID: `+itoa(hhID)+`) { id name icon } }`)
	data = gqlData(t, result)
	cats := data["categories"].([]interface{})
	if len(cats) != 1 {
		t.Errorf("expected 1 category, got %d", len(cats))
	}

	// Update category
	result = gqlRequest(t, env, `mutation {
		updateCategory(input: {id: `+itoa(catID)+`, name: "Groceries", icon: "cart"}) {
			id name icon
		}
	}`)
	data = gqlData(t, result)
	updatedCat := data["updateCategory"].(map[string]interface{})
	if updatedCat["name"] != "Groceries" {
		t.Errorf("expected name 'Groceries', got %v", updatedCat["name"])
	}
	if updatedCat["icon"] != "cart" {
		t.Errorf("expected icon 'cart', got %v", updatedCat["icon"])
	}
}

func TestGraphQLTransactions(t *testing.T) {
	env := setupTestEnv(t)

	// Setup
	result := gqlRequest(t, env, `mutation {
		createHousehold(input: {name: "Tx GQL", currency: "EUR"}) { id }
	}`)
	hhID := int(gqlData(t, result)["createHousehold"].(map[string]interface{})["id"].(float64))

	result = gqlRequest(t, env, `mutation {
		createCategory(input: {householdID: `+itoa(hhID)+`, name: "Bills"}) { id }
	}`)
	catID := int(gqlData(t, result)["createCategory"].(map[string]interface{})["id"].(float64))

	// Create transaction
	result = gqlRequest(t, env, `mutation {
		createTransaction(input: {householdID: `+itoa(hhID)+`, categoryID: `+itoa(catID)+`, amount: "-50.00", description: "Electric", date: "2026-02-10"}) {
			id amount description date householdID categoryID
		}
	}`)
	data := gqlData(t, result)
	tx := data["createTransaction"].(map[string]interface{})
	txID := int(tx["id"].(float64))
	if tx["amount"] != "-50" {
		t.Errorf("expected amount '-50', got %v", tx["amount"])
	}
	if tx["description"] != "Electric" {
		t.Errorf("expected description 'Electric', got %v", tx["description"])
	}

	// Query by month
	result = gqlRequest(t, env, `{ transactions(householdID: `+itoa(hhID)+`, month: "2026-02") { id amount description date } }`)
	data = gqlData(t, result)
	txList := data["transactions"].([]interface{})
	if len(txList) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(txList))
	}

	// Wrong month returns empty
	result = gqlRequest(t, env, `{ transactions(householdID: `+itoa(hhID)+`, month: "2026-03") { id } }`)
	data = gqlData(t, result)
	txList = data["transactions"].([]interface{})
	if len(txList) != 0 {
		t.Errorf("expected 0 transactions, got %d", len(txList))
	}

	// Update transaction
	result = gqlRequest(t, env, `mutation {
		updateTransaction(input: {id: `+itoa(txID)+`, householdID: `+itoa(hhID)+`, categoryID: `+itoa(catID)+`, amount: "-75.00", description: "Electric Updated", date: "2026-02-15"}) {
			id amount description date
		}
	}`)
	data = gqlData(t, result)
	updatedTx := data["updateTransaction"].(map[string]interface{})
	if updatedTx["amount"] != "-75" {
		t.Errorf("expected amount '-75', got %v", updatedTx["amount"])
	}
	if updatedTx["description"] != "Electric Updated" {
		t.Errorf("expected description 'Electric Updated', got %v", updatedTx["description"])
	}
}

func TestGraphQLRecurringExpenses(t *testing.T) {
	env := setupTestEnv(t)

	// Setup
	result := gqlRequest(t, env, `mutation {
		createHousehold(input: {name: "RE GQL", currency: "EUR"}) { id }
	}`)
	hhID := int(gqlData(t, result)["createHousehold"].(map[string]interface{})["id"].(float64))

	result = gqlRequest(t, env, `mutation {
		createCategory(input: {householdID: `+itoa(hhID)+`, name: "Housing"}) { id }
	}`)
	catID := int(gqlData(t, result)["createCategory"].(map[string]interface{})["id"].(float64))

	// Create
	result = gqlRequest(t, env, `mutation {
		createRecurringExpense(input: {householdID: `+itoa(hhID)+`, categoryID: `+itoa(catID)+`, name: "Internet", description: "Fiber 100Mbit", amount: "-39.99", frequency: "monthly", startDate: "2026-01-01"}) {
			id name description amount frequency active startDate
		}
	}`)
	data := gqlData(t, result)
	re := data["createRecurringExpense"].(map[string]interface{})
	reID := int(re["id"].(float64))
	if re["name"] != "Internet" {
		t.Errorf("expected name 'Internet', got %v", re["name"])
	}
	if re["description"] != "Fiber 100Mbit" {
		t.Errorf("expected description 'Fiber 100Mbit', got %v", re["description"])
	}
	if re["active"] != true {
		t.Errorf("expected active true, got %v", re["active"])
	}

	// List
	result = gqlRequest(t, env, `{ recurringExpenses(householdID: `+itoa(hhID)+`) { id name description } }`)
	data = gqlData(t, result)
	reList := data["recurringExpenses"].([]interface{})
	if len(reList) != 1 {
		t.Errorf("expected 1 recurring expense, got %d", len(reList))
	}

	// Update
	result = gqlRequest(t, env, `mutation {
		updateRecurringExpense(input: {id: `+itoa(reID)+`, categoryID: `+itoa(catID)+`, name: "Fiber Internet", description: "Upgraded plan", amount: "-49.99", frequency: "monthly", active: true, startDate: "2026-01-01"}) {
			id name description amount
		}
	}`)
	data = gqlData(t, result)
	updatedRE := data["updateRecurringExpense"].(map[string]interface{})
	if updatedRE["name"] != "Fiber Internet" {
		t.Errorf("expected name 'Fiber Internet', got %v", updatedRE["name"])
	}
	if updatedRE["description"] != "Upgraded plan" {
		t.Errorf("expected description 'Upgraded plan', got %v", updatedRE["description"])
	}
}

func TestGraphQLSummary(t *testing.T) {
	env := setupTestEnv(t)

	// Setup
	result := gqlRequest(t, env, `mutation {
		createHousehold(input: {name: "Summary GQL", currency: "EUR"}) { id }
	}`)
	hhID := int(gqlData(t, result)["createHousehold"].(map[string]interface{})["id"].(float64))

	result = gqlRequest(t, env, `mutation {
		createCategory(input: {householdID: `+itoa(hhID)+`, name: "General"}) { id }
	}`)
	catID := int(gqlData(t, result)["createCategory"].(map[string]interface{})["id"].(float64))

	// Add recurring expense
	gqlRequest(t, env, `mutation {
		createRecurringExpense(input: {householdID: `+itoa(hhID)+`, categoryID: `+itoa(catID)+`, name: "Rent", amount: "-1000", frequency: "monthly", startDate: "2026-01-01"}) { id }
	}`)

	// Add transactions
	gqlRequest(t, env, `mutation {
		createTransaction(input: {householdID: `+itoa(hhID)+`, categoryID: `+itoa(catID)+`, amount: "3000", description: "Salary", date: "2026-01-05"}) { id }
	}`)
	gqlRequest(t, env, `mutation {
		createTransaction(input: {householdID: `+itoa(hhID)+`, categoryID: `+itoa(catID)+`, amount: "-75", description: "Groceries", date: "2026-01-10"}) { id }
	}`)

	// Get summary
	result = gqlRequest(t, env, `{ monthlySummary(householdID: `+itoa(hhID)+`, month: "2026-01") {
		month householdID totalIncome totalExpenses recurringTotal oneTimeTotal
		categoryBreakdown { categoryID categoryName recurring oneTime total }
	}}`)
	data := gqlData(t, result)
	summary := data["monthlySummary"].(map[string]interface{})
	if summary["month"] != "2026-01" {
		t.Errorf("expected month '2026-01', got %v", summary["month"])
	}
	if summary["recurringTotal"] != "-1000" {
		t.Errorf("expected recurringTotal '-1000', got %v", summary["recurringTotal"])
	}
	if summary["totalIncome"] != "3000" {
		t.Errorf("expected totalIncome '3000', got %v", summary["totalIncome"])
	}
	if summary["totalExpenses"] != "-75" {
		t.Errorf("expected totalExpenses '-75', got %v", summary["totalExpenses"])
	}
	breakdown := summary["categoryBreakdown"].([]interface{})
	if len(breakdown) == 0 {
		t.Error("expected non-empty category breakdown")
	}
}

func TestGraphQLAuth(t *testing.T) {
	env := setupTestEnv(t)

	t.Run("no auth returns 401", func(t *testing.T) {
		if devmode.Enabled {
			t.Skip("dev mode uses auto-auth")
		}

		body := `{"query":"{ households { id } }"}`
		req, _ := http.NewRequest("POST", env.server.URL+"/graphql", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("executing request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401 for GraphQL without auth, got %d", resp.StatusCode)
		}
	})

	t.Run("bearer token accepted", func(t *testing.T) {
		result := gqlRequest(t, env, `{ households { id } }`)
		if _, ok := result["data"]; !ok {
			t.Error("expected data in response with bearer token")
		}
	})

	t.Run("session cookie accepted", func(t *testing.T) {
		// Create a valid session cookie
		fakeReq := httptest.NewRequest(http.MethodGet, "/", nil)
		fakeRec := httptest.NewRecorder()
		session, _ := env.sessionStore.New(fakeReq, auth.SessionName)
		session.Values[auth.SessionKeyUser] = env.userID
		session.Save(fakeReq, fakeRec)

		cookies := fakeRec.Result().Cookies()
		if len(cookies) == 0 {
			t.Fatal("no session cookie was set")
		}

		body := `{"query":"{ households { id } }"}`
		req, _ := http.NewRequest("POST", env.server.URL+"/graphql", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		for _, c := range cookies {
			req.AddCookie(c)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("executing request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200 for GraphQL with session cookie, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		if _, ok := result["data"]; !ok {
			t.Error("expected data in response with session cookie")
		}
	})
}

func TestGraphQLNoDeleteMutations(t *testing.T) {
	env := setupTestEnv(t)

	// Delete mutations should not exist in the schema.
	// gqlgen returns 422 for schema validation errors, so we check the status directly.
	mutations := []string{
		`mutation { deleteHousehold(id: 1) }`,
		`mutation { deleteCategory(id: 1) }`,
		`mutation { deleteTransaction(id: 1) }`,
		`mutation { deleteRecurringExpense(id: 1) }`,
	}

	for _, query := range mutations {
		body, _ := json.Marshal(map[string]string{"query": query})
		req, _ := http.NewRequest("POST", env.server.URL+"/graphql", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+env.token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("executing request: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusUnprocessableEntity && resp.StatusCode != http.StatusOK {
			t.Errorf("expected 422 or 200 with errors for %s, got %d", query, resp.StatusCode)
		}
	}
}

func TestGraphQLValidationErrors(t *testing.T) {
	env := setupTestEnv(t)

	// Empty household name
	result := gqlRequest(t, env, `mutation {
		createHousehold(input: {name: "", currency: "EUR"}) { id }
	}`)
	if _, hasErrors := result["errors"]; !hasErrors {
		t.Error("expected GraphQL errors for empty household name")
	}

	// Invalid currency
	result = gqlRequest(t, env, `mutation {
		createHousehold(input: {name: "Test", currency: "invalid"}) { id }
	}`)
	if _, hasErrors := result["errors"]; !hasErrors {
		t.Error("expected GraphQL errors for invalid currency")
	}

	// Invalid date format
	result = gqlRequest(t, env, `mutation {
		createHousehold(input: {name: "Test", currency: "EUR"}) { id }
	}`)
	hhID := int(gqlData(t, result)["createHousehold"].(map[string]interface{})["id"].(float64))

	result = gqlRequest(t, env, `mutation {
		createCategory(input: {householdID: `+itoa(hhID)+`, name: "Cat"}) { id }
	}`)
	catID := int(gqlData(t, result)["createCategory"].(map[string]interface{})["id"].(float64))

	result = gqlRequest(t, env, `mutation {
		createTransaction(input: {householdID: `+itoa(hhID)+`, categoryID: `+itoa(catID)+`, amount: "-50", date: "not-a-date"}) { id }
	}`)
	if _, hasErrors := result["errors"]; !hasErrors {
		t.Error("expected GraphQL errors for invalid date")
	}

	// Invalid frequency
	result = gqlRequest(t, env, `mutation {
		createRecurringExpense(input: {householdID: `+itoa(hhID)+`, categoryID: `+itoa(catID)+`, name: "Test", amount: "-50", frequency: "invalid", startDate: "2026-01-01"}) { id }
	}`)
	if _, hasErrors := result["errors"]; !hasErrors {
		t.Error("expected GraphQL errors for invalid frequency")
	}
}

func TestGraphQLFullFlow(t *testing.T) {
	env := setupTestEnv(t)

	// 1. Create household
	result := gqlRequest(t, env, `mutation {
		createHousehold(input: {name: "E2E Test", currency: "EUR", description: "End-to-end"}) { id name }
	}`)
	hhID := int(gqlData(t, result)["createHousehold"].(map[string]interface{})["id"].(float64))

	// 2. Create categories
	result = gqlRequest(t, env, `mutation {
		createCategory(input: {householdID: `+itoa(hhID)+`, name: "Income", icon: "money"}) { id }
	}`)
	incomeCatID := int(gqlData(t, result)["createCategory"].(map[string]interface{})["id"].(float64))

	result = gqlRequest(t, env, `mutation {
		createCategory(input: {householdID: `+itoa(hhID)+`, name: "Housing", icon: "house"}) { id }
	}`)
	housingCatID := int(gqlData(t, result)["createCategory"].(map[string]interface{})["id"].(float64))

	// 3. Create recurring expense
	result = gqlRequest(t, env, `mutation {
		createRecurringExpense(input: {householdID: `+itoa(hhID)+`, categoryID: `+itoa(housingCatID)+`, name: "Rent", description: "Monthly rent", amount: "-800", frequency: "monthly", startDate: "2026-01-01"}) { id name }
	}`)
	reID := int(gqlData(t, result)["createRecurringExpense"].(map[string]interface{})["id"].(float64))

	// 4. Create transactions
	result = gqlRequest(t, env, `mutation {
		createTransaction(input: {householdID: `+itoa(hhID)+`, categoryID: `+itoa(incomeCatID)+`, amount: "3000", description: "Salary", date: "2026-01-05"}) { id }
	}`)
	gqlData(t, result) // verify no errors

	result = gqlRequest(t, env, `mutation {
		createTransaction(input: {householdID: `+itoa(hhID)+`, categoryID: `+itoa(housingCatID)+`, amount: "-50", description: "Nebenkosten", date: "2026-01-15"}) { id }
	}`)
	txID := int(gqlData(t, result)["createTransaction"].(map[string]interface{})["id"].(float64))

	// 5. Update recurring expense
	result = gqlRequest(t, env, `mutation {
		updateRecurringExpense(input: {id: `+itoa(reID)+`, categoryID: `+itoa(housingCatID)+`, name: "Rent Updated", description: "Increased rent", amount: "-850", frequency: "monthly", active: true, startDate: "2026-01-01"}) { id name amount }
	}`)
	data := gqlData(t, result)
	updatedRE := data["updateRecurringExpense"].(map[string]interface{})
	if updatedRE["amount"] != "-850" {
		t.Errorf("expected amount '-850', got %v", updatedRE["amount"])
	}

	// 6. Update transaction
	result = gqlRequest(t, env, `mutation {
		updateTransaction(input: {id: `+itoa(txID)+`, householdID: `+itoa(hhID)+`, categoryID: `+itoa(housingCatID)+`, amount: "-60", description: "Nebenkosten Updated", date: "2026-01-15"}) { id amount description }
	}`)
	data = gqlData(t, result)
	updatedTx := data["updateTransaction"].(map[string]interface{})
	if updatedTx["amount"] != "-60" {
		t.Errorf("expected amount '-60', got %v", updatedTx["amount"])
	}

	// 7. Verify summary
	result = gqlRequest(t, env, `{ monthlySummary(householdID: `+itoa(hhID)+`, month: "2026-01") {
		month totalIncome totalExpenses recurringTotal oneTimeTotal
		categoryBreakdown { categoryName recurring oneTime total }
	}}`)
	data = gqlData(t, result)
	summary := data["monthlySummary"].(map[string]interface{})
	if summary["month"] != "2026-01" {
		t.Errorf("expected month '2026-01', got %v", summary["month"])
	}
	if summary["totalIncome"] != "3000" {
		t.Errorf("expected totalIncome '3000', got %v", summary["totalIncome"])
	}
	if summary["recurringTotal"] != "-850" {
		t.Errorf("expected recurringTotal '-850', got %v", summary["recurringTotal"])
	}

	// 8. Verify categories list
	result = gqlRequest(t, env, `{ categories(householdID: `+itoa(hhID)+`) { id name icon } }`)
	data = gqlData(t, result)
	cats := data["categories"].([]interface{})
	if len(cats) != 2 {
		t.Errorf("expected 2 categories, got %d", len(cats))
	}

	// 9. Verify recurring expenses list
	result = gqlRequest(t, env, `{ recurringExpenses(householdID: `+itoa(hhID)+`) { id name description amount } }`)
	data = gqlData(t, result)
	reList := data["recurringExpenses"].([]interface{})
	if len(reList) != 1 {
		t.Errorf("expected 1 recurring expense, got %d", len(reList))
	}

	// 10. Verify transactions list
	result = gqlRequest(t, env, `{ transactions(householdID: `+itoa(hhID)+`, month: "2026-01") { id amount description } }`)
	data = gqlData(t, result)
	txList := data["transactions"].([]interface{})
	if len(txList) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(txList))
	}
}
