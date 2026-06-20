package dal

import (
	"gorm.io/gorm"
)

// User struct defines the user
type User struct {
	gorm.Model
	Name          string     `gorm:"not null"`
	ThemePref     string     `gorm:"default:'auto'"`
	// Auth0Sub is the Auth0 subject (sub) claim, the stable external identity.
	// It is a pointer with a unique index so that NULLs (e.g. legacy rows) do
	// not collide on the unique constraint.
	Auth0Sub   *string    `gorm:"uniqueIndex"`
	Email      string     `gorm:"uniqueIndex;not null"`
	Activities []Activity `gorm:"foreignKey:User"`
}

// UserDal defines the interface for user data access operations
type UserDal interface {
	CreateUser(user *User) *gorm.DB
	UpdateUser(user *User) error
	FindUserById(id uint64) (*User, error)
	FindUserByEmail(email string) (*User, error)
	FindUserByAuth0Sub(sub string) (*User, error)
	FindUsers() ([]User, error)
}

// userDalImpl implements the UserDal interface
type userDalImpl struct {
	db *gorm.DB
}

// NewUserDal creates a new instance of UserDal
func NewUserDal(db *gorm.DB) UserDal {
	return &userDalImpl{db: db}
}

// CreateUser create a user entry in the user's table
func (d *userDalImpl) CreateUser(user *User) *gorm.DB {
	return d.db.Create(user)
}

// FindUserById searches the user's table with the id given
func (d *userDalImpl) FindUserById(id uint64) (*User, error) {
	var user User
	err := d.db.Model(&User{}).Preload("Activities").Take(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserByEmail searches the user's table with the email given
func (d *userDalImpl) FindUserByEmail(email string) (*User, error) {
	var user User
	err := d.db.Model(&User{}).Where("email = ?", email).Take(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserByAuth0Sub searches the user's table with the Auth0 subject given
func (d *userDalImpl) FindUserByAuth0Sub(sub string) (*User, error) {
	var user User
	err := d.db.Model(&User{}).Preload("Activities").Take(&user, "auth0_sub = ?", sub).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUsers returns all users from the database
func (d *userDalImpl) FindUsers() ([]User, error) {
	var users []User
	err := d.db.Model(&User{}).Preload("Activities").Find(&users).Error
	return users, err
}

// UpdateUser updates a user in the database
func (d *userDalImpl) UpdateUser(user *User) error {
	return d.db.Save(user).Error
}
