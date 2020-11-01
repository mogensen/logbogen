package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// ActivityParticipant is the actual record detailing that a participant has been part of an activity
type ActivityParticipant struct {
	ID            uuid.UUID `json:"id" db:"id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	ActivityID    uuid.UUID `json:"activity_id" db:"activity_id"`
	ParticipantID uuid.UUID `json:"participant_id" db:"participant_id"`
}

// String is not required by pop and may be deleted
func (a ActivityParticipant) String() string {
	ja, _ := json.Marshal(a)
	return string(ja)
}

// ActivityParticipants is not required by pop and may be deleted
type ActivityParticipants []ActivityParticipant

// String is not required by pop and may be deleted
func (a ActivityParticipants) String() string {
	ja, _ := json.Marshal(a)
	return string(ja)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (a *ActivityParticipant) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (a *ActivityParticipant) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (a *ActivityParticipant) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
