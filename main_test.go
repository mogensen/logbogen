package main

import (
	"fmt"
	"io"
	"log/slog"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/database"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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

func TestAppSetup(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer func() {
		sqlDB, err := db.DB()
		require.NoError(t, err)
		sqlDB.Close()
	}()

	// Create test configuration
	cfg := &Config{
		ListenAddr: "127.0.0.1:0", // Use port 0 to get a random available port
		ViewsPath:  "./views",
		AssetsPath: "./assets",
		Logger:     slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	// Setup app
	app, err := setupApp(cfg)
	require.NoError(t, err)

	// Start the app in a goroutine
	go func() {
		err := app.Listen(cfg.ListenAddr)
		require.NoError(t, err)
	}()

	// Wait for the app to start
	time.Sleep(500 * time.Millisecond)

	email := fmt.Sprintf("test-%d@example.com", time.Now().UnixNano())

	t.Run("Home page should be accessible", func(t *testing.T) {
		// Make a test request
		req := httptest.NewRequest("GET", "/", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), "Velkommen til Logbogen")
		require.Contains(t, string(body), "Log ind")
		require.Contains(t, string(body), "Opret bruger")
	})

	t.Run("User can signup", func(t *testing.T) {
		// Create a new user
		form := url.Values{}
		form.Add("name", "John Doe")
		form.Add("email", email)
		form.Add("password", "password")
		req := httptest.NewRequest("POST", "/auth/signup", strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		// Send the request
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		// Check the response body
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), "Brugeren er oprettet, du kan nu logge ind")
		require.Contains(t, string(body), "Login")
		require.Contains(t, string(body), "Ny bruger")
	})

	t.Run("User can login", func(t *testing.T) {
		// Create a new user
		form := url.Values{}
		form.Add("email", email)
		form.Add("password", "password")
		req := httptest.NewRequest("POST", "/auth/login", strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusFound, resp.StatusCode)
		require.Equal(t, "/", resp.Header.Get("Location"))
	})

	t.Run("User can logout", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/auth/logout", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusFound, resp.StatusCode)
		require.Equal(t, "/", resp.Header.Get("Location"))

		// Check the response body
		req = httptest.NewRequest("GET", "/", nil)
		resp, err = app.Test(req)
		require.NoError(t, err)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), "Velkommen til Logbogen")
		require.Contains(t, string(body), "Log ind")
		require.Contains(t, string(body), "Opret bruger")
	})

	// Make a test request
	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Cleanup
	err = app.Shutdown()
	require.NoError(t, err)
}
