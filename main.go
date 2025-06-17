package main

import (
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v2"
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
	routes.HomeRoutes(app, authMiddleware)
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
