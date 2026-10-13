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

func TestContentSecurityPolicy_Firebase(t *testing.T) {
	t.Run("without an auth domain nothing of Firebase is allowed", func(t *testing.T) {
		policy := ContentSecurityPolicy("", "")

		require.NotContains(t, policy, "identitytoolkit.googleapis.com")
		require.NotContains(t, policy, "securetoken.googleapis.com")
		require.NotContains(t, directive(t, policy, "script-src"), "gstatic.com")
		require.Equal(t, "'self'", directive(t, policy, "frame-src"))
	})

	t.Run("an auth domain allows the sign-in, and nothing else changes", func(t *testing.T) {
		policy := ContentSecurityPolicy("logbogen.firebaseapp.com", "")

		require.Contains(t, directive(t, policy, "script-src"), "https://www.gstatic.com")
		require.Contains(t, directive(t, policy, "connect-src"), "https://identitytoolkit.googleapis.com")
		require.Contains(t, directive(t, policy, "connect-src"), "https://securetoken.googleapis.com")
		require.Contains(t, directive(t, policy, "frame-src"), "https://logbogen.firebaseapp.com")
		require.Equal(t, "'self'", directive(t, policy, "form-action"), "logout no longer leaves the site")
		require.Equal(t, "'none'", directive(t, policy, "frame-ancestors"))
		require.Equal(t, "'self'", directive(t, policy, "base-uri"))
	})
}

func TestContentSecurityPolicy_Emulator(t *testing.T) {
	policy := ContentSecurityPolicy("logbogen.firebaseapp.com", "127.0.0.1:9099")

	require.Contains(t, directive(t, policy, "connect-src"), "http://127.0.0.1:9099")
	require.Contains(t, directive(t, policy, "frame-src"), "http://127.0.0.1:9099")
	require.NotContains(t, ContentSecurityPolicy("logbogen.firebaseapp.com", ""), "127.0.0.1")
}

func TestSecurityHeaders_SendsThePolicy(t *testing.T) {
	app := fiber.New()
	app.Use(SecurityHeaders("logbogen.firebaseapp.com", ""))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	require.NoError(t, err)

	policy := resp.Header.Get("Content-Security-Policy")
	require.Contains(t, directive(t, policy, "frame-src"), "https://logbogen.firebaseapp.com")
}

func TestSecurityHeaders_LetsThePageTalkToItsSignInPopup(t *testing.T) {
	app := fiber.New()
	app.Use(SecurityHeaders("", ""))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("ok") })

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	require.NoError(t, err)

	require.Equal(t, "same-origin-allow-popups", resp.Header.Get("Cross-Origin-Opener-Policy"))
}
