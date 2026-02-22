package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			id := c.Request().Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = uuid.New().String()
			}
			c.Request().Header.Set(echo.HeaderXRequestID, id)
			c.Response().Header().Set(echo.HeaderXRequestID, id)
			return next(c)
		}
	}
}
