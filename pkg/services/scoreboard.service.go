package services

import (
	"sort"

	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/types"
)

// UserStats represents a user with their statistics
type UserStats struct {
	User                *types.User
	AchievementsSummary []types.Achievement
	Points              int
}

// ScoreboardService handles scoreboard-related operations
type ScoreboardService struct {
	userDal dal.UserDal
}

// NewScoreboardService creates a new instance of ScoreboardService
func NewScoreboardService(userDal dal.UserDal) *ScoreboardService {
	return &ScoreboardService{
		userDal: userDal,
	}
}

// GetScoreboard returns the scoreboard data for all users
func (s *ScoreboardService) GetScoreboard(c *fiber.Ctx) error {

	users := &[]dal.User{}

	// Get all users
	err := s.userDal.FindUsers(users).Error
	if err != nil {
		return err
	}

	res, err := s.calculateUserStats(users)
	if err != nil {
		return err
	}

	return c.Render("scoreboard/show", fiber.Map{
		"Title":     "Scoreboard",
		"UserStats": res,
	})
}

func (s *ScoreboardService) calculateUserStats(users *[]dal.User) ([]UserStats, error) {
	res := []UserStats{}

	for _, user := range *users {
		activities := make([]*types.Activity, len(user.Activities))
		for i, activity := range user.Activities {
			activities[i] = types.ActivityFromDal(&activity, map[uint64]types.User{})
		}

		achievements := Achievements(activities)
		filteredAchievements := make([]types.Achievement, 0)
		for _, a := range achievements {
			if a.Level > 0 {
				filteredAchievements = append(filteredAchievements, a)
			}
		}
		userStats := UserStats{
			User:                types.UserFromDal(&user, nil),
			AchievementsSummary: filteredAchievements,
			Points:              s.summerize(activities),
		}
		res = append(res, userStats)
	}

	// Sort UserStats by Points in descending order
	sort.Slice(res, func(i, j int) bool {
		return res[i].Points > res[j].Points
	})
	return res, nil
}

// summerize calculates the points for a user based on the number of activities
func (s *ScoreboardService) summerize(activities []*types.Activity) int {
	return len(activities) * 5
}
