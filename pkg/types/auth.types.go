package types

import "github.com/mogensen/logbook/pkg/dal"

// LoginDTO defined the /login payload
type LoginDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"password"`
}

// SignupDTO defined the /login payload
type SignupDTO struct {
	LoginDTO
	Name string `json:"name" validate:"required,min=3"`
}

// UserForLogin is used for login and signup
type UserForLogin struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

// User todo
type User struct {
	ID           uint64        `json:"id"`
	Name         string        `json:"name"`
	Email        string        `json:"email"`
	Achievements []Achievement `json:"achievements"`
}

func UserFromDal(user *dal.User, achievements []Achievement) *User {
	return &User{
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		Achievements: achievements,
	}
}
