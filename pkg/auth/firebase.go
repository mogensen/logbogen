// Package auth verifies the Firebase ID tokens that browsers send after signing in.
// It is only initialized outside dev mode; in dev mode (AUTH_DEV_MODE) the app never constructs it.
package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
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
	// EmulatorHost is the host:port of the Firebase Auth emulator, empty against real Firebase.
	EmulatorHost string
}

// Identity is the subset of Firebase ID-token claims the app consumes.
type Identity struct {
	UID           string
	Email         string
	Name          string
	EmailVerified bool
}

// tokenClaims are the claims read from a Firebase ID token, whatever way it was checked.
type tokenClaims struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	AuthTime      int64  `json:"auth_time"`
}

// Verifier validates Firebase ID tokens against the project's public signing keys.
type Verifier struct {
	check func(ctx context.Context, rawToken string) (*tokenClaims, error)
	now   func() time.Time
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
	verifier := provider.Verifier(&oidc.Config{ClientID: projectID, Now: now})

	check := func(ctx context.Context, rawToken string) (*tokenClaims, error) {
		idToken, err := verifier.Verify(ctx, rawToken)
		if err != nil {
			return nil, err
		}
		var claims tokenClaims
		if err := idToken.Claims(&claims); err != nil {
			return nil, err
		}
		claims.Subject = idToken.Subject
		return &claims, nil
	}
	return &Verifier{check: check, now: now}, nil
}

// NewEmulatorVerifier builds a Verifier for the Firebase Auth emulator, whose tokens are not signed.
// It must never be used against real Firebase, as anyone could then forge a login.
func NewEmulatorVerifier(projectID string) (*Verifier, error) {
	return newEmulatorVerifier(projectID, time.Now)
}

func newEmulatorVerifier(projectID string, now func() time.Time) (*Verifier, error) {
	if projectID == "" {
		return nil, errors.New("firebase: FIREBASE_PROJECT_ID must be set")
	}

	check := func(_ context.Context, rawToken string) (*tokenClaims, error) {
		parts := strings.Split(rawToken, ".")
		if len(parts) != 3 {
			return nil, errors.New("firebase: malformed token")
		}
		payload, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, err
		}

		var claims struct {
			tokenClaims
			Issuer   string `json:"iss"`
			Audience string `json:"aud"`
			Expiry   int64  `json:"exp"`
		}
		if err := json.Unmarshal(payload, &claims); err != nil {
			return nil, err
		}
		if claims.Issuer != firebaseIssuerBase+projectID || claims.Audience != projectID {
			return nil, errors.New("firebase: token is for another project")
		}
		if now().After(time.Unix(claims.Expiry, 0)) {
			return nil, errors.New("firebase: token is expired")
		}
		return &claims.tokenClaims, nil
	}
	return &Verifier{check: check, now: now}, nil
}

// Verify checks the token signature, issuer, audience and expiry, and that the user signed in recently.
func (v *Verifier) Verify(ctx context.Context, rawToken string) (*Identity, error) {
	claims, err := v.check(ctx, rawToken)
	if err != nil {
		return nil, err
	}

	if claims.Subject == "" {
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
		UID:           claims.Subject,
		Email:         claims.Email,
		Name:          claims.Name,
		EmailVerified: claims.EmailVerified,
	}, nil
}
