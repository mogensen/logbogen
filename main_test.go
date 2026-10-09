package main

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

func TestInputBounds(t *testing.T) {
	client := NewTestClient(t)
	// The default participant ids 1 and 2 must refer to existing users
	client.CreateUser("Other User", "password")
	email := client.CreateUser("Bounds User", "password")
	client.Login(email, "password")

	t.Run("oversized comment is rejected", func(t *testing.T) {
		activity := randomActivity()
		activity.Comment = strings.Repeat("a", 2001)
		client.CreateActivity(activity)
		require.Len(t, client.GetActivities(), 0)
	})

	t.Run("out of range coordinates are rejected", func(t *testing.T) {
		activity := randomActivity()
		activity.Lat = 91
		client.CreateActivity(activity)
		activity = randomActivity()
		activity.Lng = -181
		client.CreateActivity(activity)
		require.Len(t, client.GetActivities(), 0)
	})

	t.Run("too many participants are rejected", func(t *testing.T) {
		activity := randomActivity()
		for i := uint64(0); i < 51; i++ {
			activity.ParticipantsIDs = append(activity.ParticipantsIDs, i+10)
		}
		client.CreateActivity(activity)
		require.Len(t, client.GetActivities(), 0)
	})

	t.Run("body above the limit is refused", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/activities/create", strings.NewReader(strings.Repeat("a", maxBodyBytes+1)))
		req.Header.Add("Content-Type", "application/json")
		_, err := client.app.Test(req, testRequestTimeout)
		require.ErrorContains(t, err, "body size exceeds the given limit")
	})

	t.Run("a valid activity is accepted", func(t *testing.T) {
		client.CreateActivity(randomActivity())
		require.Len(t, client.GetActivities(), 1)
	})
}

func TestCSRF(t *testing.T) {
	client := NewTestClient(t)
	client.CreateUser("Other User", "password")
	email := client.CreateUser("CSRF User", "password")
	client.Login(email, "password")
	client.CreateActivity(randomActivity())
	activities := client.GetActivities()
	require.Len(t, activities, 1)

	// rawPost sends a form POST with the session cookies but no automatic CSRF token
	rawPost := func(path string, header http.Header, form url.Values) *http.Response {
		req := httptest.NewRequest("POST", path, strings.NewReader(form.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		for key, values := range header {
			req.Header[key] = values
		}
		u, _ := url.Parse("http://localhost")
		for _, cookie := range client.jar.Cookies(u) {
			req.AddCookie(cookie)
		}
		resp, err := client.app.Test(req, testRequestTimeout)
		require.NoError(t, err)
		return resp
	}

	t.Run("POST without a token is forbidden", func(t *testing.T) {
		resp := rawPost("/activities/"+activities[0].ID.String()+"/delete", nil, url.Values{})
		require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
		require.Len(t, client.GetActivities(), 1)
	})

	t.Run("POST with a wrong token is forbidden", func(t *testing.T) {
		resp := rawPost("/users/theme", http.Header{"X-Csrf-Token": {"wrong"}}, url.Values{})
		require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})

	t.Run("form field token is accepted", func(t *testing.T) {
		form := url.Values{"_csrf": {client.csrfToken()}}
		resp := rawPost("/activities/"+activities[0].ID.String()+"/delete", nil, form)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		require.Len(t, client.GetActivities(), 0)
	})

	t.Run("logout and clone no longer accept GET", func(t *testing.T) {
		require.Equal(t, fiber.StatusMethodNotAllowed, client.Get("/auth/logout").StatusCode)
		require.Equal(t, fiber.StatusMethodNotAllowed, client.Get("/activities/clone/"+uuid.NewString()).StatusCode)
		// Still logged in
		require.Equal(t, fiber.StatusOK, client.Get("/activities/list").StatusCode)
	})

	t.Run("forms and layout carry the token", func(t *testing.T) {
		token := client.csrfToken()
		for _, path := range []string{"/", "/activities/create", "/certifications/create"} {
			body := client.GetResponseBody(client.Get(path))
			client.AssertResponseContains(body, `<meta name="csrf-token" content="`+token+`">`)
		}
		for _, path := range []string{"/activities/create", "/certifications/create"} {
			body := client.GetResponseBody(client.Get(path))
			client.AssertResponseContains(body, `name="_csrf" value="`+token+`"`)
		}
	})
}
