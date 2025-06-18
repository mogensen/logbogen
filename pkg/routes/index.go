package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/config"
	"github.com/mogensen/logbook/pkg/utils/middleware"
)

func HomeRoutes(app fiber.Router, authMiddleware *middleware.AuthMiddleware) {
	app.Get("/", authMiddleware.User, IndexPage)
}

// IndexPage handles the root route of the application
func IndexPage(c *fiber.Ctx) error {
	return c.Render("home/index", fiber.Map{
		"Categories": config.AllActivityCategories,
		"Types":      config.AllActivityTypes,
	})
}
