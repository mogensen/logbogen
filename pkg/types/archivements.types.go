package types

import "fmt"

// Achievement describes an achievement a user can gain
type Achievement struct {
	Type  ActivityType
	Level int
}

func (a Achievement) ImageSlug() string {
	return fmt.Sprintf("achievements/%s-%d.png", a.Type.ID, a.Level)
}

func (a Achievement) Name() string {
	return fmt.Sprintf("%s Level %d", a.Type.Name, a.Level)
}
