package config

import (
	_ "embed"
	"log"

	"gopkg.in/yaml.v3"
)

//go:embed data/activities.yaml
var activitiesYAML []byte

// Activity category constants
const (
	Climbing string = "climbing"
	Sailing  string = "sailing"
	Other    string = "other"
)

type ActivityCategory struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

type ActivityType struct {
	ID       string `yaml:"id"`
	Name     string `yaml:"name"`
	Category string `yaml:"category"`
}

type activitiesConfig struct {
	Categories []ActivityCategory `yaml:"categories"`
	Types      []ActivityType     `yaml:"types"`
}

var (
	// AllActivityCategories defines all available activity categories
	AllActivityCategories []ActivityCategory
	// AllActivityTypes defines all available activity types
	AllActivityTypes []ActivityType
)

func init() {
	// Parse the YAML
	var config activitiesConfig
	if err := yaml.Unmarshal(activitiesYAML, &config); err != nil {
		log.Fatalf("Error parsing activities config: %v", err)
	}

	// Set the global variables
	AllActivityCategories = config.Categories
	AllActivityTypes = config.Types
}

// CategoryByID returns a pointer to the ActivityCategory with the given ID
func CategoryByID(id string) *ActivityCategory {
	for _, c := range AllActivityCategories {
		if c.ID == id {
			return &c
		}
	}
	return nil
}

// ActivityTypeByID returns a pointer to the ActivityType with the given ID
func ActivityTypeByID(id string) *ActivityType {
	for _, t := range AllActivityTypes {
		if t.ID == id {
			return &t
		}
	}
	return &ActivityType{
		ID:       Other,
		Name:     id,
		Category: Other,
	}
}
