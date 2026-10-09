package database

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type logRow struct {
	ID   uint
	Name string
}

// loggedDB returns a database that logs through slog into the buffer at the given level
func loggedDB(t *testing.T, level slog.Level, slow time.Duration) (*gorm.DB, *bytes.Buffer) {
	var buf bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: level})))
	t.Cleanup(func() { slog.SetDefault(previous) })

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: &gormLogger{slow: slow}})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&logRow{}))
	buf.Reset()
	return db, &buf
}

func entries(t *testing.T, buf *bytes.Buffer) []map[string]any {
	var result []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		var entry map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &entry), line)
		result = append(result, entry)
	}
	return result
}

func TestGormLogger(t *testing.T) {
	t.Run("queries are debug lines with the sql, rows and latency", func(t *testing.T) {
		db, buf := loggedDB(t, slog.LevelDebug, time.Hour)

		require.NoError(t, db.Create(&logRow{Name: "a"}).Error)

		lines := entries(t, buf)
		require.Len(t, lines, 1)
		require.Equal(t, "query", lines[0]["msg"])
		require.Equal(t, "DEBUG", lines[0]["level"])
		require.Contains(t, lines[0]["sql"], "INSERT INTO `log_rows`")
		require.EqualValues(t, 1, lines[0]["rows"])
		require.GreaterOrEqual(t, lines[0]["latency_ms"], 0.0)
	})

	t.Run("queries are not logged at info level", func(t *testing.T) {
		db, buf := loggedDB(t, slog.LevelInfo, time.Hour)

		require.NoError(t, db.Create(&logRow{Name: "a"}).Error)

		require.Empty(t, entries(t, buf))
	})

	t.Run("a missing record is not an error", func(t *testing.T) {
		db, buf := loggedDB(t, slog.LevelInfo, time.Hour)

		err := db.First(&logRow{}, "id = ?", 99).Error

		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
		require.Empty(t, entries(t, buf))
	})

	t.Run("a failed query is an error line with the cause", func(t *testing.T) {
		db, buf := loggedDB(t, slog.LevelInfo, time.Hour)

		require.Error(t, db.Exec("SELECT * FROM missing_table").Error)

		lines := entries(t, buf)
		require.Len(t, lines, 1)
		require.Equal(t, "query failed", lines[0]["msg"])
		require.Equal(t, "ERROR", lines[0]["level"])
		require.Contains(t, lines[0]["error"], "no such table")
		require.Equal(t, "SELECT * FROM missing_table", lines[0]["sql"])
	})

	t.Run("a slow query is a warning", func(t *testing.T) {
		db, buf := loggedDB(t, slog.LevelInfo, time.Nanosecond)

		require.NoError(t, db.Create(&logRow{Name: "a"}).Error)

		lines := entries(t, buf)
		require.Len(t, lines, 1)
		require.Equal(t, "slow query", lines[0]["msg"])
		require.Equal(t, "WARN", lines[0]["level"])
	})
}

func TestNewGormLogger_IsUsedForTheConnection(t *testing.T) {
	require.IsType(t, &gormLogger{}, NewGormLogger())
	require.Equal(t, 200*time.Millisecond, NewGormLogger().(*gormLogger).slow)
}
