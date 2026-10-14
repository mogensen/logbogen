package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mogensen/logbook/pkg/config"
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

func TestImportingThePackageCreatesNoDatabaseFile(t *testing.T) {
	if config.DB != config.DefaultDB {
		t.Skip("DB is set, so the default path says nothing")
	}
	require.NoFileExists(t, config.DB)
	require.NoDirExists(t, filepath.Dir(config.DB))
}

func TestEnsureDir(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "nested", "data", "logbog.db")

	require.NoError(t, ensureDir(dbPath))
	info, err := os.Stat(filepath.Dir(dbPath))
	require.NoError(t, err)
	require.True(t, info.IsDir())

	// Running it again on an existing directory is fine
	require.NoError(t, ensureDir(dbPath))
}
