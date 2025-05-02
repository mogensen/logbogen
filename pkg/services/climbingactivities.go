package services

import (
	"errors"
	"time"

	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/mogensen/logbook/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func CreateClimbingActivityPage(c *fiber.Ctx) error {
	// Render the create climbing activity page
	return c.Render("climbingactivities/create", fiber.Map{})
}

// CreateClimbingActivity is responsible for create ClimbingActivity
func CreateClimbingActivity(c *fiber.Ctx) error {
	b := new(types.CreateDTO)

	if err := utils.ParseBodyAndValidate(c, b); err != nil {
		return err
	}

	d := &dal.ClimbingActivity{
		Date:         b.ClimbingActivity.Date,
		Lat:          b.ClimbingActivity.Lat,
		Lng:          b.ClimbingActivity.Lng,
		Location:     b.ClimbingActivity.Location,
		Type:         b.ClimbingActivity.Type.String(),
		Role:         b.ClimbingActivity.Role,
		Comment:      b.ClimbingActivity.Comment,
		Participants: b.ClimbingActivity.Participants,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := dal.CreateClimbingActivity(d).Error; err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return c.JSON(&types.ClimbingActivityCreate{
		ClimbingActivity: &types.ClimbingActivity{
			ID:           d.ID,
			User:         d.User,
			Date:         d.Date,
			Lat:          d.Lat,
			Lng:          d.Lng,
			Location:     d.Location,
			Type:         types.MapClimbingType(d.Type),
			Role:         d.Role,
			Comment:      d.Comment,
			Participants: []uint64(d.Participants),
			CreatedAt:    d.CreatedAt,
			UpdatedAt:    d.UpdatedAt,
		},
	})
}

// GetClimbingActivitys returns the ClimbingActivitys list
func GetClimbingActivitys(c *fiber.Ctx) error {
	d := &[]types.ClimbingActivity{}

	err := dal.FindClimbingActivitysByUser(d, utils.GetUser(c)).Error
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return c.Render("climbingactivities/list", fiber.Map{
		"ClimbingActivities": d,
	})
}

// GetClimbingActivity return a single ClimbingActivity
func GetClimbingActivity(c *fiber.Ctx) error {
	ClimbingActivityID := c.Params("ClimbingActivityID")

	if ClimbingActivityID == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid ClimbingActivityID")
	}

	d := &types.ClimbingActivity{}

	err := dal.FindClimbingActivityByUser(d, ClimbingActivityID, utils.GetUser(c)).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(&types.ClimbingActivityCreate{})
	}

	return c.JSON(&types.ClimbingActivityCreate{
		ClimbingActivity: d,
	})
}

// DeleteClimbingActivity deletes a single ClimbingActivity
func DeleteClimbingActivity(c *fiber.Ctx) error {
	ClimbingActivityID := c.Params("ClimbingActivityID")

	if ClimbingActivityID == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid ClimbingActivityID")
	}

	res := dal.DeleteClimbingActivity(ClimbingActivityID, utils.GetUser(c))
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
	ClimbingActivityID := c.Params("ClimbingActivityID")

	if ClimbingActivityID == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid ClimbingActivityID")
	}

	if err := utils.ParseBodyAndValidate(c, b); err != nil {
		return err
	}

	d := &dal.ClimbingActivity{
		Date:         b.ClimbingActivity.Date,
		Lat:          b.ClimbingActivity.Lat,
		Lng:          b.ClimbingActivity.Lng,
		Location:     b.ClimbingActivity.Location,
		Type:         b.ClimbingActivity.Type.String(),
		Role:         b.ClimbingActivity.Role,
		Comment:      b.ClimbingActivity.Comment,
		Participants: b.ClimbingActivity.Participants,
		UpdatedAt:    time.Now(),
	}
	err := dal.UpdateClimbingActivity(ClimbingActivityID, utils.GetUser(c), d).Error
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return c.JSON(&types.MsgResponse{
		Message: "ClimbingActivity successfully updated",
	})
}
