package routes

import (
	"sort"

	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/services"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/mogensen/logbook/pkg/utils/middleware"
)

// UserStats represents a user with their statistics
type UserStats struct {
	User                *types.User
	AchievementsSummary []types.Achievement
	Points              int
}

func ScoreboardRoutes(app *fiber.App) {
	scoreboard := app.Group("/scoreboard").Use(middleware.Auth)

	scoreboard.Get("/", middleware.User, func(c *fiber.Ctx) error {
		users := &[]dal.User{}

		// Get all users
		err := dal.FindUsers(users).Error
		if err != nil {
			return err
		}

		res := []UserStats{}

		for _, user := range *users {
			activities := make([]*types.Activity, len(user.Activities))
			for i, activity := range user.Activities {
				activities[i] = services.MapActivityFromDal(&activity, nil)
			}

			achievements := services.Achievements(activities)
			filteredAchievements := make([]types.Achievement, 0)
			for _, a := range achievements {
				if a.Level > 0 {
					filteredAchievements = append(filteredAchievements, a)
				}
			}
			userStats := UserStats{
				User:                types.UserFromDal(&user, nil),
				AchievementsSummary: filteredAchievements,
				Points:              summerize(achievements),
			}
			res = append(res, userStats)
		}

		// Sort UserStats by Points in descending order
		sort.Slice(res, func(i, j int) bool {
			return res[i].Points > res[j].Points
		})

		return c.Render("scoreboard/show", fiber.Map{
			"Title":     "Scoreboard",
			"UserStats": res,
		})
	})
}

func summerize(achievements []types.Achievement) int {
	points := 0
	for _, achievement := range achievements {
		points += achievement.Level * 5
	}
	return points
}
