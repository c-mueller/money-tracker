package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps the MCP server and API client.
type Server struct {
	mcpServer *mcp.Server
	client    *Client
}

// NewServer creates a new MCP server with all tools, resources, and prompts registered.
func NewServer(client *Client, version string) *Server {
	s := &Server{
		mcpServer: mcp.NewServer(&mcp.Implementation{
			Name:    "money-tracker",
			Version: version,
		}, nil),
		client: client,
	}

	s.registerHouseholdTools()
	s.registerCategoryTools()
	s.registerTransactionTools()
	s.registerRecurringExpenseTools()
	s.registerScheduleOverrideTools()
	s.registerSummaryTools()
	s.registerResources()
	s.registerPrompts()

	return s
}

// Run starts the MCP server on stdio transport.
func (s *Server) Run(ctx context.Context) error {
	return s.mcpServer.Run(ctx, &mcp.StdioTransport{})
}

// InProcessTransport returns a pair of in-memory transports for testing.
// The caller should use the returned transport to connect an MCP client,
// while the server-side transport is connected automatically.
func (s *Server) InProcessTransport() mcp.Transport {
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	go s.mcpServer.Connect(context.Background(), serverTransport, nil) //nolint:errcheck
	return clientTransport
}

// --- Household Tools ---

type listHouseholdsArgs struct{}

type createHouseholdArgs struct {
	Name        string `json:"name" jsonschema:"required,Name of the household"`
	Description string `json:"description,omitempty" jsonschema:"Description of the household"`
	Currency    string `json:"currency" jsonschema:"required,ISO 4217 currency code (e.g. EUR)"`
	Icon        string `json:"icon,omitempty" jsonschema:"Material icon name"`
}

type updateHouseholdArgs struct {
	ID          int    `json:"id" jsonschema:"required,Household ID"`
	Name        string `json:"name,omitempty" jsonschema:"New name"`
	Description string `json:"description,omitempty" jsonschema:"New description"`
	Currency    string `json:"currency,omitempty" jsonschema:"New ISO 4217 currency code"`
	Icon        string `json:"icon,omitempty" jsonschema:"New material icon name"`
}

type deleteHouseholdArgs struct {
	ID int `json:"id" jsonschema:"required,Household ID"`
}

func (s *Server) registerHouseholdTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_households",
		Description: "List all households for the authenticated user",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args listHouseholdsArgs) (*mcp.CallToolResult, any, error) {
		households, err := s.client.ListHouseholds()
		if err != nil {
			return nil, nil, err
		}
		return textResult(households)
	})

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "create_household",
		Description: "Create a new household",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args createHouseholdArgs) (*mcp.CallToolResult, any, error) {
		household, err := s.client.CreateHousehold(toMap(args))
		if err != nil {
			return nil, nil, err
		}
		return textResult(household)
	})

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "update_household",
		Description: "Update an existing household",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args updateHouseholdArgs) (*mcp.CallToolResult, any, error) {
		m := toMap(args)
		delete(m, "id")
		household, err := s.client.UpdateHousehold(args.ID, m)
		if err != nil {
			return nil, nil, err
		}
		return textResult(household)
	})

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "delete_household",
		Description: "Delete a household and all its data (categories, transactions, recurring expenses)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args deleteHouseholdArgs) (*mcp.CallToolResult, any, error) {
		if err := s.client.DeleteHousehold(args.ID); err != nil {
			return nil, nil, err
		}
		return confirmResult("Household deleted")
	})
}

// --- Category Tools ---

type listCategoriesArgs struct {
	HouseholdID int `json:"household_id" jsonschema:"required,Household ID"`
}

type createCategoryArgs struct {
	HouseholdID int    `json:"household_id" jsonschema:"required,Household ID"`
	Name        string `json:"name" jsonschema:"required,Category name"`
	Icon        string `json:"icon,omitempty" jsonschema:"Material icon name"`
}

type updateCategoryArgs struct {
	HouseholdID int    `json:"household_id" jsonschema:"required,Household ID"`
	CategoryID  int    `json:"category_id" jsonschema:"required,Category ID"`
	Name        string `json:"name,omitempty" jsonschema:"New name"`
	Icon        string `json:"icon,omitempty" jsonschema:"New material icon name"`
}

type deleteCategoryArgs struct {
	HouseholdID int `json:"household_id" jsonschema:"required,Household ID"`
	CategoryID  int `json:"category_id" jsonschema:"required,Category ID"`
}

func (s *Server) registerCategoryTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_categories",
		Description: "List categories for a household",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args listCategoriesArgs) (*mcp.CallToolResult, any, error) {
		categories, err := s.client.ListCategories(args.HouseholdID)
		if err != nil {
			return nil, nil, err
		}
		return textResult(categories)
	})

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "create_category",
		Description: "Create a new category in a household",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args createCategoryArgs) (*mcp.CallToolResult, any, error) {
		m := toMap(args)
		delete(m, "household_id")
		category, err := s.client.CreateCategory(args.HouseholdID, m)
		if err != nil {
			return nil, nil, err
		}
		return textResult(category)
	})

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "update_category",
		Description: "Update a category in a household",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args updateCategoryArgs) (*mcp.CallToolResult, any, error) {
		m := toMap(args)
		delete(m, "household_id")
		delete(m, "category_id")
		category, err := s.client.UpdateCategory(args.HouseholdID, args.CategoryID, m)
		if err != nil {
			return nil, nil, err
		}
		return textResult(category)
	})

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "delete_category",
		Description: "Delete a category from a household",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args deleteCategoryArgs) (*mcp.CallToolResult, any, error) {
		if err := s.client.DeleteCategory(args.HouseholdID, args.CategoryID); err != nil {
			return nil, nil, err
		}
		return confirmResult("Category deleted")
	})
}

// --- Transaction Tools ---

type listTransactionsArgs struct {
	HouseholdID int    `json:"household_id" jsonschema:"required,Household ID"`
	Month       string `json:"month,omitempty" jsonschema:"Month in YYYY-MM format (default: current month)"`
}

type createTransactionArgs struct {
	HouseholdID int    `json:"household_id" jsonschema:"required,Household ID"`
	CategoryID  int    `json:"category_id" jsonschema:"required,Category ID"`
	Amount      string `json:"amount" jsonschema:"required,Decimal amount as string (negative=expense positive=income)"`
	Description string `json:"description" jsonschema:"required,Transaction description"`
	Date        string `json:"date" jsonschema:"required,Date in YYYY-MM-DD format"`
}

type updateTransactionArgs struct {
	HouseholdID   int    `json:"household_id" jsonschema:"required,Household ID"`
	TransactionID int    `json:"transaction_id" jsonschema:"required,Transaction ID"`
	CategoryID    int    `json:"category_id,omitempty" jsonschema:"New category ID"`
	Amount        string `json:"amount,omitempty" jsonschema:"New decimal amount"`
	Description   string `json:"description,omitempty" jsonschema:"New description"`
	Date          string `json:"date,omitempty" jsonschema:"New date in YYYY-MM-DD format"`
}

type deleteTransactionArgs struct {
	HouseholdID   int `json:"household_id" jsonschema:"required,Household ID"`
	TransactionID int `json:"transaction_id" jsonschema:"required,Transaction ID"`
}

func (s *Server) registerTransactionTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_transactions",
		Description: "List transactions for a household in a given month",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args listTransactionsArgs) (*mcp.CallToolResult, any, error) {
		txs, err := s.client.ListTransactions(args.HouseholdID, args.Month)
		if err != nil {
			return nil, nil, err
		}
		return textResult(txs)
	})

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "create_transaction",
		Description: "Create a new transaction in a household",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args createTransactionArgs) (*mcp.CallToolResult, any, error) {
		m := toMap(args)
		delete(m, "household_id")
		tx, err := s.client.CreateTransaction(args.HouseholdID, m)
		if err != nil {
			return nil, nil, err
		}
		return textResult(tx)
	})

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "update_transaction",
		Description: "Update a transaction in a household",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args updateTransactionArgs) (*mcp.CallToolResult, any, error) {
		m := toMap(args)
		delete(m, "household_id")
		delete(m, "transaction_id")
		tx, err := s.client.UpdateTransaction(args.HouseholdID, args.TransactionID, m)
		if err != nil {
			return nil, nil, err
		}
		return textResult(tx)
	})

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "delete_transaction",
		Description: "Delete a transaction from a household",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args deleteTransactionArgs) (*mcp.CallToolResult, any, error) {
		if err := s.client.DeleteTransaction(args.HouseholdID, args.TransactionID); err != nil {
			return nil, nil, err
		}
		return confirmResult("Transaction deleted")
	})
}

// --- Recurring Expense Tools ---

type listRecurringExpensesArgs struct {
	HouseholdID int `json:"household_id" jsonschema:"required,Household ID"`
}

type createRecurringExpenseArgs struct {
	HouseholdID int     `json:"household_id" jsonschema:"required,Household ID"`
	CategoryID  int     `json:"category_id" jsonschema:"required,Category ID"`
	Name        string  `json:"name" jsonschema:"required,Name of the recurring entry"`
	Description string  `json:"description,omitempty" jsonschema:"Description"`
	Amount      string  `json:"amount" jsonschema:"required,Decimal amount (negative=expense positive=income)"`
	Frequency   string  `json:"frequency" jsonschema:"required,Frequency: daily|weekday|weekly|biweekly|monthly|quarterly|yearly"`
	StartDate   string  `json:"start_date" jsonschema:"required,Start date in YYYY-MM-DD format"`
	EndDate     *string `json:"end_date,omitempty" jsonschema:"End date in YYYY-MM-DD format (empty=indefinite)"`
	Active      *bool   `json:"active,omitempty" jsonschema:"Whether the entry is active (default: true)"`
}

type updateRecurringExpenseArgs struct {
	HouseholdID int     `json:"household_id" jsonschema:"required,Household ID"`
	RecurringID int     `json:"recurring_id" jsonschema:"required,Recurring expense ID"`
	CategoryID  int     `json:"category_id,omitempty" jsonschema:"New category ID"`
	Name        string  `json:"name,omitempty" jsonschema:"New name"`
	Description string  `json:"description,omitempty" jsonschema:"New description"`
	Amount      string  `json:"amount,omitempty" jsonschema:"New decimal amount"`
	Frequency   string  `json:"frequency,omitempty" jsonschema:"New frequency"`
	Active      *bool   `json:"active,omitempty" jsonschema:"Active status"`
	StartDate   string  `json:"start_date,omitempty" jsonschema:"New start date"`
	EndDate     *string `json:"end_date,omitempty" jsonschema:"New end date"`
}

type deleteRecurringExpenseArgs struct {
	HouseholdID int `json:"household_id" jsonschema:"required,Household ID"`
	RecurringID int `json:"recurring_id" jsonschema:"required,Recurring expense ID"`
}

func (s *Server) registerRecurringExpenseTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_recurring_expenses",
		Description: "List recurring expenses for a household",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args listRecurringExpensesArgs) (*mcp.CallToolResult, any, error) {
		expenses, err := s.client.ListRecurringExpenses(args.HouseholdID)
		if err != nil {
			return nil, nil, err
		}
		return textResult(expenses)
	})

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "create_recurring_expense",
		Description: "Create a new recurring expense in a household",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args createRecurringExpenseArgs) (*mcp.CallToolResult, any, error) {
		m := toMap(args)
		delete(m, "household_id")
		expense, err := s.client.CreateRecurringExpense(args.HouseholdID, m)
		if err != nil {
			return nil, nil, err
		}
		return textResult(expense)
	})

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "update_recurring_expense",
		Description: "Update a recurring expense in a household",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args updateRecurringExpenseArgs) (*mcp.CallToolResult, any, error) {
		m := toMap(args)
		delete(m, "household_id")
		delete(m, "recurring_id")
		expense, err := s.client.UpdateRecurringExpense(args.HouseholdID, args.RecurringID, m)
		if err != nil {
			return nil, nil, err
		}
		return textResult(expense)
	})

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "delete_recurring_expense",
		Description: "Delete a recurring expense from a household",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args deleteRecurringExpenseArgs) (*mcp.CallToolResult, any, error) {
		if err := s.client.DeleteRecurringExpense(args.HouseholdID, args.RecurringID); err != nil {
			return nil, nil, err
		}
		return confirmResult("Recurring expense deleted")
	})
}

// --- Schedule Override Tools ---

type listScheduleOverridesArgs struct {
	HouseholdID int `json:"household_id" jsonschema:"required,Household ID"`
	RecurringID int `json:"recurring_id" jsonschema:"required,Recurring expense ID"`
}

type createScheduleOverrideArgs struct {
	HouseholdID   int    `json:"household_id" jsonschema:"required,Household ID"`
	RecurringID   int    `json:"recurring_id" jsonschema:"required,Recurring expense ID"`
	EffectiveDate string `json:"effective_date" jsonschema:"required,Effective date in YYYY-MM-DD format"`
	Amount        string `json:"amount" jsonschema:"required,New decimal amount from this date"`
	Frequency     string `json:"frequency" jsonschema:"required,New frequency from this date"`
}

type updateScheduleOverrideArgs struct {
	HouseholdID   int    `json:"household_id" jsonschema:"required,Household ID"`
	RecurringID   int    `json:"recurring_id" jsonschema:"required,Recurring expense ID"`
	OverrideID    int    `json:"override_id" jsonschema:"required,Override ID"`
	EffectiveDate string `json:"effective_date,omitempty" jsonschema:"New effective date"`
	Amount        string `json:"amount,omitempty" jsonschema:"New amount"`
	Frequency     string `json:"frequency,omitempty" jsonschema:"New frequency"`
}

type deleteScheduleOverrideArgs struct {
	HouseholdID int `json:"household_id" jsonschema:"required,Household ID"`
	RecurringID int `json:"recurring_id" jsonschema:"required,Recurring expense ID"`
	OverrideID  int `json:"override_id" jsonschema:"required,Override ID"`
}

func (s *Server) registerScheduleOverrideTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_schedule_overrides",
		Description: "List schedule overrides for a recurring expense",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args listScheduleOverridesArgs) (*mcp.CallToolResult, any, error) {
		overrides, err := s.client.ListScheduleOverrides(args.HouseholdID, args.RecurringID)
		if err != nil {
			return nil, nil, err
		}
		return textResult(overrides)
	})

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "create_schedule_override",
		Description: "Create a schedule override for a recurring expense (changes amount/frequency from a specific date)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args createScheduleOverrideArgs) (*mcp.CallToolResult, any, error) {
		m := toMap(args)
		delete(m, "household_id")
		delete(m, "recurring_id")
		override, err := s.client.CreateScheduleOverride(args.HouseholdID, args.RecurringID, m)
		if err != nil {
			return nil, nil, err
		}
		return textResult(override)
	})

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "update_schedule_override",
		Description: "Update a schedule override",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args updateScheduleOverrideArgs) (*mcp.CallToolResult, any, error) {
		m := toMap(args)
		delete(m, "household_id")
		delete(m, "recurring_id")
		delete(m, "override_id")
		override, err := s.client.UpdateScheduleOverride(args.HouseholdID, args.RecurringID, args.OverrideID, m)
		if err != nil {
			return nil, nil, err
		}
		return textResult(override)
	})

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "delete_schedule_override",
		Description: "Delete a schedule override",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args deleteScheduleOverrideArgs) (*mcp.CallToolResult, any, error) {
		if err := s.client.DeleteScheduleOverride(args.HouseholdID, args.RecurringID, args.OverrideID); err != nil {
			return nil, nil, err
		}
		return confirmResult("Schedule override deleted")
	})
}

// --- Summary Tool ---

type getMonthlySummaryArgs struct {
	HouseholdID int    `json:"household_id" jsonschema:"required,Household ID"`
	Month       string `json:"month,omitempty" jsonschema:"Month in YYYY-MM format (default: current month)"`
}

func (s *Server) registerSummaryTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "get_monthly_summary",
		Description: "Get a monthly financial summary for a household including income, expenses, recurring totals, and category breakdown",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args getMonthlySummaryArgs) (*mcp.CallToolResult, any, error) {
		summary, err := s.client.GetSummary(args.HouseholdID, args.Month)
		if err != nil {
			return nil, nil, err
		}
		return textResult(summary)
	})
}

// --- Resources ---

func (s *Server) registerResources() {
	s.mcpServer.AddResource(&mcp.Resource{
		URI:         "money-tracker://households",
		Name:        "Households Overview",
		Description: "List of all households with basic info",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		households, err := s.client.ListHouseholds()
		if err != nil {
			return nil, err
		}
		text, err := toJSON(households)
		if err != nil {
			return nil, err
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     text,
			}},
		}, nil
	})

	s.mcpServer.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "money-tracker://households/{household_id}/summary/{month}",
		Name:        "Monthly Summary",
		Description: "Monthly financial summary for a specific household",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		// Parse household_id and month from the URI
		var householdID int
		var month string
		_, err := fmt.Sscanf(req.Params.URI, "money-tracker://households/%d/summary/%s", &householdID, &month)
		if err != nil {
			return nil, fmt.Errorf("invalid resource URI: %s", req.Params.URI)
		}
		summary, err := s.client.GetSummary(householdID, month)
		if err != nil {
			return nil, err
		}
		text, err := toJSON(summary)
		if err != nil {
			return nil, err
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{
				URI:      req.Params.URI,
				MIMEType: "application/json",
				Text:     text,
			}},
		}, nil
	})
}

// --- Prompts ---

func (s *Server) registerPrompts() {
	s.mcpServer.AddPrompt(&mcp.Prompt{
		Name:        "monthly_report",
		Description: "Generate a formatted monthly financial report",
		Arguments: []*mcp.PromptArgument{
			{Name: "household_id", Description: "Household ID", Required: true},
			{Name: "month", Description: "Month in YYYY-MM format (default: current month)", Required: false},
		},
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		householdID, month, err := parseHouseholdMonth(req.Params.Arguments)
		if err != nil {
			return nil, err
		}
		if month == "" {
			month = time.Now().Format("2006-01")
		}

		summary, err := s.client.GetSummary(householdID, month)
		if err != nil {
			return nil, err
		}
		txs, err := s.client.ListTransactions(householdID, month)
		if err != nil {
			return nil, err
		}
		recurring, err := s.client.ListRecurringExpenses(householdID)
		if err != nil {
			return nil, err
		}

		summaryJSON, _ := toJSON(summary)
		txsJSON, _ := toJSON(txs)
		recurringJSON, _ := toJSON(recurring)

		return &mcp.GetPromptResult{
			Description: fmt.Sprintf("Monthly financial report for %s", month),
			Messages: []*mcp.PromptMessage{{
				Role: "user",
				Content: &mcp.TextContent{
					Text: fmt.Sprintf(`Please create a well-formatted monthly financial report for %s.

## Monthly Summary
%s

## Transactions
%s

## Recurring Expenses
%s

Please format this as a clear, readable report with:
- Overall income vs expenses summary
- Category-by-category breakdown
- Notable transactions
- Recurring expense overview
- Net result for the month`, month, summaryJSON, txsJSON, recurringJSON),
				},
			}},
		}, nil
	})

	s.mcpServer.AddPrompt(&mcp.Prompt{
		Name:        "budget_analysis",
		Description: "Analyze spending patterns and provide recommendations",
		Arguments: []*mcp.PromptArgument{
			{Name: "household_id", Description: "Household ID", Required: true},
			{Name: "months", Description: "Number of months to analyze (default: 3)", Required: false},
		},
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		householdID, _, err := parseHouseholdMonth(req.Params.Arguments)
		if err != nil {
			return nil, err
		}
		monthsStr := req.Params.Arguments["months"]
		months := 3
		if monthsStr != "" {
			_, _ = fmt.Sscanf(monthsStr, "%d", &months)
		}

		var summaries []string
		now := time.Now()
		for i := 0; i < months; i++ {
			m := now.AddDate(0, -i, 0).Format("2006-01")
			summary, err := s.client.GetSummary(householdID, m)
			if err != nil {
				continue
			}
			j, _ := toJSON(summary)
			summaries = append(summaries, j)
		}

		return &mcp.GetPromptResult{
			Description: fmt.Sprintf("Budget analysis for the last %d months", months),
			Messages: []*mcp.PromptMessage{{
				Role: "user",
				Content: &mcp.TextContent{
					Text: fmt.Sprintf(`Please analyze the spending patterns for the last %d months and provide recommendations.

## Monthly Summaries
%s

Please provide:
- Spending trends over time
- Categories with increasing/decreasing spend
- Suggestions for savings
- Comparison of recurring vs one-time expenses
- Overall financial health assessment`, months, formatSummaries(summaries)),
				},
			}},
		}, nil
	})

	s.mcpServer.AddPrompt(&mcp.Prompt{
		Name:        "categorize_transaction",
		Description: "Suggest the best category for a transaction based on its description",
		Arguments: []*mcp.PromptArgument{
			{Name: "household_id", Description: "Household ID", Required: true},
			{Name: "description", Description: "Transaction description", Required: true},
			{Name: "amount", Description: "Transaction amount", Required: true},
		},
	}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		householdID, _, err := parseHouseholdMonth(req.Params.Arguments)
		if err != nil {
			return nil, err
		}
		description := req.Params.Arguments["description"]
		amount := req.Params.Arguments["amount"]

		categories, err := s.client.ListCategories(householdID)
		if err != nil {
			return nil, err
		}
		categoriesJSON, _ := toJSON(categories)

		return &mcp.GetPromptResult{
			Description: "Category suggestion for a transaction",
			Messages: []*mcp.PromptMessage{{
				Role: "user",
				Content: &mcp.TextContent{
					Text: fmt.Sprintf(`Based on the following transaction, suggest the most appropriate category.

## Transaction
- Description: %s
- Amount: %s

## Available Categories
%s

Please suggest the best matching category and explain why.`, description, amount, categoriesJSON),
				},
			}},
		}, nil
	})
}

// --- Helpers ---

func textResult(v any) (*mcp.CallToolResult, any, error) {
	text, err := toJSON(v)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}, nil, nil
}

func confirmResult(msg string) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
	}, nil, nil
}

func toJSON(v any) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling JSON: %w", err)
	}
	return string(b), nil
}

func toMap(v any) map[string]any {
	b, _ := json.Marshal(v)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	return m
}

func parseHouseholdMonth(args map[string]string) (int, string, error) {
	householdIDStr := args["household_id"]
	var householdID int
	if _, err := fmt.Sscanf(householdIDStr, "%d", &householdID); err != nil {
		return 0, "", fmt.Errorf("invalid household_id: %s", householdIDStr)
	}
	return householdID, args["month"], nil
}

func formatSummaries(summaries []string) string {
	result := ""
	for i, s := range summaries {
		result += fmt.Sprintf("### Month %d\n%s\n\n", i+1, s)
	}
	return result
}
