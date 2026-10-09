package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSessionCookieFlags(t *testing.T) {
	require.True(t, SessionStore.CookieSecure, "cookie must be Secure by default")
	require.True(t, SessionStore.CookieHTTPOnly)
	require.Equal(t, "Lax", SessionStore.CookieSameSite)

	SetCookieSecure(false)
	t.Cleanup(func() { SetCookieSecure(true) })
	require.False(t, SessionStore.CookieSecure)
}
