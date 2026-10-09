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

	// DEVMODE enables the local dev login bypass (no Auth0 tenant needed).
	// Keep this false in production.
	DEVMODE = getEnvBool("AUTH_DEV_MODE", false)

	// ALLOWDEVMODE lets the dev login run on a non loopback address. Never set this in production.
	ALLOWDEVMODE = getEnvBool("ALLOW_DEV_MODE", false)

	// AUTH0DOMAIN is the Auth0 tenant domain, e.g. "my-tenant.eu.auth0.com".
	AUTH0DOMAIN = getEnv("AUTH0_DOMAIN", "")
	// AUTH0CLIENTID is the Auth0 application client ID.
	AUTH0CLIENTID = getEnv("AUTH0_CLIENT_ID", "")
	// AUTH0CLIENTSECRET is the Auth0 application client secret.
	AUTH0CLIENTSECRET = getEnv("AUTH0_CLIENT_SECRET", "")
	// AUTH0CALLBACKURL is the OIDC redirect URL registered in Auth0,
	// e.g. "http://localhost:3000/auth/callback".
	AUTH0CALLBACKURL = getEnv("AUTH0_CALLBACK_URL", "")
	// AUTH0LOGOUTRETURNURL is where Auth0 returns the user after logout,
	// e.g. "http://localhost:3000/".
	AUTH0LOGOUTRETURNURL = getEnv("AUTH0_LOGOUT_RETURN_URL", "")
)

func getEnv(name string, fallback string) string {
	if value, exists := os.LookupEnv(name); exists {
		return value
	}

	// An empty fallback is a legitimate default (e.g. unset Auth0 vars in dev mode).
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
