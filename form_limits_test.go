package main

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/stretchr/testify/require"
)

// activityForm is what the browser posts for the activity form
func activityForm(overrides map[string]string) url.Values {
	form := url.Values{
		"date":     {time.Now().AddDate(0, 0, -3).Format("2006-01-02")},
		"lat":      {"55.676098"},
		"lng":      {"12.568337"},
		"category": {"climbing"},
		"type":     {"boulder"},
		"role":     {"Deltager"},
		"comment":  {"Kommentar"},
	}
	for key, value := range overrides {
		form.Set(key, value)
	}
	return form
}

func newFormClient(t *testing.T) *TestClient {
	client := NewTestClient(t)
	client.CreateUser("Other User", "password")
	email := client.CreateUser("Form User", "password")
	client.Login(email, "password")
	return client
}

func TestFormsCarryTheLimitsOfTheValidation(t *testing.T) {
	limit := func(t *testing.T, field string) int {
		max, err := types.ActivityMax(field)
		require.NoError(t, err)
		return max
	}
	client := newFormClient(t)
	client.CreateActivity(randomActivity())
	activityID := client.GetActivities()[0].ID.String()

	for name, path := range map[string]string{"create": "/activities/create", "edit": "/activities/" + activityID + "/edit"} {
		t.Run("the "+name+" form", func(t *testing.T) {
			body := client.GetResponseBody(client.Get(path))

			client.AssertResponseContains(body, fmt.Sprintf(`name="role" maxlength="%d"`, limit(t, "Role")))
			client.AssertResponseContains(body, fmt.Sprintf(`name="otherType" maxlength="%d"`, limit(t, "OtherType")))
			client.AssertResponseContains(body, fmt.Sprintf(`maxlength="%d" data-counter="comment-count"`, limit(t, "Comment")))
			client.AssertResponseContains(body, fmt.Sprintf(`<span id="comment-count">0</span>/%d`, limit(t, "Comment")))
			client.AssertResponseContains(body, fmt.Sprintf(`data-max="%d"`, limit(t, "ParticipantsIDs")))
			client.AssertResponseContains(body, `<script src="/js/field-limits.js"></script>`)
		})
	}

	t.Run("the certification form", func(t *testing.T) {
		body := client.GetResponseBody(client.Get("/certifications/create"))

		provider, err := types.CertificationMax("Provider")
		require.NoError(t, err)
		other, err := types.CertificationMax("OtherType")
		require.NoError(t, err)
		participants, err := types.CertificationMax("ParticipantsIDs")
		require.NoError(t, err)
		client.AssertResponseContains(body, fmt.Sprintf(`name="provider" maxlength="%d"`, provider))
		client.AssertResponseContains(body, fmt.Sprintf(`name="otherType" maxlength="%d"`, other))
		client.AssertResponseContains(body, fmt.Sprintf(`data-max="%d"`, participants))
	})
}

func TestLineBreaksDoNotCountTwice(t *testing.T) {
	client := newFormClient(t)

	// Under the limit with one character per line break, which is how the browser counts, but over it as submitted
	comment := strings.Repeat(strings.Repeat("a", 18)+"\r\n", 105)
	require.Greater(t, len(comment), 2000)
	require.LessOrEqual(t, len(strings.ReplaceAll(comment, "\r\n", "\n")), 2000)

	resp := client.Post("/activities/create", activityForm(map[string]string{"comment": comment}))

	require.Equal(t, fiber.StatusFound, resp.StatusCode, "a comment that fits in the form is accepted")
	activities := client.GetActivities()
	require.Len(t, activities, 1)
	require.NotContains(t, activities[0].Comment, "\r", "stored with plain line breaks")
	require.Equal(t, 105, strings.Count(activities[0].Comment, "\n"))
}

func TestTooLongInputShowsTheFormAgainInsteadOfAnErrorPage(t *testing.T) {
	client := newFormClient(t)
	client.CreateActivity(randomActivity())
	existing := client.GetActivities()[0]

	assertFormWithMessage := func(t *testing.T, resp string, message string) {
		client.AssertResponseContains(resp, `role="alert">`+message+`</div>`)
		client.AssertResponseContains(resp, `id="activity-form"`)
		client.AssertResponseNotContains(resp, "Der er sket en fejl")
	}

	t.Run("creating with a too long comment", func(t *testing.T) {
		resp := client.Post("/activities/create", activityForm(map[string]string{
			"comment": strings.Repeat("a", 2001), "role": "Min rolle",
		}))

		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		body := client.GetResponseBody(resp)
		assertFormWithMessage(t, body, "Kommentaren må højst være 2000 tegn lang.")
		client.AssertResponseContains(body, "Min rolle")
		client.AssertResponseContains(body, `action="/activities/create"`)
		require.Len(t, client.GetActivities(), 1, "nothing was created")
	})

	t.Run("updating with a too long role keeps the chosen category and type", func(t *testing.T) {
		resp := client.Post("/activities/"+existing.ID.String(), activityForm(map[string]string{
			"role": strings.Repeat("r", 101), "category": "sailing", "type": "kayak", "comment": "Min ændring",
		}))

		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		body := client.GetResponseBody(resp)
		assertFormWithMessage(t, body, "Rollen må højst være 100 tegn lang.")
		client.AssertResponseContains(body, `action="/activities/`+existing.ID.String()+`"`)
		client.AssertResponseContains(body, `data-current="sailing"`)
		client.AssertResponseContains(body, `data-current="kayak"`)
		client.AssertResponseContains(body, "Min ændring")
		require.Equal(t, existing.Role, client.GetActivities()[0].Role, "nothing was changed")
	})

	t.Run("too many participants", func(t *testing.T) {
		form := activityForm(nil)
		for i := range 51 {
			form.Add("participants", fmt.Sprint(i+10))
		}

		resp := client.Post("/activities/create", form)

		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		assertFormWithMessage(t, client.GetResponseBody(resp), "Der kan højst vælges 50 deltagere.")
	})

	t.Run("a position outside the map", func(t *testing.T) {
		resp := client.Post("/activities/create", activityForm(map[string]string{"lng": "190.5"}))

		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		assertFormWithMessage(t, client.GetResponseBody(resp), "Positionen på kortet er ugyldig.")
	})

	t.Run("several problems are all listed, without the submitted text", func(t *testing.T) {
		resp := client.Post("/activities/create", activityForm(map[string]string{
			"comment": strings.Repeat("hemmelig", 400), "role": strings.Repeat("r", 101),
		}))

		body := client.GetResponseBody(resp)
		client.AssertResponseContains(body, "Kommentaren må højst være 2000 tegn lang.")
		client.AssertResponseContains(body, "Rollen må højst være 100 tegn lang.")
		client.AssertResponseNotContains(body, `role="alert">`+strings.Repeat("hemmelig", 5))
	})
}
