package services

import (
	"github.com/mogensen/logbook/pkg/types"
)

func Archivements(activities []*types.Activity) []types.Achievement {
	activityCounts := make(map[types.ActivityType]int)

	// Count the number of activities for each climb type
	for _, activity := range activities {
		activityCounts[activity.Type]++
	}

	// Create a slice to store the achievements
	var achievements []types.Achievement

	for id := range types.ActivityTypeNames {
		if counts, ok := activityCounts[id]; ok {
			achievements = append(achievements, types.Achievement{
				Level: int(counts/5) + 1,
				Type:  id,
			})
		} else {
			achievements = append(achievements, types.Achievement{
				Type:  id,
				Level: 0,
			})
		}
	}

	return achievements
}
