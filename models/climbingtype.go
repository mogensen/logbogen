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
	// Wall Climbing
	Wall ClimbingType = "WALL"
	// Other Climbing
	Other ClimbingType = "OTHER"
)

func (ct ClimbingType) String() string {
	switch ct {
	case Tree:
		return "tree"
	case Rock:
		return "rock"
	case Boulder:
		return "boulder"
	case Ice:
		return "ice"
	case HighRope:
		return "highrope"
	case Wall:
		return "wall"
	case Other:
		return "other"
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
var ClimbingTypes = []ClimbingType{Tree, Rock, Boulder, Ice, HighRope, Wall, Other}
