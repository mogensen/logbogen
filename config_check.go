package main

import (
	"errors"
	"fmt"
	"net"
	"net/url"

	"github.com/mogensen/logbook/pkg/auth"
)

// isLoopbackHost reports whether host only accepts connections from the local machine
func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// validateAuthConfig refuses to start with an unsafe authentication setup. Dev mode lets anyone log in
// as any email, so it must stay on a loopback address unless allowDevMode is set explicitly.
func validateAuthConfig(cfg *Config, auth0 auth.Config, allowDevMode bool) error {
	if cfg.DevMode {
		host, _, err := net.SplitHostPort(cfg.ListenAddr)
		if err != nil {
			return fmt.Errorf("invalid listen address %q: %w", cfg.ListenAddr, err)
		}
		if !isLoopbackHost(host) && !allowDevMode {
			return fmt.Errorf("AUTH_DEV_MODE lets anyone log in as any user and cannot listen on %q; "+
				"bind to a loopback address or set ALLOW_DEV_MODE=true", cfg.ListenAddr)
		}
		cfg.Logger.Warn("Dev login bypass is enabled, never use this in production")
		return nil
	}

	var missing []string
	for name, value := range map[string]string{
		"AUTH0_DOMAIN":            auth0.Domain,
		"AUTH0_CLIENT_ID":         auth0.ClientID,
		"AUTH0_CLIENT_SECRET":     auth0.ClientSecret,
		"AUTH0_CALLBACK_URL":      auth0.CallbackURL,
		"AUTH0_LOGOUT_RETURN_URL": auth0.LogoutReturnURL,
	} {
		if value == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required settings: %v", missing)
	}

	for name, value := range map[string]string{
		"AUTH0_CALLBACK_URL":      auth0.CallbackURL,
		"AUTH0_LOGOUT_RETURN_URL": auth0.LogoutReturnURL,
	} {
		u, err := url.Parse(value)
		if err != nil || u.Host == "" {
			return fmt.Errorf("%s is not a valid URL", name)
		}
		if u.Scheme != "https" && !isLoopbackHost(u.Hostname()) {
			return errors.New(name + " must use https unless it points to localhost")
		}
	}
	return nil
}

// formActionOrigins lists the origins a form post may redirect to. Logout posts a form and then redirects
// to Auth0 to end the Auth0 session too, which the form-action policy would otherwise block.
func formActionOrigins(devMode bool, auth0Domain string) []string {
	if devMode || auth0Domain == "" {
		return nil
	}
	return []string{"https://" + auth0Domain}
}
