// Package auth verifies the Firebase ID tokens that browsers send after signing in.
// It is only initialized outside dev mode; in dev mode (AUTH_DEV_MODE) the app never constructs it.
package auth

import (
	"context"
	"errors"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
)

const (
	firebaseIssuerBase = "https://securetoken.google.com/"
	// maxAuthAge is how long after the sign-in a Firebase ID token may still be exchanged for a session.
	maxAuthAge = 5 * time.Minute
)

// ErrStaleLogin means the user signed in too long ago for the token to start a session.
var ErrStaleLogin = errors.New("firebase: sign-in is too old")

// Config holds the public Firebase web app settings. None of them are secrets.
type Config struct {
	ProjectID  string
	APIKey     string
	AuthDomain string
}

// Identity is the subset of Firebase ID-token claims the app consumes.
type Identity struct {
	UID           string
	Email         string
	Name          string
	EmailVerified bool
}

// Verifier validates Firebase ID tokens against the project's public signing keys.
type Verifier struct {
	verifier *oidc.IDTokenVerifier
	now      func() time.Time
}

// NewVerifier discovers the signing keys for the Firebase project and builds a Verifier.
func NewVerifier(ctx context.Context, projectID string) (*Verifier, error) {
	return newVerifier(ctx, firebaseIssuerBase+projectID, projectID, time.Now)
}

func newVerifier(ctx context.Context, issuer, projectID string, now func() time.Time) (*Verifier, error) {
	if projectID == "" {
		return nil, errors.New("firebase: FIREBASE_PROJECT_ID must be set")
	}

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}

	return &Verifier{
		verifier: provider.Verifier(&oidc.Config{ClientID: projectID, Now: now}),
		now:      now,
	}, nil
}

// Verify checks the token signature, issuer, audience and expiry, and that the user signed in recently.
func (v *Verifier) Verify(ctx context.Context, rawToken string) (*Identity, error) {
	idToken, err := v.verifier.Verify(ctx, rawToken)
	if err != nil {
		return nil, err
	}

	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		AuthTime      int64  `json:"auth_time"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}

	if idToken.Subject == "" {
		return nil, errors.New("firebase: token has no subject")
	}
	if claims.AuthTime == 0 {
		return nil, errors.New("firebase: token has no auth_time")
	}

	age := v.now().Sub(time.Unix(claims.AuthTime, 0))
	if age > maxAuthAge {
		return nil, ErrStaleLogin
	}
	if age < -time.Minute {
		return nil, errors.New("firebase: auth_time is in the future")
	}

	return &Identity{
		UID:           idToken.Subject,
		Email:         claims.Email,
		Name:          claims.Name,
		EmailVerified: claims.EmailVerified,
	}, nil
}
