package types

// Achievements calculates the achievements based on the activities provided.
// It counts the number of activities for each type and assigns a level to each achievement.
// The level is determined by dividing the activity count by 5 plus 1. If no activities
// are recorded for a type, the level is set to 0. The function returns a slice of achievements
// for all activity
func Achievements(activities []*Activity) []Achievement {
	activityCounts := make(map[string]int)

	// Count the number of activities for each type
	for _, activity := range activities {
		activityCounts[activity.Type.ID]++
	}

	// Create a slice to store the achievements
	var achievements []Achievement

	// Create achievements for all activity types
	for _, activityType := range AllActivityTypes {
		count := activityCounts[activityType.ID]
		level := 0
		if count > 0 {
			level = (count-1)/5 + 1
		}
		achievements = append(achievements, Achievement{
			Type:  activityType,
			Level: level,
		})
	}

	return achievements
}
