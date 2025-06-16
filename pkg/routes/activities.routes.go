package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/services"
	"github.com/mogensen/logbook/pkg/utils/middleware"
)

// ActivitiesRoutes contains all routes relative to /activities
func ActivitiesRoutes(app fiber.Router, activityService *services.ActivityService, authMiddleware *middleware.AuthMiddleware) {
	r := app.Group("/activities").Use(authMiddleware.Auth)
	r.Get("/types", activityService.GetActivityTypes)
	r.Get("/categories", activityService.GetActivityCategories)

	r.Get("/create", activityService.CreateActivityPage)
	r.Post("/create", activityService.CreateActivity)
	r.Get("/list", activityService.GetActivities)
	r.Get("/pending", activityService.GetPendingActivitiesForUser)
	r.Get("/:ActivityID", activityService.GetActivity)
	r.Get("/:ActivityID/edit", activityService.EditActivity)
	r.Post("/:ActivityID", activityService.UpdateActivity)
	r.Post("/:ActivityID/delete", activityService.DeleteActivity)

	r.Get("/clone/:ActivityID", activityService.CloneActivity)
}
