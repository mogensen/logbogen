package types

import (
	"time"

	"github.com/mogensen/logbook/pkg/dal"
)

// UserForLogin is a lightweight user projection used in listings
type UserForLogin struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// User todo
type User struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	// The rest is for pages only. Users appear in other users' JSON as owners and participants.
	Email        string        `json:"-"`
	ThemePref    string        `json:"-"`
	MemberSince  time.Time     `json:"-"`
	Activities   []*Activity   `json:"-"`
	Achievements []Achievement `json:"-"`
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
		ThemePref:    user.ThemePref,
		MemberSince:  user.Model.CreatedAt,
		Activities:   activities,
		Achievements: achievements,
	}
}
