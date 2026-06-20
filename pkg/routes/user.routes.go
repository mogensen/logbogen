package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/services"
	"github.com/mogensen/logbook/pkg/utils/middleware"
)

// UserRoutes containes all the auth routes
func UserRoutes(app fiber.Router, userService *services.UserService, authMiddleware *middleware.AuthMiddleware) {
	r := app.Group("/users").Use(authMiddleware.Auth)
	r.Get("/list", userService.GetUsersHandler)
	r.Get("/:UserID", userService.GetUserHandler)
	r.Post("/theme", userService.UpdateThemeHandler)
}
