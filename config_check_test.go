package main

import (
	"io"
	"log/slog"
	"testing"

	"github.com/mogensen/logbook/pkg/auth"
	"github.com/stretchr/testify/require"
)

func validFirebase() auth.Config {
	return auth.Config{
		ProjectID:  "logbogen",
		APIKey:     "api-key",
		AuthDomain: "logbogen.firebaseapp.com",
	}
}

func TestValidateAuthConfig(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	none := func() auth.Config { return auth.Config{} }

	tests := []struct {
		name         string
		devMode      bool
		listenAddr   string
		allowDevMode bool
		firebase     func() auth.Config
		wantErr      string
	}{
		{name: "dev mode on loopback", devMode: true, listenAddr: "127.0.0.1:3000", firebase: none},
		{name: "dev mode on localhost", devMode: true, listenAddr: "localhost:3000", firebase: none},
		{name: "dev mode on all interfaces", devMode: true, listenAddr: ":3000", firebase: none, wantErr: "AUTH_DEV_MODE"},
		{name: "dev mode on a public address", devMode: true, listenAddr: "0.0.0.0:3000", firebase: none, wantErr: "AUTH_DEV_MODE"},
		{name: "dev mode override", devMode: true, listenAddr: "0.0.0.0:3000", allowDevMode: true, firebase: none},
		{name: "dev mode with a bad address", devMode: true, listenAddr: "nonsense", firebase: none, wantErr: "invalid listen address"},
		{name: "complete production settings", listenAddr: "127.0.0.1:3000", firebase: validFirebase},
		{name: "missing api key", listenAddr: "127.0.0.1:3000", firebase: func() auth.Config {
			f := validFirebase()
			f.APIKey = ""
			return f
		}, wantErr: "FIREBASE_API_KEY"},
		{name: "missing project id", listenAddr: "127.0.0.1:3000", firebase: func() auth.Config {
			f := validFirebase()
			f.ProjectID = ""
			return f
		}, wantErr: "FIREBASE_PROJECT_ID"},
		{name: "no firebase settings", listenAddr: "127.0.0.1:3000", firebase: none, wantErr: "missing required settings"},
		{name: "auth domain with a scheme", listenAddr: "127.0.0.1:3000", firebase: func() auth.Config {
			f := validFirebase()
			f.AuthDomain = "https://logbogen.firebaseapp.com"
			return f
		}, wantErr: "FIREBASE_AUTH_DOMAIN must be a host name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{DevMode: tt.devMode, ListenAddr: tt.listenAddr, Firebase: tt.firebase(), Logger: logger}
			err := validateAuthConfig(cfg, tt.allowDevMode)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestValidateAuthConfig_Emulator(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name       string
		listenAddr string
		emulator   string
		wantErr    string
	}{
		{name: "loopback emulator on a loopback address", listenAddr: "127.0.0.1:3000", emulator: "127.0.0.1:9099"},
		{name: "localhost emulator", listenAddr: "localhost:3000", emulator: "localhost:9099"},
		{name: "emulator on another machine", listenAddr: "127.0.0.1:3000", emulator: "10.0.0.5:9099", wantErr: "loopback"},
		{name: "app on all interfaces", listenAddr: ":3000", emulator: "127.0.0.1:9099", wantErr: "loopback"},
		{name: "emulator without a port", listenAddr: "127.0.0.1:3000", emulator: "127.0.0.1", wantErr: "FIREBASE_AUTH_EMULATOR_HOST"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			firebase := validFirebase()
			firebase.EmulatorHost = tt.emulator
			cfg := &Config{ListenAddr: tt.listenAddr, Firebase: firebase, Logger: logger}

			err := validateAuthConfig(cfg, false)

			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.wantErr)
		})
	}

	t.Run("dev mode does not use the emulator", func(t *testing.T) {
		cfg := &Config{DevMode: true, ListenAddr: "127.0.0.1:3000", Logger: logger,
			Firebase: auth.Config{EmulatorHost: "10.0.0.5:9099"}}

		require.NoError(t, validateAuthConfig(cfg, false))
	})
}

func TestCSPFirebase(t *testing.T) {
	firebase := validFirebase()
	firebase.EmulatorHost = "127.0.0.1:9099"

	domain, emulator := cspFirebase(&Config{Firebase: firebase})
	require.Equal(t, "logbogen.firebaseapp.com", domain)
	require.Equal(t, "127.0.0.1:9099", emulator)

	domain, emulator = cspFirebase(&Config{DevMode: true, Firebase: firebase})
	require.Empty(t, domain, "dev mode signs in locally")
	require.Empty(t, emulator)
}
