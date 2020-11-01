package grifts

import (
	"math/rand"
)

func randType() string {
	types := []string{"Træklatring", "Vægklatring", "Kursus", "Kæmpegynge"}
	return types[rand.Intn(len(types))]
}

func randRole() string {
	roles := []string{"Instruktør", "Klatring med deltagere", "Rekreativ træklatring"}
	return roles[rand.Intn(len(roles))]
}
