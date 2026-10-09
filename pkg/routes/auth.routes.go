package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/mogensen/logbook/pkg/services"
)

// AuthRoutes contains all the auth routes. When devMode is true the local
// dev-login bypass route is registered as well. Every route is rate limited to maxRequests per window and client IP.
func AuthRoutes(app fiber.Router, authService *services.AuthService, devMode bool, maxRequests int, window time.Duration) {
	r := app.Group("/auth")

	r.Use(limiter.New(limiter.Config{
		Max:        maxRequests,
		Expiration: window,
		LimitReached: func(c *fiber.Ctx) error {
			return fiber.NewError(fiber.StatusTooManyRequests, "For mange forsøg. Prøv igen senere.")
		},
	}))

	r.Get("/login", authService.LoginPageHandler)
	r.Get("/callback", authService.CallbackHandler)
	r.Post("/logout", authService.LogoutHandler)

	if devMode {
		r.Post("/dev-login", authService.DevLoginHandler)
	}
}
