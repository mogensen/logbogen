package utils

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"log/slog"
	"sort"

	"github.com/gofiber/fiber/v2"
)

// RequestIDLocalsKey is where middleware.RequestID stores the id of the current request.
const RequestIDLocalsKey = "requestID"

// NewRequestID returns a short random id that ties the log lines of one request together.
func NewRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(b)
}

// RequestID returns the id of the current request, empty when middleware.RequestID has not run.
func RequestID(c *fiber.Ctx) string {
	id, _ := c.Locals(RequestIDLocalsKey).(string)
	return id
}

// NewLogger creates a structured logger writing JSON lines, or key=value text lines for the format "text".
func NewLogger(w io.Writer, format string, level slog.Level) *slog.Logger {
	options := &slog.HandlerOptions{Level: level}
	if format == "text" {
		return slog.New(slog.NewTextHandler(w, options))
	}
	return slog.New(slog.NewJSONHandler(w, options))
}

// Log returns the default logger with the request id and, when logged in, the user id of the request,
// so every line a handler logs can be tied to one request and one user.
func Log(c *fiber.Ctx) *slog.Logger {
	logger := slog.Default()
	if id := RequestID(c); id != "" {
		logger = logger.With("request_id", id)
	}
	if user := GetUser(c); user != nil {
		logger = logger.With("user_id", user.ID)
	}
	return logger
}

// LogEvent logs one line for a business action, such as an activity being created, with everything needed
// to answer who did what to which record in the fields.
func LogEvent(c *fiber.Ctx, event string, fields map[string]any) {
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	attrs := make([]any, 0, len(keys))
	for _, key := range keys {
		attrs = append(attrs, slog.Any(key, fields[key]))
	}
	Log(c).Info(event, attrs...)
}
