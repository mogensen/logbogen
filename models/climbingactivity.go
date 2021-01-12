package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/gofrs/uuid"
)

// Climbingactivity is used by pop to map your climbingactivities database table to your go code.
type Climbingactivity struct {
	ID           uuid.UUID    `json:"id" db:"id"`
	UserID       uuid.UUID    `db:"user_id"`
	User         User         `belongs_to:"user_id"`
	Date         time.Time    `json:"date" db:"date"`
	Lat          float64      `json:"lat" db:"lat"`
	Lng          float64      `json:"lng" db:"lng"`
	Location     string       `json:"location" db:"location"`
	Type         ClimbingType `json:"type" db:"type"`
	OtherType    string       `json:"other_type" db:"other_type"`
	Role         string       `json:"role" db:"role"`
	Comment      string       `json:"comment" db:"comment"`
	Participants Users        `many_to_many:"participants_climbingactivities" json:"participants"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at" db:"updated_at"`
}

// String is not required by pop and may be deleted
func (c Climbingactivity) String() string {
	jc, _ := json.Marshal(c)
	return string(jc)
}

// Climbingactivities is not required by pop and may be deleted
type Climbingactivities []Climbingactivity

// String is not required by pop and may be deleted
func (c Climbingactivities) String() string {
	jc, _ := json.Marshal(c)
	return string(jc)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (c *Climbingactivity) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.TimeIsPresent{Field: c.Date, Name: "Date"},
		&validators.StringIsPresent{Field: c.Role, Name: "Role"},
	), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (c *Climbingactivity) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (c *Climbingactivity) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
