package database

import (
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/sqlite3"
	"github.com/mogensen/logbook/pkg/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB is the underlying database connection
var DB *gorm.DB
var SessionStore *session.Store

// ensureDir creates the directory that holds the database file when it is missing
func ensureDir(dbPath string) error {
	return os.MkdirAll(filepath.Dir(dbPath), 0o755)
}

func init() {
	if err := ensureDir(config.DB); err != nil {
		slog.Error("Failed to create the database directory", "path", config.DB, "error", err)
	}

	// Init sessions store
	storage := sqlite3.New(sqlite3.Config{
		Database:        config.DB,
		Table:           "sessions",
		Reset:           false,
		GCInterval:      10 * time.Second,
		MaxOpenConns:    100,
		MaxIdleConns:    100,
		ConnMaxLifetime: 1 * time.Second,
	})

	// Initialize a session store
	sessConfig := session.Config{
		Storage:        storage,
		Expiration:     30 * time.Minute, // Expire sessions after 30 minutes of inactivity
		KeyLookup:      "cookie:session", // Removed __Host- prefix
		CookieSecure:   true,
		CookieHTTPOnly: true,
		CookieSameSite: "Lax",
	}
	SessionStore = session.New(sessConfig)
}

// SetCookieSecure toggles the Secure flag on the session cookie. Only dev mode over plain HTTP should turn it off.
func SetCookieSecure(secure bool) {
	SessionStore.CookieSecure = secure
}

// Connect initiate the database connection and migrate all the tables
func Connect() error {
	if err := ensureDir(config.DB); err != nil {
		return err
	}

	db, err := gorm.Open(sqlite.Open(config.DB), &gorm.Config{
		NowFunc: func() time.Time { return time.Now().Local() },
		Logger:  NewGormLogger(),
	})
	if err != nil {
		return err
	}

	// Setting the database connection to use in routes
	DB = db

	slog.Info("Database connection established")
	return nil
}

// Migrate migrates all the database tables
func Migrate(tables ...interface{}) error {
	return DB.AutoMigrate(tables...)
}
