package achievementevents

import (
	"fmt"
	"logbogen/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/events"
	"github.com/gobuffalo/pop/v5/slices"
)

// Init registers all event listeners
func init() {
	// if you want to give your listener a nice name to identify itself
	events.NamedListen("climbingactivity listener", func(e events.Event) {

		if e.Kind != "buffalo:route:finished" {
			return
		}

		c := e.Payload["context"].(*buffalo.DefaultContext)
		ri := e.Payload["route"].(buffalo.RouteInfo)

		if ri.ResourceName != "ClimbingactivitiesResource" {
			return
		}

		user, ok := c.Value("current_user").(*models.User)
		if !ok {
			return
		}

		switch ri.Method {
		case "DELETE":
			fmt.Printf("### Event: ClimbingActivity : deleted -> %s\n", user.Name)
		case "POST":
			fmt.Printf("### Event: ClimbingActivity : created -> %s\n", user.Name)
		case "PUT":
			fmt.Printf("### Event: ClimbingActivity : updated -> %s\n", user.Name)
		default:
			return
		}

		updateAchievementsForUser(user)
	})

	events.NamedListen("achievement updater", func(e events.Event) {
		if e.Kind != "logbogen:achievements:updateall" {
			return
		}

		users := &models.Users{}
		// Retrieve all Users
		if err := models.DB.All(users); err != nil {
			fmt.Printf("### Event: ERROR -> %v\n", err)
			return
		}
		for _, v := range *users {
			updateAchievementsForUser(&v)
		}
	})

}

func updateAchievementsForUser(user *models.User) {
	fmt.Printf("### Event: updating achievements -> %s\n", user.Name)

	climbingactivities := &models.Climbingactivities{}
	q := models.DB.Where("user_id = ?", user.ID)
	// Retrieve all Climbingactivities from the DB
	if err := q.All(climbingactivities); err != nil {
		fmt.Printf("### Event: ERROR getting activities -> %v\n", err)
		return
	}

	a := &models.Achievement{}
	q = models.DB.Where("user_id = ?", user.ID)
	exists, err := q.Exists(a)
	if err != nil {
		fmt.Printf("### Event: ERROR on exists -> %v\n", err)
		return
	}
	if exists {
		if err = q.First(a); err != nil {
			fmt.Printf("### Event: ERROR on first -> %v\n", err)
		}
	}

	earned := []UnlockableAchievement{}
	possibleAchievement := getAllPossibleAchievements()

	for _, possible := range possibleAchievement {
		if possible.Evaluate(climbingactivities) {
			earned = append(earned, possible)
		}
	}

	a.UserID = user.ID
	a.Data = slices.Map{
		"earned": earned,
	}

	if err = models.DB.Save(a); err != nil {
		fmt.Printf("### Event: ERROR saving -> %v\n", err)
	}
}

func getAllPossibleAchievements() []UnlockableAchievement {
	possibleAchievements := []UnlockableAchievement{}
	for _, cType := range models.ClimbingTypes {
		for level := 0; level < 5; level++ {
			a := NewNumberOfClimbsAchievement(cType, level)
			possibleAchievements = append(possibleAchievements, a)
		}
	}
	return possibleAchievements
}
