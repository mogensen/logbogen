package achievementevents

import "logbogen/models"

type UnlockableArchievement interface {
	Evaluate(climbs *models.Climbingactivities) bool
	GetSlug() string
	GetDescription() string
	GetLevel() int
}
