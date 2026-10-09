package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/mocks"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestScoreboardService_Summarize(t *testing.T) {
	// Arrange
	mockUserDal := new(mocks.UserDalMock)
	service := NewScoreboardService(mockUserDal)
	activities := []*types.Activity{
		{ID: uuid.New()},
		{ID: uuid.New()},
		{ID: uuid.New()},
	}

	// Act
	points := service.summarize(activities)

	// Assert
	assert.Equal(t, 15, points) // 3 activities * 5 points
}

func TestScoreboardService_calculateUserStats(t *testing.T) {
	mockUserDal := new(mocks.UserDalMock)
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

func TestScoreboardService_calculateUserStats_Ranks(t *testing.T) {
	service := NewScoreboardService(new(mocks.UserDalMock))
	userWith := func(id uint, name string, activities int) dal.User {
		user := dal.User{Model: gorm.Model{ID: id}, Name: name, Email: name + "@test.com"}
		for i := 0; i < activities; i++ {
			user.Activities = append(user.Activities,
				dal.Activity{ID: uuid.New(), User: uint64Ptr(uint64(id)), Type: "rock", Category: types.Climbing})
		}
		return user
	}

	users := []dal.User{
		userWith(1, "Dina", 1),
		userWith(2, "Anna", 3),
		userWith(3, "Carl", 1),
		userWith(4, "Bo", 3),
		userWith(5, "Eve", 0),
	}

	stats, err := service.calculateUserStats(&users)
	require.NoError(t, err)

	var names []string
	var ranks []int
	for _, stat := range stats {
		names = append(names, stat.User.Name)
		ranks = append(ranks, stat.Rank)
	}
	require.Equal(t, []string{"Anna", "Bo", "Carl", "Dina", "Eve"}, names, "equal points are ordered by name")
	require.Equal(t, []int{1, 1, 3, 3, 5}, ranks, "the top place is 1 and equal points share it")
}
