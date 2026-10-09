package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/config"
	"github.com/mogensen/logbook/pkg/services"
	"github.com/mogensen/logbook/pkg/utils"
	"github.com/mogensen/logbook/pkg/utils/middleware"
)

func HomeRoutes(app fiber.Router, userService *services.UserService, authMiddleware *middleware.AuthMiddleware) {
	app.Get("/", authMiddleware.User, IndexPage(userService))
}

func IndexPage(userService *services.UserService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := utils.GetUser(c)

		data := fiber.Map{}
		if user != nil {
			data = userService.ProfileData(user, user)
			// The heatmap goes below the map on the front page, not in the profile header
			data["HeatmapBelow"] = true
		}
		if user == nil {
			data["User"] = user
		}
		data["Categories"] = config.AllActivityCategories
		data["Types"] = config.AllActivityTypes
		return c.Render("home/index", data)
	}
}
