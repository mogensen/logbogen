package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/buffalo/binding"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

type UsersImage struct {
	ID        uuid.UUID    `json:"id" db:"id"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
	UserID    uuid.UUID    `json:"user_id" db:"user_id"`
	Image     binding.File `db:"-" form:"image"`
	ImageData []byte       `db:"image_data"`
}

// String is not required by pop and may be deleted
func (p UsersImage) String() string {
	jp, _ := json.Marshal(p)
	return string(jp)
}

// UsersImages is not required by pop and may be deleted
type UsersImages []UsersImage

// String is not required by pop and may be deleted
func (p UsersImages) String() string {
	jp, _ := json.Marshal(p)
	return string(jp)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (p *UsersImage) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (p *UsersImage) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (p *UsersImage) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

func (p *UsersImage) AfterCreate(tx *pop.Connection) error {

	return nil
}
