package services

import (
	"testing"
	"time"

	"github.com/mogensen/logbook/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestGetUserActivityHeatmap_FutureActivities(t *testing.T) {
	now := time.Now()
	future := now.AddDate(3, 0, 0)
	user := &types.User{Activities: []*types.Activity{{Date: types.Date(future)}}}

	rows := (&UserService{}).GetUserActivityHeatmap(user)

	// From the first activity year up to the latest one, even when that is in the future
	require.NotEmpty(t, rows)
	require.Equal(t, future.Year(), rows[0].Year)
	require.Equal(t, future.Year(), rows[len(rows)-1].Year)
}

func TestGetUserActivityHeatmap_SpansFromFirstActivityToToday(t *testing.T) {
	now := time.Now()
	past := now.AddDate(-2, 0, 0)
	user := &types.User{Activities: []*types.Activity{{Date: types.Date(past)}}}

	rows := (&UserService{}).GetUserActivityHeatmap(user)

	require.Len(t, rows, 3)
	require.Equal(t, past.Year(), rows[0].Year)
	require.Equal(t, now.Year(), rows[2].Year)
}
