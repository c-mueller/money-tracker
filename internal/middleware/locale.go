package middleware

import (
	"icekalt.dev/money-tracker/internal/i18n"

	"github.com/labstack/echo/v4"
)

const LocaleContextKey = "locale"

// Locale returns middleware that sets the locale on the Echo context.
// It checks the Accept-Language header first, falling back to the server default.
func Locale(defaultLocale i18n.Locale) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			locale := defaultLocale
			if header := c.Request().Header.Get("Accept-Language"); header != "" {
				if parsed, ok := i18n.ParseAcceptLanguage(header); ok {
					locale = parsed
				}
			}
			c.Set(LocaleContextKey, locale)
			return next(c)
		}
	}
}
