//go:build !integration
// +build !integration

package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementations and helpers would be needed for a full test, but here's a basic structure:
func TestGetPendingActivitiesForUser_Filtering(t *testing.T) {
	// Setup
	userID := uint64(1)
	otherUserID := uint64(2)
	date := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	pending := dal.Activity{
		ID:       uuid.New(),
		User:     &otherUserID,
		Date:     date,
		Category: "climbing",
		Type:     "boulder",
	}
	logged := dal.Activity{
		ID:       uuid.New(),
		User:     &userID,
		Date:     date,
		Category: "climbing",
		Type:     "boulder",
	}

	// Call filtering logic directly (extract to helper for testability if needed)
	pendingActivities := []dal.Activity{pending}
	loggedActivities := []dal.Activity{logged}
	sut := ActivityService{}

	filtered, err := sut.filterPending(loggedActivities, pendingActivities, map[uint64]types.User{})
	require.NoError(t, err)

	assert.Len(t, filtered, 0, "Should filter out pending activity if already logged")
}

func TestFilterPending_Case1_NoLoggedActivities_AllPendingReturned(t *testing.T) {
	otherUserID := uint64(2)
	date := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	pending := dal.Activity{
		ID:       uuid.New(),
		User:     &otherUserID,
		Date:     date,
		Category: "climbing",
		Type:     "boulder",
	}
	pending2 := dal.Activity{
		ID:       uuid.New(),
		User:     &otherUserID,
		Date:     date.AddDate(0, 0, 1),
		Category: "climbing",
		Type:     "lead",
	}
	pendingActivities := []dal.Activity{pending, pending2}
	loggedActivities := []dal.Activity{}

	sut := ActivityService{}
	filtered, err := sut.filterPending(loggedActivities, pendingActivities, map[uint64]types.User{})
	require.NoError(t, err)
	assert.Len(t, filtered, 2)
}

func TestFilterPending_Case2_SomePendingMatchLogged_FilteredOut(t *testing.T) {
	userID := uint64(1)
	otherUserID := uint64(2)
	date := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	pending1 := dal.Activity{
		ID:       uuid.New(),
		User:     &otherUserID,
		Date:     date,
		Category: "climbing",
		Type:     "boulder",
	}
	pending2 := dal.Activity{
		ID:       uuid.New(),
		User:     &otherUserID,
		Date:     date.AddDate(0, 0, 1),
		Category: "climbing",
		Type:     "lead",
	}
	logged := dal.Activity{
		ID:       uuid.New(),
		User:     &userID,
		Date:     date,
		Category: "climbing",
		Type:     "boulder",
	}
	pendingActivities := []dal.Activity{pending1, pending2}
	loggedActivities := []dal.Activity{logged}

	sut := ActivityService{}
	filtered, err := sut.filterPending(loggedActivities, pendingActivities, map[uint64]types.User{})
	require.NoError(t, err)
	assert.Len(t, filtered, 1)
	assert.Equal(t, pending2.ID, filtered[0].ID)
}

func TestFilterPending_Case3_NoPendingMatchLogged_AllReturned(t *testing.T) {
	userID := uint64(1)
	otherUserID := uint64(2)
	date := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	pending1 := dal.Activity{
		ID:       uuid.New(),
		User:     &otherUserID,
		Date:     date,
		Category: "climbing",
		Type:     "boulder",
	}
	pending2 := dal.Activity{
		ID:       uuid.New(),
		User:     &otherUserID,
		Date:     date.AddDate(0, 0, 1),
		Category: "climbing",
		Type:     "lead",
	}
	logged := dal.Activity{
		ID:       uuid.New(),
		User:     &userID,
		Date:     date,
		Category: "climbing",
		Type:     "lead",
	}
	pendingActivities := []dal.Activity{pending1, pending2}
	loggedActivities := []dal.Activity{logged}

	sut := ActivityService{}
	filtered, err := sut.filterPending(loggedActivities, pendingActivities, map[uint64]types.User{})
	require.NoError(t, err)
	assert.Len(t, filtered, 2)
}

func TestFilterPending_Case4_EdgeCases(t *testing.T) {
	userID := uint64(1)
	otherUserID := uint64(2)
	baseDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	// Multiple activities on same date but different type/category
	pending1 := dal.Activity{
		ID:       uuid.New(),
		User:     &otherUserID,
		Date:     baseDate,
		Category: "climbing",
		Type:     "boulder",
	}
	pending2 := dal.Activity{
		ID:       uuid.New(),
		User:     &otherUserID,
		Date:     baseDate,
		Category: "climbing",
		Type:     "lead",
	}
	pending3 := dal.Activity{
		ID:       uuid.New(),
		User:     &otherUserID,
		Date:     baseDate,
		Category: "sailing",
		Type:     "dinghy",
	}
	logged := dal.Activity{
		ID:       uuid.New(),
		User:     &userID,
		Date:     baseDate,
		Category: "climbing",
		Type:     "boulder",
	}
	pendingActivities := []dal.Activity{pending1, pending2, pending3}
	loggedActivities := []dal.Activity{logged}

	sut := ActivityService{}
	filtered, err := sut.filterPending(loggedActivities, pendingActivities, map[uint64]types.User{})
	require.NoError(t, err)
	assert.Len(t, filtered, 2)
	ids := []uuid.UUID{filtered[0].ID, filtered[1].ID}
	assert.Contains(t, ids, pending2.ID)
	assert.Contains(t, ids, pending3.ID)

	// Activities with similar but not identical dates
	pending4 := dal.Activity{
		ID:       uuid.New(),
		User:     &otherUserID,
		Date:     baseDate.Add(24 * time.Hour), // next day
		Category: "climbing",
		Type:     "boulder",
	}
	logged2 := dal.Activity{
		ID:       uuid.New(),
		User:     &userID,
		Date:     baseDate.Add(48 * time.Hour), // two days later
		Category: "climbing",
		Type:     "boulder",
	}
	pendingActivities2 := []dal.Activity{pending4}
	loggedActivities2 := []dal.Activity{logged2}
	filtered2, err := sut.filterPending(loggedActivities2, pendingActivities2, map[uint64]types.User{})
	require.NoError(t, err)
	assert.Len(t, filtered2, 1)
	assert.Equal(t, pending4.ID, filtered2[0].ID)

	// Empty pending activities
	filtered3, err := sut.filterPending(loggedActivities, []dal.Activity{}, map[uint64]types.User{})
	require.NoError(t, err)
	assert.Len(t, filtered3, 0)

	// Empty logged activities
	filtered4, err := sut.filterPending([]dal.Activity{}, pendingActivities, map[uint64]types.User{})
	require.NoError(t, err)
	assert.Len(t, filtered4, 3)
}
