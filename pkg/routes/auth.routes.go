package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/services"
)

// AuthRoutes containes all the auth routes
func AuthRoutes(app fiber.Router, authService *services.AuthService) {
	r := app.Group("/auth")

	r.Get("/login", authService.LoginPageHandler)
	r.Get("/signup", authService.SignupPageHandler)
	r.Post("/signup", authService.SignupHandler)
	r.Post("/login", authService.LoginHandler)
	r.Get("/logout", authService.LogoutHandler)
}
