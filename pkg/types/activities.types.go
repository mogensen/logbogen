package types

import (
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ClimbingType is used to define a type of climbing
type ClimbingType string

func (c ClimbingType) String() string {
	return string(c)
}

const (
	// Tree Climbing
	Tree ClimbingType = "tree"
	// Rock Climbing
	Rock ClimbingType = "rock"
	// Boulder Climbing
	Boulder ClimbingType = "boulder"
	// Ice Climbing
	Ice ClimbingType = "ice"
	// HighRope Climbing
	HighRope ClimbingType = "highrope"
	// Wall Climbing
	Wall ClimbingType = "wall"
	// Other Climbing
	Other ClimbingType = "other"
)

func MapClimbingType(t string) ClimbingType {
	lower := ClimbingType(strings.ToLower(t))
	if slices.Contains(ClimbingTypes, lower) {
		return lower
	}
	// If the type is not found, return Other
	return Other
}

// ClimbingTypes is all the avaiable types of climbing
var ClimbingTypes = []ClimbingType{Tree, Rock, Boulder, Ice, HighRope, Wall, Other}
var Names = map[ClimbingType]string{
	Tree:     "Træklatring",
	Rock:     "Klippklatring",
	Boulder:  "Bouldering",
	Ice:      "Isklatring",
	HighRope: "Høje Seil",
	Wall:     "Vægklatring",
	Other:    "Andet",
}

// ClimbingActivity struct contains the ClimbingActivity field which should be returned in a
type ClimbingActivity struct {
	ID              uuid.UUID    `json:"id" form:"id"`
	UserId          *uint64      `form:"user"`
	User            User         `json:"user"`
	Date            Date         `json:"date" form:"date"`
	Lat             float64      `json:"lat" form:"lat"`
	Lng             float64      `json:"lng" form:"lng"`
	Location        string       `json:"location" form:"location"`
	Type            ClimbingType `json:"type" form:"type"`
	OtherType       string       `json:"otherType" form:"otherType"`
	Role            string       `json:"role" form:"role"`
	Comment         string       `json:"comment" form:"comment"`
	ParticipantsIDs []uint64     `form:"participants"`
	Participants    []User       `json:"participants"`
	CreatedAt       time.Time    `json:"createdAt" form:"createdAt"`
	UpdatedAt       time.Time    `json:"updatedAt" form:"updatedAt"`
}

func (c *ClimbingActivity) TypeStr() string {
	if c.Type == Other {
		return c.OtherType
	}
	return Names[c.Type]

}

func (c *ClimbingActivity) Title() string {
	return c.TypeStr() + " nær " + c.Location
}

// CreateDTO struct defines the /ClimbingActivity/create payload
type CreateDTO struct {
	*ClimbingActivity
}

// ClimbingActivityCreate struct defines the /ClimbingActivity/create
type ClimbingActivityCreate struct {
	ClimbingActivity *ClimbingActivity `json:"climbingActivity"`
}

// ClimbingActivities defines the ClimbingActivities list
type ClimbingActivities struct {
	ClimbingActivitys *[]ClimbingActivity `json:"climbingActivities"`
}
