package types

import "github.com/mogensen/logbook/pkg/dal"

// UserForLogin is a lightweight user projection used in listings
type UserForLogin struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// User todo
type User struct {
	ID           uint64        `json:"id"`
	Name         string        `json:"name"`
	Email        string        `json:"email"`
	Activities   []*Activity   `json:"activities"`
	Achievements []Achievement `json:"achievements"`
}

func UserFromDal(user *dal.User) *User {
	activities := make([]*Activity, len(user.Activities))
	for i, activity := range user.Activities {
		activities[i] = ActivityFromDal(&activity, map[uint64]User{})
	}

	achievements := Achievements(activities)

	return &User{
		ID:           uint64(user.ID),
		Name:         user.Name,
		Email:        user.Email,
		Activities:   activities,
		Achievements: achievements,
	}
}
