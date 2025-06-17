package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/database"
	"github.com/mogensen/logbook/pkg/routes"
	"github.com/mogensen/logbook/pkg/services"
	"github.com/mogensen/logbook/pkg/utils"
	"github.com/mogensen/logbook/pkg/utils/middleware"
	slogfiber "github.com/samber/slog-fiber"
)

// Config holds the application configuration
type Config struct {
	ListenAddr  string
	DatabaseURL string
	ViewsPath   string
	AssetsPath  string
	CertFile    string
	KeyFile     string
	Logger      *slog.Logger
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		ListenAddr: "127.0.0.1:3000",
		ViewsPath:  "./views",
		AssetsPath: "./assets",
		Logger:     slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}
}

// setupApp creates and configures the Fiber application
func setupApp(cfg *Config) (*fiber.App, error) {
	err := database.Migrate(&dal.User{}, &dal.Activity{})
	if err != nil {
		cfg.Logger.Error("Error migrating database", "error", err)
		return nil, err
	}

	// HTML templates
	engine := html.New(cfg.ViewsPath, ".html")
	engine.AddFunc("current_user", utils.GetUser)
	engine.AddFunc("is_current_user", utils.IsCurrentUser)
	engine.AddFunc("fmtDate", utils.FormatDate)
	engine.AddFunc("fmtDateHuman", utils.FormatDateHuman)
	engine.AddFunc("is_same_user", utils.IsSameUser)
	engine.AddFunc("json", utils.ToJSON)
	engine.AddFunc("firstSix", utils.FirstSix)
	engine.AddFunc("userImage", utils.UserImage)
	engine.AddFunc("ctxActivity", utils.CtxActivity)

	// Create a Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler:      utils.ErrorHandler,
		Views:             engine,
		ViewsLayout:       "layouts/main",
		PassLocalsToViews: true,
	})

	app.Use(slogfiber.New(cfg.Logger))
	app.Use(recover.New())

	app.Static("/", cfg.AssetsPath)

	// CSRF Error handler
	csrfMiddleware := setupCsrfMiddleware()

	// Data Layer
	userDal := dal.NewUserDal(database.DB)
	activityDal := dal.NewActivityService(database.DB)

	// Services
	weatherService := services.NewWeatherService()
	activitiesService := services.NewActivityService(userDal, activityDal, weatherService)
	scoreboardService := services.NewScoreboardService(userDal)
	authService := services.NewAuthService(userDal)

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Route for the root path
	routes.HomeRoutes(app, csrfMiddleware, authMiddleware)
	routes.AuthRoutes(app, authService, authMiddleware)
	routes.ActivitiesRoutes(app, activitiesService, authMiddleware)
	routes.ScoreboardRoutes(app, scoreboardService, authMiddleware)

	return app, nil
}

func main() {
	cfg := DefaultConfig()

	// Connect to database
	err := database.Connect()
	if err != nil {
		cfg.Logger.Error("Error connecting to database", "error", err)
		return
	}

	app, err := setupApp(cfg)
	if err != nil {
		cfg.Logger.Error("Error setting up app", "error", err)
		return
	}

	app.Listen(cfg.ListenAddr)
}

func setupCsrfMiddleware() fiber.Handler {
	csrfErrorHandler := func(c *fiber.Ctx, err error) error {
		// Log the error so we can track who is trying to perform CSRF attacks
		slog.Warn("CSRF Error detected",
			"error", err,
			"url", c.OriginalURL(),
			"ip", c.IP(),
		)

		// check accepted content types
		switch c.Accepts("html", "json") {
		case "json":
			// Return a 403 Forbidden response for JSON requests
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "403 Forbidden",
			})
		case "html":
			// Return a 403 Forbidden response for HTML requests
			return c.Status(fiber.StatusForbidden).Render("error", fiber.Map{
				"Title":     "Error",
				"Error":     "403 Forbidden",
				"ErrorCode": "403",
			})
		default:
			// Return a 403 Forbidden response for all other requests
			return c.Status(fiber.StatusForbidden).SendString("403 Forbidden")
		}
	}

	// Configure the CSRF middleware
	csrfConfig := csrf.Config{
		Session:        database.SessionStore,
		KeyLookup:      "form:csrf",   // In this example, we will be using a hidden input field to store the CSRF token
		CookieName:     "__Host-csrf", // Recommended to use the __Host- prefix when serving the app over TLS
		CookieSameSite: "Lax",         // Recommended to set this to Lax or Strict
		CookieSecure:   true,          // Recommended to set to true when serving the app over TLS
		CookieHTTPOnly: true,          // Recommended, otherwise if using JS framework recomend: false and KeyLookup: "header:X-CSRF-Token"
		ContextKey:     "csrf",
		ErrorHandler:   csrfErrorHandler,
		Expiration:     30 * time.Minute,
	}
	csrfMiddleware := csrf.New(csrfConfig)
	return csrfMiddleware
}
