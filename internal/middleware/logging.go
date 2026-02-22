package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func Logger(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			if err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()
			latency := time.Since(start)

			fields := []zap.Field{
				zap.String("method", req.Method),
				zap.String("uri", req.RequestURI),
				zap.Int("status", res.Status),
				zap.Duration("latency", latency),
				zap.String("request_id", req.Header.Get(echo.HeaderXRequestID)),
			}

			n := res.Status
			switch {
			case n >= 500:
				logger.Error("server error", fields...)
			case n >= 400:
				logger.Warn("client error", fields...)
			default:
				logger.Info("request", fields...)
			}

			return nil
		}
	}
}
