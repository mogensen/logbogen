package types

import "fmt"

// Achievement describes an archivement a user can gain
type Achievement struct {
	ClimbType ClimbingType
	Level     int
}

func (a Achievement) ImageSlug() string {
	return fmt.Sprintf("achievements/%s-%d.png", a.ClimbType, a.Level)
}

func (a Achievement) Name() string {
	return fmt.Sprintf("%s Level %d", a.ClimbType, a.Level)
}
