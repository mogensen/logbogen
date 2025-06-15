package dal

import (
	"github.com/mogensen/logbook/pkg/database"
	"gorm.io/gorm"
)

// User struct defines the user
type User struct {
	gorm.Model
	Name       string     `gorm:"not null"`
	Email      string     `gorm:"uniqueIndex;not null"`
	Password   string     `gorm:"not null"`
	Activities []Activity `gorm:"foreignKey:User"`
}

// CreateUser create a user entry in the user's table
func CreateUser(user *User) *gorm.DB {
	return database.DB.Create(user)
}

// FindUserByOd searches the user's table with the id given
func FindUserById(dest interface{}, id uint64) *gorm.DB {
	return findUser(dest, "id = ?", id)
}

// FindUserByEmail searches the user's table with the email given
func FindUserByEmail(dest interface{}, email string) *gorm.DB {
	return findUser(dest, "email = ?", email)
}

// findUser searches the user's table with the condition given
func findUser(dest interface{}, conds ...interface{}) *gorm.DB {
	return database.DB.Model(&User{}).Take(dest, conds...)
}

// FindUser searches the user's table with the condition given
func FindUsers(dest interface{}, conds ...interface{}) *gorm.DB {
	return database.DB.Model(&User{}).Preload("Activities").Find(dest, conds...)
}
