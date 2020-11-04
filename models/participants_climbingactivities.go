package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// ParticipantsClimbingactivity is the actual record detailing that a participant has been part of an activity
type ParticipantsClimbingactivity struct {
	ID            uuid.UUID `json:"id" db:"id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	ActivityID    uuid.UUID `json:"activity_id" db:"climbingactivity_id"`
	ParticipantID uuid.UUID `json:"participant_id" db:"user_id"`
}

// String is not required by pop and may be deleted
func (a ParticipantsClimbingactivity) String() string {
	ja, _ := json.Marshal(a)
	return string(ja)
}

// ParticipantsClimbingactivities is not required by pop and may be deleted
type ParticipantsClimbingactivities []ParticipantsClimbingactivity

// String is not required by pop and may be deleted
func (a ParticipantsClimbingactivities) String() string {
	ja, _ := json.Marshal(a)
	return string(ja)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (a *ParticipantsClimbingactivity) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (a *ParticipantsClimbingactivity) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (a *ParticipantsClimbingactivity) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
