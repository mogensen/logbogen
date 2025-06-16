package dal

import (
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

// UserDal defines the interface for user data access operations
type UserDal interface {
	CreateUser(user *User) *gorm.DB
	FindUserById(id uint64) (*User, error)
	FindUserByEmail(dest interface{}, email string) *gorm.DB
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
func (d *userDalImpl) FindUserByEmail(dest interface{}, email string) *gorm.DB {
	return d.findUser(dest, "email = ?", email)
}

// findUser searches the user's table with the condition given
func (d *userDalImpl) findUser(dest interface{}, conds ...interface{}) *gorm.DB {
	return d.db.Model(&User{}).Preload("Activities").Take(dest, conds...)
}

// FindUsers returns all users from the database
func (d *userDalImpl) FindUsers() ([]User, error) {
	var users []User
	err := d.db.Model(&User{}).Preload("Activities").Find(&users).Error
	return users, err
}
