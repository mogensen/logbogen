package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSessionCookieFlags(t *testing.T) {
	require.True(t, SessionStore.CookieSecure, "cookie must be Secure by default")
	require.True(t, SessionStore.CookieHTTPOnly)
	require.Equal(t, "Lax", SessionStore.CookieSameSite)

	SetCookieSecure(false)
	t.Cleanup(func() { SetCookieSecure(true) })
	require.False(t, SessionStore.CookieSecure)
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

func TestRenameLegacyColumns_KeepsTheIdentityOfExistingUsers(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec("CREATE TABLE users (id integer primary key, auth0_sub text)").Error)
	require.NoError(t, db.Exec("INSERT INTO users (id, auth0_sub) VALUES (1, 'dev|a@example.com')").Error)

	require.NoError(t, renameLegacyColumns(db))
	require.NoError(t, renameLegacyColumns(db), "running it again is a no-op")

	var uid string
	require.NoError(t, db.Raw("SELECT firebase_uid FROM users WHERE id = 1").Scan(&uid).Error)
	require.Equal(t, "dev|a@example.com", uid)
	require.False(t, db.Migrator().HasColumn("users", "auth0_sub"))
}
