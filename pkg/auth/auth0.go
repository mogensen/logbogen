// Package auth wraps the Auth0 OIDC Authorization Code flow used to
// authenticate users. It is only initialized when the app runs against a real
// Auth0 tenant; in dev mode (AUTH_DEV_MODE) the app never constructs it.
package auth

import (
	"context"
	"errors"
	"net/url"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// Config holds the Auth0 application settings needed to drive the OIDC flow.
type Config struct {
	Domain          string
	ClientID        string
	ClientSecret    string
	CallbackURL     string
	LogoutReturnURL string
}

// Authenticator bundles the OIDC provider, OAuth2 config and ID-token verifier.
type Authenticator struct {
	cfg      Config
	oauth    oauth2.Config
	verifier *oidc.IDTokenVerifier
}

// Claims is the subset of ID-token claims the app consumes.
type Claims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// New discovers the Auth0 OIDC endpoints and builds an Authenticator.
func New(ctx context.Context, cfg Config) (*Authenticator, error) {
	if cfg.Domain == "" || cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.CallbackURL == "" {
		return nil, errors.New("auth0: AUTH0_DOMAIN, AUTH0_CLIENT_ID, AUTH0_CLIENT_SECRET and AUTH0_CALLBACK_URL must be set")
	}

	provider, err := oidc.NewProvider(ctx, "https://"+cfg.Domain+"/")
	if err != nil {
		return nil, err
	}

	oauthConfig := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.CallbackURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &Authenticator{
		cfg:      cfg,
		oauth:    oauthConfig,
		verifier: provider.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
	}, nil
}

// AuthCodeURL builds the Auth0 authorize URL, binding the login to a CSRF
// state value and an ID-token nonce.
func (a *Authenticator) AuthCodeURL(state, nonce string) string {
	return a.oauth.AuthCodeURL(state, oidc.Nonce(nonce))
}

// Exchange trades the authorization code for tokens.
func (a *Authenticator) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return a.oauth.Exchange(ctx, code)
}

// VerifyIDToken validates the ID token attached to the token response and
// returns the parsed claims. It checks the token signature, audience and the
// nonce bound at login time.
func (a *Authenticator) VerifyIDToken(ctx context.Context, token *oauth2.Token, expectedNonce string) (*Claims, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("auth0: no id_token field in oauth2 token")
	}

	idToken, err := a.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, err
	}

	if idToken.Nonce != expectedNonce {
		return nil, errors.New("auth0: id token nonce did not match")
	}

	var claims Claims
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}
	return &claims, nil
}

// LogoutURL builds the Auth0 RP-initiated logout URL that clears the Auth0
// session and returns the user to LogoutReturnURL.
func (a *Authenticator) LogoutURL() string {
	logout := "https://" + a.cfg.Domain + "/v2/logout"
	q := url.Values{}
	q.Set("client_id", a.cfg.ClientID)
	if a.cfg.LogoutReturnURL != "" {
		q.Set("returnTo", a.cfg.LogoutReturnURL)
	}
	return logout + "?" + q.Encode()
}
