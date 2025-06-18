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
	// Get all users
	users, err := s.userDal.FindUsers()
	if err != nil {
		return err
	}

	res, err := s.calculateUserStats(&users)
	if err != nil {
		return err
	}

	return c.Render("scoreboard/show", fiber.Map{"UserStats": res})
}

func (s *ScoreboardService) calculateUserStats(users *[]dal.User) ([]UserStats, error) {
	res := []UserStats{}

	for _, dalUser := range *users {
		user := types.UserFromDal(&dalUser)

		filteredAchievements := make([]types.Achievement, 0)
		for _, a := range user.Achievements {
			if a.Level > 0 {
				filteredAchievements = append(filteredAchievements, a)
			}
		}
		userStats := UserStats{
			User:                user,
			AchievementsSummary: filteredAchievements,
			Points:              s.summarize(user.Activities),
		}
		res = append(res, userStats)
	}

	// Sort UserStats by Points in descending order
	sort.Slice(res, func(i, j int) bool {
		return res[i].Points > res[j].Points
	})
	return res, nil
}

// summarize calculates the points for a user based on the number of activities
func (s *ScoreboardService) summarize(activities []*types.Activity) int {
	return len(activities) * 5
}
