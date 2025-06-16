package mocks

import (
	"github.com/mogensen/logbook/pkg/dal"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockUserDal is a mock implementation of the UserDal interface
type MockUserDal struct {
	mock.Mock
}

func (m *MockUserDal) CreateUser(user *dal.User) *gorm.DB {
	args := m.Called(user)
	return args.Get(0).(*gorm.DB)
}

func (m *MockUserDal) FindUserById(id uint64) (*dal.User, error) {
	args := m.Called(id)
	var user *dal.User
	if args.Get(0) != nil {
		user = args.Get(0).(*dal.User)
	}
	return user, args.Error(1)
}

func (m *MockUserDal) FindUserByEmail(dest interface{}, email string) *gorm.DB {
	args := m.Called(dest, email)
	return args.Get(0).(*gorm.DB)
}

func (m *MockUserDal) FindUsers() ([]dal.User, error) {
	args := m.Called()
	return args.Get(0).([]dal.User), args.Error(1)
}
