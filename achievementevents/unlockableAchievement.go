package achievementevents

import "logbogen/models"

type UnlockableAchievement interface {
	Evaluate(climbs *models.Climbingactivities) bool
	GetName() string
	GetLevel() int
}
