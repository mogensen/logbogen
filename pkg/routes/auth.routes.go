package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/services"
	"github.com/mogensen/logbook/pkg/utils/middleware"
)

// AuthRoutes containes all the auth routes
func AuthRoutes(app fiber.Router, authService *services.AuthService, authMiddleware *middleware.AuthMiddleware) {
	r := app.Group("/auth")

	r.Get("/signup", authService.SignupPageHandler)
	r.Post("/signup", authService.SignupHandler)
	r.Post("/login", authService.LoginHandler)
	r.Get("/logout", authService.LogoutHandler)

	r = app.Group("/users").Use(authMiddleware.Auth)
	r.Get("/list", authService.GetUsersHandler)
	r.Get("/:UserID", authService.GetUserHandler)
}
