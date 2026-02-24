//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"testing"

	mcppkg "icekalt.dev/money-tracker/internal/mcp"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// setupMCPEnv creates a test environment and returns an MCP server session
// connected via in-process transport. This tests the full stack:
// MCP client → tool handler → HTTP client → REST API → DB
func setupMCPEnv(t *testing.T) (*testEnv, *mcp.ClientSession) {
	t.Helper()

	env := setupTestEnv(t)

	apiClient := mcppkg.NewClient(env.server.URL, env.token)
	mcpServer := mcppkg.NewServer(apiClient, "test")

	mcpClient := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "v0.0.1",
	}, nil)

	session, err := mcpClient.Connect(context.Background(), mcpServer.InProcessTransport(), nil)
	if err != nil {
		t.Fatalf("failed to connect MCP client: %v", err)
	}
	t.Cleanup(func() {
		session.Close()
	})

	return env, session
}

func callTool(t *testing.T, session *mcp.ClientSession, name string, args map[string]any) string {
	t.Helper()

	argsJSON, err := json.Marshal(args)
	if err != nil {
		t.Fatalf("marshaling tool args: %v", err)
	}

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      name,
		Arguments: json.RawMessage(argsJSON),
	})
	if err != nil {
		t.Fatalf("calling tool %s: %v", name, err)
	}
	if result.IsError {
		t.Fatalf("tool %s returned error: %v", name, toolText(result))
	}
	return toolText(result)
}

func callToolExpectError(t *testing.T, session *mcp.ClientSession, name string, args map[string]any) string {
	t.Helper()

	argsJSON, err := json.Marshal(args)
	if err != nil {
		t.Fatalf("marshaling tool args: %v", err)
	}

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      name,
		Arguments: json.RawMessage(argsJSON),
	})
	if err != nil {
		// Network/protocol errors count as expected errors
		return err.Error()
	}
	if result.IsError {
		return toolText(result)
	}
	return toolText(result)
}

func toolText(result *mcp.CallToolResult) string {
	for _, c := range result.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			return tc.Text
		}
	}
	return ""
}

func parseJSONArray(t *testing.T, text string) []map[string]any {
	t.Helper()
	var result []map[string]any
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		t.Fatalf("parsing JSON array: %v (text: %s)", err, text)
	}
	return result
}

func parseJSONObject(t *testing.T, text string) map[string]any {
	t.Helper()
	var result map[string]any
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		t.Fatalf("parsing JSON object: %v (text: %s)", err, text)
	}
	return result
}

// --- Full Flow Test ---

func TestMCPFullFlow(t *testing.T) {
	_, session := setupMCPEnv(t)

	// 1. List households (empty)
	text := callTool(t, session, "list_households", nil)
	households := parseJSONArray(t, text)
	if len(households) != 0 {
		t.Errorf("expected 0 households, got %d", len(households))
	}

	// 2. Create household
	text = callTool(t, session, "create_household", map[string]any{
		"name":     "Test Haushalt",
		"currency": "EUR",
	})
	hh := parseJSONObject(t, text)
	hhID := int(hh["id"].(float64))
	if hh["name"] != "Test Haushalt" {
		t.Errorf("expected name 'Test Haushalt', got %v", hh["name"])
	}
	if hh["currency"] != "EUR" {
		t.Errorf("expected currency 'EUR', got %v", hh["currency"])
	}

	// 3. List households (1 result)
	text = callTool(t, session, "list_households", nil)
	households = parseJSONArray(t, text)
	if len(households) != 1 {
		t.Errorf("expected 1 household, got %d", len(households))
	}

	// 4. Create category
	text = callTool(t, session, "create_category", map[string]any{
		"household_id": hhID,
		"name":         "Miete",
		"icon":         "home",
	})
	cat := parseJSONObject(t, text)
	catID := int(cat["id"].(float64))
	if cat["name"] != "Miete" {
		t.Errorf("expected category name 'Miete', got %v", cat["name"])
	}

	// 5. Create transaction
	text = callTool(t, session, "create_transaction", map[string]any{
		"household_id": hhID,
		"category_id":  catID,
		"amount":       "-50.00",
		"description":  "Nebenkosten",
		"date":         "2026-01-15",
	})
	tx := parseJSONObject(t, text)
	if tx["amount"] != "-50" {
		t.Errorf("expected amount '-50', got %v", tx["amount"])
	}
	if tx["description"] != "Nebenkosten" {
		t.Errorf("expected description 'Nebenkosten', got %v", tx["description"])
	}

	// 6. Create recurring expense
	text = callTool(t, session, "create_recurring_expense", map[string]any{
		"household_id": hhID,
		"category_id":  catID,
		"name":         "Kaltmiete",
		"amount":       "-800.00",
		"frequency":    "monthly",
		"start_date":   "2026-01-01",
	})
	re := parseJSONObject(t, text)
	reID := int(re["id"].(float64))
	if re["name"] != "Kaltmiete" {
		t.Errorf("expected name 'Kaltmiete', got %v", re["name"])
	}

	// 7. Get monthly summary
	text = callTool(t, session, "get_monthly_summary", map[string]any{
		"household_id": hhID,
		"month":        "2026-01",
	})
	summary := parseJSONObject(t, text)
	if summary["month"] != "2026-01" {
		t.Errorf("expected month '2026-01', got %v", summary["month"])
	}

	// 8. Create schedule override
	text = callTool(t, session, "create_schedule_override", map[string]any{
		"household_id":   hhID,
		"recurring_id":   reID,
		"effective_date": "2026-06-01",
		"amount":         "-900.00",
		"frequency":      "monthly",
	})
	override := parseJSONObject(t, text)
	ovID := int(override["id"].(float64))
	if override["amount"] != "-900" {
		t.Errorf("expected override amount '-900', got %v", override["amount"])
	}

	// 9. Delete override, then recurring, then household
	// (cascade delete from household doesn't cover overrides)
	callTool(t, session, "delete_schedule_override", map[string]any{
		"household_id": hhID,
		"recurring_id": reID,
		"override_id":  ovID,
	})
	callTool(t, session, "delete_recurring_expense", map[string]any{
		"household_id": hhID,
		"recurring_id": reID,
	})

	text = callTool(t, session, "delete_household", map[string]any{
		"id": hhID,
	})
	if text != "Household deleted" {
		t.Errorf("expected 'Household deleted', got %v", text)
	}

	// 10. Verify empty
	text = callTool(t, session, "list_households", nil)
	households = parseJSONArray(t, text)
	if len(households) != 0 {
		t.Errorf("expected 0 households after delete, got %d", len(households))
	}
}

// --- Household CRUD ---

func TestMCPHouseholdCRUD(t *testing.T) {
	_, session := setupMCPEnv(t)

	// Create
	text := callTool(t, session, "create_household", map[string]any{
		"name":        "CRUD Test",
		"currency":    "USD",
		"description": "Test desc",
		"icon":        "wallet",
	})
	hh := parseJSONObject(t, text)
	hhID := int(hh["id"].(float64))
	if hh["name"] != "CRUD Test" {
		t.Errorf("expected name 'CRUD Test', got %v", hh["name"])
	}
	if hh["currency"] != "USD" {
		t.Errorf("expected currency 'USD', got %v", hh["currency"])
	}

	// Update
	text = callTool(t, session, "update_household", map[string]any{
		"id":       hhID,
		"name":     "Updated",
		"currency": "EUR",
	})
	hh = parseJSONObject(t, text)
	if hh["name"] != "Updated" {
		t.Errorf("expected name 'Updated', got %v", hh["name"])
	}

	// Delete
	text = callTool(t, session, "delete_household", map[string]any{"id": hhID})
	if text != "Household deleted" {
		t.Errorf("expected confirmation, got %v", text)
	}

	// Verify gone
	text = callTool(t, session, "list_households", nil)
	list := parseJSONArray(t, text)
	if len(list) != 0 {
		t.Errorf("expected 0 households, got %d", len(list))
	}
}

// --- Category CRUD ---

func TestMCPCategoryCRUD(t *testing.T) {
	_, session := setupMCPEnv(t)

	// Setup: create household
	text := callTool(t, session, "create_household", map[string]any{
		"name": "Cat Test", "currency": "EUR",
	})
	hh := parseJSONObject(t, text)
	hhID := int(hh["id"].(float64))

	// Create category
	text = callTool(t, session, "create_category", map[string]any{
		"household_id": hhID,
		"name":         "Lebensmittel",
		"icon":         "shopping_cart",
	})
	cat := parseJSONObject(t, text)
	catID := int(cat["id"].(float64))
	if cat["name"] != "Lebensmittel" {
		t.Errorf("expected 'Lebensmittel', got %v", cat["name"])
	}

	// List categories
	text = callTool(t, session, "list_categories", map[string]any{"household_id": hhID})
	cats := parseJSONArray(t, text)
	if len(cats) != 1 {
		t.Errorf("expected 1 category, got %d", len(cats))
	}

	// Update category
	text = callTool(t, session, "update_category", map[string]any{
		"household_id": hhID,
		"category_id":  catID,
		"name":         "Essen & Trinken",
	})
	cat = parseJSONObject(t, text)
	if cat["name"] != "Essen & Trinken" {
		t.Errorf("expected 'Essen & Trinken', got %v", cat["name"])
	}

	// Delete category
	text = callTool(t, session, "delete_category", map[string]any{
		"household_id": hhID,
		"category_id":  catID,
	})
	if text != "Category deleted" {
		t.Errorf("expected confirmation, got %v", text)
	}

	// Verify gone
	text = callTool(t, session, "list_categories", map[string]any{"household_id": hhID})
	cats = parseJSONArray(t, text)
	if len(cats) != 0 {
		t.Errorf("expected 0 categories, got %d", len(cats))
	}
}

// --- Transaction CRUD ---

func TestMCPTransactionCRUD(t *testing.T) {
	_, session := setupMCPEnv(t)

	// Setup
	text := callTool(t, session, "create_household", map[string]any{
		"name": "Tx Test", "currency": "EUR",
	})
	hh := parseJSONObject(t, text)
	hhID := int(hh["id"].(float64))

	text = callTool(t, session, "create_category", map[string]any{
		"household_id": hhID, "name": "Essen",
	})
	cat := parseJSONObject(t, text)
	catID := int(cat["id"].(float64))

	// Create transaction
	text = callTool(t, session, "create_transaction", map[string]any{
		"household_id": hhID,
		"category_id":  catID,
		"amount":       "-25.50",
		"description":  "Supermarkt",
		"date":         "2026-02-10",
	})
	tx := parseJSONObject(t, text)
	txID := int(tx["id"].(float64))
	if tx["description"] != "Supermarkt" {
		t.Errorf("expected 'Supermarkt', got %v", tx["description"])
	}

	// List transactions for month
	text = callTool(t, session, "list_transactions", map[string]any{
		"household_id": hhID,
		"month":        "2026-02",
	})
	txs := parseJSONArray(t, text)
	if len(txs) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(txs))
	}

	// List transactions for different month (empty)
	text = callTool(t, session, "list_transactions", map[string]any{
		"household_id": hhID,
		"month":        "2026-03",
	})
	txs = parseJSONArray(t, text)
	if len(txs) != 0 {
		t.Errorf("expected 0 transactions for 2026-03, got %d", len(txs))
	}

	// Update transaction
	text = callTool(t, session, "update_transaction", map[string]any{
		"household_id":   hhID,
		"transaction_id": txID,
		"amount":         "-30.00",
		"description":    "REWE Einkauf",
		"category_id":    catID,
		"date":           "2026-02-10",
	})
	tx = parseJSONObject(t, text)
	if tx["description"] != "REWE Einkauf" {
		t.Errorf("expected 'REWE Einkauf', got %v", tx["description"])
	}

	// Delete transaction
	text = callTool(t, session, "delete_transaction", map[string]any{
		"household_id":   hhID,
		"transaction_id": txID,
	})
	if text != "Transaction deleted" {
		t.Errorf("expected confirmation, got %v", text)
	}
}

// --- Recurring Expense CRUD ---

func TestMCPRecurringExpenseCRUD(t *testing.T) {
	_, session := setupMCPEnv(t)

	// Setup
	text := callTool(t, session, "create_household", map[string]any{
		"name": "Recurring Test", "currency": "EUR",
	})
	hh := parseJSONObject(t, text)
	hhID := int(hh["id"].(float64))

	text = callTool(t, session, "create_category", map[string]any{
		"household_id": hhID, "name": "Wohnen",
	})
	cat := parseJSONObject(t, text)
	catID := int(cat["id"].(float64))

	// Create recurring expense
	text = callTool(t, session, "create_recurring_expense", map[string]any{
		"household_id": hhID,
		"category_id":  catID,
		"name":         "Internet",
		"description":  "Fiber 1Gbit",
		"amount":       "-49.99",
		"frequency":    "monthly",
		"start_date":   "2026-01-01",
	})
	re := parseJSONObject(t, text)
	reID := int(re["id"].(float64))
	if re["name"] != "Internet" {
		t.Errorf("expected 'Internet', got %v", re["name"])
	}
	if re["frequency"] != "monthly" {
		t.Errorf("expected 'monthly', got %v", re["frequency"])
	}

	// List recurring expenses
	text = callTool(t, session, "list_recurring_expenses", map[string]any{"household_id": hhID})
	res := parseJSONArray(t, text)
	if len(res) != 1 {
		t.Errorf("expected 1 recurring expense, got %d", len(res))
	}

	// Update recurring expense
	text = callTool(t, session, "update_recurring_expense", map[string]any{
		"household_id": hhID,
		"recurring_id": reID,
		"name":         "Glasfaser",
		"amount":       "-59.99",
		"frequency":    "monthly",
		"category_id":  catID,
		"start_date":   "2026-01-01",
		"active":       true,
	})
	re = parseJSONObject(t, text)
	if re["name"] != "Glasfaser" {
		t.Errorf("expected 'Glasfaser', got %v", re["name"])
	}

	// Delete
	text = callTool(t, session, "delete_recurring_expense", map[string]any{
		"household_id": hhID,
		"recurring_id": reID,
	})
	if text != "Recurring expense deleted" {
		t.Errorf("expected confirmation, got %v", text)
	}

	// Verify gone
	text = callTool(t, session, "list_recurring_expenses", map[string]any{"household_id": hhID})
	res = parseJSONArray(t, text)
	if len(res) != 0 {
		t.Errorf("expected 0 recurring expenses, got %d", len(res))
	}
}

// --- Schedule Override CRUD ---

func TestMCPScheduleOverrideCRUD(t *testing.T) {
	_, session := setupMCPEnv(t)

	// Setup
	text := callTool(t, session, "create_household", map[string]any{
		"name": "Override Test", "currency": "EUR",
	})
	hh := parseJSONObject(t, text)
	hhID := int(hh["id"].(float64))

	text = callTool(t, session, "create_category", map[string]any{
		"household_id": hhID, "name": "Wohnen",
	})
	cat := parseJSONObject(t, text)
	catID := int(cat["id"].(float64))

	text = callTool(t, session, "create_recurring_expense", map[string]any{
		"household_id": hhID,
		"category_id":  catID,
		"name":         "Miete",
		"amount":       "-800.00",
		"frequency":    "monthly",
		"start_date":   "2026-01-01",
	})
	re := parseJSONObject(t, text)
	reID := int(re["id"].(float64))

	// Create override
	text = callTool(t, session, "create_schedule_override", map[string]any{
		"household_id":   hhID,
		"recurring_id":   reID,
		"effective_date": "2026-07-01",
		"amount":         "-850.00",
		"frequency":      "monthly",
	})
	ov := parseJSONObject(t, text)
	ovID := int(ov["id"].(float64))
	if ov["amount"] != "-850" {
		t.Errorf("expected '-850', got %v", ov["amount"])
	}

	// List overrides
	text = callTool(t, session, "list_schedule_overrides", map[string]any{
		"household_id": hhID,
		"recurring_id": reID,
	})
	ovs := parseJSONArray(t, text)
	if len(ovs) != 1 {
		t.Errorf("expected 1 override, got %d", len(ovs))
	}

	// Update override
	text = callTool(t, session, "update_schedule_override", map[string]any{
		"household_id":   hhID,
		"recurring_id":   reID,
		"override_id":    ovID,
		"effective_date": "2026-07-01",
		"amount":         "-900.00",
		"frequency":      "monthly",
	})
	ov = parseJSONObject(t, text)
	if ov["amount"] != "-900" {
		t.Errorf("expected '-900', got %v", ov["amount"])
	}

	// Delete override
	text = callTool(t, session, "delete_schedule_override", map[string]any{
		"household_id": hhID,
		"recurring_id": reID,
		"override_id":  ovID,
	})
	if text != "Schedule override deleted" {
		t.Errorf("expected confirmation, got %v", text)
	}
}

// --- Summary Test ---

func TestMCPMonthlySummary(t *testing.T) {
	_, session := setupMCPEnv(t)

	// Setup: household + category + transactions + recurring
	text := callTool(t, session, "create_household", map[string]any{
		"name": "Summary Test", "currency": "EUR",
	})
	hh := parseJSONObject(t, text)
	hhID := int(hh["id"].(float64))

	text = callTool(t, session, "create_category", map[string]any{
		"household_id": hhID, "name": "Gehalt",
	})
	incomeCat := parseJSONObject(t, text)
	incomeCatID := int(incomeCat["id"].(float64))

	text = callTool(t, session, "create_category", map[string]any{
		"household_id": hhID, "name": "Miete",
	})
	expenseCat := parseJSONObject(t, text)
	expenseCatID := int(expenseCat["id"].(float64))

	// Income transaction
	callTool(t, session, "create_transaction", map[string]any{
		"household_id": hhID,
		"category_id":  incomeCatID,
		"amount":       "3000.00",
		"description":  "Gehalt Januar",
		"date":         "2026-01-01",
	})

	// Expense transaction
	callTool(t, session, "create_transaction", map[string]any{
		"household_id": hhID,
		"category_id":  expenseCatID,
		"amount":       "-50.00",
		"description":  "Nebenkosten",
		"date":         "2026-01-15",
	})

	// Recurring expense
	callTool(t, session, "create_recurring_expense", map[string]any{
		"household_id": hhID,
		"category_id":  expenseCatID,
		"name":         "Kaltmiete",
		"amount":       "-800.00",
		"frequency":    "monthly",
		"start_date":   "2026-01-01",
	})

	// Get summary
	text = callTool(t, session, "get_monthly_summary", map[string]any{
		"household_id": hhID,
		"month":        "2026-01",
	})
	summary := parseJSONObject(t, text)

	if summary["month"] != "2026-01" {
		t.Errorf("expected month '2026-01', got %v", summary["month"])
	}
	if summary["one_time_income"] != "3000" {
		t.Errorf("expected one_time_income '3000', got %v", summary["one_time_income"])
	}
	if summary["one_time_expenses"] != "-50" {
		t.Errorf("expected one_time_expenses '-50', got %v", summary["one_time_expenses"])
	}
	if summary["recurring_total"] != "-800" {
		t.Errorf("expected recurring_total '-800', got %v", summary["recurring_total"])
	}

	// Verify category breakdown exists
	breakdown, ok := summary["category_breakdown"].([]any)
	if !ok || len(breakdown) == 0 {
		t.Errorf("expected non-empty category_breakdown")
	}
}

// --- MCP Resources Test ---

func TestMCPResources(t *testing.T) {
	_, session := setupMCPEnv(t)

	// Setup: create a household with data
	text := callTool(t, session, "create_household", map[string]any{
		"name": "Resource Test", "currency": "EUR",
	})
	hh := parseJSONObject(t, text)
	hhID := int(hh["id"].(float64))

	text = callTool(t, session, "create_category", map[string]any{
		"household_id": hhID, "name": "Test",
	})
	cat := parseJSONObject(t, text)
	catID := int(cat["id"].(float64))

	callTool(t, session, "create_transaction", map[string]any{
		"household_id": hhID,
		"category_id":  catID,
		"amount":       "-100.00",
		"description":  "Test Tx",
		"date":         "2026-01-15",
	})

	// Read households resource
	result, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "money-tracker://households",
	})
	if err != nil {
		t.Fatalf("reading households resource: %v", err)
	}
	if len(result.Contents) != 1 {
		t.Fatalf("expected 1 resource content, got %d", len(result.Contents))
	}
	if result.Contents[0].MIMEType != "application/json" {
		t.Errorf("expected application/json, got %s", result.Contents[0].MIMEType)
	}

	var resourceHouseholds []map[string]any
	if err := json.Unmarshal([]byte(result.Contents[0].Text), &resourceHouseholds); err != nil {
		t.Fatalf("parsing households resource: %v", err)
	}
	if len(resourceHouseholds) != 1 {
		t.Errorf("expected 1 household in resource, got %d", len(resourceHouseholds))
	}

	// Read summary resource template
	result, err = session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "money-tracker://households/" + itoa(hhID) + "/summary/2026-01",
	})
	if err != nil {
		t.Fatalf("reading summary resource: %v", err)
	}
	if len(result.Contents) != 1 {
		t.Fatalf("expected 1 resource content, got %d", len(result.Contents))
	}

	var summaryResource map[string]any
	if err := json.Unmarshal([]byte(result.Contents[0].Text), &summaryResource); err != nil {
		t.Fatalf("parsing summary resource: %v", err)
	}
	if summaryResource["month"] != "2026-01" {
		t.Errorf("expected month '2026-01', got %v", summaryResource["month"])
	}
}

// --- MCP Prompts Test ---

func TestMCPPrompts(t *testing.T) {
	_, session := setupMCPEnv(t)

	// Setup
	text := callTool(t, session, "create_household", map[string]any{
		"name": "Prompt Test", "currency": "EUR",
	})
	hh := parseJSONObject(t, text)
	hhID := int(hh["id"].(float64))

	callTool(t, session, "create_category", map[string]any{
		"household_id": hhID, "name": "Essen",
	})

	// Test monthly_report prompt
	promptResult, err := session.GetPrompt(context.Background(), &mcp.GetPromptParams{
		Name: "monthly_report",
		Arguments: map[string]string{
			"household_id": itoa(hhID),
			"month":        "2026-01",
		},
	})
	if err != nil {
		t.Fatalf("getting monthly_report prompt: %v", err)
	}
	if len(promptResult.Messages) != 1 {
		t.Fatalf("expected 1 prompt message, got %d", len(promptResult.Messages))
	}
	if promptResult.Messages[0].Role != "user" {
		t.Errorf("expected role 'user', got %s", promptResult.Messages[0].Role)
	}

	// Test budget_analysis prompt
	promptResult, err = session.GetPrompt(context.Background(), &mcp.GetPromptParams{
		Name: "budget_analysis",
		Arguments: map[string]string{
			"household_id": itoa(hhID),
			"months":       "2",
		},
	})
	if err != nil {
		t.Fatalf("getting budget_analysis prompt: %v", err)
	}
	if len(promptResult.Messages) != 1 {
		t.Fatalf("expected 1 prompt message, got %d", len(promptResult.Messages))
	}

	// Test categorize_transaction prompt
	promptResult, err = session.GetPrompt(context.Background(), &mcp.GetPromptParams{
		Name: "categorize_transaction",
		Arguments: map[string]string{
			"household_id": itoa(hhID),
			"description":  "REWE Einkauf",
			"amount":       "-45.50",
		},
	})
	if err != nil {
		t.Fatalf("getting categorize_transaction prompt: %v", err)
	}
	if len(promptResult.Messages) != 1 {
		t.Fatalf("expected 1 prompt message, got %d", len(promptResult.Messages))
	}

	// Verify prompt content contains the description
	tc, ok := promptResult.Messages[0].Content.(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", promptResult.Messages[0].Content)
	}
	if tc.Text == "" {
		t.Errorf("expected non-empty prompt text")
	}
}

// --- Tool Listing Test ---

func TestMCPToolListing(t *testing.T) {
	_, session := setupMCPEnv(t)

	tools, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("listing tools: %v", err)
	}

	expectedTools := []string{
		"list_households", "create_household", "update_household", "delete_household",
		"list_categories", "create_category", "update_category", "delete_category",
		"list_transactions", "create_transaction", "update_transaction", "delete_transaction",
		"list_recurring_expenses", "create_recurring_expense", "update_recurring_expense", "delete_recurring_expense",
		"list_schedule_overrides", "create_schedule_override", "update_schedule_override", "delete_schedule_override",
		"get_monthly_summary",
	}

	toolNames := make(map[string]bool)
	for _, tool := range tools.Tools {
		toolNames[tool.Name] = true
	}

	for _, expected := range expectedTools {
		if !toolNames[expected] {
			t.Errorf("missing expected tool: %s", expected)
		}
	}

	if len(tools.Tools) != len(expectedTools) {
		t.Errorf("expected %d tools, got %d", len(expectedTools), len(tools.Tools))
	}
}
