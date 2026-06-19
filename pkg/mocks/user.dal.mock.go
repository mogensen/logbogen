package mocks

import (
	"github.com/mogensen/logbook/pkg/dal"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// UserDalMock is a mock implementation of dal.UserDal
type UserDalMock struct {
	mock.Mock
}

// CreateUser mocks the CreateUser method
func (m *UserDalMock) CreateUser(user *dal.User) *gorm.DB {
	args := m.Called(user)
	return args.Get(0).(*gorm.DB)
}

// FindUserById mocks the FindUserById method
func (m *UserDalMock) FindUserById(id uint64) (*dal.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dal.User), args.Error(1)
}

// FindUserByEmail mocks the FindUserByEmail method
func (m *UserDalMock) FindUserByEmail(email string) (*dal.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dal.User), args.Error(1)
}

// FindUserByAuth0Sub mocks the FindUserByAuth0Sub method
func (m *UserDalMock) FindUserByAuth0Sub(sub string) (*dal.User, error) {
	args := m.Called(sub)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dal.User), args.Error(1)
}

// FindUsers mocks the FindUsers method
func (m *UserDalMock) FindUsers() ([]dal.User, error) {
	args := m.Called()
	return args.Get(0).([]dal.User), args.Error(1)
}
