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

// ActivityService handles activity related operations
type ActivityService struct {
	activityDal dal.ActivityDal
	userDal     dal.UserDal
	weatherSvc  *WeatherService
}

// NewActivityService creates a new instance of ActivityService
func NewActivityService(userDal dal.UserDal, activityDal dal.ActivityDal, weatherSvc *WeatherService) *ActivityService {
	return &ActivityService{
		userDal:     userDal,
		activityDal: activityDal,
		weatherSvc:  weatherSvc,
	}
}

// CreateActivityPage renders the create activity page
func (s *ActivityService) CreateActivityPage(c *fiber.Ctx) error {
	// Render the create activity page
	return c.Render("activities/create", fiber.Map{
		"Activity": &types.Activity{
			Date: types.Date(time.Now()),
		},
	})
}

// CreateActivity is responsible for creating an Activity
func (s *ActivityService) CreateActivity(c *fiber.Ctx) error {
	b := new(types.CreateDTO)

	if err := utils.ParseBodyAndValidate(c, b); err != nil {
		return c.Render("activities/create", fiber.Map{
			"Activity": b.Activity,
			"error":    err.Message,
		})
	}

	t := b.Activity.TypeID
	if b.CategoryID == types.Other {
		t = b.OtherType
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
		Type:         t,
	}

	geo, _ := ReverseGeocode(b.Activity.Lat, b.Activity.Lng)
	activity.Location = geo.SimpleDisplayName()

	_, err := s.activityDal.CreateActivity(activity)
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return c.Redirect("/activities/" + activity.ID.String())
}

// GetActivities returns the Activities list
func (s *ActivityService) GetActivities(c *fiber.Ctx) error {
	res, err := s.GetActivitiesForUser(utils.GetUser(c).ID)
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

func (s *ActivityService) GetActivitiesForUser(userId uint64) ([]*types.Activity, error) {
	activities, err := s.activityDal.FindActivitiesByUser(userId)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusConflict, err.Error())
	}

	userMap, err := s.getUserMap()
	if err != nil {
		return nil, err
	}

	res := make([]*types.Activity, len(activities))
	for i, activity := range activities {
		res[i] = types.ActivityFromDal(&activity, userMap)
	}
	return res, nil
}

func (s *ActivityService) GetPendingActivitiesForUser(c *fiber.Ctx) error {
	userID := utils.GetUser(c).ID

	// Fetch pending activities (where user is a participant, but not the owner)
	pendingActivities, err := s.activityDal.FindPendingActivitiesByUser(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	// Fetch all activities logged by the user
	loggedActivities, err := s.activityDal.FindActivitiesByUser(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	userMap, err := s.getUserMap()
	if err != nil {
		return err
	}

	filtered, err := s.filterPending(loggedActivities, pendingActivities, userMap)
	if err != nil {
		return err
	}

	accept := c.Accepts("html", "json")
	if accept == "json" {
		return c.JSON(filtered)
	}

	return c.Render("activities/pending", fiber.Map{
		"Activities": &filtered,
	})
}

func (s *ActivityService) filterPending(loggedActivities []dal.Activity, pendingActivities []dal.Activity, userMap map[uint64]types.User) ([]*types.Activity, error) {
	loggedSet := make(map[string]struct{})
	for _, a := range loggedActivities {
		loggedSet[acticityHash(a)] = struct{}{}
	}

	filtered := make([]*types.Activity, 0, len(pendingActivities))
	for _, activity := range pendingActivities {
		if _, exists := loggedSet[acticityHash(activity)]; !exists {
			filtered = append(filtered, types.ActivityFromDal(&activity, userMap))
		}
	}
	return filtered, nil
}

func acticityHash(a dal.Activity) string {
	return a.Category + "|" + a.Type + "|" + a.Date.Format("2006-01-02")
}

// GetActivity return a single Activity
func (s *ActivityService) GetActivity(c *fiber.Ctx) error {
	activityID := c.Params("ActivityID")

	if activityID == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid ActivityID")
	}

	activity, err := s.activityDal.FindActivityByUser(activityID, utils.GetUser(c).ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(&types.CreateDTO{})
	}
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	userMap, err := s.getUserMap()
	if err != nil {
		return err
	}

	res := types.ActivityFromDal(&activity, userMap)

	// Get weather data for the activity
	weather, err := s.weatherSvc.GetWeather(activity.Lat, activity.Lng, activity.Date)
	if err != nil {
		slog.Error("Error getting weather data", "error", err)
	}

	// Create a weather view model
	type WeatherViewModel struct {
		Temperature float64
		Icon        string
	}

	var weatherVM *WeatherViewModel
	if weather != nil && len(weather.Daily.Temperature2mMax) > 0 && len(weather.Daily.WeatherCode) > 0 {
		weatherVM = &WeatherViewModel{
			Temperature: weather.Daily.Temperature2mMax[0],
			Icon:        GetWeatherIcon(weather.Daily.WeatherCode[0]),
		}
	}

	slog.Info("Weather", "weather", weatherVM)

	return c.Render("activities/show", fiber.Map{
		"Activity": res,
		"Weather":  weatherVM,
	})
}

// EditActivity return a single Activity
func (s *ActivityService) EditActivity(c *fiber.Ctx) error {
	activityID := c.Params("ActivityID")

	if activityID == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid ActivityID")
	}

	activity, err := s.activityDal.FindActivityByUser(activityID, utils.GetUser(c).ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(&types.CreateDTO{})
	}
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	userMap, err := s.getUserMap()
	if err != nil {
		return err
	}

	res := types.ActivityFromDal(&activity, userMap)

	return c.Render("activities/edit", fiber.Map{"Activity": res})
}

// DeleteActivity deletes a single Activity
func (s *ActivityService) DeleteActivity(c *fiber.Ctx) error {
	activityID := c.Params("ActivityID")

	if activityID == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid ActivityID")
	}

	err := s.activityDal.DeleteActivity(activityID, utils.GetUser(c).ID)
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return c.JSON(&types.MsgResponse{
		Message: "Activity successfully deleted",
	})
}

// UpdateActivity updates an Activity
func (s *ActivityService) UpdateActivity(c *fiber.Ctx) error {
	b := new(types.CreateDTO)
	activityID := c.Params("ActivityID")

	if activityID == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid ActivityID")
	}

	if err := utils.ParseBodyAndValidate(c, b); err != nil {
		return err
	}

	t := b.Activity.TypeID
	if b.CategoryID == types.Other {
		t = b.OtherType
	}

	activity := &dal.Activity{
		Date:         time.Time(b.Activity.Date),
		Lat:          b.Activity.Lat,
		Lng:          b.Activity.Lng,
		Location:     b.Activity.Location,
		Category:     b.Activity.CategoryID,
		Role:         b.Activity.Role,
		Comment:      b.Activity.Comment,
		Participants: b.Activity.ParticipantsIDs,
		UpdatedAt:    time.Now(),
		User:         &utils.GetUser(c).ID,
		Type:         t,
	}

	geo, _ := ReverseGeocode(b.Activity.Lat, b.Activity.Lng)
	activity.Location = geo.SimpleDisplayName()

	_, err := s.activityDal.UpdateActivity(activityID, utils.GetUser(c).ID, activity)
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return c.Redirect("/activities/" + activityID)
}

// CloneActivity clones an Activity
func (s *ActivityService) CloneActivity(c *fiber.Ctx) error {
	activityID := c.Params("ActivityID")

	if activityID == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid ActivityID")
	}

	activity, err := s.activityDal.FindActivityToClone(activityID, utils.GetUser(c).ID)
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	// Create a new activity with the same data but new ID and user
	newActivity := &dal.Activity{
		ID:           uuid.New(),
		Date:         activity.Date,
		Lat:          activity.Lat,
		Lng:          activity.Lng,
		Location:     activity.Location,
		Category:     activity.Category,
		Role:         activity.Role,
		Comment:      activity.Comment,
		Participants: s.participantsForClone(activity.Participants, utils.GetUser(c).ID, *activity.User),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		User:         &utils.GetUser(c).ID,
		Type:         activity.Type,
	}

	_, err = s.activityDal.CreateActivity(newActivity)
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return c.Redirect("/activities/" + newActivity.ID.String())
}

func (s *ActivityService) participantsForClone(participants []uint64, currentUser uint64, originalUser uint64) []uint64 {
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

func (s *ActivityService) getUserMap() (map[uint64]types.User, error) {
	users, err := s.userDal.FindUsers()
	if err != nil {
		return nil, err
	}

	userMap := make(map[uint64]types.User)
	for _, user := range users {
		userMap[uint64(user.ID)] = *types.UserFromDal(&user)
	}
	return userMap, nil
}

// GetActivityTypes returns all activity types
func (s *ActivityService) GetActivityTypes(c *fiber.Ctx) error {
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

// GetActivityCategories returns all activity categories
func (s *ActivityService) GetActivityCategories(c *fiber.Ctx) error {
	return c.JSON(types.AllActivityCategories)
}
