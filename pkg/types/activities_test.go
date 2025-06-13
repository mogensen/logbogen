package types

import "testing"

func TestActivityTypeMappings(t *testing.T) {
	// Test that all ActivityType constants have a name in ActivityTypeNames
	for _, activityType := range allActivityTypes {
		if name, exists := ActivityTypeNames[activityType]; !exists {
			t.Errorf("ActivityType %s is missing from ActivityTypeNames map", activityType)
		} else if name == "" {
			t.Errorf("ActivityType %s has an empty name in ActivityTypeNames map", activityType)
		}
	}

	// Test that all ActivityType constants are mapped in Categories
	for _, activityType := range allActivityTypes {
		found := false
		for _, categoryMap := range Categories {
			if _, exists := categoryMap[activityType]; exists {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ActivityType %s is not mapped in any category in Categories map", activityType)
		}
	}

	// Test that all mapped types in Categories have corresponding ActivityTypeNames
	for category, typeMap := range Categories {
		for activityType := range typeMap {
			if name, exists := ActivityTypeNames[activityType]; !exists {
				t.Errorf("ActivityType %s in category %s is missing from ActivityTypeNames map", activityType, category)
			} else if name == "" {
				t.Errorf("ActivityType %s in category %s has an empty name in ActivityTypeNames map", activityType, category)
			}
		}
	}
}
