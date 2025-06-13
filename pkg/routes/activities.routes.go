package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/services"
	"github.com/mogensen/logbook/pkg/utils/middleware"
)

// ActivitiesRoutes contains all routes relative to /activities
func ActivitiesRoutes(app fiber.Router) {
	r := app.Group("/activities").Use(middleware.Auth)
	r.Get("/types", services.GetActivityTypes)
	r.Get("/categories", services.GetActivityCategories)

	r.Get("/create", services.CreateActivityPage)
	r.Post("/create", services.CreateActivity)
	r.Get("/list", services.GetActivities)
	r.Get("/pending", services.GetPendingActivitiesForUser)
	r.Get("/:ActivityID", services.GetActivity)
	r.Get("/:ActivityID/edit", services.EditActivity)
	r.Post("/:ActivityID", services.UpdateActivity)
	r.Post("/:ActivityID/delete", services.DeleteActivity)

	r.Get("/clone/:ActivityID", services.CloneActivity)
}
