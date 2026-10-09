package utils

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/stretchr/testify/require"
)

// logLines parses the JSON lines a logger wrote
func logLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	var lines []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		var entry map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &entry), line)
		lines = append(lines, entry)
	}
	return lines
}

// useLogger makes the logger the default one for the test
func useLogger(t *testing.T, logger *slog.Logger) {
	previous := slog.Default()
	slog.SetDefault(logger)
	t.Cleanup(func() { slog.SetDefault(previous) })
}

func TestNewLogger(t *testing.T) {
	t.Run("writes JSON lines by default", func(t *testing.T) {
		var buf bytes.Buffer
		NewLogger(&buf, "json", slog.LevelInfo).Info("hello", "user_id", 7)

		lines := logLines(t, &buf)
		require.Len(t, lines, 1)
		require.Equal(t, "hello", lines[0]["msg"])
		require.Equal(t, "INFO", lines[0]["level"])
		require.EqualValues(t, 7, lines[0]["user_id"])
		require.Contains(t, lines[0], "time")
	})

	t.Run("an unknown format is JSON too", func(t *testing.T) {
		var buf bytes.Buffer
		NewLogger(&buf, "yaml", slog.LevelInfo).Info("hello")

		require.Len(t, logLines(t, &buf), 1)
	})

	t.Run("text format is key=value", func(t *testing.T) {
		var buf bytes.Buffer
		NewLogger(&buf, "text", slog.LevelInfo).Info("hello", "user_id", 7)

		require.Contains(t, buf.String(), "msg=hello")
		require.Contains(t, buf.String(), "user_id=7")
		require.NotContains(t, buf.String(), "{")
	})

	t.Run("lines below the level are dropped", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger(&buf, "json", slog.LevelWarn)
		logger.Info("quiet")
		logger.Debug("quieter")
		logger.Warn("loud")

		lines := logLines(t, &buf)
		require.Len(t, lines, 1)
		require.Equal(t, "loud", lines[0]["msg"])
	})
}

func TestNewRequestID(t *testing.T) {
	first, second := NewRequestID(), NewRequestID()

	require.Len(t, first, 16)
	require.Regexp(t, "^[0-9a-f]{16}$", first)
	require.NotEqual(t, first, second)
}

func TestLogAndLogEvent_CarryTheRequestAndUser(t *testing.T) {
	var buf bytes.Buffer
	useLogger(t, NewLogger(&buf, "json", slog.LevelInfo))

	app := fiber.New()
	app.Get("/anonymous", func(c *fiber.Ctx) error {
		c.Locals(RequestIDLocalsKey, "req-1")
		Log(c).Info("something happened")
		return c.SendString("ok")
	})
	app.Get("/member", func(c *fiber.Ctx) error {
		c.Locals(RequestIDLocalsKey, "req-2")
		c.Locals("USER", &types.User{ID: 42})
		LogEvent(c, "activity_created", map[string]any{"category": "climbing", "activity_id": "abc", "participants": 2})
		return c.SendString("ok")
	})
	for _, path := range []string{"/anonymous", "/member"} {
		_, err := app.Test(httptest.NewRequest("GET", path, nil))
		require.NoError(t, err)
	}

	lines := logLines(t, &buf)
	require.Len(t, lines, 2)

	require.Equal(t, "something happened", lines[0]["msg"])
	require.Equal(t, "req-1", lines[0]["request_id"])
	require.NotContains(t, lines[0], "user_id", "no user, no user id")

	require.Equal(t, "activity_created", lines[1]["msg"])
	require.Equal(t, "INFO", lines[1]["level"])
	require.Equal(t, "req-2", lines[1]["request_id"])
	require.EqualValues(t, 42, lines[1]["user_id"])
	require.Equal(t, "climbing", lines[1]["category"])
	require.Equal(t, "abc", lines[1]["activity_id"])
	require.EqualValues(t, 2, lines[1]["participants"])
}

func TestLog_WithoutRequestIDMiddleware(t *testing.T) {
	var buf bytes.Buffer
	useLogger(t, NewLogger(&buf, "json", slog.LevelInfo))

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		Log(c).Info("plain")
		return nil
	})
	_, err := app.Test(httptest.NewRequest("GET", "/", nil))
	require.NoError(t, err)

	lines := logLines(t, &buf)
	require.Len(t, lines, 1)
	require.NotContains(t, lines[0], "request_id")
}
