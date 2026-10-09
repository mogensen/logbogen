package config

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSplitList(t *testing.T) {
	require.Empty(t, splitList(""))
	require.Equal(t, []string{"127.0.0.1"}, splitList("127.0.0.1"))
	require.Equal(t, []string{"10.0.0.1", "10.0.0.2"}, splitList(" 10.0.0.1, ,10.0.0.2 "))
}

func TestDefaultListenAddr(t *testing.T) {
	require.Equal(t, "127.0.0.1:3000", DefaultListenAddr)
}

func TestDefaultDB(t *testing.T) {
	require.Equal(t, "./data/logbog.db", DefaultDB)
	require.Equal(t, DefaultDB, getEnv("DB_NOT_SET_IN_TESTS", DefaultDB))
}

func TestGetEnvPositiveInt(t *testing.T) {
	const name = "TEST_POSITIVE_INT"

	t.Run("unset uses the default", func(t *testing.T) {
		require.Equal(t, 30, getEnvPositiveInt(name, 30))
	})

	t.Run("a valid value overrides the default", func(t *testing.T) {
		t.Setenv(name, "100000")
		require.Equal(t, 100000, getEnvPositiveInt(name, 30))
	})

	for _, invalid := range []string{"", "abc", "0", "-5", "1.5"} {
		t.Run("invalid value "+invalid+" keeps the default", func(t *testing.T) {
			t.Setenv(name, invalid)
			require.Equal(t, 30, getEnvPositiveInt(name, 30))
		})
	}
}

func TestGetEnvPositiveDuration(t *testing.T) {
	const name = "TEST_POSITIVE_DURATION"

	t.Run("unset uses the default", func(t *testing.T) {
		require.Equal(t, time.Minute, getEnvPositiveDuration(name, time.Minute))
	})

	t.Run("a valid value overrides the default", func(t *testing.T) {
		t.Setenv(name, "90s")
		require.Equal(t, 90*time.Second, getEnvPositiveDuration(name, time.Minute))
	})

	for _, invalid := range []string{"", "soon", "0s", "-1m", "10"} {
		t.Run("invalid value "+invalid+" keeps the default", func(t *testing.T) {
			t.Setenv(name, invalid)
			require.Equal(t, time.Minute, getEnvPositiveDuration(name, time.Minute))
		})
	}
}

func TestAuthRateLimitDefaults(t *testing.T) {
	require.Equal(t, 30, DefaultAuthRateLimit)
	require.Equal(t, time.Minute, DefaultAuthRateWindow)
}

func TestGetEnvChoice(t *testing.T) {
	const name = "TEST_CHOICE"

	t.Run("unset uses the default", func(t *testing.T) {
		require.Equal(t, "json", getEnvChoice(name, "json", "json", "text"))
	})

	t.Run("an allowed value is used, whatever its case", func(t *testing.T) {
		t.Setenv(name, " Text ")
		require.Equal(t, "text", getEnvChoice(name, "json", "json", "text"))
	})

	t.Run("a value that is not allowed keeps the default", func(t *testing.T) {
		t.Setenv(name, "xml")
		require.Equal(t, "json", getEnvChoice(name, "json", "json", "text"))
	})
}

func TestGetEnvLogLevel(t *testing.T) {
	const name = "TEST_LOG_LEVEL"

	t.Run("unset uses the default", func(t *testing.T) {
		require.Equal(t, slog.LevelInfo, getEnvLogLevel(name, slog.LevelInfo))
	})

	for value, want := range map[string]slog.Level{
		"debug": slog.LevelDebug, "INFO": slog.LevelInfo, "warn": slog.LevelWarn, "Error": slog.LevelError,
	} {
		t.Run("level "+value, func(t *testing.T) {
			t.Setenv(name, value)
			require.Equal(t, want, getEnvLogLevel(name, slog.LevelInfo))
		})
	}

	for _, invalid := range []string{"", "verbose", "42x"} {
		t.Run("invalid value "+invalid+" keeps the default", func(t *testing.T) {
			t.Setenv(name, invalid)
			require.Equal(t, slog.LevelWarn, getEnvLogLevel(name, slog.LevelWarn))
		})
	}
}
