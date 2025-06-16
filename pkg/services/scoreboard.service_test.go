package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// MockUserDal is a mock implementation of dal.UserDal
type MockUserDal struct {
	mock.Mock
}

func (m *MockUserDal) CreateUser(user *dal.User) *gorm.DB {
	args := m.Called(user)
	return args.Get(0).(*gorm.DB)
}

func (m *MockUserDal) FindUserById(dest interface{}, id uint64) *gorm.DB {
	args := m.Called(dest, id)
	return args.Get(0).(*gorm.DB)
}

func (m *MockUserDal) FindUserByEmail(dest interface{}, email string) *gorm.DB {
	args := m.Called(dest, email)
	return args.Get(0).(*gorm.DB)
}

func (m *MockUserDal) FindUsers(dest interface{}, conds ...interface{}) *gorm.DB {
	args := m.Called(dest, conds)
	return args.Get(0).(*gorm.DB)
}

func TestScoreboardService_Summerize(t *testing.T) {
	// Arrange
	mockUserDal := new(MockUserDal)
	service := NewScoreboardService(mockUserDal)
	activities := []*types.Activity{
		{ID: uuid.New()},
		{ID: uuid.New()},
		{ID: uuid.New()},
	}

	// Act
	points := service.summerize(activities)

	// Assert
	assert.Equal(t, 15, points) // 3 activities * 5 points
}

func TestScoreboardService_calculateUserStats(t *testing.T) {
	mockUserDal := new(MockUserDal)
	service := NewScoreboardService(mockUserDal)

	users := []dal.User{
		{
			Model: gorm.Model{ID: 1},
			Name:  "User1",
			Email: "user1@test.com",
			Activities: []dal.Activity{
				{Model: gorm.Model{ID: 1}, ID: uuid.New(), User: uint64Ptr(1), Type: "rock", Category: types.Climbing},
				{Model: gorm.Model{ID: 2}, ID: uuid.New(), User: uint64Ptr(1), Type: "rock", Category: types.Climbing},
			},
		},
		{
			Model: gorm.Model{ID: 2},
			Name:  "User2",
			Email: "user2@test.com",
			Activities: []dal.Activity{
				{Model: gorm.Model{ID: 3}, ID: uuid.New(), User: uint64Ptr(2), Type: "rock", Category: types.Climbing},
				{Model: gorm.Model{ID: 4}, ID: uuid.New(), User: uint64Ptr(2), Type: "kayak", Category: types.Sailing},
				{Model: gorm.Model{ID: 5}, ID: uuid.New(), User: uint64Ptr(2), Type: "kayak", Category: types.Sailing},
			},
		},
	}

	stast, err := service.calculateUserStats(&users)
	require.NoError(t, err)

	assert.Equal(t, 15, stast[0].Points)
	assert.Equal(t, 10, stast[1].Points)
	assert.Equal(t, "User2", stast[0].User.Name)
	assert.Equal(t, "User1", stast[1].User.Name)
	assert.Equal(t, 2, len(stast[0].AchievementsSummary))
	assert.Equal(t, 1, len(stast[1].AchievementsSummary))

}

// Helper function to create a pointer to uint64
func uint64Ptr(n uint64) *uint64 {
	return &n
}
