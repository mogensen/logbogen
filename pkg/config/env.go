package config

import (
	"fmt"
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
)

var (
	// PORT returns the server listening port
	PORT = getEnv("PORT", "5000")
	// DB returns the name of the sqlite database
	DB = getEnv("DB", "fiber.db")

	// DEVMODE enables the local dev login bypass (no Auth0 tenant needed).
	// Keep this false in production.
	DEVMODE = getEnvBool("AUTH_DEV_MODE", false)

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
