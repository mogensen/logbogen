package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// ParticipantsClimbingactivities is the actual record detailing that a participant has been part of an activity
type ParticipantsClimbingactivities struct {
	ID            uuid.UUID `json:"id" db:"id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	ActivityID    uuid.UUID `json:"activity_id" db:"climbingactivity_id"`
	ParticipantID uuid.UUID `json:"participant_id" db:"user_id"`
}

// String is not required by pop and may be deleted
func (a ParticipantsClimbingactivities) String() string {
	ja, _ := json.Marshal(a)
	return string(ja)
}

// ActivityParticipants is not required by pop and may be deleted
type ActivityParticipants []ParticipantsClimbingactivities

// String is not required by pop and may be deleted
func (a ActivityParticipants) String() string {
	ja, _ := json.Marshal(a)
	return string(ja)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (a *ParticipantsClimbingactivities) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (a *ParticipantsClimbingactivities) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (a *ParticipantsClimbingactivities) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
