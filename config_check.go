package main

import (
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
)

// isLoopbackHost reports whether host only accepts connections from the local machine
func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// validateProxyConfig refuses a proxy header without trusted proxies, since any client could then set its own IP.
func validateProxyConfig(cfg *Config) error {
	if cfg.ProxyHeader != "" && len(cfg.TrustedProxies) == 0 {
		return errors.New("PROXY_HEADER is set without TRUSTED_PROXIES, so clients could spoof their IP; " +
			"list the proxies in TRUSTED_PROXIES")
	}
	return nil
}

// validateAuthConfig refuses to start with an unsafe authentication setup. Dev mode lets anyone log in
// as any email, so it must stay on a loopback address unless allowDevMode is set explicitly.
func validateAuthConfig(cfg *Config, allowDevMode bool) error {
	firebase := cfg.Firebase
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

	if firebase.EmulatorHost != "" {
		if err := validateEmulator(cfg); err != nil {
			return err
		}
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

// validateEmulator keeps the Firebase Auth emulator, which trusts unsigned tokens, on this machine.
func validateEmulator(cfg *Config) error {
	emulatorHost, _, err := net.SplitHostPort(cfg.Firebase.EmulatorHost)
	if err != nil {
		return fmt.Errorf("invalid FIREBASE_AUTH_EMULATOR_HOST %q: %w", cfg.Firebase.EmulatorHost, err)
	}
	listenHost, _, err := net.SplitHostPort(cfg.ListenAddr)
	if err != nil {
		return fmt.Errorf("invalid listen address %q: %w", cfg.ListenAddr, err)
	}
	if !isLoopbackHost(emulatorHost) || !isLoopbackHost(listenHost) {
		return errors.New("FIREBASE_AUTH_EMULATOR_HOST trusts unsigned tokens and needs a loopback emulator " +
			"and listen address")
	}
	cfg.Logger.Warn("Firebase Auth emulator is enabled, never use this in production")
	return nil
}

// cspFirebase returns the Firebase auth domain and emulator host the CSP must allow, none in dev mode.
func cspFirebase(cfg *Config) (authDomain, emulatorHost string) {
	if cfg.DevMode {
		return "", ""
	}
	return cfg.Firebase.AuthDomain, cfg.Firebase.EmulatorHost
}
