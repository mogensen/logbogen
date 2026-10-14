package middleware

import (
	"log/slog"
	"path"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/utils"
)

// AccessLog logs one structured line per request, without the query string since it can hold tokens.
// Static files are logged at debug level, and a handler error goes through the error handler here, once.
func AccessLog(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		if err := c.Next(); err != nil {
			if handlerErr := c.App().Config().ErrorHandler(c, err); handlerErr != nil {
				_ = c.SendStatus(fiber.StatusInternalServerError)
			}
		}

		// Streamed files, like static assets, have no body in memory but do announce their length
		size := len(c.Response().Body())
		if size == 0 {
			size = max(0, c.Response().Header.ContentLength())
		}

		attrs := []slog.Attr{
			slog.String("request_id", utils.RequestID(c)),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.String("route", c.Route().Path),
			slog.Int("status", c.Response().StatusCode()),
			slog.Float64("latency_ms", float64(time.Since(start).Microseconds())/1000),
			slog.Int("bytes", size),
			slog.String("ip", c.IP()),
		}
		if user := utils.GetUser(c); user != nil {
			attrs = append(attrs, slog.Uint64("user_id", user.ID))
		}

		level := slog.LevelInfo
		if path.Ext(c.Path()) != "" {
			level = slog.LevelDebug
		}
		logger.LogAttrs(c.UserContext(), level, "request", attrs...)
		return nil
	}
}
