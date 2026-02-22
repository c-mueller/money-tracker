package middleware

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func Recovery(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("panic recovered",
						zap.String("error", fmt.Sprintf("%v", r)),
						zap.String("uri", c.Request().RequestURI),
					)
					err = echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
				}
			}()
			return next(c)
		}
	}
}
