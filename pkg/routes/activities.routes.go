package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/services"
	"github.com/mogensen/logbook/pkg/utils/middleware"
)

// TodoRoutes contains all routes relative to /todo
func ActivitiesRoutes(app fiber.Router) {
	r := app.Group("/activities").Use(middleware.Auth)

	r.Get("/create", services.CreateClimbingActivityPage)
	r.Post("/create", services.CreateClimbingActivity)
	r.Get("/list", services.GetClimbingActivitys)
	r.Get("/:ActivityID", services.GetClimbingActivity)
	r.Get("/:ActivityID/edit", services.EditClimbingActivity)
	r.Post("/:ActivityID", services.UpdateClimbingActivity)
	r.Delete("/:ActivityID", services.DeleteClimbingActivity)
}
