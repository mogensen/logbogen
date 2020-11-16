package achievementevents

import (
	"fmt"
	"logbogen/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/events"
)

// Init registers all event listeners
func Init() {
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

		climbingactivities := &models.Climbingactivities{}
		q := models.DB.Where("user_id = ?", user.ID)
		// Retrieve all Climbingactivities from the DB
		if err := q.All(climbingactivities); err != nil {
			return
		}
		earned := []UnlockableArchievement{}

		possibleArchievements := getAllPossibleArchievements()

		for _, possible := range possibleArchievements {
			if possible.Evaluate(climbingactivities) {
				earned = append(earned, possible)
			}
		}
		for _, archivement := range earned {
			fmt.Printf("##### %s has earned '%s' %s\n", user.Name, archivement.GetDescription(), archivement.GetSlug())
		}
	})
}

func getAllPossibleArchievements() []UnlockableArchievement {
	possibleArchievements := []UnlockableArchievement{}
	for _, v := range models.ClimbingTypes {
		for i := 0; i < 10; i++ {
			possibleArchievements = append(possibleArchievements, &NumberOfClimbsArchievement{ClimbType: v, Level: i})
		}
	}
	return possibleArchievements
}
