package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/mogensen/logbook/pkg/auth"
	"github.com/mogensen/logbook/pkg/config"
	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/database"
	"github.com/mogensen/logbook/pkg/routes"
	"github.com/mogensen/logbook/pkg/services"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/mogensen/logbook/pkg/utils"
	"github.com/mogensen/logbook/pkg/utils/middleware"
)

// maxBodyBytes caps the size of a request body
const maxBodyBytes = 64 * 1024

// csrfToken reads the CSRF token from the X-Csrf-Token header, falling back to the _csrf form field
func csrfToken(c *fiber.Ctx) (string, error) {
	if token := c.Get("X-Csrf-Token"); token != "" {
		return token, nil
	}
	if token := c.FormValue("_csrf"); token != "" {
		return token, nil
	}
	return "", csrf.ErrTokenNotFound
}

// Config holds the application configuration
type Config struct {
	ListenAddr  string
	DatabaseURL string
	ViewsPath   string
	AssetsPath  string
	CertFile    string
	KeyFile     string
	DevMode     bool
	// Firebase holds the public settings of the Firebase web app, unused in dev mode
	Firebase auth.Config
	// ProxyHeader and TrustedProxies make c.IP() resolve the real client behind a reverse proxy
	// Weather looks up the weather of an activity, nil uses the public Open-Meteo API
	Weather services.WeatherProvider
	// Geocoder names the place of an activity, nil skips geocoding
	Geocoder services.Geocoder
	// AuthRateLimit is the number of auth requests a client IP may make per AuthRateWindow
	AuthRateLimit  int
	AuthRateWindow time.Duration
	ProxyHeader    string
	TrustedProxies []string
	Logger         *slog.Logger
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		ListenAddr: config.LISTENADDR,
		ViewsPath:  "./views",
		AssetsPath: "./assets",
		DevMode:    config.DEVMODE,
		Firebase: auth.Config{
			ProjectID:    config.FIREBASEPROJECTID,
			APIKey:       config.FIREBASEAPIKEY,
			AuthDomain:   config.FIREBASEAUTHDOMAIN,
			EmulatorHost: config.FIREBASEAUTHEMULATORHOST,
		},

		Geocoder:       services.NewNominatimGeocoder(),
		AuthRateLimit:  config.AUTHRATELIMIT,
		AuthRateWindow: config.AUTHRATEWINDOW,
		ProxyHeader:    config.PROXYHEADER,
		TrustedProxies: config.TRUSTEDPROXIES,
		Logger:         utils.NewLogger(os.Stdout, config.LOGFORMAT, config.LOGLEVEL),
	}
}

// setupApp creates and configures the Fiber application
func setupApp(cfg *Config) (*fiber.App, error) {
	if err := validateAuthConfig(cfg, config.ALLOWDEVMODE); err != nil {
		cfg.Logger.Error("Invalid configuration", "error", err)
		return nil, err
	}

	err := database.Migrate(&dal.User{}, &dal.Activity{}, &dal.Certification{})
	if err != nil {
		cfg.Logger.Error("Error migrating database", "error", err)
		return nil, err
	}

	database.SetCookieSecure(!cfg.DevMode)

	// HTML templates
	engine := html.New(cfg.ViewsPath, ".html")
	engine.AddFunc("activityMax", types.ActivityMax)
	engine.AddFunc("certificationMax", types.CertificationMax)
	engine.AddFunc("current_user", utils.GetUser)
	engine.AddFunc("is_current_user", utils.IsCurrentUser)
	engine.AddFunc("fmtDate", utils.FormatDate)
	engine.AddFunc("fmtDateHuman", utils.FormatDateHuman)
	engine.AddFunc("fmtDateLong", utils.FormatDateLong)
	engine.AddFunc("is_same_user", utils.IsSameUser)
	engine.AddFunc("firstSix", utils.FirstSix)
	engine.AddFunc("userImage", utils.UserImage)
	engine.AddFunc("ctxActivity", utils.CtxActivity)

	// Create a Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler:      utils.ErrorHandler,
		Views:             engine,
		ViewsLayout:       "layouts/main",
		PassLocalsToViews: true,
		// The startup banner is not structured, the app logs when it starts listening
		DisableStartupMessage: true,
		BodyLimit:             maxBodyBytes,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          30 * time.Second,
		IdleTimeout:           60 * time.Second,

		// Fiber trusts ProxyHeader from anyone when the proxy check is off, so the header is dropped without proxies
		ProxyHeader:             proxyHeader(cfg),
		EnableTrustedProxyCheck: len(cfg.TrustedProxies) > 0,
		TrustedProxies:          cfg.TrustedProxies,
	})

	// The request id comes first so every log line of a request has it, then one access line per request
	app.Use(middleware.RequestID)
	app.Use(middleware.AccessLog(cfg.Logger))
	app.Use(recover.New(recover.Config{EnableStackTrace: true, StackTraceHandler: logPanic}))

	// Nothing starting with a dot (.DS_Store, .env) may be served from the assets tree
	app.Use(middleware.BlockDotfiles)
	app.Static("/", cfg.AssetsPath)
	// Registered after Static, which resets the response headers for "/"
	app.Use(middleware.SecurityHeaders(cspFirebase(cfg)))

	// CSRF tokens are checked on every unsafe request, from the header (fetch/ajax) or the _csrf form field
	app.Use(csrf.New(csrf.Config{
		Extractor:      csrfToken,
		CookieName:     "csrf_",
		CookieSameSite: "Lax",
		CookieSecure:   !cfg.DevMode,
		CookieHTTPOnly: true,
		Expiration:     time.Hour,
		ContextKey:     "csrf",
	}))

	// Data Layer
	userDal := dal.NewUserDal(database.DB)
	activityDal := dal.NewActivityService(database.DB)
	certificationDal := dal.NewCertificationService(database.DB)

	// Firebase token verifier (skipped in dev mode, which uses the dev-login bypass)
	var verifier services.TokenVerifier
	if !cfg.DevMode {
		if cfg.Firebase.EmulatorHost != "" {
			cfg.Logger.Warn("Using the Firebase Auth emulator, which trusts unsigned tokens")
			verifier, err = auth.NewEmulatorVerifier(cfg.Firebase.ProjectID)
		} else {
			verifier, err = auth.NewVerifier(context.Background(), cfg.Firebase.ProjectID)
		}
		if err != nil {
			cfg.Logger.Error("Error initializing Firebase token verifier", "error", err)
			return nil, err
		}
	}

	// Services
	weatherService := cfg.Weather
	if weatherService == nil {
		weatherService = services.NewWeatherService()
	}
	activitiesService := services.NewActivityService(userDal, activityDal, weatherService, cfg.Geocoder)
	scoreboardService := services.NewScoreboardService(userDal)
	authService := services.NewAuthService(userDal, verifier, cfg.Firebase, cfg.DevMode)
	userService := services.NewUserService(userDal, certificationDal)
	certificationsService := services.NewCertificationService(certificationDal, userDal)

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)
	themeMiddleware := middleware.ThemeMiddleware()

	// Apply user middleware globally so theme middleware can read the user
	app.Use(authMiddleware.User)
	app.Use(themeMiddleware)

	// Route for the root path
	routes.HomeRoutes(app, userService, authMiddleware)
	routes.AuthRoutes(app, authService, cfg.DevMode, cfg.AuthRateLimit, cfg.AuthRateWindow)
	routes.UserRoutes(app, userService, authMiddleware)
	routes.ActivitiesRoutes(app, activitiesService, authMiddleware)
	routes.ScoreboardRoutes(app, scoreboardService, authMiddleware)
	routes.CertificationRoutes(app, certificationsService, authMiddleware)
	if cfg.DevMode {
		routes.DevRoutes(app)
	}

	return app, nil
}

// proxyHeader is the configured header, or empty without trusted proxies (warning if one was set)
func proxyHeader(cfg *Config) string {
	if len(cfg.TrustedProxies) > 0 {
		return cfg.ProxyHeader
	}
	if cfg.ProxyHeader != "" {
		cfg.Logger.Warn("PROXY_HEADER is ignored without TRUSTED_PROXIES", "header", cfg.ProxyHeader)
	}
	return ""
}

// logPanic logs a recovered panic with its stack, in the same structured form as the other errors
func logPanic(c *fiber.Ctx, recovered any) {
	utils.Log(c).Error("panic recovered", "panic", fmt.Sprint(recovered), "stack", string(debug.Stack()))
}

func main() {
	cfg := DefaultConfig()
	slog.SetDefault(cfg.Logger)

	// Connect to database
	err := database.Connect()
	if err != nil {
		cfg.Logger.Error("Error connecting to database", "error", err)
		os.Exit(1)
	}

	app, err := setupApp(cfg)
	if err != nil {
		cfg.Logger.Error("Error setting up app", "error", err)
		os.Exit(1)
	}

	cfg.Logger.Info("listening", "addr", cfg.ListenAddr, "dev_mode", cfg.DevMode)
	if err := app.Listen(cfg.ListenAddr); err != nil {
		cfg.Logger.Error("Server stopped", "error", err)
		os.Exit(1)
	}
}
