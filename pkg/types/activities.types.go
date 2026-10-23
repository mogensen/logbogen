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

// Weather is the stored weather of the day of an activity
type Weather struct {
	Temperature float64 `json:"temperature"`
	Code        int     `json:"code"`
}

// Description names the weather in Danish from its WMO code
func (w Weather) Description() string {
	switch {
	case w.Code == 0:
		return "Klart"
	case w.Code >= 1 && w.Code <= 3:
		return "Halvskyet"
	case w.Code >= 45 && w.Code <= 48:
		return "Tåge"
	case w.Code >= 51 && w.Code <= 57:
		return "Støvregn"
	case w.Code >= 61 && w.Code <= 67, w.Code >= 80 && w.Code <= 82:
		return "Regn"
	case w.Code >= 71 && w.Code <= 77, w.Code >= 85 && w.Code <= 86:
		return "Sne"
	case w.Code >= 95:
		return "Torden"
	}
	return ""
}

// Activity struct contains all activity fields
type Activity struct {
	ID       uuid.UUID `json:"id" form:"id"`
	UserId   *uint64   `form:"user"`
	User     User      `json:"user"`
	Date     Date      `json:"date" form:"date"`
	Lat      float64   `json:"lat" form:"lat" validate:"gte=-90,lte=90"`
	Lng      float64   `json:"lng" form:"lng" validate:"gte=-180,lte=180"`
	Location string    `json:"location" form:"location" validate:"max=200"`

	CategoryID string   `form:"category" validate:"required,max=50"`
	Category   Category `json:"category"`

	TypeID    string       `form:"type" validate:"required_unless=CategoryID other,max=50"`
	Type      ActivityType `json:"type"`
	OtherType string       `json:"otherType" form:"otherType" validate:"required_if=Type other,max=100"`

	Role            string    `json:"role" form:"role" validate:"max=100"`
	Comment         string    `json:"comment" form:"comment" validate:"max=2000"`
	ParticipantsIDs []uint64  `form:"participants" validate:"max=50"`
	Participants    []User    `json:"participants"`
	Weather         *Weather  `json:"weather,omitempty" form:"-"`
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

	var weather *Weather
	if activity.WeatherTemperature != nil && activity.WeatherCode != nil {
		weather = &Weather{Temperature: *activity.WeatherTemperature, Code: *activity.WeatherCode}
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
		Weather:         weather,
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
