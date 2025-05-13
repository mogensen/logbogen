package services

import (
	"github.com/mogensen/logbook/pkg/types"
)

func Archivements(activities []*types.ClimbingActivity) []types.Achievement {
	activityCounts := make(map[types.ClimbingType]int)

	// Count the number of activities for each climb type
	for _, activity := range activities {
		activityCounts[activity.Type]++
	}

	// Create a slice to store the achievements
	var achievements []types.Achievement

	for _, climbType := range types.ClimbingTypes {
		if counts, ok := activityCounts[climbType]; ok {
			achievements = append(achievements, types.Achievement{
				ClimbType: climbType,
				Level:     int(counts/5) + 1,
			})
		} else {
			achievements = append(achievements, types.Achievement{
				ClimbType: climbType,
				Level:     0,
			})
		}
	}

	return achievements
}
