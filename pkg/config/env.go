package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

// DefaultListenAddr only accepts local connections, containers set LISTEN_ADDR to listen on all interfaces
const DefaultListenAddr = "127.0.0.1:3000"

// DefaultAuthRateLimit and DefaultAuthRateWindow are the production limits on auth requests per client IP
const (
	DefaultAuthRateLimit  = 30
	DefaultAuthRateWindow = time.Minute
)

// DefaultDB is the sqlite file used when DB is not set, inside a directory that can be mounted as a volume
const DefaultDB = "./data/logbog.db"

var (
	// LISTENADDR is the host and port the server listens on.
	LISTENADDR = getEnv("LISTEN_ADDR", DefaultListenAddr)
	// PORT returns the server listening port
	PORT = getEnv("PORT", "5000")
	// DB returns the path of the sqlite database, which also holds the sessions
	DB = getEnv("DB", DefaultDB)

	// PROXYHEADER is the header carrying the client IP when behind a reverse proxy, e.g. "X-Forwarded-For".
	PROXYHEADER = getEnv("PROXY_HEADER", "")
	// TRUSTEDPROXIES is a comma separated list of proxy IPs allowed to set PROXYHEADER.
	TRUSTEDPROXIES = splitList(getEnv("TRUSTED_PROXIES", ""))

	// AUTHRATELIMIT is the number of login, callback and logout requests allowed per client IP per window.
	// An unset or invalid value keeps the strict default.
	AUTHRATELIMIT = getEnvPositiveInt("AUTH_RATE_LIMIT", DefaultAuthRateLimit)
	// AUTHRATEWINDOW is the period the limit applies to. An unset or invalid value keeps the strict default.
	AUTHRATEWINDOW = getEnvPositiveDuration("AUTH_RATE_WINDOW", DefaultAuthRateWindow)

	// LOGFORMAT is "json" for structured JSON lines, or "text" for key=value lines that are easier to read locally.
	LOGFORMAT = getEnvChoice("LOG_FORMAT", "json", "json", "text")
	// LOGLEVEL is the lowest level that is logged: debug, info, warn or error.
	LOGLEVEL = getEnvLogLevel("LOG_LEVEL", slog.LevelInfo)

	// DEVMODE enables the local dev login bypass (no Firebase project needed).
	// Keep this false in production.
	DEVMODE = getEnvBool("AUTH_DEV_MODE", false)

	// ALLOWDEVMODE lets the dev login run on a non loopback address. Never set this in production.
	ALLOWDEVMODE = getEnvBool("ALLOW_DEV_MODE", false)

	// FIREBASEPROJECTID is the Firebase project ID, which ID tokens are issued for.
	FIREBASEPROJECTID = getEnv("FIREBASE_PROJECT_ID", "")
	// FIREBASEAPIKEY is the public web API key of the Firebase web app.
	FIREBASEAPIKEY = getEnv("FIREBASE_API_KEY", "")
	// FIREBASEAUTHDOMAIN is the host of the Firebase auth handler, e.g. "my-project.firebaseapp.com".
	FIREBASEAUTHDOMAIN = getEnv("FIREBASE_AUTH_DOMAIN", "")
)

func getEnv(name string, fallback string) string {
	if value, exists := os.LookupEnv(name); exists {
		return value
	}

	// An empty fallback is a legitimate default (e.g. unset Firebase vars in dev mode).
	return fallback
}

func getEnvBool(name string, fallback bool) bool {
	value, exists := os.LookupEnv(name)
	if !exists {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		panic(fmt.Sprintf(`Environment variable %v must be a boolean :: %v`, name, value))
	}
	return parsed
}

func splitList(value string) []string {
	var items []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

// getEnvPositiveInt reads a positive integer and falls back, with a warning, when it is unset or invalid
func getEnvPositiveInt(name string, fallback int) int {
	value, exists := os.LookupEnv(name)
	if !exists {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		slog.Warn("Ignoring invalid setting, using the default", "name", name, "value", value, "default", fallback)
		return fallback
	}
	return parsed
}

// getEnvPositiveDuration reads a positive duration like "30s", and falls back with a warning when it is invalid
func getEnvPositiveDuration(name string, fallback time.Duration) time.Duration {
	value, exists := os.LookupEnv(name)
	if !exists {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		slog.Warn("Ignoring invalid setting, using the default", "name", name, "value", value, "default", fallback)
		return fallback
	}
	return parsed
}

// getEnvChoice reads one of the allowed values and falls back, with a warning, when it is unset or not allowed
func getEnvChoice(name, fallback string, allowed ...string) string {
	value, exists := os.LookupEnv(name)
	if !exists {
		return fallback
	}

	value = strings.ToLower(strings.TrimSpace(value))
	for _, option := range allowed {
		if value == option {
			return value
		}
	}
	slog.Warn("Ignoring invalid setting, using the default", "name", name, "value", value, "allowed", allowed)
	return fallback
}

// getEnvLogLevel reads a log level name and falls back, with a warning, when it is unset or invalid
func getEnvLogLevel(name string, fallback slog.Level) slog.Level {
	value, exists := os.LookupEnv(name)
	if !exists {
		return fallback
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(strings.TrimSpace(value))); err != nil {
		slog.Warn("Ignoring invalid setting, using the default", "name", name, "value", value, "default", fallback.String())
		return fallback
	}
	return level
}
