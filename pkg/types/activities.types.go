package types

import (
	"time"

	"github.com/google/uuid"
)

// ActivityCategory defines the category of an activity
type ActivityCategory string

const (
	Climbing ActivityCategory = "climbing"
	Sailing  ActivityCategory = "sailing"
)

// ActivityType is used to define a type of activity
type ActivityType string

func (a ActivityType) String() string {
	return string(a)
}

func (a ActivityType) Name() string {
	return ActivityTypeNames[a]
}

const (
	// Climbing Types
	Tree     ActivityType = "tree"
	Rock     ActivityType = "rock"
	Boulder  ActivityType = "boulder"
	Ice      ActivityType = "ice"
	HighRope ActivityType = "highrope"
	Wall     ActivityType = "wall"
	// Sailing Types
	Kayak       ActivityType = "kayak"
	Canoe       ActivityType = "canoe"
	Sail        ActivityType = "sail"
	PaddleBoard ActivityType = "paddle-board"
	// Other
	Other ActivityType = "other"
)

var Categories = map[ActivityCategory]map[ActivityType]string{
	Climbing: {
		Tree:     ActivityTypeNames[Tree],
		Rock:     ActivityTypeNames[Rock],
		Boulder:  ActivityTypeNames[Boulder],
		Ice:      ActivityTypeNames[Ice],
		HighRope: ActivityTypeNames[HighRope],
		Wall:     ActivityTypeNames[Wall],
		Other:    ActivityTypeNames[Other],
	},
	Sailing: {
		Kayak:       ActivityTypeNames[Kayak],
		Canoe:       ActivityTypeNames[Canoe],
		Sail:        ActivityTypeNames[Sail],
		PaddleBoard: ActivityTypeNames[PaddleBoard],
		Other:       ActivityTypeNames[Other],
	},
}

// ActivityTypeNames is a map of activity types to their names
var ActivityTypeNames = map[ActivityType]string{
	Tree:        "Træklatring",
	Rock:        "Klippeklatring",
	Boulder:     "Bouldering",
	Ice:         "Isklatring",
	HighRope:    "High Rope",
	Wall:        "Vægklatring",
	Kayak:       "Kajak",
	Canoe:       "Kano",
	Sail:        "Sejlbåd",
	PaddleBoard: "Paddleboard",
	Other:       "Anden",
}

// Activity struct contains all activity fields
type Activity struct {
	ID              uuid.UUID        `json:"id" form:"id"`
	UserId          *uint64          `form:"user"`
	User            User             `json:"user"`
	Date            Date             `json:"date" form:"date"`
	Lat             float64          `json:"lat" form:"lat"`
	Lng             float64          `json:"lng" form:"lng"`
	Location        string           `json:"location" form:"location"`
	Category        ActivityCategory `json:"category" form:"category"`
	Type            ActivityType     `json:"type" form:"type" validate:"required"`
	OtherType       string           `json:"otherType" form:"otherType" validate:"required_if=Type other"`
	Role            string           `json:"role" form:"role"`
	Comment         string           `json:"comment" form:"comment"`
	ParticipantsIDs []uint64         `form:"participants"`
	Participants    []User           `json:"participants"`
	CreatedAt       time.Time        `json:"createdAt" form:"createdAt"`
	UpdatedAt       time.Time        `json:"updatedAt" form:"updatedAt"`
}

func (a *Activity) TypeStr() string {
	if a.Type == Other {
		return a.OtherType
	}
	return ActivityTypeNames[a.Type]
}

func (a *Activity) Title() string {
	return a.TypeStr() + " nær " + a.Location
}

// CreateDTO struct defines the /Activity/create payload
type CreateDTO struct {
	*Activity
}

// Activities defines the Activities list
type Activities struct {
	Activities *[]Activity `json:"activities"`
}
