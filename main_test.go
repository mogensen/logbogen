package main

import (
	"errors"
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

func TestSecurityHeaders(t *testing.T) {
	client := NewTestClient(t)
	email := client.CreateUser("Header User", "password")
	client.Login(email, "password")

	for _, path := range []string{"/", "/activities/list", "/does-not-exist"} {
		resp := client.Get(path)
		require.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"), path)
		require.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"), path)
		require.Equal(t, "strict-origin-when-cross-origin", resp.Header.Get("Referrer-Policy"), path)
		csp := resp.Header.Get("Content-Security-Policy")
		require.Contains(t, csp, "default-src 'self'", path)
		require.Contains(t, csp, "frame-ancestors 'none'", path)
		require.Contains(t, csp, "https://unpkg.com", path)
		// The activity page loads its weather icons from here
		require.Contains(t, csp, "https://openweathermap.org", path)
		// Gravatar redirects the avatar fallback through this image CDN
		require.Contains(t, csp, "https://*.wp.com", path)
	}
}

func TestProxyConfig(t *testing.T) {
	clientIP := func(t *testing.T, opts ...func(*Config)) string {
		client := NewTestClient(t, opts...)
		client.app.Get("/ip", func(c *fiber.Ctx) error { return c.SendString(c.IP()) })
		req := httptest.NewRequest("GET", "/ip", nil)
		req.Header.Set("X-Forwarded-For", "203.0.113.7")
		resp, err := client.app.Test(req, testRequestTimeout)
		require.NoError(t, err)
		return client.GetResponseBody(resp)
	}

	t.Run("forwarded header is ignored by default", func(t *testing.T) {
		require.NotEqual(t, "203.0.113.7", clientIP(t))
	})

	t.Run("forwarded header is used for a trusted proxy", func(t *testing.T) {
		got := clientIP(t, func(cfg *Config) {
			cfg.ProxyHeader = "X-Forwarded-For"
			cfg.TrustedProxies = []string{"0.0.0.0"}
		})
		require.Equal(t, "203.0.113.7", got)
	})

	t.Run("forwarded header is ignored for an untrusted proxy", func(t *testing.T) {
		got := clientIP(t, func(cfg *Config) {
			cfg.ProxyHeader = "X-Forwarded-For"
			cfg.TrustedProxies = []string{"10.0.0.1"}
		})
		require.NotEqual(t, "203.0.113.7", got)
	})
}

func TestErrorPagesHideInternals(t *testing.T) {
	client := NewTestClient(t)
	client.app.Get("/boom", func(c *fiber.Ctx) error {
		return errors.New("Error 1146: table users_secret doesn't exist")
	})
	client.app.Get("/wrapped", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusInternalServerError, "sql: connection refused to db-host")
	})
	client.app.Get("/conflict", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusConflict, "Already exists")
	})

	t.Run("plain errors are hidden and return 500", func(t *testing.T) {
		resp := client.Get("/boom")
		require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		body := client.GetResponseBody(resp)
		client.AssertResponseNotContains(body, "users_secret")
		client.AssertResponseContains(body, "Der opstod en fejl.")
	})

	t.Run("5xx fiber errors are hidden", func(t *testing.T) {
		resp := client.Get("/wrapped")
		require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		client.AssertResponseNotContains(client.GetResponseBody(resp), "db-host")
	})

	t.Run("4xx fiber errors keep their message and status", func(t *testing.T) {
		resp := client.Get("/conflict")
		require.Equal(t, fiber.StatusConflict, resp.StatusCode)
		client.AssertResponseContains(client.GetResponseBody(resp), "Already exists")
	})

	t.Run("unknown routes return 404", func(t *testing.T) {
		require.Equal(t, fiber.StatusNotFound, client.Get("/no-such-page").StatusCode)
	})

	t.Run("validation errors do not echo field names or constraints", func(t *testing.T) {
		email := client.CreateUser("Error User", "password")
		client.Login(email, "password")
		activity := randomActivity()
		activity.Comment = strings.Repeat("secret-value", 200)
		resp := client.CreateActivity(activity)
		body := client.GetResponseBody(resp)
		client.AssertResponseContains(body, `role="alert">Invalid input</div>`)
		client.AssertResponseNotContains(body, "constraint")
		client.AssertResponseNotContains(body, "`Comment`")
	})

	t.Run("missing records return 404 without driver text", func(t *testing.T) {
		resp := client.Post("/activities/"+uuid.NewString()+"/delete", url.Values{})
		body := client.GetResponseBody(resp)
		client.AssertResponseNotContains(body, "record not found")
		client.AssertResponseNotContains(body, "SELECT")
	})
}

func TestAuthRateLimit(t *testing.T) {
	client := NewTestClient(t, func(cfg *Config) { cfg.AuthRateLimit = 3 })

	for i := 0; i < 3; i++ {
		require.Equal(t, fiber.StatusOK, client.Get("/auth/login").StatusCode)
	}

	resp := client.Get("/auth/login")
	require.Equal(t, fiber.StatusTooManyRequests, resp.StatusCode)
	client.AssertResponseContains(client.GetResponseBody(resp), "For mange forsøg")

	t.Run("other pages are not limited", func(t *testing.T) {
		require.Equal(t, fiber.StatusOK, client.Get("/").StatusCode)
	})
}

func TestAuthRateLimitIsPerClientIP(t *testing.T) {
	client := NewTestClient(t, func(cfg *Config) {
		cfg.AuthRateLimit = 1
		cfg.ProxyHeader = "X-Forwarded-For"
		cfg.TrustedProxies = []string{"0.0.0.0"}
	})

	getAs := func(ip string) int {
		req := httptest.NewRequest("GET", "/auth/login", nil)
		req.Header.Set("X-Forwarded-For", ip)
		resp, err := client.app.Test(req, testRequestTimeout)
		require.NoError(t, err)
		return resp.StatusCode
	}

	require.Equal(t, fiber.StatusOK, getAs("203.0.113.1"))
	require.Equal(t, fiber.StatusTooManyRequests, getAs("203.0.113.1"))
	// A second client behind the same proxy is not locked out
	require.Equal(t, fiber.StatusOK, getAs("203.0.113.2"))
}

func TestUserTextIsEscaped(t *testing.T) {
	const payload = `<script>alert("xss")</script>`
	const escaped = `&lt;script&gt;`

	client := NewTestClient(t)
	client.CreateUser("Other User", "password")
	client.devLogin(payload, "xss-user@example.com")

	activity := randomActivity()
	activity.CategoryID = types.Other
	activity.TypeID = ""
	activity.OtherType = payload
	activity.Role = payload
	activity.Comment = payload
	client.CreateActivity(activity)
	activities := client.GetActivities()
	require.Len(t, activities, 1)
	id := activities[0].ID.String()

	paths := []string{
		"/activities/" + id,
		"/activities/" + id + "/edit",
		"/activities/list",
		"/activities/pending",
		"/users/" + "2",
		"/scoreboard",
		"/",
	}
	for _, path := range paths {
		resp := client.Get(path)
		require.Equal(t, fiber.StatusOK, resp.StatusCode, path)
		body := client.GetResponseBody(resp)
		client.AssertResponseNotContains(body, payload)
		client.AssertResponseNotContains(body, `<script>alert`)
		if path == "/activities/"+id {
			client.AssertResponseContains(body, escaped)
		}
	}
}

func TestActivitiesSurviveGeocoderFailure(t *testing.T) {
	client := NewTestClient(t, func(cfg *Config) {
		cfg.Geocoder = fakeGeocoder{err: errors.New("geocode API returned status 429")}
	})
	client.CreateUser("Other User", "password")
	email := client.CreateUser("Geocode User", "password")
	client.Login(email, "password")

	t.Run("create falls back to the submitted location", func(t *testing.T) {
		activity := randomActivity()
		activity.Location = "Havnen"
		client.CreateActivity(activity)

		activities := client.GetActivities()
		require.Len(t, activities, 1)
		require.Equal(t, "Havnen", activities[0].Location)
	})

	t.Run("update falls back to the coordinates", func(t *testing.T) {
		activity := client.GetActivities()[0]
		activity.Location = ""
		activity.Lat = 55.5
		activity.Lng = 12.25
		client.UpdateActivity(activity)

		activities := client.GetActivities()
		require.Len(t, activities, 1)
		require.Equal(t, "55.5000, 12.2500", activities[0].Location)
	})
}

func TestAuthRateLimitDefaultAndOverride(t *testing.T) {
	hit := func(client *TestClient, n int) (limitedAt int) {
		for i := 1; i <= n; i++ {
			if client.Get("/auth/login").StatusCode == fiber.StatusTooManyRequests {
				return i
			}
		}
		return 0
	}

	t.Run("the production default returns 429", func(t *testing.T) {
		client := NewTestClient(t, func(cfg *Config) {
			cfg.AuthRateLimit = config.DefaultAuthRateLimit
			cfg.AuthRateWindow = config.DefaultAuthRateWindow
		})
		require.Equal(t, config.DefaultAuthRateLimit+1, hit(client, config.DefaultAuthRateLimit+5))
	})

	t.Run("the test override lifts the limit", func(t *testing.T) {
		client := NewTestClient(t)
		require.Zero(t, hit(client, config.DefaultAuthRateLimit*3))
	})

	t.Run("requests are allowed again after the window", func(t *testing.T) {
		client := NewTestClient(t, func(cfg *Config) {
			cfg.AuthRateLimit = 1
			cfg.AuthRateWindow = time.Second
		})
		require.Equal(t, fiber.StatusOK, client.Get("/auth/login").StatusCode)
		require.Equal(t, fiber.StatusTooManyRequests, client.Get("/auth/login").StatusCode)
		time.Sleep(1200 * time.Millisecond)
		require.Equal(t, fiber.StatusOK, client.Get("/auth/login").StatusCode)
	})
}

func TestActivityPageShowsWeatherFromTheWeatherService(t *testing.T) {
	client := NewTestClient(t)
	client.CreateUser("Other User", "password")
	email := client.CreateUser("Weather User", "password")
	client.Login(email, "password")
	activity := randomActivity()
	activity.Date = types.Date(time.Now().AddDate(0, 0, -10))
	client.CreateActivity(activity)
	activities := client.GetActivities()
	require.Len(t, activities, 1)

	body := client.GetResponseBody(client.Get("/activities/" + activities[0].ID.String()))
	client.AssertResponseContains(body, "12.5°C")
}

func TestStoredWeather(t *testing.T) {
	daysAgo := func(n int) types.Date { return types.Date(time.Now().AddDate(0, 0, -n)) }

	newClient := func(t *testing.T, weather *fakeWeather) *TestClient {
		client := NewTestClient(t, func(cfg *Config) { cfg.Weather = weather })
		client.CreateUser("Other User", "password")
		email := client.CreateUser("Weather User", "password")
		client.Login(email, "password")
		return client
	}
	create := func(client *TestClient, date types.Date) types.Activity {
		activity := randomActivity()
		activity.Date = date
		client.CreateActivity(activity)
		activities := client.GetActivities()
		return activities[len(activities)-1]
	}
	view := func(client *TestClient, activity types.Activity) string {
		resp := client.Get("/activities/" + activity.ID.String())
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		return client.GetResponseBody(resp)
	}

	t.Run("a past activity stores the weather when it is created and never looks it up again", func(t *testing.T) {
		weather := &fakeWeather{temperature: 17.5, code: 3}
		client := newClient(t, weather)

		activity := create(client, daysAgo(10))
		require.Equal(t, 1, weather.calls())
		require.Equal(t, &types.Weather{Temperature: 17.5, Code: 3}, activity.Weather)

		client.AssertResponseContains(view(client, activity), "17.5")
		view(client, activity)
		view(client, activity)
		client.GetActivities()
		require.Equal(t, 1, weather.calls(), "page views and lists must not call the weather API")
	})

	t.Run("a past activity without stored weather is looked up once and stored", func(t *testing.T) {
		weather := &fakeWeather{err: errors.New("status 429")}
		client := newClient(t, weather)

		activity := create(client, daysAgo(10))
		require.Nil(t, activity.Weather, "a failing lookup must not block saving")
		require.Equal(t, 1, weather.calls())

		// The API is down during a view too, so nothing is shown or stored
		view(client, activity)
		require.Equal(t, 2, weather.calls())
		require.Nil(t, client.GetActivities()[0].Weather)

		weather.set(func(f *fakeWeather) { f.err = nil; f.temperature = 8.25; f.code = 61 })
		client.AssertResponseContains(view(client, activity), "8.25")
		require.Equal(t, 3, weather.calls())
		require.Equal(t, &types.Weather{Temperature: 8.25, Code: 61}, client.GetActivities()[0].Weather)

		view(client, activity)
		view(client, activity)
		require.Equal(t, 3, weather.calls(), "stored weather is served from the database")
	})

	t.Run("weather the API has not computed yet is not stored as zero", func(t *testing.T) {
		weather := &fakeWeather{missing: true}
		client := newClient(t, weather)

		activity := create(client, daysAgo(2))
		view(client, activity)

		require.Nil(t, client.GetActivities()[0].Weather)
	})

	t.Run("today is looked up on every view and never stored", func(t *testing.T) {
		weather := &fakeWeather{temperature: 21.5, code: 1}
		client := newClient(t, weather)

		activity := create(client, daysAgo(0))
		require.Zero(t, weather.calls(), "no lookup when saving today")

		client.AssertResponseContains(view(client, activity), "21.5")
		view(client, activity)
		require.Equal(t, 2, weather.calls())
		require.Nil(t, client.GetActivities()[0].Weather)
	})

	t.Run("a future activity has no weather", func(t *testing.T) {
		weather := &fakeWeather{temperature: 30, code: 0}
		client := newClient(t, weather)

		activity := create(client, types.Date(time.Now().AddDate(0, 0, 5)))
		view(client, activity)

		require.Zero(t, weather.calls())
		require.Nil(t, client.GetActivities()[0].Weather)
	})

	t.Run("an update refetches only when the day or place changed", func(t *testing.T) {
		weather := &fakeWeather{temperature: 10, code: 2}
		client := newClient(t, weather)
		activity := create(client, daysAgo(10))
		require.Equal(t, 1, weather.calls())

		// Only the comment changes
		activity.Comment = "Edited"
		client.UpdateActivity(activity)
		require.Equal(t, 1, weather.calls())
		require.Equal(t, &types.Weather{Temperature: 10, Code: 2}, client.GetActivities()[0].Weather)

		// A new day gets the weather of that day
		weather.set(func(f *fakeWeather) { f.temperature = 12.5; f.code = 63 })
		activity.Date = daysAgo(20)
		client.UpdateActivity(activity)
		require.Equal(t, 2, weather.calls())
		require.Equal(t, &types.Weather{Temperature: 12.5, Code: 63}, client.GetActivities()[0].Weather)

		// A new place gets the weather of that place
		weather.set(func(f *fakeWeather) { f.temperature = 14; f.code = 80 })
		activity.Lat = 56.1
		client.UpdateActivity(activity)
		require.Equal(t, 3, weather.calls())
		require.Equal(t, &types.Weather{Temperature: 14, Code: 80}, client.GetActivities()[0].Weather)
	})

	t.Run("an update to a day without weather clears the old weather", func(t *testing.T) {
		weather := &fakeWeather{temperature: 10, code: 2}
		client := newClient(t, weather)
		activity := create(client, daysAgo(10))
		require.NotNil(t, activity.Weather)

		activity.Date = types.Date(time.Now().AddDate(0, 0, 3))
		client.UpdateActivity(activity)
		require.Nil(t, client.GetActivities()[0].Weather)

		// A failed lookup for the new day clears it too
		activity.Date = daysAgo(10)
		client.UpdateActivity(activity)
		require.NotNil(t, client.GetActivities()[0].Weather)
		weather.set(func(f *fakeWeather) { f.err = errors.New("down") })
		activity.Date = daysAgo(11)
		client.UpdateActivity(activity)
		require.Nil(t, client.GetActivities()[0].Weather)
	})

	t.Run("a failing weather service does not block saving or the page", func(t *testing.T) {
		weather := &fakeWeather{err: errors.New("connection refused")}
		client := newClient(t, weather)

		activity := create(client, daysAgo(10))
		activity.Comment = "Edited"
		activity.Date = daysAgo(12)
		client.UpdateActivity(activity)

		activities := client.GetActivities()
		require.Len(t, activities, 1)
		require.Equal(t, "Edited", activities[0].Comment)
		client.AssertResponseContains(view(client, activities[0]), "Edited")
	})
}

func TestFooterCreditsTheImageAuthors(t *testing.T) {
	client := NewTestClient(t)
	body := client.GetResponseBody(client.Get("/"))

	for _, credit := range []string{
		"Hammock icons created by Icon.doit - Flaticon",
		"Mushroom icons created by Freepik - Flaticon",
		"Berries icons created by Freepik - Flaticon",
		"Spices icons created by Freepik - Flaticon",
		"Harvest icons created by iconixar - Flaticon",
		"Reminder icons created by Freepik - Flaticon",
		"Kayak icons created by Freepik - Flaticon",
		"Photo by Laura Chouette",
		"Weather data by Open-Meteo.com",
	} {
		client.AssertResponseContains(body, credit)
	}
}
