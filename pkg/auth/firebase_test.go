package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/require"
)

const testProject = "logbogen-test"

var testNow = time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)

// fakeFirebase serves the OIDC discovery document and signing keys for a Firebase project and signs tokens.
type fakeFirebase struct {
	server *httptest.Server
	key    *rsa.PrivateKey
	signer jose.Signer
}

func newFakeFirebase(t *testing.T) *fakeFirebase {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key},
		(&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", "test-key"))
	require.NoError(t, err)

	f := &fakeFirebase{key: key, signer: signer}
	mux := http.NewServeMux()
	mux.HandleFunc("/"+testProject+"/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                f.issuer(),
			"jwks_uri":                              f.server.URL + "/jwks",
			"id_token_signing_alg_values_supported": []string{"RS256"},
		}))
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		jwk := jose.JSONWebKey{Key: &key.PublicKey, KeyID: "test-key", Algorithm: "RS256", Use: "sig"}
		require.NoError(t, json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{jwk}}))
	})
	f.server = httptest.NewServer(mux)
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakeFirebase) issuer() string { return f.server.URL + "/" + testProject }

func (f *fakeFirebase) verifier(t *testing.T) *Verifier {
	v, err := newVerifier(context.Background(), f.issuer(), testProject, func() time.Time { return testNow })
	require.NoError(t, err)
	return v
}

// token signs a valid token and lets the caller change the claims before signing
func (f *fakeFirebase) token(t *testing.T, change func(claims map[string]any)) string {
	claims := map[string]any{
		"iss":            f.issuer(),
		"aud":            testProject,
		"sub":            "uid-1",
		"iat":            testNow.Add(-time.Minute).Unix(),
		"exp":            testNow.Add(time.Hour).Unix(),
		"auth_time":      testNow.Add(-time.Minute).Unix(),
		"email":          "jane@example.com",
		"email_verified": true,
		"name":           "Jane Doe",
	}
	if change != nil {
		change(claims)
	}
	raw, err := jwt.Signed(f.signer).Claims(claims).Serialize()
	require.NoError(t, err)
	return raw
}

func TestVerifier_AcceptsAValidToken(t *testing.T) {
	f := newFakeFirebase(t)

	identity, err := f.verifier(t).Verify(context.Background(), f.token(t, nil))

	require.NoError(t, err)
	require.Equal(t, &Identity{UID: "uid-1", Email: "jane@example.com", Name: "Jane Doe", EmailVerified: true}, identity)
}

func TestVerifier_ReportsAnUnverifiedEmail(t *testing.T) {
	f := newFakeFirebase(t)

	identity, err := f.verifier(t).Verify(context.Background(), f.token(t, func(c map[string]any) {
		c["email_verified"] = false
	}))

	require.NoError(t, err)
	require.False(t, identity.EmailVerified)
}

func TestVerifier_RejectsBadTokens(t *testing.T) {
	f := newFakeFirebase(t)
	v := f.verifier(t)

	tests := map[string]func(c map[string]any){
		"wrong issuer":        func(c map[string]any) { c["iss"] = "https://securetoken.google.com/other" },
		"wrong audience":      func(c map[string]any) { c["aud"] = "other-project" },
		"expired":             func(c map[string]any) { c["exp"] = testNow.Add(-time.Hour).Unix() },
		"no subject":          func(c map[string]any) { delete(c, "sub") },
		"no auth_time":        func(c map[string]any) { delete(c, "auth_time") },
		"auth_time in future": func(c map[string]any) { c["auth_time"] = testNow.Add(time.Hour).Unix() },
	}
	for name, change := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := v.Verify(context.Background(), f.token(t, change))
			require.Error(t, err)
		})
	}
}

func TestVerifier_RejectsAStaleLogin(t *testing.T) {
	f := newFakeFirebase(t)

	_, err := f.verifier(t).Verify(context.Background(), f.token(t, func(c map[string]any) {
		c["auth_time"] = testNow.Add(-time.Hour).Unix()
	}))

	require.ErrorIs(t, err, ErrStaleLogin)
}

func TestVerifier_RejectsATokenSignedWithAnotherKey(t *testing.T) {
	f := newFakeFirebase(t)
	other := newFakeFirebase(t)

	_, err := f.verifier(t).Verify(context.Background(), other.token(t, func(c map[string]any) {
		c["iss"] = f.issuer()
	}))

	require.Error(t, err)
}

func TestVerifier_RejectsGarbage(t *testing.T) {
	f := newFakeFirebase(t)

	_, err := f.verifier(t).Verify(context.Background(), "not-a-token")

	require.Error(t, err)
}

func TestNewVerifier_NeedsAProjectID(t *testing.T) {
	_, err := NewVerifier(context.Background(), "")

	require.ErrorContains(t, err, "FIREBASE_PROJECT_ID")
}
