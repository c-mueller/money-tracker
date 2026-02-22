package middleware

import (
	"github.com/labstack/echo/v4"
	"icekalt.dev/money-tracker/internal/service"
)

const UserIDContextKey = "user_id"

func InjectUserID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if userID, ok := c.Get(UserIDContextKey).(int); ok {
				ctx := service.WithUserID(c.Request().Context(), userID)
				c.SetRequest(c.Request().WithContext(ctx))
			}
			return next(c)
		}
	}
}
