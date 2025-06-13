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

func CreateActivityPage(c *fiber.Ctx) error {
	// Render the create activity page
	return c.Render("activities/create", fiber.Map{
		"Activity": &types.Activity{
			Date: types.Date(time.Now()),
		},
	})
}

// CreateActivity is responsible for creating an Activity
func CreateActivity(c *fiber.Ctx) error {
	b := new(types.CreateDTO)

	if err := utils.ParseBodyAndValidate(c, b); err != nil {
		return c.Render("activities/create", fiber.Map{
			"Activity": b.Activity,
			"error":    err.Message,
		})
	}

	activity := &dal.Activity{
		ID:           uuid.New(),
		Date:         time.Time(b.Activity.Date),
		Lat:          b.Activity.Lat,
		Lng:          b.Activity.Lng,
		Location:     b.Activity.Location,
		Category:     b.Activity.CategoryID,
		Role:         b.Activity.Role,
		Comment:      b.Activity.Comment,
		Participants: b.Activity.ParticipantsIDs,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		User:         &utils.GetUser(c).ID,
		Type:         b.Activity.TypeID,
		OtherType:    b.Activity.OtherType,
	}

	geo, _ := ReverseGeocode(b.Activity.Lat, b.Activity.Lng)
	activity.Location = geo.SimpleDisplayName()

	if err := dal.CreateActivity(activity).Error; err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return c.Redirect("/activities/" + activity.ID.String())
}

// GetActivities returns the Activities list
func GetActivities(c *fiber.Ctx) error {
	res, err := GetActivitiesForUser(utils.GetUser(c).ID)
	if err != nil {
		return err
	}

	accept := c.Accepts("html", "json")
	if accept == "json" {
		return c.JSON(res)
	}

	return c.Render("activities/list", fiber.Map{
		"Activities": &res,
	})
}

func GetActivitiesForUser(userId uint64) ([]*types.Activity, error) {
	activities := []dal.Activity{}

	err := dal.FindActivitiesByUser(&activities, userId).Error
	if err != nil {
		return nil, fiber.NewError(fiber.StatusConflict, err.Error())
	}

	userMap, err := getUserMap()
	if err != nil {
		return nil, err
	}

	res := make([]*types.Activity, len(activities))
	for i, activity := range activities {
		res[i] = mapActivityFromDal(&activity, userMap)
	}
	return res, nil
}

func GetPendingActivitiesForUser(c *fiber.Ctx) error {
	activities := []dal.Activity{}

	err := dal.FindPendingActivitiesByUser(&activities, utils.GetUser(c).ID).Error
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	userMap, err := getUserMap()
	if err != nil {
		return err
	}

	res := make([]*types.Activity, len(activities))
	for i, activity := range activities {
		res[i] = mapActivityFromDal(&activity, userMap)
	}

	accept := c.Accepts("html", "json")
	if accept == "json" {
		return c.JSON(res)
	}

	return c.Render("activities/pending", fiber.Map{
		"Activities": &res,
	})
}

// GetActivity return a single Activity
func GetActivity(c *fiber.Ctx) error {
	activityID := c.Params("ActivityID")

	if activityID == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid ActivityID")
	}

	activity := &dal.Activity{}

	err := dal.FindActivityByUser(activity, activityID, utils.GetUser(c).ID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(&types.CreateDTO{})
	}

	userMap, err := getUserMap()
	if err != nil {
		return err
	}

	res := mapActivityFromDal(activity, userMap)

	return c.Render("activities/show", fiber.Map{
		"Activity": res,
	})
}

// EditActivity return a single Activity
func EditActivity(c *fiber.Ctx) error {
	activityID := c.Params("ActivityID")

	if activityID == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid ActivityID")
	}

	activity := &dal.Activity{}

	err := dal.FindActivityByUser(activity, activityID, utils.GetUser(c).ID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(&types.CreateDTO{})
	}

	userMap, err := getUserMap()
	if err != nil {
		return err
	}

	res := mapActivityFromDal(activity, userMap)

	return c.Render("activities/edit", fiber.Map{"Activity": res})
}

// DeleteActivity deletes a single Activity
func DeleteActivity(c *fiber.Ctx) error {
	activityID := c.Params("ActivityID")

	if activityID == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid ActivityID")
	}

	res := dal.DeleteActivity(activityID, utils.GetUser(c).ID)
	if res.RowsAffected == 0 {
		return fiber.NewError(fiber.StatusConflict, "Unable to delete Activity")
	}

	err := res.Error
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return c.JSON(&types.MsgResponse{
		Message: "Activity successfully deleted",
	})
}

// UpdateActivity updates an Activity
func UpdateActivity(c *fiber.Ctx) error {
	b := new(types.CreateDTO)
	activityID := c.Params("ActivityID")

	if activityID == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid ActivityID")
	}

	if err := utils.ParseBodyAndValidate(c, b); err != nil {
		return err
	}

	activity := &dal.Activity{
		Date:         time.Time(b.Activity.Date),
		Lat:          b.Activity.Lat,
		Lng:          b.Activity.Lng,
		Location:     b.Activity.Location,
		Category:     b.Activity.Category.ID,
		Role:         b.Activity.Role,
		Comment:      b.Activity.Comment,
		Participants: b.Activity.ParticipantsIDs,
		UpdatedAt:    time.Now(),
		User:         &utils.GetUser(c).ID,
		Type:         b.Activity.TypeID,
		OtherType:    b.Activity.OtherType,
	}

	geo, _ := ReverseGeocode(b.Activity.Lat, b.Activity.Lng)
	activity.Location = geo.SimpleDisplayName()

	err := dal.UpdateActivity(activityID, utils.GetUser(c).ID, activity).Error
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return c.Redirect("/activities/" + activityID)
}

// CloneActivity clones an Activity
func CloneActivity(c *fiber.Ctx) error {
	activityID := c.Params("ActivityID")

	if activityID == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid ActivityID")
	}

	activity := &dal.Activity{}
	err := dal.FindActivityToClone(activity, activityID, utils.GetUser(c).ID).Error
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	// Create a new activity with the same data
	newActivity := &dal.Activity{
		ID:           uuid.New(),
		Date:         activity.Date,
		Lat:          activity.Lat,
		Lng:          activity.Lng,
		Location:     activity.Location,
		Category:     activity.Category,
		Role:         activity.Role,
		Comment:      activity.Comment,
		Participants: participantsForClone(activity.Participants, utils.GetUser(c).ID, *activity.User),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		User:         &utils.GetUser(c).ID,
		Type:         activity.Type,
		OtherType:    activity.OtherType,
	}

	if err := dal.CreateActivity(newActivity).Error; err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return c.Redirect("/activities/" + newActivity.ID.String())
}

func participantsForClone(participants []uint64, currentUser uint64, originalUser uint64) []uint64 {
	// Remove the original user and add the current user
	newParticipants := make([]uint64, 0, len(participants))
	for _, p := range participants {
		if p != originalUser {
			newParticipants = append(newParticipants, p)
		}
	}
	newParticipants = append(newParticipants, currentUser)
	return newParticipants
}

func getUserMap() (map[uint64]types.User, error) {
	users := &[]dal.User{}
	err := dal.FindUsers(users).Error
	if err != nil {
		return nil, err
	}

	userMap := make(map[uint64]types.User)
	for _, user := range *users {
		userMap[user.ID] = *types.UserFromDal(&user, nil)
	}
	return userMap, nil
}

func mapActivityFromDal(activity *dal.Activity, userMap map[uint64]types.User) *types.Activity {
	participants := make([]types.User, 0, len(activity.Participants))
	for _, p := range activity.Participants {
		if user, ok := userMap[p]; ok {
			participants = append(participants, user)
		}
	}

	return &types.Activity{
		ID:              activity.ID,
		Date:            types.Date(activity.Date),
		Lat:             activity.Lat,
		Lng:             activity.Lng,
		Location:        activity.Location,
		CategoryID:      activity.Category,
		Category:        *types.CategoryByID(activity.Category),
		Type:            *types.ActivityTypeByID(activity.Type),
		TypeID:          activity.Type,
		OtherType:       activity.OtherType,
		Role:            activity.Role,
		Comment:         activity.Comment,
		Participants:    participants,
		ParticipantsIDs: activity.Participants,
		CreatedAt:       activity.CreatedAt,
		UpdatedAt:       activity.UpdatedAt,
		User:            userMap[*activity.User],
	}
}

// GetActivityTypes returns the activity types for a given category
func GetActivityTypes(c *fiber.Ctx) error {
	category := c.Query("category")
	if category == "" {
		// If no category is specified, return all categories and their types
		return c.JSON(types.AllActivityTypes)
	}

	categoryTypes := make([]types.ActivityType, 0, len(types.AllActivityTypes))
	for _, activityType := range types.AllActivityTypes {
		if activityType.Category == category {
			categoryTypes = append(categoryTypes, activityType)
		}
	}

	return c.JSON(categoryTypes)
}

// GetActivityCategories returns all available activity categories
func GetActivityCategories(c *fiber.Ctx) error {
	return c.JSON(types.AllActivityCategories)
}
