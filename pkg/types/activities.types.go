package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/mogensen/logbook/pkg/config"
	"github.com/mogensen/logbook/pkg/dal"
)

const (
	Climbing string = "climbing"
	Sailing  string = "sailing"
	Foraging string = "foraging"
	Other    string = "other"
)

// Use config package's types
type Category = config.Category

type ActivityType = config.ActivityType
type CertificationType = config.CertificationType

// Use config package's variables
var (
	AllActivityCategories = config.AllActivityCategories
	AllActivityTypes      = config.AllActivityTypes
)

// Activity struct contains all activity fields
type Activity struct {
	ID       uuid.UUID `json:"id" form:"id"`
	UserId   *uint64   `form:"user"`
	User     User      `json:"user"`
	Date     Date      `json:"date" form:"date"`
	Lat      float64   `json:"lat" form:"lat"`
	Lng      float64   `json:"lng" form:"lng"`
	Location string    `json:"location" form:"location"`

	CategoryID string   `form:"category" validate:"required"`
	Category   Category `json:"category"`

	TypeID    string       `form:"type" validate:"required_unless=CategoryID other"`
	Type      ActivityType `json:"type"`
	OtherType string       `json:"otherType" form:"otherType" validate:"required_if=Type other"`

	Role            string    `json:"role" form:"role"`
	Comment         string    `json:"comment" form:"comment"`
	ParticipantsIDs []uint64  `form:"participants"`
	Participants    []User    `json:"participants"`
	CreatedAt       time.Time `json:"createdAt" form:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt" form:"updatedAt"`
}

func (a *Activity) Title() string {
	return a.Type.Name + " nær " + a.Location
}

// ActivityFromDal converts a DAL activity to a types Activity
func ActivityFromDal(activity *dal.Activity, userMap map[uint64]User) *Activity {
	participants := make([]User, 0, len(activity.Participants))
	for _, p := range activity.Participants {
		if user, ok := userMap[p]; ok {
			participants = append(participants, user)
		}
	}

	return &Activity{
		ID:              activity.ID,
		Date:            Date(activity.Date),
		Lat:             activity.Lat,
		Lng:             activity.Lng,
		Location:        activity.Location,
		CategoryID:      activity.Category,
		Category:        *config.CategoryByID(activity.Category),
		Type:            *config.ActivityTypeByID(activity.Type),
		TypeID:          activity.Type,
		Role:            activity.Role,
		Comment:         activity.Comment,
		Participants:    participants,
		ParticipantsIDs: activity.Participants,
		CreatedAt:       activity.CreatedAt,
		UpdatedAt:       activity.UpdatedAt,
		User:            userMap[*activity.User],
	}
}

// CreateDTO struct defines the /Activity/create payload
type CreateDTO struct {
	*Activity
}

// Activities defines the Activities list
type Activities struct {
	Activities *[]Activity `json:"activities"`
}
