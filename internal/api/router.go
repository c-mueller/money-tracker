package api

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"icekalt.dev/money-tracker/internal/auth"
	"icekalt.dev/money-tracker/internal/devmode"
	gql "icekalt.dev/money-tracker/internal/graphql"
	mw "icekalt.dev/money-tracker/internal/middleware"
)

func (s *Server) setupRoutes() {
	// Template renderer
	renderer, err := NewTemplateRenderer(s.i18nBundle, s.defaultLocale)
	if err != nil {
		s.logger.Fatal("failed to setup templates: " + err.Error())
	}
	s.echo.Renderer = renderer
	s.renderer = renderer

	// Static files
	s.setupStatic()

	// Auth routes (no auth middleware)
	if s.authHandler != nil {
		s.echo.GET("/auth/login", s.authHandler.HandleLogin)
		s.echo.GET("/auth/callback", s.authHandler.HandleCallback)
		s.echo.GET("/auth/logout", s.authHandler.HandleLogout)
	}

	// Auth middleware for all protected routes
	authMW := mw.Auth(s.sessionStore, s.services.APIToken, s.devUserID)

	// Locale middleware
	localeMW := mw.Locale(s.defaultLocale)

	// --- API Routes ---
	apiGroup := s.echo.Group("/api/v1")
	apiGroup.GET("/health", s.handleHealth)
	apiGroup.GET("/openapi.yaml", s.handleOpenAPISpec)
	apiGroup.Use(authMW)

	// Households
	apiGroup.GET("/households", s.handleListHouseholds)
	apiGroup.POST("/households", s.handleCreateHousehold)
	apiGroup.PUT("/households/:id", s.handleUpdateHousehold)
	apiGroup.DELETE("/households/:id", s.handleDeleteHousehold)

	// Categories
	apiGroup.GET("/households/:id/categories", s.handleListCategories)
	apiGroup.POST("/households/:id/categories", s.handleCreateCategory)
	apiGroup.PUT("/households/:id/categories/:categoryId", s.handleUpdateCategory)
	apiGroup.DELETE("/households/:id/categories/:categoryId", s.handleDeleteCategory)

	// Transactions
	apiGroup.GET("/households/:id/transactions", s.handleListTransactions)
	apiGroup.POST("/households/:id/transactions", s.handleCreateTransaction)
	apiGroup.PUT("/households/:id/transactions/:transactionId", s.handleUpdateTransaction)
	apiGroup.DELETE("/households/:id/transactions/:transactionId", s.handleDeleteTransaction)

	// Recurring Expenses
	apiGroup.GET("/households/:id/recurring-expenses", s.handleListRecurringExpenses)
	apiGroup.POST("/households/:id/recurring-expenses", s.handleCreateRecurringExpense)
	apiGroup.PUT("/households/:id/recurring-expenses/:recurringId", s.handleUpdateRecurringExpense)
	apiGroup.DELETE("/households/:id/recurring-expenses/:recurringId", s.handleDeleteRecurringExpense)

	// Summary
	apiGroup.GET("/households/:id/summary", s.handleGetSummary)

	// API Tokens
	apiGroup.GET("/tokens", s.handleListTokens)
	apiGroup.POST("/tokens", s.handleCreateToken)
	apiGroup.DELETE("/tokens/:tokenId", s.handleDeleteToken)

	// --- GraphQL ---
	gqlHandler := handler.NewDefaultServer(gql.NewExecutableSchema(gql.Config{
		Resolvers: &gql.Resolver{
			HouseholdSvc:        s.services.Household,
			CategorySvc:         s.services.Category,
			TransactionSvc:      s.services.Transaction,
			RecurringExpenseSvc: s.services.RecurringExpense,
			SummarySvc:          s.services.Summary,
		},
	}))

	graphqlGroup := s.echo.Group("/graphql")
	graphqlGroup.Use(authMW)
	graphqlGroup.POST("", echo.WrapHandler(gqlHandler))

	// Playground (behind auth so it works with browser sessions)
	if devmode.Enabled {
		playgroundGroup := s.echo.Group("/playground")
		playgroundGroup.Use(authMW)
		playgroundGroup.GET("", echo.WrapHandler(playground.Handler("GraphQL", "/graphql")))
	}

	// --- Swagger UI (behind auth) ---
	swaggerGroup := s.echo.Group("/swagger")
	swaggerGroup.Use(authMW)
	swaggerGroup.GET("", s.handleSwaggerUI)

	// --- Web Routes ---
	webGroup := s.echo.Group("")
	webGroup.Use(localeMW)
	webGroup.Use(authMW)

	webGroup.GET("/", s.handleWebDashboard)
	webGroup.GET("/households/new", s.handleWebHouseholdNew)
	webGroup.POST("/households", s.handleWebHouseholdCreate)
	webGroup.GET("/households/:id", s.handleWebHouseholdDetail)
	webGroup.GET("/households/:id/transactions/new", s.handleWebTransactionNew)
	webGroup.POST("/households/:id/transactions", s.handleWebTransactionCreate)
	webGroup.GET("/households/:id/transactions/:transactionId/edit", s.handleWebTransactionEdit)
	webGroup.POST("/households/:id/transactions/:transactionId", s.handleWebTransactionUpdate)
	webGroup.GET("/households/:id/settings", s.handleWebHouseholdSettings)
	webGroup.POST("/households/:id/settings", s.handleWebHouseholdSettingsUpdate)
	webGroup.GET("/households/:id/categories", s.handleWebCategoryList)
	webGroup.POST("/households/:id/categories", s.handleWebCategoryCreate)
	webGroup.GET("/households/:id/categories/:categoryId/edit", s.handleWebCategoryEdit)
	webGroup.POST("/households/:id/categories/:categoryId", s.handleWebCategoryUpdate)
	webGroup.GET("/households/:id/recurring", s.handleWebRecurringList)
	webGroup.GET("/households/:id/recurring/new", s.handleWebRecurringNew)
	webGroup.POST("/households/:id/recurring", s.handleWebRecurringCreate)
	webGroup.GET("/households/:id/recurring/:recurringId/edit", s.handleWebRecurringEdit)
	webGroup.POST("/households/:id/recurring/:recurringId", s.handleWebRecurringUpdate)
	webGroup.GET("/tokens", s.handleWebTokenList)
	webGroup.POST("/tokens", s.handleWebTokenCreate)
}

// SetupAuth configures authentication for the server.
func (s *Server) SetupAuth(oidcCfg *auth.OIDCConfig, store sessions.Store, devUserID int) {
	s.sessionStore = store
	s.devUserID = devUserID
	if oidcCfg != nil {
		s.authHandler = NewAuthHandler(oidcCfg, store, s.services)
	}
}
