package types

import (
	"os"
	"path/filepath"
	"testing"
)

func TestActivityTypeMappings(t *testing.T) {
	// Test that all activity types have valid names and categories
	for _, activityType := range AllActivityTypes {
		if activityType.Name == "" {
			t.Errorf("ActivityType %s has an empty name", activityType.ID)
		}
		if activityType.Category == "" {
			t.Errorf("ActivityType %s has an empty category", activityType.ID)
		}
	}

	// Test that all activity types have valid categories
	for _, activityType := range AllActivityTypes {
		found := false
		for _, category := range AllActivityCategories {
			if activityType.Category == category.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ActivityType %s has invalid category %s", activityType.ID, activityType.Category)
		}
	}

	// Test that all activity types have unique IDs
	seenIDs := make(map[string]bool)
	for _, activityType := range AllActivityTypes {
		if seenIDs[activityType.ID] {
			t.Errorf("Duplicate activity type ID found: %s", activityType.ID)
		}
		seenIDs[activityType.ID] = true
	}
}

func TestActivityTypeImages(t *testing.T) {
	// Get the absolute path to the activities images directory
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	// Navigate to the project root (assuming we're in pkg/types)
	projectRoot := filepath.Join(cwd, "..", "..")
	imageDir := filepath.Join(projectRoot, "assets", "images", "activities")

	// Check if the directory exists
	if _, err := os.Stat(imageDir); os.IsNotExist(err) {
		t.Fatalf("Activities images directory not found at: %s", imageDir)
	}

	// Check that each activity type has a corresponding SVG image
	for _, activityType := range AllActivityTypes {
		imagePath := filepath.Join(imageDir, activityType.ID+".svg")
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			t.Errorf("Missing image for activity type %s: %s", activityType.ID, imagePath)
		}
	}

	// Check that there are no extra images that don't correspond to activity types
	entries, err := os.ReadDir(imageDir)
	if err != nil {
		t.Fatalf("Failed to read activities directory: %v", err)
	}

	// Create a map of valid activity type IDs
	validIDs := make(map[string]bool)
	for _, activityType := range AllActivityTypes {
		validIDs[activityType.ID] = true
	}

	// Add special cases
	validIDs["other"] = true // The "other" category has its own image

	// Check each file in the directory
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Skip non-SVG files
		if filepath.Ext(entry.Name()) != ".svg" {
			continue
		}

		// Get the activity type ID from the filename (without extension)
		id := entry.Name()[:len(entry.Name())-4] // remove .svg extension

		// Skip known legacy files
		if id == "absail" {
			continue
		}

		// Check if this ID corresponds to a valid activity type
		if !validIDs[id] {
			t.Errorf("Found image for unknown activity type: %s", id)
		}
	}
}
