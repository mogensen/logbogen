package achievementevents

import (
	"fmt"
	"logbogen/models"
	"strings"
)

type NumberOfClimbsArchievement struct {
	ClimbType models.ClimbingType
	Level     int
	Name      string
}

func (ta *NumberOfClimbsArchievement) Evaluate(climbs *models.Climbingactivities) bool {
	count := 0
	for _, c := range *climbs {
		if c.Type == ta.ClimbType {
			count++
		}
	}
	return count >= ta.Level
}

func (ta *NumberOfClimbsArchievement) GetDescription() string {
	return fmt.Sprintf("%s level %d", ta.ClimbType.String(), ta.Level)
}

func (ta *NumberOfClimbsArchievement) GetLevel() int {
	return ta.Level
}

func (ta *NumberOfClimbsArchievement) GetSlug() string {
	return fmt.Sprintf("numberofclimbs-%s-%d", strings.ToLower(ta.ClimbType.String()), ta.Level)
}
