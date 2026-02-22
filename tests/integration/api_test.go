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

	// 1. Health check
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

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
