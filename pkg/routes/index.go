package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/utils/middleware"
)

func HomeRoutes(app fiber.Router, csrfMiddleware fiber.Handler, authMiddleware *middleware.AuthMiddleware) {
	app.Get("/", csrfMiddleware, authMiddleware.User, IndexPage)
}

// IndexPage handles the root route of the application
func IndexPage(c *fiber.Ctx) error {
	return c.Render("home/index", fiber.Map{})
}
