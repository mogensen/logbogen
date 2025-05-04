package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/mogensen/logbook/pkg/utils"
)

func indexPage(c *fiber.Ctx) error {
	if utils.GetUser(c) == nil {
		return c.Render("index", fiber.Map{})
	}

	activities := &[]types.ClimbingActivity{}

	err := dal.FindClimbingActivitiesByUser(activities, utils.GetUser(c)).Error
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	// render the root page as HTML
	return c.Render("index", fiber.Map{
		"ClimbingActivities": *activities,
	})
}
