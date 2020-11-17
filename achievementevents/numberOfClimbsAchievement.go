package achievementevents

import (
	"fmt"
	"logbogen/models"
)

type NumberOfClimbsAchievement struct {
	ClimbType models.ClimbingType
	Level     int
	Name      string
	ImageSlug string
}

func NewNumberOfClimbsAchievement(cType models.ClimbingType, level int) *NumberOfClimbsAchievement {
	return &NumberOfClimbsAchievement{
		ClimbType: cType,
		Level:     level,
		ImageSlug: fmt.Sprintf("achievements/%s-%d.png", cType.String(), level),
		Name:      fmt.Sprintf("%s Level %d", cType.String(), level),
	}
}

func (ta *NumberOfClimbsAchievement) Evaluate(climbs *models.Climbingactivities) bool {
	count := 0
	for _, c := range *climbs {
		if c.Type == ta.ClimbType {
			count++
		}
	}
	switch ta.Level {
	case 0:
		return count == 0
	case 1:
		return count >= 1 && count < 5
	case 2:
		return count >= 5 && count < 10
	case 3:
		return count >= 10 && count < 15
	case 4:
		return count >= 15 && count < 20
	case 5:
		return count >= 20
	default:
		return false
	}
}

func (ta *NumberOfClimbsAchievement) GetName() string {
	return fmt.Sprintf("%s level %d", ta.ClimbType.String(), ta.Level)
}

func (ta *NumberOfClimbsAchievement) GetLevel() int {
	return ta.Level
}
