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

func TestCategoryFilters(t *testing.T) {
	activity := func(category string) *types.Activity { return &types.Activity{ID: uuid.New(), CategoryID: category} }
	activities := []*types.Activity{
		activity("sailing"), activity("climbing"), activity("sailing"), activity("climbing"), activity("sailing"),
	}

	t.Run("lists the categories in use, in the configured order, with counts", func(t *testing.T) {
		filters := categoryFilters(activities, "")

		require.Equal(t, []CategoryFilter{
			{ID: "climbing", Name: "Klatring", Count: 2},
			{ID: "sailing", Name: "Sejlads", Count: 3},
		}, filters)
	})

	t.Run("marks the selected category", func(t *testing.T) {
		filters := categoryFilters(activities, "sailing")

		require.False(t, filters[0].Active)
		require.True(t, filters[1].Active)
	})

	t.Run("keeps a selected category without activities so it can be left again", func(t *testing.T) {
		filters := categoryFilters(activities, "camping")

		require.Len(t, filters, 3)
		require.Equal(t, CategoryFilter{ID: "camping", Name: "Camping", Count: 0, Active: true}, filters[2])
	})

	t.Run("no activities, no filter buttons", func(t *testing.T) {
		require.Empty(t, categoryFilters(nil, ""))
	})
}

func TestFilterByCategory(t *testing.T) {
	climbing := &types.Activity{ID: uuid.New(), CategoryID: "climbing"}
	sailing := &types.Activity{ID: uuid.New(), CategoryID: "sailing"}
	activities := []*types.Activity{climbing, sailing, climbing}

	require.Equal(t, activities, filterByCategory(activities, ""), "no category shows everything")
	require.Equal(t, []*types.Activity{climbing, climbing}, filterByCategory(activities, "climbing"))
	require.Equal(t, []*types.Activity{sailing}, filterByCategory(activities, "sailing"))
	require.Empty(t, filterByCategory(activities, "camping"))
}

func TestValidCategory(t *testing.T) {
	require.Equal(t, "climbing", validCategory("climbing"))
	require.Empty(t, validCategory(""))
	require.Empty(t, validCategory("not-a-category"), "an unknown category is ignored")
	require.Empty(t, validCategory("<script>"))
}
