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
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/database"
	"github.com/mogensen/logbook/pkg/services"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/mogensen/logbook/pkg/utils"
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
func NewTestClient(t *testing.T, opts ...func(*Config)) *TestClient {
	app := newTestApp(t, opts...)

	// Create cookie jar
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	return &TestClient{
		app: app,
		jar: jar,
		t:   t,
	}
}

// newTestApp creates the app on an in-memory database. DevMode enables the local dev-login bypass so
// the suite never reaches Firebase, and the geocoder and weather services are fakes.
func newTestApp(t *testing.T, opts ...func(*Config)) *fiber.App {
	db := setupTestDB(t)

	cfg := &Config{
		ListenAddr: "127.0.0.1:0",
		ViewsPath:  "./views",
		AssetsPath: "./assets",
		DevMode:    true,
		// Tests never call the public geocoding API
		Geocoder: fakeGeocoder{name: "København, Danmark"},
		// Tests log in many times from the same address
		AuthRateLimit:  100000,
		AuthRateWindow: time.Minute,
		Logger:         utils.NewLogger(io.Discard, "json", slog.LevelInfo),
	}

	// Tests never call the public weather API either
	weatherServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"daily":{"temperature_2m_max":[12.5],"weathercode":[3]},"current":{"temperature_2m":12.5,"weathercode":3}}`))
	}))
	t.Cleanup(weatherServer.Close)
	cfg.Weather = services.NewWeatherServiceWithBaseURLs(weatherServer.URL, weatherServer.URL)

	for _, opt := range opts {
		opt(cfg)
	}

	app, err := setupApp(cfg)
	require.NoError(t, err)

	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	})

	return app
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

// csrfToken returns the CSRF token cookie, visiting the home page first when none has been issued yet
func (tc *TestClient) csrfToken() string {
	u, _ := url.Parse("http://localhost")
	for range 2 {
		for _, cookie := range tc.jar.Cookies(u) {
			if cookie.Name == "csrf_" {
				return cookie.Value
			}
		}
		tc.Get("/")
	}
	require.Fail(tc.t, "no CSRF token was issued")
	return ""
}

// Post makes a POST request to the specified path with form data
func (tc *TestClient) Post(path string, formData url.Values) *http.Response {
	req := httptest.NewRequest("POST", path, strings.NewReader(formData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("X-Csrf-Token", tc.csrfToken())

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
func (tc *TestClient) PostJSON(path string, data any) *http.Response {
	jsonData, err := json.Marshal(data)
	require.NoError(tc.t, err)

	req := httptest.NewRequest("POST", path, strings.NewReader(string(jsonData)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Csrf-Token", tc.csrfToken())

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
// establishes the session, mirroring the Firebase session endpoint.
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
// ignored — Firebase owns credentials, the dev bypass keys users by email.
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
	resp := tc.Post("/auth/logout", url.Values{})
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

// fakeGeocoder answers every lookup with a fixed place, or an error when err is set
type fakeGeocoder struct {
	name string
	err  error
}

func (f fakeGeocoder) ReverseGeocode(lat, lng float64) (*services.GeocodeResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &services.GeocodeResponse{Name: f.name}, nil
}

// fakeWeather answers weather lookups with fixed values and records the requested dates
type fakeWeather struct {
	mu          sync.Mutex
	dates       []string
	temperature float64
	code        int
	err         error
	missing     bool
}

func (f *fakeWeather) GetWeather(lat, lng float64, date time.Time) (*services.WeatherResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.dates = append(f.dates, date.Format("2006-01-02"))
	if f.err != nil {
		return nil, f.err
	}

	response := &services.WeatherResponse{}
	if f.missing {
		response.Daily.Temperature2mMax = []*float64{nil}
		response.Daily.WeatherCode = []*int{nil}
		return response, nil
	}
	response.Daily.Temperature2mMax = []*float64{&f.temperature}
	response.Daily.WeatherCode = []*int{&f.code}
	return response, nil
}

func (f *fakeWeather) calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.dates)
}

func (f *fakeWeather) set(fn func(*fakeWeather)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fn(f)
}
