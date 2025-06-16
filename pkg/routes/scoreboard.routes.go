package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/services"
	"github.com/mogensen/logbook/pkg/utils/middleware"
)

// ScoreboardRoutes contains all routes relative to /scoreboard
func ScoreboardRoutes(app fiber.Router, scoreboardService *services.ScoreboardService, authMiddleware *middleware.AuthMiddleware) {
	r := app.Group("/scoreboard").Use(authMiddleware.Auth)
	r.Get("/", authMiddleware.User, scoreboardService.GetScoreboard)
}
