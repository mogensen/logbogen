package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/config"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/mogensen/logbook/pkg/utils"
	"github.com/stretchr/testify/require"
)

// logCapture collects everything the app logs, as the JSON lines a production run would write
type logCapture struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (l *logCapture) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.buf.Write(p)
}

func (l *logCapture) text() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.buf.String()
}

// entries parses every logged line, failing on a line that is not JSON
func (l *logCapture) entries(t *testing.T) []map[string]any {
	var entries []map[string]any
	for line := range strings.SplitSeq(strings.TrimSpace(l.text()), "\n") {
		if line == "" {
			continue
		}
		var entry map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &entry), "not structured: %s", line)
		entries = append(entries, entry)
	}
	return entries
}

// find returns the lines with the message
func (l *logCapture) find(t *testing.T, msg string) []map[string]any {
	var found []map[string]any
	for _, entry := range l.entries(t) {
		if entry["msg"] == msg {
			found = append(found, entry)
		}
	}
	return found
}

// newLoggedClient is a test client whose app and default logger write structured lines into the capture
func newLoggedClient(t *testing.T, level slog.Level) (*TestClient, *logCapture) {
	logs := &logCapture{}
	logger := utils.NewLogger(logs, "json", level)

	previous := slog.Default()
	slog.SetDefault(logger)
	t.Cleanup(func() { slog.SetDefault(previous) })

	client := NewTestClient(t, func(cfg *Config) { cfg.Logger = logger })
	return client, logs
}

func TestRequestsAreLoggedWithAnIDAndTheUser(t *testing.T) {
	client, logs := newLoggedClient(t, slog.LevelInfo)
	client.CreateUser("Other User", "password")
	email := client.CreateUser("Logged User", "password")
	client.Login(email, "password")

	resp := client.Get("/activities/list")

	id := resp.Header.Get("X-Request-Id")
	require.Regexp(t, "^[0-9a-f]{16}$", id)
	var line map[string]any
	for _, entry := range logs.find(t, "request") {
		if entry["request_id"] == id {
			line = entry
		}
	}
	require.NotNil(t, line, "the id sent to the client is on the access line")
	require.Equal(t, "GET", line["method"])
	require.Equal(t, "/activities/list", line["path"])
	require.EqualValues(t, 200, line["status"])
	require.EqualValues(t, 2, line["user_id"])
	require.Contains(t, line, "latency_ms")

	t.Run("every logged line is structured", func(t *testing.T) {
		require.NotEmpty(t, logs.entries(t))
	})

	t.Run("static files stay out of the info level", func(t *testing.T) {
		before := len(logs.find(t, "request"))

		require.Equal(t, fiber.StatusOK, client.Get("/css/application.css").StatusCode)

		require.Equal(t, before, len(logs.find(t, "request")))
	})
}

func TestBusinessActionsAreLoggedAsEvents(t *testing.T) {
	client, logs := newLoggedClient(t, slog.LevelInfo)
	client.CreateUser("Other User", "password")
	email := client.CreateUser("Event User", "password")
	client.Login(email, "password")

	t.Run("logging in", func(t *testing.T) {
		events := logs.find(t, "user_logged_in")

		require.NotEmpty(t, events)
		last := events[len(events)-1]
		require.Equal(t, "dev", last["login_method"])
		require.EqualValues(t, 2, last["user_id"])
		require.NotContains(t, logs.text(), email, "the email is not logged")
	})

	activity := randomActivity()
	activity.Date = types.Date(time.Now().AddDate(0, 0, -3))
	client.CreateActivity(activity)
	created := client.GetActivities()[0]

	t.Run("creating an activity", func(t *testing.T) {
		events := logs.find(t, "activity_created")

		require.Len(t, events, 1)
		require.Equal(t, created.ID.String(), events[0]["activity_id"])
		require.Equal(t, activity.CategoryID, events[0]["category"])
		require.Equal(t, activity.TypeID, events[0]["type"])
		require.EqualValues(t, 2, events[0]["user_id"], "who did it")
		require.NotEmpty(t, events[0]["request_id"])
		require.Equal(t, "INFO", events[0]["level"])
	})

	t.Run("updating an activity", func(t *testing.T) {
		created.Comment = "Edited"
		created.Date = types.Date(time.Now().AddDate(0, 0, -9))
		client.UpdateActivity(created)

		events := logs.find(t, "activity_updated")
		require.Len(t, events, 1)
		require.Equal(t, created.ID.String(), events[0]["activity_id"])
		require.Equal(t, true, events[0]["day_or_place_changed"])
	})

	t.Run("changing the theme", func(t *testing.T) {
		require.Equal(t, fiber.StatusOK, client.PostJSON("/users/theme", map[string]string{"theme": "dark"}).StatusCode)

		events := logs.find(t, "theme_changed")
		require.Len(t, events, 1)
		require.Equal(t, "dark", events[0]["theme"])
	})

	t.Run("a certification", func(t *testing.T) {
		userID := uint64(2)
		cert := types.Certification{
			UserID:     &userID,
			CategoryID: config.AllCertificationCategories[0].ID,
			TypeID:     config.AllCertificationTypes[0].ID,
			StartDate:  types.Date(time.Now().AddDate(-1, 0, 0)),
			EndDate:    types.Date(time.Now().AddDate(1, 0, 0)),
		}
		require.Equal(t, fiber.StatusFound, client.PostJSON("/certifications/create", cert).StatusCode)

		events := logs.find(t, "certification_created")
		require.Len(t, events, 1)
		require.Equal(t, cert.TypeID, events[0]["type"])
		certID, _ := events[0]["certification_id"].(string)
		require.NotEmpty(t, certID)

		require.Equal(t, fiber.StatusFound, client.Post("/certifications/"+certID+"/delete", url.Values{}).StatusCode)
		require.Len(t, logs.find(t, "certification_deleted"), 1)
	})

	t.Run("deleting an activity", func(t *testing.T) {
		require.Equal(t, fiber.StatusOK, client.Post("/activities/"+created.ID.String()+"/delete", url.Values{}).StatusCode)

		events := logs.find(t, "activity_deleted")
		require.Len(t, events, 1)
		require.Equal(t, created.ID.String(), events[0]["activity_id"])
	})

	t.Run("logging out", func(t *testing.T) {
		client.Logout()

		events := logs.find(t, "user_logged_out")
		require.Len(t, events, 1)
		require.EqualValues(t, 2, events[0]["user_id"])
	})
}

func TestFailuresAreLoggedWithTheIDShownToTheUser(t *testing.T) {
	client, logs := newLoggedClient(t, slog.LevelInfo)
	client.app.Get("/boom", func(c *fiber.Ctx) error {
		return errors.New("Error 1146: table users_secret does not exist")
	})
	client.app.Get("/panic", func(c *fiber.Ctx) error { panic("something exploded") })

	t.Run("an error shows its id on the page and logs the cause under it", func(t *testing.T) {
		resp := client.Get("/boom")
		id := resp.Header.Get("X-Request-Id")
		body := client.GetResponseBody(resp)

		require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		client.AssertResponseContains(body, `<dd class="error-id">`+id+"</dd>")
		client.AssertResponseNotContains(body, "users_secret")

		failed := logs.find(t, "request failed")
		require.Len(t, failed, 1)
		require.Equal(t, id, failed[0]["request_id"])
		require.Equal(t, "ERROR", failed[0]["level"])
		require.Contains(t, failed[0]["error"], "users_secret", "the cause is in the log, not on the page")
		require.EqualValues(t, 500, failed[0]["status"])
	})

	t.Run("a client error has no error line but the page shows the request id", func(t *testing.T) {
		before := len(logs.find(t, "request failed"))

		resp := client.Get("/no-such-page")

		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
		client.AssertResponseContains(client.GetResponseBody(resp), resp.Header.Get("X-Request-Id"))
		require.Equal(t, before, len(logs.find(t, "request failed")))
	})

	t.Run("a panic is logged with its stack", func(t *testing.T) {
		resp := client.Get("/panic")
		id := resp.Header.Get("X-Request-Id")

		require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		panics := logs.find(t, "panic recovered")
		require.Len(t, panics, 1)
		require.Equal(t, id, panics[0]["request_id"])
		require.Equal(t, "something exploded", panics[0]["panic"])
		require.Contains(t, panics[0]["stack"], "logging_test.go")
		client.AssertResponseNotContains(client.GetResponseBody(resp), "something exploded")
	})
}

func TestSecretsAreNotLogged(t *testing.T) {
	client, logs := newLoggedClient(t, slog.LevelDebug)

	client.Post("/auth/session", url.Values{"idToken": {"secret-token"}})
	client.Post("/auth/dev-login", url.Values{"email": {"dev@example.com"}})

	text := logs.text()
	require.NotContains(t, text, "secret-token")
	require.NotContains(t, text, "csrf_")
	require.NotEmpty(t, logs.entries(t))
}
