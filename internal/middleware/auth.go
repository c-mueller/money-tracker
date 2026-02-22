package middleware

import (
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"icekalt.dev/money-tracker/internal/auth"
	"icekalt.dev/money-tracker/internal/devmode"
	"icekalt.dev/money-tracker/internal/service"
)

func Auth(store sessions.Store, tokenSvc *service.APITokenService, devUserID int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Dev mode: auto-auth
			if devmode.Enabled {
				c.Set(UserIDContextKey, devUserID)
				ctx := service.WithUserID(c.Request().Context(), devUserID)
				c.SetRequest(c.Request().WithContext(ctx))
				return next(c)
			}

			// Check Bearer token
			authHeader := c.Request().Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				apiToken, err := tokenSvc.ValidateToken(c.Request().Context(), token)
				if err != nil {
					return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
				}
				c.Set(UserIDContextKey, apiToken.UserID)
				ctx := service.WithUserID(c.Request().Context(), apiToken.UserID)
				c.SetRequest(c.Request().WithContext(ctx))
				return next(c)
			}

			// Check session cookie
			session, err := store.Get(c.Request(), auth.SessionName)
			if err == nil {
				if userID, ok := session.Values[auth.SessionKeyUser].(int); ok {
					c.Set(UserIDContextKey, userID)
					ctx := service.WithUserID(c.Request().Context(), userID)
					c.SetRequest(c.Request().WithContext(ctx))
					return next(c)
				}
			}

			return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
		}
	}
}
