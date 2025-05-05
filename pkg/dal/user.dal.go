package dal

import (
	"github.com/mogensen/logbook/pkg/database"
	"gorm.io/gorm"
)

// User struct defines the user
type User struct {
	gorm.Model
	Name               string             `gorm:"not null"`
	Email              string             `gorm:"uniqueIndex;not null"`
	Password           string             `gorm:"not null"`
	ClimbingActivities []ClimbingActivity `gorm:"foreignKey:User"`
}

// CreateUser create a user entry in the user's table
func CreateUser(user *User) *gorm.DB {
	return database.DB.Create(user)
}

// FindUser searches the user's table with the condition given
func FindUser(dest interface{}, conds ...interface{}) *gorm.DB {
	return database.DB.Model(&User{}).Take(dest, conds...)
}

// FindUserByEmail searches the user's table with the email given
func FindUserByEmail(dest interface{}, email string) *gorm.DB {
	return FindUser(dest, "email = ?", email)
}

// FindUser searches the user's table with the condition given
func FindUsers(dest interface{}, conds ...interface{}) *gorm.DB {
	return database.DB.Model(&User{}).Find(dest, conds...)
}
