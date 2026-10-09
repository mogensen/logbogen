package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// directive returns the value of one directive of a policy
func directive(t *testing.T, policy, name string) string {
	for _, part := range strings.Split(policy, "; ") {
		if strings.HasPrefix(part, name+" ") {
			return strings.TrimPrefix(part, name+" ")
		}
	}
	require.Failf(t, "directive missing", "%s is not in %q", name, policy)
	return ""
}

func TestContentSecurityPolicy_FormAction(t *testing.T) {
	t.Run("only the app itself by default", func(t *testing.T) {
		require.Equal(t, "'self'", directive(t, ContentSecurityPolicy(), "form-action"))
	})

	t.Run("lists the origins a form post may redirect to", func(t *testing.T) {
		policy := ContentSecurityPolicy("https://tenant.eu.auth0.com")

		require.Equal(t, "'self' https://tenant.eu.auth0.com", directive(t, policy, "form-action"))
		// The other directives are untouched
		require.Equal(t, "'none'", directive(t, policy, "frame-ancestors"))
		require.Equal(t, "'self'", directive(t, policy, "base-uri"))
	})
}

func TestSecurityHeaders_SendsThePolicyWithTheFormActionOrigins(t *testing.T) {
	app := fiber.New()
	app.Use(SecurityHeaders("https://tenant.eu.auth0.com"))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	require.NoError(t, err)

	policy := resp.Header.Get("Content-Security-Policy")
	require.Equal(t, "'self' https://tenant.eu.auth0.com", directive(t, policy, "form-action"))
}
