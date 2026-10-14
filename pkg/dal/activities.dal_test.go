package dal

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestActivityService_CreateActivity(t *testing.T) {
	db := setupTestDB(t)
	service := NewActivityService(db)

	userID := uint64(1)
	activity := &Activity{
		ID:           uuid.New(),
		User:         &userID,
		Date:         time.Now(),
		Lat:          55.676098,
		Lng:          12.568337,
		Location:     "Copenhagen",
		Type:         "Training",
		Category:     "Fitness",
		Role:         "Participant",
		Comment:      "Test activity",
		Participants: datatypes.JSONSlice[uint64]{userID},
	}

	_, err := service.CreateActivity(activity)
	assert.NoError(t, err)

	found, err := service.FindActivity("id = ?", activity.ID)
	assert.NoError(t, err)
	assert.Equal(t, activity.ID, found.ID)
	assert.Equal(t, activity.Location, found.Location)
}

func TestActivityService_FindActivityByUser(t *testing.T) {
	db := setupTestDB(t)
	service := NewActivityService(db)

	userID := uint64(1)
	activity := &Activity{
		ID:           uuid.New(),
		User:         &userID,
		Date:         time.Now(),
		Lat:          55.676098,
		Lng:          12.568337,
		Location:     "Copenhagen",
		Type:         "Training",
		Category:     "Fitness",
		Role:         "Participant",
		Comment:      "Test activity",
		Participants: datatypes.JSONSlice[uint64]{userID},
	}

	_, err := service.CreateActivity(activity)
	assert.NoError(t, err)

	found, err := service.FindActivityByUser(activity.ID.String(), userID)
	assert.NoError(t, err)
	assert.Equal(t, activity.ID, found.ID)
}

func TestActivityService_FindActivitiesByUser(t *testing.T) {
	db := setupTestDB(t)
	service := NewActivityService(db)

	userID := uint64(1)
	activity1 := &Activity{
		ID:           uuid.New(),
		User:         &userID,
		Date:         time.Now(),
		Lat:          55.676098,
		Lng:          12.568337,
		Location:     "Copenhagen",
		Type:         "Training",
		Category:     "Fitness",
		Role:         "Participant",
		Comment:      "Test activity 1",
		Participants: datatypes.JSONSlice[uint64]{userID},
	}

	activity2 := &Activity{
		ID:           uuid.New(),
		User:         &userID,
		Date:         time.Now(),
		Lat:          55.676098,
		Lng:          12.568337,
		Location:     "Copenhagen",
		Type:         "Training",
		Category:     "Fitness",
		Role:         "Participant",
		Comment:      "Test activity 2",
		Participants: datatypes.JSONSlice[uint64]{userID},
	}

	_, err := service.CreateActivity(activity1)
	assert.NoError(t, err)
	_, err = service.CreateActivity(activity2)
	assert.NoError(t, err)

	activities, err := service.FindActivitiesByUser(userID)
	assert.NoError(t, err)
	assert.Len(t, activities, 2)
}

func TestActivityService_UpdateActivity(t *testing.T) {
	db := setupTestDB(t)
	service := NewActivityService(db)

	userID := uint64(1)
	activity := &Activity{
		ID:           uuid.New(),
		User:         &userID,
		Date:         time.Now(),
		Lat:          55.676098,
		Lng:          12.568337,
		Location:     "Copenhagen",
		Type:         "Training",
		Category:     "Fitness",
		Role:         "Participant",
		Comment:      "Test activity",
		Participants: datatypes.JSONSlice[uint64]{userID},
	}

	_, err := service.CreateActivity(activity)
	assert.NoError(t, err)

	updateData := map[string]interface{}{
		"comment": "Updated comment",
	}
	updated, err := service.UpdateActivity(activity.ID.String(), userID, updateData)
	assert.NoError(t, err)
	assert.Equal(t, "Updated comment", updated.Comment)
}

func TestActivityService_DeleteActivity(t *testing.T) {
	db := setupTestDB(t)
	service := NewActivityService(db)

	userID := uint64(1)
	activity := &Activity{
		ID:           uuid.New(),
		User:         &userID,
		Date:         time.Now(),
		Lat:          55.676098,
		Lng:          12.568337,
		Location:     "Copenhagen",
		Type:         "Training",
		Category:     "Fitness",
		Role:         "Participant",
		Comment:      "Test activity",
		Participants: datatypes.JSONSlice[uint64]{userID},
	}

	_, err := service.CreateActivity(activity)
	assert.NoError(t, err)

	err = service.DeleteActivity(activity.ID.String(), userID)
	assert.NoError(t, err)

	_, err = service.FindActivity("id = ?", activity.ID)
	assert.Error(t, err)
}

func TestActivityService_FindPendingActivitiesByUser(t *testing.T) {
	db := setupTestDB(t)
	service := NewActivityService(db)

	userID1 := uint64(1)
	userID2 := uint64(2)

	activity := &Activity{
		ID:           uuid.New(),
		User:         &userID1,
		Date:         time.Now(),
		Lat:          55.676098,
		Lng:          12.568337,
		Location:     "Copenhagen",
		Type:         "Training",
		Category:     "Fitness",
		Role:         "Participant",
		Comment:      "Test activity",
		Participants: datatypes.JSONSlice[uint64]{userID1, userID2},
	}

	_, err := service.CreateActivity(activity)
	assert.NoError(t, err)

	pendingActivities, err := service.FindPendingActivitiesByUser(userID2)
	assert.NoError(t, err)
	assert.Len(t, pendingActivities, 1)
	assert.Equal(t, activity.ID, pendingActivities[0].ID)
}

func TestActivityService_FindActivityToClone(t *testing.T) {
	db := setupTestDB(t)
	service := NewActivityService(db)

	userID1 := uint64(1)
	userID2 := uint64(2)

	// Create an activity owned by user1 with user2 as participant
	activity := &Activity{
		ID:           uuid.New(),
		User:         &userID1,
		Date:         time.Now(),
		Lat:          55.676098,
		Lng:          12.568337,
		Location:     "Copenhagen",
		Type:         "Training",
		Category:     "Fitness",
		Role:         "Participant",
		Comment:      "Test activity",
		Participants: datatypes.JSONSlice[uint64]{userID1, userID2},
	}

	_, err := service.CreateActivity(activity)
	assert.NoError(t, err)

	// Test finding activity to clone as user2
	found, err := service.FindActivityToClone(activity.ID.String(), userID2)
	assert.NoError(t, err)
	assert.Equal(t, activity.ID, found.ID)
	assert.Equal(t, userID1, *found.User)
	assert.Contains(t, found.Participants, userID2)

	// Test that user1 cannot find their own activity to clone
	_, err = service.FindActivityToClone(activity.ID.String(), userID1)
	assert.Error(t, err)
}

func TestActivityService_UpdateWeather(t *testing.T) {
	db := setupTestDB(t)
	service := NewActivityService(db)
	userID := uint64(1)
	otherUserID := uint64(2)
	activity := &Activity{ID: uuid.New(), User: &userID, Date: time.Now(), Location: "Home", Type: "t", Category: "c", Role: "r", Comment: "c"}
	_, err := service.CreateActivity(activity)
	require.NoError(t, err)

	temperature, code := 0.0, 0

	t.Run("stores the weather, including a real zero", func(t *testing.T) {
		require.NoError(t, service.UpdateWeather(activity.ID.String(), userID, &temperature, &code))

		stored, err := service.FindActivityByUser(activity.ID.String(), userID)
		require.NoError(t, err)
		require.NotNil(t, stored.WeatherTemperature)
		require.NotNil(t, stored.WeatherCode)
		require.Zero(t, *stored.WeatherTemperature)
		require.Zero(t, *stored.WeatherCode)
		require.NotNil(t, stored.WeatherFetchedAt)
	})

	t.Run("another user cannot change it", func(t *testing.T) {
		other, otherCode := 99.0, 99
		require.NoError(t, service.UpdateWeather(activity.ID.String(), otherUserID, &other, &otherCode))

		stored, err := service.FindActivityByUser(activity.ID.String(), userID)
		require.NoError(t, err)
		require.Zero(t, *stored.WeatherTemperature)
	})

	t.Run("a general update leaves the weather alone", func(t *testing.T) {
		_, err := service.UpdateActivity(activity.ID.String(), userID, &Activity{Comment: "Edited"})
		require.NoError(t, err)

		stored, err := service.FindActivityByUser(activity.ID.String(), userID)
		require.NoError(t, err)
		require.NotNil(t, stored.WeatherTemperature)
	})

	t.Run("nil clears everything, and a half value clears too", func(t *testing.T) {
		require.NoError(t, service.UpdateWeather(activity.ID.String(), userID, &temperature, nil))

		stored, err := service.FindActivityByUser(activity.ID.String(), userID)
		require.NoError(t, err)
		require.Nil(t, stored.WeatherTemperature)
		require.Nil(t, stored.WeatherCode)
		require.Nil(t, stored.WeatherFetchedAt)
	})
}
