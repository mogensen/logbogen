package database

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/sqlite3"
	"github.com/mogensen/logbook/pkg/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the underlying database connection
var DB *gorm.DB
var SessionStore *session.Store

func init() {
	// Init sessions store
	storage := sqlite3.New(sqlite3.Config{
		Database:        "./fiber.db",
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
		Expiration:     30 * time.Minute,        // Expire sessions after 30 minutes of inactivity
		KeyLookup:      "cookie:__Host-session", // Recommended to use the __Host- prefix when serving the app over TLS
		CookieSecure:   true,
		CookieHTTPOnly: true,
		CookieSameSite: "Lax",
	}
	SessionStore = session.New(sessConfig)
}

// Connect initiate the database connection and migrate all the tables
func Connect() error {
	db, err := gorm.Open(sqlite.Open(config.DB), &gorm.Config{
		NowFunc: func() time.Time { return time.Now().Local() },
		Logger:  logger.Default.LogMode(logger.Info),
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
