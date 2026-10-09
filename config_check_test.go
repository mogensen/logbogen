package main

import (
	"io"
	"log/slog"
	"testing"

	"github.com/mogensen/logbook/pkg/auth"
	"github.com/stretchr/testify/require"
)

func validAuth0() auth.Config {
	return auth.Config{
		Domain:          "tenant.eu.auth0.com",
		ClientID:        "id",
		ClientSecret:    "secret",
		CallbackURL:     "https://logbogen.example.com/auth/callback",
		LogoutReturnURL: "https://logbogen.example.com/",
	}
}

func TestValidateAuthConfig(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name         string
		devMode      bool
		listenAddr   string
		allowDevMode bool
		auth0        func() auth.Config
		wantErr      string
	}{
		{name: "dev mode on loopback", devMode: true, listenAddr: "127.0.0.1:3000", auth0: func() auth.Config { return auth.Config{} }},
		{name: "dev mode on localhost", devMode: true, listenAddr: "localhost:3000", auth0: func() auth.Config { return auth.Config{} }},
		{name: "dev mode on all interfaces", devMode: true, listenAddr: ":3000", auth0: func() auth.Config { return auth.Config{} }, wantErr: "AUTH_DEV_MODE"},
		{name: "dev mode on a public address", devMode: true, listenAddr: "0.0.0.0:3000", auth0: func() auth.Config { return auth.Config{} }, wantErr: "AUTH_DEV_MODE"},
		{name: "dev mode override", devMode: true, listenAddr: "0.0.0.0:3000", allowDevMode: true, auth0: func() auth.Config { return auth.Config{} }},
		{name: "dev mode with a bad address", devMode: true, listenAddr: "nonsense", auth0: func() auth.Config { return auth.Config{} }, wantErr: "invalid listen address"},
		{name: "complete production settings", listenAddr: "127.0.0.1:3000", auth0: validAuth0},
		{name: "localhost callback over http", listenAddr: "127.0.0.1:3000", auth0: func() auth.Config {
			a := validAuth0()
			a.CallbackURL = "http://localhost:3000/auth/callback"
			a.LogoutReturnURL = "http://localhost:3000/"
			return a
		}},
		{name: "missing client secret", listenAddr: "127.0.0.1:3000", auth0: func() auth.Config {
			a := validAuth0()
			a.ClientSecret = ""
			return a
		}, wantErr: "AUTH0_CLIENT_SECRET"},
		{name: "no auth0 settings", listenAddr: "127.0.0.1:3000", auth0: func() auth.Config { return auth.Config{} }, wantErr: "missing required settings"},
		{name: "http callback on a public host", listenAddr: "127.0.0.1:3000", auth0: func() auth.Config {
			a := validAuth0()
			a.CallbackURL = "http://logbogen.example.com/auth/callback"
			return a
		}, wantErr: "AUTH0_CALLBACK_URL must use https"},
		{name: "callback that is not a url", listenAddr: "127.0.0.1:3000", auth0: func() auth.Config {
			a := validAuth0()
			a.LogoutReturnURL = "not a url"
			return a
		}, wantErr: "AUTH0_LOGOUT_RETURN_URL is not a valid URL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{DevMode: tt.devMode, ListenAddr: tt.listenAddr, Logger: logger}
			err := validateAuthConfig(cfg, tt.auth0(), tt.allowDevMode)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestFormActionOrigins(t *testing.T) {
	require.Equal(t, []string{"https://tenant.eu.auth0.com"}, formActionOrigins(false, "tenant.eu.auth0.com"))
	require.Empty(t, formActionOrigins(true, "tenant.eu.auth0.com"), "dev mode logs out locally")
	require.Empty(t, formActionOrigins(false, ""))
}
