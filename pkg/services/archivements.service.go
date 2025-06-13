package services

import (
	"github.com/mogensen/logbook/pkg/types"
)

func Archivements(activities []*types.Activity) []types.Achievement {
	activityCounts := make(map[string]int)

	// Count the number of activities for each climb type
	for _, activity := range activities {
		activityCounts[activity.Type.ID]++
	}

	// Create a slice to store the achievements
	var achievements []types.Achievement

	for _, a := range types.AllActivityTypes {
		if counts, ok := activityCounts[a.ID]; ok {
			achievements = append(achievements, types.Achievement{
				Level: int(counts/5) + 1,
				Type:  a,
			})
		} else {
			achievements = append(achievements, types.Achievement{
				Type:  a,
				Level: 0,
			})
		}
	}

	return achievements
}
