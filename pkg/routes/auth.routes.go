package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/services"
)

// AuthRoutes containes all the auth routes
func AuthRoutes(app fiber.Router) {
	r := app.Group("/auth")

	r.Get("/signup", services.SignupPage)
	r.Post("/signup", services.Signup)
	r.Post("/login", services.Login)
	r.Get("/logout", services.Logout)
}
