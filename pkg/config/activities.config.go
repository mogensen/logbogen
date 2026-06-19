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

type Category struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

type ActivityType struct {
	ID       string `yaml:"id"`
	Name     string `yaml:"name"`
	Category string `yaml:"category"`
}

type activitiesConfig struct {
	Categories []Category     `yaml:"categories"`
	Types      []ActivityType `yaml:"types"`
}

var (
	// AllActivityCategories defines all available activity categories
	AllActivityCategories []Category
	// AllActivityTypes defines all available activity types
	AllActivityTypes []ActivityType
)

func init() {
	// Parse the YAML
	var activityConfig activitiesConfig
	if err := yaml.Unmarshal(activitiesYAML, &activityConfig); err != nil {
		log.Fatalf("Error parsing activities config: %v", err)
	}
	var certConfig CertificationsConfig
	if err := yaml.Unmarshal(certificationsYAML, &certConfig); err != nil {
		log.Fatalf("Error parsing activities config: %v", err)
	}

	// Set the global variables
	AllActivityCategories = activityConfig.Categories
	AllActivityTypes = activityConfig.Types
	AllCertificationCategories = certConfig.Categories
	AllCertificationTypes = certConfig.Certifications
}

// CategoryByID returns a pointer to the ActivityCategory with the given ID
func CategoryByID(id string) *Category {
	for _, c := range AllActivityCategories {
		if c.ID == id {
			return &c
		}
	}
	return nil
}

func CertificationCategoriesByID(id string) *Category {
	for _, c := range AllCertificationCategories {
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

// CertificationTypeByID returns a pointer to the CertificationType with the given ID
func CertificationTypeByID(id string) *CertificationType {
	for _, t := range AllCertificationTypes {
		if t.ID == id {
			return &t
		}
	}
	return &CertificationType{
		ID:       Other,
		Name:     id,
		Category: Other,
	}
}
