package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/mogensen/logbook/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func CreateClimbingActivityPage(c *fiber.Ctx) error {
	// Render the create climbing activity page
	return c.Render("climbingactivities/create", fiber.Map{
		"ClimbingTypes": types.ClimbingTypes,
	})
}

// CreateClimbingActivity is responsible for create ClimbingActivity
func CreateClimbingActivity(c *fiber.Ctx) error {
	b := new(types.CreateDTO)

	if err := utils.ParseBodyAndValidate(c, b); err != nil {
		return err
	}

	activity := &dal.ClimbingActivity{
		ID:           uuid.New(),
		Date:         time.Time(b.ClimbingActivity.Date),
		Lat:          b.ClimbingActivity.Lat,
		Lng:          b.ClimbingActivity.Lng,
		Location:     b.ClimbingActivity.Location,
		Type:         b.ClimbingActivity.Type.String(),
		OtherType:    b.ClimbingActivity.OtherType,
		Role:         b.ClimbingActivity.Role,
		Comment:      b.ClimbingActivity.Comment,
		Participants: b.ClimbingActivity.Participants,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		User:         utils.GetUser(c),
	}

	geo, _ := ReverseGeocode(b.ClimbingActivity.Lat, b.ClimbingActivity.Lng)
	activity.Location = geo.SimpleDisplayName()

	if err := dal.CreateClimbingActivity(activity).Error; err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return c.Redirect("/activities/" + activity.ID.String())
}

// GetClimbingActivitys returns the ClimbingActivitys list
func GetClimbingActivitys(c *fiber.Ctx) error {
	activities := []types.ClimbingActivity{}

	err := dal.FindClimbingActivitiesByUser(&activities, utils.GetUser(c)).Error
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	accept := c.Accepts("html", "json")

	if accept == "json" {
		return c.JSON(activities)
	}

	return c.Render("climbingactivities/list", fiber.Map{
		"ClimbingActivities": &activities,
	})
}

// GetClimbingActivity return a single ClimbingActivity
func GetClimbingActivity(c *fiber.Ctx) error {
	activityID := c.Params("ActivityID")

	if activityID == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid ActivityID")
	}

	activity := &types.ClimbingActivity{}

	err := dal.FindClimbingActivityByUser(activity, activityID, utils.GetUser(c)).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(&types.ClimbingActivityCreate{})
	}

	return c.Render("climbingactivities/show", fiber.Map{
		"ClimbingActivity": activity,
	})
}

// EditClimbingActivity return a single ClimbingActivity
func EditClimbingActivity(c *fiber.Ctx) error {
	activityID := c.Params("ActivityID")

	if activityID == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid ActivityID")
	}

	activity := &types.ClimbingActivity{}

	err := dal.FindClimbingActivityByUser(activity, activityID, utils.GetUser(c)).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(&types.ClimbingActivityCreate{})
	}

	return c.Render("climbingactivities/edit", fiber.Map{
		"ClimbingActivity": activity,
		"ClimbingTypes":    types.ClimbingTypes,
	})
}

// DeleteClimbingActivity deletes a single ClimbingActivity
func DeleteClimbingActivity(c *fiber.Ctx) error {
	activityID := c.Params("ActivityID")

	if activityID == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid ActivityID")
	}

	res := dal.DeleteClimbingActivity(activityID, utils.GetUser(c))
	if res.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusConflict, "Unable to delete ClimbingActivity")
	}

	err := res.Error
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return c.JSON(&types.MsgResponse{
		Message: "ClimbingActivity successfully deleted",
	})
}

// UpdateClimbingActivity ClimbingActivity
func UpdateClimbingActivity(c *fiber.Ctx) error {
	b := new(types.CreateDTO)
	activityID := c.Params("ActivityID")

	if activityID == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid ActivityID")
	}

	if err := utils.ParseBodyAndValidate(c, b); err != nil {
		return err
	}

	activity := &dal.ClimbingActivity{
		Date:         time.Time(b.ClimbingActivity.Date),
		Lat:          b.ClimbingActivity.Lat,
		Lng:          b.ClimbingActivity.Lng,
		Location:     b.ClimbingActivity.Location,
		Type:         b.ClimbingActivity.Type.String(),
		OtherType:    b.ClimbingActivity.OtherType,
		Role:         b.ClimbingActivity.Role,
		Comment:      b.ClimbingActivity.Comment,
		Participants: b.ClimbingActivity.Participants,
		UpdatedAt:    time.Now(),
		User:         utils.GetUser(c),
	}

	geo, _ := ReverseGeocode(b.ClimbingActivity.Lat, b.ClimbingActivity.Lng)
	activity.Location = geo.SimpleDisplayName()

	err := dal.UpdateClimbingActivity(activityID, utils.GetUser(c), activity).Error
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return c.Redirect("/activities/" + activityID)
}
