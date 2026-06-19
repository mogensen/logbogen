package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/services"
)

// AuthRoutes contains all the auth routes. When devMode is true the local
// dev-login bypass route is registered as well.
func AuthRoutes(app fiber.Router, authService *services.AuthService, devMode bool) {
	r := app.Group("/auth")

	r.Get("/login", authService.LoginPageHandler)
	r.Get("/callback", authService.CallbackHandler)
	r.Get("/logout", authService.LogoutHandler)

	if devMode {
		r.Post("/dev-login", authService.DevLoginHandler)
	}
}
