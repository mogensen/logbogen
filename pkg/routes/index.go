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
		activityBars := []services.ActivityMonthCategoryBar{}
		user := utils.GetUser(c)
		if user != nil {
			activityBars = userService.GetUserActivityBars(user)
		}
		return c.Render("home/index", fiber.Map{
			"User":         utils.GetUser(c),
			"ActivityBars": activityBars,
			"Categories":   config.AllActivityCategories,
			"Types":        config.AllActivityTypes,
		})
	}
}
