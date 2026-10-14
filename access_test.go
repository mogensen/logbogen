package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/database"
	"github.com/stretchr/testify/require"
)

// secondClient is another browser on the same app and database
func secondClient(t *testing.T, first *TestClient) *TestClient {
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	return &TestClient{app: first.app, jar: jar, t: t}
}

func userID(t *testing.T, email string) uint64 {
	var user dal.User
	require.NoError(t, database.DB.Where("email = ?", email).Take(&user).Error)
	return uint64(user.ID)
}

func (tc *TestClient) getJSON(path string) string {
	req := httptest.NewRequest("GET", path, nil)
	req.Header.Add("Accept", "application/json")
	u, _ := url.Parse("http://localhost")
	for _, cookie := range tc.jar.Cookies(u) {
		req.AddCookie(cookie)
	}
	resp, err := tc.app.Test(req, testRequestTimeout)
	require.NoError(tc.t, err)
	return tc.GetResponseBody(resp)
}

func activityID(t *testing.T, resp *http.Response) string {
	require.Equal(t, fiber.StatusFound, resp.StatusCode)
	return resp.Header.Get("Location")[len("/activities/"):]
}

func TestUsersCannotReadOrChangeEachOthersData(t *testing.T) {
	alice := NewTestClient(t)
	bob := secondClient(t, alice)
	alice.devLogin("Alice", "alice@example.com")
	bob.devLogin("Bob", "bob@example.com")
	bobID := userID(t, "bob@example.com")

	secret := randomActivity()
	secret.Comment = "bob-secret-comment"
	secret.Lat, secret.Lng = 61.1111, 7.7777
	secret.ParticipantsIDs = nil
	bobActivity := activityID(t, bob.CreateActivity(secret))

	shared := randomActivity()
	shared.Comment = "alice-comment"
	shared.ParticipantsIDs = []uint64{bobID}
	activityID(t, alice.CreateActivity(shared))

	t.Run("another user cannot open, edit, change or delete an activity", func(t *testing.T) {
		for _, path := range []string{"/activities/" + bobActivity, "/activities/" + bobActivity + "/edit"} {
			body := alice.GetResponseBody(alice.Get(path))
			require.NotContains(t, body, "bob-secret-comment", path)
		}

		update := secret
		update.ID = uuid.MustParse(bobActivity)
		update.Comment = "changed-by-alice"
		alice.UpdateActivity(update)
		alice.Post("/activities/"+bobActivity+"/delete", url.Values{})
		alice.Post("/activities/clone/"+bobActivity, url.Values{})

		require.Contains(t, bob.GetResponseBody(bob.Get("/activities/"+bobActivity)), "bob-secret-comment")
		require.NotContains(t, bob.getJSON("/activities/list"), "changed-by-alice")
		require.NotContains(t, alice.getJSON("/activities/list"), "bob-secret-comment")
	})

	t.Run("the activity JSON of a user holds nothing of the participants beyond their name", func(t *testing.T) {
		body := alice.getJSON("/activities/list")
		require.NotContains(t, body, "bob-secret-comment")
		require.NotContains(t, body, "61.1111")
		require.NotContains(t, body, "bob@example.com")
	})

	t.Run("the pending JSON holds nothing of the owner beyond what is shared", func(t *testing.T) {
		body := bob.getJSON("/activities/pending")
		require.Contains(t, body, "alice-comment")
		require.NotContains(t, body, "alice@example.com")
	})

	t.Run("a participant list cannot leak the other activities of the owner", func(t *testing.T) {
		other := randomActivity()
		other.Comment = "alice-private-comment"
		other.ParticipantsIDs = nil
		other.Lat = 49.9999
		activityID(t, alice.CreateActivity(other))

		body := bob.getJSON("/activities/pending")
		require.NotContains(t, body, "alice-private-comment")
		require.NotContains(t, body, "49.9999")
	})

	t.Run("a certification cannot be moved to another user", func(t *testing.T) {
		cert := url.Values{"category": {"other"}, "otherType": {"alice-cert"}, "provider": {"p"},
			"start_date": {"2026-01-01"}, "end_date": {"2028-01-01"}}
		resp := alice.Post("/certifications/create", cert)
		require.Equal(t, fiber.StatusFound, resp.StatusCode)
		certPath := resp.Header.Get("Location")

		cert.Set("user_id", fmt.Sprint(bobID))
		alice.Post(certPath+"/edit", cert)

		require.NotContains(t, bob.GetResponseBody(bob.Get("/certifications/list")), "alice-cert")
		require.Contains(t, alice.GetResponseBody(alice.Get("/certifications/list")), "alice-cert")
	})

	t.Run("another user cannot open, edit or delete a certification", func(t *testing.T) {
		cert := url.Values{"category": {"other"}, "otherType": {"bob-cert"}, "provider": {"p"},
			"start_date": {"2026-01-01"}, "end_date": {"2028-01-01"}}
		resp := bob.Post("/certifications/create", cert)
		certPath := resp.Header.Get("Location")

		require.NotContains(t, alice.GetResponseBody(alice.Get(certPath)), "bob-cert")
		require.NotContains(t, alice.GetResponseBody(alice.Get(certPath+"/edit")), "bob-cert")
		alice.Post(certPath+"/delete", url.Values{})
		require.Contains(t, bob.GetResponseBody(bob.Get("/certifications/list")), "bob-cert")
	})
}

func TestCertificationParticipantsMustBeRealUsers(t *testing.T) {
	alice := NewTestClient(t)
	alice.devLogin("Alice", "alice@example.com")
	bobClient := secondClient(t, alice)
	bobClient.devLogin("Bob", "bob@example.com")
	bobID := userID(t, "bob@example.com")
	form := func(participants ...uint64) url.Values {
		values := url.Values{"category": {"other"}, "otherType": {"team-cert"}, "provider": {"p"},
			"start_date": {"2026-01-01"}, "end_date": {"2028-01-01"}}
		for _, id := range participants {
			values.Add("participants", fmt.Sprint(id))
		}
		return values
	}

	t.Run("an unknown participant is refused and nothing is stored", func(t *testing.T) {
		resp := alice.Post("/certifications/create", form(bobID, 9999))
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		require.Contains(t, alice.GetResponseBody(resp), "unknown participant")
		require.NotContains(t, alice.GetResponseBody(alice.Get("/certifications/list")), "team-cert")
	})

	t.Run("a real participant is accepted and shown by name", func(t *testing.T) {
		resp := alice.Post("/certifications/create", form(bobID))
		require.Equal(t, fiber.StatusFound, resp.StatusCode)
		require.Contains(t, alice.GetResponseBody(alice.Get(resp.Header.Get("Location"))), "Bob")
	})

	t.Run("an edit with an unknown participant is refused and changes nothing", func(t *testing.T) {
		resp := alice.Post("/certifications/create", form(bobID))
		path := resp.Header.Get("Location")

		edit := alice.Post(path+"/edit", form(9999))

		require.Equal(t, fiber.StatusBadRequest, edit.StatusCode)
		require.Contains(t, alice.GetResponseBody(alice.Get(path)), "Bob")
	})
}
