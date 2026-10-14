package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/mogensen/logbook/pkg/utils"
	"github.com/stretchr/testify/require"
)

func lines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	var entries []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		var entry map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &entry), line)
		entries = append(entries, entry)
	}
	return entries
}

// logApp is an app with the request id and access log in front, and routes for the cases to log
func logApp(buf *bytes.Buffer, level slog.Level) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var fe *fiber.Error
			if errors.As(err, &fe) {
				code = fe.Code
			}
			return c.Status(code).SendString("error page")
		},
	})
	app.Use(RequestID)
	app.Use(AccessLog(utils.NewLogger(buf, "json", level)))
	app.Use(recover.New())

	app.Get("/ok", func(c *fiber.Ctx) error { return c.SendString("hello") })
	app.Get("/activities/:id", func(c *fiber.Ctx) error {
		c.Locals("USER", &types.User{ID: 9})
		return c.SendString("an activity")
	})
	app.Get("/missing", func(c *fiber.Ctx) error { return fiber.NewError(fiber.StatusNotFound, "no such thing") })
	app.Get("/broken", func(c *fiber.Ctx) error { return errors.New("table users has no column secret") })
	app.Get("/panic", func(c *fiber.Ctx) error { panic("something exploded") })
	app.Get("/session-check", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusFound) })
	app.Get("/css/app.css", func(c *fiber.Ctx) error { return c.SendString("body{}") })
	return app
}

func get(t *testing.T, app *fiber.App, target string) string {
	resp, err := app.Test(httptest.NewRequest("GET", target, nil))
	require.NoError(t, err)
	return resp.Header.Get("X-Request-Id")
}

func TestRequestID(t *testing.T) {
	var buf bytes.Buffer
	app := logApp(&buf, slog.LevelInfo)

	first, second := get(t, app, "/ok"), get(t, app, "/ok")

	require.Regexp(t, "^[0-9a-f]{16}$", first)
	require.NotEqual(t, first, second, "every request has its own id")
}

func TestAccessLog(t *testing.T) {
	t.Run("one structured line per request", func(t *testing.T) {
		var buf bytes.Buffer
		app := logApp(&buf, slog.LevelInfo)

		id := get(t, app, "/ok")

		entries := lines(t, &buf)
		require.Len(t, entries, 1)
		entry := entries[0]
		require.Equal(t, "request", entry["msg"])
		require.Equal(t, "INFO", entry["level"])
		require.Equal(t, id, entry["request_id"], "the logged id is the one sent to the client")
		require.Equal(t, "GET", entry["method"])
		require.Equal(t, "/ok", entry["path"])
		require.Equal(t, "/ok", entry["route"])
		require.EqualValues(t, 200, entry["status"])
		require.EqualValues(t, len("hello"), entry["bytes"])
		require.GreaterOrEqual(t, entry["latency_ms"], 0.0)
		require.NotEmpty(t, entry["ip"])
		require.NotContains(t, entry, "user_id")
	})

	t.Run("the route is logged next to the path, and the user when known", func(t *testing.T) {
		var buf bytes.Buffer
		app := logApp(&buf, slog.LevelInfo)

		get(t, app, "/activities/1234")

		entry := lines(t, &buf)[0]
		require.Equal(t, "/activities/1234", entry["path"])
		require.Equal(t, "/activities/:id", entry["route"], "a route that does not depend on the id")
		require.EqualValues(t, 9, entry["user_id"])
	})

	t.Run("the query string is never logged", func(t *testing.T) {
		var buf bytes.Buffer
		app := logApp(&buf, slog.LevelInfo)

		get(t, app, "/session-check?token=secret-token&state=secret-state&error=access_denied")

		require.NotContains(t, buf.String(), "secret")
		require.NotContains(t, buf.String(), "access_denied")
		require.Equal(t, "/session-check", lines(t, &buf)[0]["path"])
	})

	t.Run("an error is logged with the status the client gets", func(t *testing.T) {
		var buf bytes.Buffer
		app := logApp(&buf, slog.LevelInfo)

		get(t, app, "/missing")
		get(t, app, "/broken")

		entries := lines(t, &buf)
		require.Len(t, entries, 2, "one line each, the error is not handled twice")
		require.EqualValues(t, 404, entries[0]["status"])
		require.EqualValues(t, 500, entries[1]["status"])
		require.NotContains(t, buf.String(), "secret", "the error text is not part of the access line")
	})

	t.Run("a panic is recovered and logged as a 500", func(t *testing.T) {
		var buf bytes.Buffer
		app := logApp(&buf, slog.LevelInfo)

		resp, err := app.Test(httptest.NewRequest("GET", "/panic", nil))
		require.NoError(t, err)

		require.Equal(t, 500, resp.StatusCode)
		entries := lines(t, &buf)
		require.Len(t, entries, 1)
		require.EqualValues(t, 500, entries[0]["status"])
	})

	t.Run("static files are logged at debug level only", func(t *testing.T) {
		var info bytes.Buffer
		get(t, logApp(&info, slog.LevelInfo), "/css/app.css")
		require.Empty(t, lines(t, &info), "hidden at info level")

		var debug bytes.Buffer
		get(t, logApp(&debug, slog.LevelDebug), "/css/app.css")
		entries := lines(t, &debug)
		require.Len(t, entries, 1)
		require.Equal(t, "DEBUG", entries[0]["level"])
	})
}
