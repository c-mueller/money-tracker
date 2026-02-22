package i18n

var deTranslations = map[string]string{
	// Navigation & Layout
	"dashboard":   "Dashboard",
	"login":       "Anmelden",
	"logout":      "Abmelden",
	"api_tokens":  "API-Tokens",
	"money_tracker": "Money Tracker",

	// Common actions
	"save":   "Speichern",
	"cancel": "Abbrechen",
	"delete": "Löschen",
	"edit":   "Bearbeiten",
	"add":    "Hinzufügen",
	"create": "Erstellen",
	"open":   "Öffnen",
	"copy":   "Kopieren",
	"copied": "Kopiert!",
	"today":  "Heute",

	// Common labels
	"name":        "Name",
	"amount":      "Betrag",
	"description": "Beschreibung",
	"category":    "Kategorie",
	"date":        "Datum",
	"frequency":   "Frequenz",
	"type":        "Typ",
	"currency":    "Währung",
	"icon":        "Icon",
	"active":      "Aktiv",
	"inactive":    "Inaktiv",
	"select":      "Auswählen…",

	// Household
	"new_household":        "Neuer Haushalt",
	"edit_household":       "Haushalt bearbeiten",
	"household":            "Haushalt",
	"no_households_empty":  "Noch keine Haushalte vorhanden. Erstelle deinen ersten Haushalt.",
	"custom_currency":      "Benutzerdefiniert…",
	"delete_household":     "Haushalt löschen",

	// Transactions
	"transactions":          "Transaktionen",
	"new_transaction":       "Neue Transaktion",
	"edit_transaction":      "Transaktion bearbeiten",
	"add_transaction":       "Transaktion hinzufügen",
	"no_transactions_month": "Keine Transaktionen in diesem Monat.",
	"expense":               "Ausgabe",
	"income":                "Einnahmen",

	// Recurring
	"recurring":                "Wiederkehrend",
	"new_recurring":            "Neue wiederkehrende Transaktion",
	"edit_recurring":           "Wiederkehrende Transaktion bearbeiten",
	"add_recurring":            "Wiederkehrende Transaktion hinzufügen",
	"no_recurring_empty":       "Noch keine wiederkehrenden Transaktionen.",
	"start_date":               "Startdatum",
	"end_date":                 "Enddatum",
	"end_date_optional":        "optional",

	// Categories
	"categories":              "Kategorien",
	"edit_category":            "Kategorie bearbeiten",
	"new_category_placeholder": "Neuer Kategoriename",
	"new_category":             "+ Neue Kategorie…",
	"category_name":            "Kategoriename",
	"no_categories_empty":      "Noch keine Kategorien.",

	// Settings
	"settings":     "Einstellungen",
	"danger_zone":  "Gefahrenzone",

	// Summary
	"one_time":  "Einmalig",
	"month":     "Monat",

	// Tokens
	"token_name_placeholder": "Tokenname",
	"new_token_created":      "Neuer Token erstellt! Kopiere ihn jetzt — er wird nicht erneut angezeigt.",
	"no_tokens_empty":        "Noch keine API-Tokens.",
	"created":                "Erstellt",
	"last_used":              "Zuletzt verwendet",
	"never":                  "Nie",

	// Login
	"sign_in_message": "Melde dich an, um deine Haushaltsbudgets zu verwalten.",
	"sign_in_oidc":    "Mit OIDC anmelden",

	// Confirmations
	"delete_transaction_confirm": "Diese Transaktion löschen?",
	"delete_category_confirm":    "Kategorie '%s' löschen?",
	"delete_recurring_confirm":   "'%s' löschen?",
	"delete_token_confirm":       "Token '%s' löschen?",
	"delete_household_confirm":   "Diesen Haushalt und alle Daten löschen?",

	// Errors (JS)
	"error_prefix": "Fehler: ",
}
