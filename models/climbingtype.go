package models

import (
	"strings"
)

// ClimbingType is used to define a type of climbing
type ClimbingType string

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
	// Other Climbing
	Other ClimbingType = "OTHER"
)

func (ct ClimbingType) String() string {
	switch ct {
	case Tree:
		return "Tree"
	case Rock:
		return "Rock"
	case Boulder:
		return "Boulder"
	case Ice:
		return "Ice"
	case HighRope:
		return "HighRope"
	case Other:
		return "Other"
	default:
		return ""
	}
}

func (c ClimbingType) SelectLabel() string {
	return c.String()
}
func (c ClimbingType) SelectValue() interface{} {
	return strings.ToUpper(c.String())
}

// ClimbingTypes is all the avaiable types of climbing
var ClimbingTypes = []ClimbingType{Tree, Rock, Boulder, Ice, HighRope, Other}
