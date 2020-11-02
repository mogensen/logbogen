package grifts

import (
	"logbogen/models"
	"math/rand"
)

func randType() models.ClimbingType {
	return models.ClimbingTypes[rand.Intn(len(models.ClimbingTypes))]
}
func randOtherType() string {
	types := []string{"Vertikalklatring", "Rappelling"}
	return types[rand.Intn(len(types))]
}

func randRole() string {
	roles := []string{"Instruktør", "Klatring med deltagere", "Rekreativ klatring med andre på samme niveau"}
	return roles[rand.Intn(len(roles))]
}
