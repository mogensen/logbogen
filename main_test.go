package main

import (
	"math/rand"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/config"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestAppSetup(t *testing.T) {
	client := NewTestClient(t)

	t.Run("Home page should be accessible", func(t *testing.T) {
		resp := client.Get("/")
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		body := client.GetResponseBody(resp)
		client.AssertResponseContains(body, "Velkommen til Logbogen")
		client.AssertResponseContains(body, "Log ind")
	})

	t.Run("User can be created via dev login", func(t *testing.T) {
		email := client.CreateUser("John Doe", "password")
		require.NotEmpty(t, email)
	})

	t.Run("User can login", func(t *testing.T) {
		email := client.CreateUser("Jane Doe", "password")
		client.Login(email, "password")
	})

	t.Run("User can list activities", func(t *testing.T) {
		email := client.CreateUser("Activity User", "password")
		client.Login(email, "password")

		activities := client.GetActivities()
		require.Len(t, activities, 0)
	})

	t.Run("User can create activity", func(t *testing.T) {
		email := client.CreateUser("Activity User", "password")
		client.Login(email, "password")
		activities := client.GetActivities()
		require.Len(t, activities, 0)

		activity := randomActivity()
		client.CreateActivity(activity)

		// Verify activity is created
		activities = client.GetActivities()
		require.Len(t, activities, 1)
		require.Equal(t, "Test comment", activities[0].Comment)
		require.Equal(t, "Test role", activities[0].Role)
		require.Equal(t, []uint64{1, 2}, activities[0].ParticipantsIDs)
		// Type is mapped to the correct name
		require.Equal(t, *config.ActivityTypeByID(activity.TypeID), activities[0].Type)
		// Geocoded location
		require.Equal(t, "København, Danmark", activities[0].Location)
	})

	t.Run("User can edit activity", func(t *testing.T) {
		email := client.CreateUser("Activity User", "password")
		client.Login(email, "password")
		activities := client.GetActivities()
		require.Len(t, activities, 0)

		activity := randomActivity()
		client.CreateActivity(activity)

		// Verify activity is created
		activities = client.GetActivities()
		require.Len(t, activities, 1)
		require.Equal(t, activity.Comment, activities[0].Comment)
		require.Equal(t, activity.Role, activities[0].Role)
		require.Equal(t, activity.ParticipantsIDs, activities[0].ParticipantsIDs)
		require.Equal(t, *config.ActivityTypeByID(activity.TypeID), activities[0].Type)
		require.Equal(t, "København, Danmark", activities[0].Location)

		// Edit activity
		a := activities[0]

		a.Comment = "Edited comment"
		a.Role = "Edited role"
		a.ParticipantsIDs = []uint64{3}
		client.UpdateActivity(a)

		// Verify activity is updated
		activities = client.GetActivities()
		require.Len(t, activities, 1)
		require.Equal(t, "Edited comment", activities[0].Comment)
		require.Equal(t, "Edited role", activities[0].Role)
		require.Equal(t, []uint64{3}, activities[0].ParticipantsIDs)
	})

	t.Run("User can logout", func(t *testing.T) {
		email := client.CreateUser("Logout User", "password")
		client.Login(email, "password")
		client.Logout()

		// Verify logout worked
		resp := client.Get("/")
		body := client.GetResponseBody(resp)
		client.AssertResponseContains(body, "Velkommen til Logbogen")
		client.AssertResponseContains(body, "Log ind")
	})
}

func randomActivity() types.Activity {
	return types.Activity{
		Date:            randomDate(),
		Lat:             55.676098,
		Lng:             12.568337,
		CategoryID:      types.AllActivityCategories[0].ID,
		TypeID:          types.AllActivityTypes[0].ID,
		Comment:         "Test comment",
		Role:            "Test role",
		ParticipantsIDs: []uint64{1, 2},
	}
}

func randomDate() types.Date {
	return types.Date(time.Now().AddDate(0, 0, rand.Intn(3650))) // Random date up to 10 years ago
}
