package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/database"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const testRequestTimeout = -1 // disable Fiber's default 1s timeout for external API calls

func setupTestDB(t *testing.T) *gorm.DB {
	// Create an in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Get the underlying *sql.DB
	sqlDB, err := db.DB()
	require.NoError(t, err)

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	database.DB = db
	return db
}

// TestClient represents a test client for making HTTP requests
type TestClient struct {
	app *fiber.App
	jar *cookiejar.Jar
	t   *testing.T
}

// NewTestClient creates a new test client with an in-memory database
func NewTestClient(t *testing.T) *TestClient {
	// Setup test database
	db := setupTestDB(t)

	// Create test configuration. DevMode enables the local dev-login bypass so
	// the suite never reaches Auth0.
	cfg := &Config{
		ListenAddr: "127.0.0.1:0",
		ViewsPath:  "./views",
		AssetsPath: "./assets",
		DevMode:    true,
		Logger:     slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	// Setup app
	app, err := setupApp(cfg)
	require.NoError(t, err)

	// Create cookie jar
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	client := &TestClient{
		app: app,
		jar: jar,
		t:   t,
	}

	// Cleanup function
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	})

	return client
}

// Get makes a GET request to the specified path
func (tc *TestClient) Get(path string) *http.Response {
	req := httptest.NewRequest("GET", path, nil)

	// Add cookies
	u, _ := url.Parse("http://localhost")
	for _, cookie := range tc.jar.Cookies(u) {
		req.AddCookie(cookie)
	}

	resp, err := tc.app.Test(req, testRequestTimeout)
	require.NoError(tc.t, err)

	// Update cookies
	tc.jar.SetCookies(u, resp.Cookies())

	return resp
}

// Post makes a POST request to the specified path with form data
func (tc *TestClient) Post(path string, formData url.Values) *http.Response {
	req := httptest.NewRequest("POST", path, strings.NewReader(formData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Add cookies
	u, _ := url.Parse("http://localhost")
	for _, cookie := range tc.jar.Cookies(u) {
		req.AddCookie(cookie)
	}

	resp, err := tc.app.Test(req, testRequestTimeout)
	require.NoError(tc.t, err)

	// Update cookies
	tc.jar.SetCookies(u, resp.Cookies())

	return resp
}

// PostJSON makes a POST request to the specified path with JSON data
func (tc *TestClient) PostJSON(path string, data interface{}) *http.Response {
	jsonData, err := json.Marshal(data)
	require.NoError(tc.t, err)

	req := httptest.NewRequest("POST", path, strings.NewReader(string(jsonData)))
	req.Header.Add("Content-Type", "application/json")

	// Add cookies
	u, _ := url.Parse("http://localhost")
	for _, cookie := range tc.jar.Cookies(u) {
		req.AddCookie(cookie)
	}

	resp, err := tc.app.Test(req, testRequestTimeout)
	require.NoError(tc.t, err)

	// Update cookies
	tc.jar.SetCookies(u, resp.Cookies())

	return resp
}

// devLogin drives the dev-login bypass: it upserts a user by email and
// establishes the session, mirroring the Auth0 callback.
func (tc *TestClient) devLogin(name, email string) {
	form := url.Values{}
	form.Add("name", name)
	form.Add("email", email)

	resp := tc.Post("/auth/dev-login", form)
	require.Equal(tc.t, fiber.StatusFound, resp.StatusCode)
	require.Equal(tc.t, "/", resp.Header.Get("Location"))
}

// CreateUser creates (via first dev-login) a new user and returns the email
// used. The password argument is retained for call-site compatibility and
// ignored — Auth0 owns credentials, the dev bypass keys users by email.
func (tc *TestClient) CreateUser(name, password string) string {
	email := fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())
	tc.devLogin(name, email)
	return email
}

// Login logs in an existing user by email via the dev-login bypass.
func (tc *TestClient) Login(email, password string) {
	tc.devLogin("", email)
}

// Logout logs out the current user
func (tc *TestClient) Logout() {
	resp := tc.Get("/auth/logout")
	require.Equal(tc.t, fiber.StatusFound, resp.StatusCode)
	require.Equal(tc.t, "/", resp.Header.Get("Location"))
}

// GetActivities retrieves the list of activities as JSON
func (tc *TestClient) GetActivities() []types.Activity {
	req := httptest.NewRequest("GET", "/activities/list", nil)
	req.Header.Add("Accept", "application/json")

	// Add cookies
	u, _ := url.Parse("http://localhost")
	for _, cookie := range tc.jar.Cookies(u) {
		req.AddCookie(cookie)
	}

	resp, err := tc.app.Test(req, testRequestTimeout)
	require.NoError(tc.t, err)
	require.Equal(tc.t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(tc.t, err)
	resp.Body.Close()

	var activities []types.Activity
	err = json.Unmarshal(body, &activities)
	require.NoError(tc.t, err)

	return activities
}

// CreateActivity creates a new activity
func (tc *TestClient) CreateActivity(activity types.Activity) *http.Response {
	return tc.PostJSON("/activities/create", activity)
}

func (tc *TestClient) UpdateActivity(activity types.Activity) *http.Response {
	return tc.PostJSON("/activities/"+activity.ID.String(), activity)
}

// GetResponseBody reads and returns the response body as a string
func (tc *TestClient) GetResponseBody(resp *http.Response) string {
	body, err := io.ReadAll(resp.Body)
	require.NoError(tc.t, err)
	resp.Body.Close()
	return string(body)
}

// AssertResponseContains checks if the response body contains the expected text
func (tc *TestClient) AssertResponseContains(body string, expected string) {
	require.Contains(tc.t, body, expected)
}

// AssertResponseNotContains checks if the response body does not contain the expected text
func (tc *TestClient) AssertResponseNotContains(body string, expected string) {
	require.NotContains(tc.t, body, expected)
}

// TestClientIntegration demonstrates how to use the TestClient
func TestClientIntegration(t *testing.T) {
	client := NewTestClient(t)

	t.Run("Complete user workflow", func(t *testing.T) {
		// Test home page
		resp := client.Get("/")
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		client.AssertResponseContains(client.GetResponseBody(resp), "Velkommen til Logbogen")

		// Create user
		email := client.CreateUser("John Doe", "password")
		require.NotEmpty(t, email)

		// Login
		client.Login(email, "password")

		// Check activities (should be empty)
		activities := client.GetActivities()
		require.Len(t, activities, 0)

		// Logout
		client.Logout()

		// Verify logout worked
		resp = client.Get("/")
		client.AssertResponseContains(client.GetResponseBody(resp), "Log ind")
	})
}

// TestClientReusability demonstrates that the client can be reused across tests
func TestClientReusability(t *testing.T) {
	client := NewTestClient(t)

	t.Run("Multiple users can be created", func(t *testing.T) {
		// Create first user
		email1 := client.CreateUser("User 1", "password1")
		client.Login(email1, "password1")
		client.Logout()

		// Create second user
		email2 := client.CreateUser("User 2", "password2")
		client.Login(email2, "password2")
		client.Logout()

		require.NotEqual(t, email1, email2)
	})
}
