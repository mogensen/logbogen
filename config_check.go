package main

import (
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"

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
func validateAuthConfig(cfg *Config, firebase auth.Config, allowDevMode bool) error {
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
		"FIREBASE_PROJECT_ID":  firebase.ProjectID,
		"FIREBASE_API_KEY":     firebase.APIKey,
		"FIREBASE_AUTH_DOMAIN": firebase.AuthDomain,
	} {
		if value == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("missing required settings: %v", missing)
	}

	if strings.ContainsAny(firebase.AuthDomain, "/:? ") {
		return errors.New("FIREBASE_AUTH_DOMAIN must be a host name like my-project.firebaseapp.com")
	}
	return nil
}

// cspAuthDomain is the Firebase auth domain the CSP must allow, none in dev mode which signs in locally.
func cspAuthDomain(devMode bool, firebase auth.Config) string {
	if devMode {
		return ""
	}
	return firebase.AuthDomain
}
