package types

import (
	"github.com/google/uuid"
	"github.com/mogensen/logbook/pkg/config"
	"github.com/mogensen/logbook/pkg/dal"
)

// Certification represents a user certification
// Provider: e.g. "Falk" or "Dansk Træklatrenævn"
type Certification struct {
	ID              uuid.UUID         `json:"id" form:"id"`
	UserID          *uint64           `json:"user_id" form:"user_id"`
	CategoryID      string            `form:"category" validate:"required"`
	Category        Category          `json:"category"`
	TypeID          string            `form:"type" validate:"required_unless=CategoryID other"`
	Type            CertificationType `json:"type"`
	OtherType       string            `json:"otherType" form:"otherType" validate:"required_if=Type other"`
	Provider        string            `json:"provider" form:"provider"`
	StartDate       Date              `json:"start_date" form:"start_date"`
	EndDate         Date              `json:"end_date" form:"end_date"`
	ParticipantsIDs []uint64          `form:"participants"`
	Participants    []User            `json:"participants"`
}

// Mapper functions
func CertificationFromDB(db dal.Certification, userMap map[uint64]User) Certification {
	participants := make([]User, 0, len(db.Participants))
	for _, p := range db.Participants {
		if user, ok := userMap[p]; ok {
			participants = append(participants, user)
		}
	}

	return Certification{
		ID:              db.ID,
		UserID:          db.UserID,
		Provider:        db.Provider,
		CategoryID:      db.Category,
		Category:        *config.CertificationCategoriesByID(db.Category),
		TypeID:          db.Type,
		Type:            *config.CertificationTypeByID(db.Type),
		StartDate:       Date(db.StartDate),
		EndDate:         Date(db.EndDate),
		Participants:    participants,
		ParticipantsIDs: db.Participants,
	}
}

func (c Certification) ToDB() dal.Certification {
	typeID := c.TypeID
	if c.CategoryID == "other" && c.OtherType != "" {
		typeID = c.OtherType
	}

	return dal.Certification{
		ID:           c.ID,
		UserID:       c.UserID,
		Provider:     c.Provider,
		StartDate:    c.StartDate.Time(),
		EndDate:      c.EndDate.Time(),
		Participants: c.ParticipantsIDs,
		Category:     c.CategoryID,
		Type:         typeID,
	}
}
