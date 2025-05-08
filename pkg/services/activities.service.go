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
		"ClimbingActivity": &types.ClimbingActivity{
			Date: types.Date(time.Now()),
		},
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
		Participants: b.ClimbingActivity.ParticipantsIDs,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		User:         &utils.GetUser(c).ID,
	}

	geo, _ := ReverseGeocode(b.ClimbingActivity.Lat, b.ClimbingActivity.Lng)
	activity.Location = geo.SimpleDisplayName()

	if err := dal.CreateClimbingActivity(activity).Error; err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return c.Redirect("/activities/" + activity.ID.String())
}

// GetClimbingActivities returns the ClimbingActivitys list
func GetClimbingActivities(c *fiber.Ctx) error {
	activities := []dal.ClimbingActivity{}

	err := dal.FindClimbingActivitiesByUser(&activities, utils.GetUser(c).ID).Error
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	userMap, err := getUserMap()
	if err != nil {
		return err
	}

	res := make([]*types.ClimbingActivity, len(activities))
	for i, activity := range activities {
		res[i] = mapActivityFromDal(&activity, userMap)
	}

	accept := c.Accepts("html", "json")
	if accept == "json" {
		return c.JSON(res)
	}

	return c.Render("climbingactivities/list", fiber.Map{
		"ClimbingActivities": &res,
	})
}

func getUserMap() (map[uint64]types.User, error) {
	users := &[]types.User{}

	err := dal.FindUsers(users).Error
	if err != nil {
		return nil, fiber.NewError(fiber.StatusConflict, err.Error())
	}

	userMap := make(map[uint64]types.User)
	for i, user := range *users {
		userMap[user.ID] = (*users)[i]

	}
	return userMap, nil
}

// GetClimbingActivity return a single ClimbingActivity
func GetClimbingActivity(c *fiber.Ctx) error {
	activityID := c.Params("ActivityID")

	if activityID == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid ActivityID")
	}

	activity := &dal.ClimbingActivity{}

	err := dal.FindClimbingActivityByUser(activity, activityID, utils.GetUser(c).ID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(&types.ClimbingActivityCreate{})
	}

	userMap, err := getUserMap()
	if err != nil {
		return err
	}

	res := mapActivityFromDal(activity, userMap)

	return c.Render("climbingactivities/show", fiber.Map{
		"ClimbingActivity": res,
	})
}

func mapActivityFromDal(activity *dal.ClimbingActivity, userMap map[uint64]types.User) *types.ClimbingActivity {

	participants := make([]types.User, len(activity.Participants))
	for i, participant := range activity.Participants {
		if user, ok := userMap[participant]; ok {
			participants[i] = user
		}
	}

	return &types.ClimbingActivity{
		ID:              activity.ID,
		Date:            types.Date(activity.Date),
		Lat:             activity.Lat,
		Lng:             activity.Lng,
		Location:        activity.Location,
		Type:            types.ClimbingType(activity.Type),
		OtherType:       activity.OtherType,
		Role:            activity.Role,
		Comment:         activity.Comment,
		Participants:    participants,
		CreatedAt:       activity.CreatedAt,
		UpdatedAt:       activity.UpdatedAt,
		User:            activity.User,
		ParticipantsIDs: activity.Participants,
	}
}

// EditClimbingActivity return a single ClimbingActivity
func EditClimbingActivity(c *fiber.Ctx) error {
	activityID := c.Params("ActivityID")

	if activityID == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid ActivityID")
	}

	activity := &dal.ClimbingActivity{}

	err := dal.FindClimbingActivityByUser(activity, activityID, utils.GetUser(c).ID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(&types.ClimbingActivityCreate{})
	}

	userMap, err := getUserMap()
	if err != nil {
		return err
	}

	res := mapActivityFromDal(activity, userMap)

	return c.Render("climbingactivities/edit", fiber.Map{
		"ClimbingActivity": res,
		"ClimbingTypes":    types.ClimbingTypes,
	})
}

// DeleteClimbingActivity deletes a single ClimbingActivity
func DeleteClimbingActivity(c *fiber.Ctx) error {
	activityID := c.Params("ActivityID")

	if activityID == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid ActivityID")
	}

	res := dal.DeleteClimbingActivity(activityID, utils.GetUser(c).ID)
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
		Participants: b.ClimbingActivity.ParticipantsIDs,
		UpdatedAt:    time.Now(),
		User:         &utils.GetUser(c).ID,
	}

	geo, _ := ReverseGeocode(b.ClimbingActivity.Lat, b.ClimbingActivity.Lng)
	activity.Location = geo.SimpleDisplayName()

	err := dal.UpdateClimbingActivity(activityID, utils.GetUser(c).ID, activity).Error
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return c.Redirect("/activities/" + activityID)
}
