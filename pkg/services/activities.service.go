package services

import (
	"errors"
	"log/slog"
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
		"ClimbingTypes": types.ClimbingTypeNames,
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
	res, err := GetActivitiesForUser(utils.GetUser(c).ID)
	if err != nil {
		return err
	}

	accept := c.Accepts("html", "json")
	if accept == "json" {
		return c.JSON(res)
	}

	return c.Render("climbingactivities/list", fiber.Map{
		"ClimbingActivities": &res,
	})
}

func GetActivitiesForUser(userId uint64) ([]*types.ClimbingActivity, error) {
	activities := []dal.ClimbingActivity{}

	err := dal.FindClimbingActivitiesByUser(&activities, userId).Error
	if err != nil {
		return nil, fiber.NewError(fiber.StatusConflict, err.Error())
	}

	userMap, err := getUserMap()
	if err != nil {
		return nil, err
	}

	res := make([]*types.ClimbingActivity, len(activities))
	for i, activity := range activities {
		res[i] = mapActivityFromDal(&activity, userMap)
	}
	return res, nil
}

func GetPendingActivitiesForUser(c *fiber.Ctx) error {
	activities := []dal.ClimbingActivity{}

	err := dal.FindPendingActivitiesByUser(&activities, utils.GetUser(c).ID).Error
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

	return c.Render("climbingactivities/pending", fiber.Map{
		"ClimbingActivities": &res,
	})
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
		"ClimbingTypes":    types.ClimbingTypeNames,
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

// CloneClimbingActivity clones a ClimbingActivity
func CloneClimbingActivity(c *fiber.Ctx) error {
	activityID := c.Params("ActivityID")

	if activityID == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid ActivityID")
	}

	activity := &dal.ClimbingActivity{}

	err := dal.FindClimbingActivityToClone(activity, activityID, utils.GetUser(c).ID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "ClimbingActivity not found")
	}

	// Clone the activity
	clonedActivity := &dal.ClimbingActivity{}
	clonedActivity.Date = activity.Date
	clonedActivity.Lat = activity.Lat
	clonedActivity.Lng = activity.Lng
	clonedActivity.Location = activity.Location
	clonedActivity.Comment = activity.Comment
	clonedActivity.Type = activity.Type
	clonedActivity.OtherType = activity.OtherType
	clonedActivity.Role = activity.Role
	clonedActivity.Participants = participantsForClone(activity.Participants, utils.GetUser(c).ID, *activity.User)
	clonedActivity.User = &utils.GetUser(c).ID

	userMap, err := getUserMap()
	if err != nil {
		return err
	}

	res := mapActivityFromDal(clonedActivity, userMap)

	return c.Render("climbingactivities/create", fiber.Map{
		"ClimbingActivity": res,
		"ClimbingTypes":    types.ClimbingTypeNames,
	})
}

func participantsForClone(participants []uint64, currentUser uint64, originalUser uint64) []uint64 {
	// Add original user as participant
	res := []uint64{originalUser}

	// Remove self from participants
	for _, r := range participants {
		uID := r
		if uID != currentUser {
			res = append(res, uID)
		}
	}
	slog.Error("participantsForClone", "res", res, "currentUser", currentUser, "originalUser", originalUser)
	return res
}

func getUserMap() (map[uint64]types.User, error) {
	dalUsers := &[]dal.User{}

	err := dal.FindUsers(dalUsers).Error
	if err != nil {
		return nil, fiber.NewError(fiber.StatusConflict, err.Error())
	}

	users := []types.User{}
	for _, u := range *dalUsers {
		users = append(users, *types.UserFromDal(&u, nil))
	}

	userMap := make(map[uint64]types.User)
	for i, user := range users {
		userMap[user.ID] = (users)[i]

	}
	return userMap, nil
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
		UserId:          activity.User,
		User:            userMap[*activity.User],
		ParticipantsIDs: activity.Participants,
	}
}
