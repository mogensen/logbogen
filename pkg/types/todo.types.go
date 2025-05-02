package types

import (
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
	Tree ClimbingType = "TREE"
	// Rock Climbing
	Rock ClimbingType = "ROCK"
	// Boulder Climbing
	Boulder ClimbingType = "BOULDER"
	// Ice Climbing
	Ice ClimbingType = "ICE"
	// HighRope Climbing
	HighRope ClimbingType = "HIGHROPE"
	// Wall Climbing
	Wall ClimbingType = "WALL"
	// Other Climbing
	Other ClimbingType = "OTHER"
)

func MapClimbingType(t string) ClimbingType {
	switch strings.ToUpper(t) {
	case "TREE":
		return Tree
	case "ROCK":
		return Rock
	case "BOULDER":
		return Boulder
	case "ICE":
		return Ice
	case "HIGHROPE":
		return HighRope
	case "WALL":
		return Wall
	case "OTHER":
		return Other
	default:
		return Other
	}
}

// ClimbingTypes is all the avaiable types of climbing
var ClimbingTypes = []ClimbingType{Tree, Rock, Boulder, Ice, HighRope, Wall, Other}

// ClimbingActivity struct contains the ClimbingActivity field which should be returned in a
type ClimbingActivity struct {
	ID           uuid.UUID    `json:"iD"`
	User         *uint        `json:"user"`
	Date         time.Time    `json:"date"`
	Lat          float64      `json:"lat"`
	Lng          float64      `json:"lng"`
	Location     string       `json:"location"`
	Type         ClimbingType `json:"type"`
	Role         string       `json:"role"`
	Comment      string       `json:"comment"`
	Participants []uint64     `json:"participants"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`
}

// CreateDTO struct defines the /ClimbingActivity/create payload
type CreateDTO struct {
	ClimbingActivity *ClimbingActivity `json:"climbingActivity"`
}

// ClimbingActivityCreate struct defines the /ClimbingActivity/create
type ClimbingActivityCreate struct {
	ClimbingActivity *ClimbingActivity `json:"climbingActivity"`
}

// ClimbingActivities defines the ClimbingActivities list
type ClimbingActivities struct {
	ClimbingActivitys *[]ClimbingActivity `json:"climbingActivities"`
}
