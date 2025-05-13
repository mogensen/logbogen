package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/services"
	"github.com/mogensen/logbook/pkg/utils/middleware"
)

// AuthRoutes containes all the auth routes
func AuthRoutes(app fiber.Router) {
	r := app.Group("/auth")

	r.Get("/signup", services.SignupPage)
	r.Post("/signup", services.Signup)
	r.Post("/login", services.Login)
	r.Get("/logout", services.Logout)

	r = app.Group("/users").Use(middleware.Auth)
	r.Get("/list", services.GetUsers)
	r.Get("/:UserID", services.GetUser)
}
