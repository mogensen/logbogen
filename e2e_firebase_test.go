//go:build e2e

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/mogensen/logbook/pkg/auth"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// These tests sign in through the real login page against the Firebase Auth emulator. It is started with
// npx (it needs Node and network access) unless one already runs on the port, and its project is a demo one.

const (
	emulatorHost    = "127.0.0.1:9099"
	emulatorProject = "demo-logbogen"
)

func emulatorUp() bool {
	conn, err := net.DialTimeout("tcp", emulatorHost, time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// authEmulator makes sure the Firebase Auth emulator is running
func authEmulator(t *testing.T) {
	if emulatorUp() {
		return
	}

	dir := t.TempDir()
	config := `{"emulators":{"auth":{"host":"127.0.0.1","port":9099},"ui":{"enabled":false}}}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "firebase.json"), []byte(config), 0o644))

	cmd := exec.Command("npx", "-y", "firebase-tools@latest", "emulators:start", "--only", "auth", "--project", emulatorProject)
	cmd.Dir = dir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	require.NoError(t, cmd.Start(), "the Firebase Auth emulator needs Node and npx")
	t.Cleanup(func() {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
		_ = cmd.Wait()
	})

	deadline := time.Now().Add(150 * time.Second)
	for !emulatorUp() {
		require.True(t, time.Now().Before(deadline), "the Firebase Auth emulator did not start")
		time.Sleep(time.Second)
	}
}

// clearEmulatorAccounts forgets the accounts of earlier runs, which the emulator keeps while it runs
func clearEmulatorAccounts(t *testing.T) {
	req, err := http.NewRequest(http.MethodDelete, "http://"+emulatorHost+"/emulator/v1/projects/"+emulatorProject+"/accounts", nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// firebaseServer starts the app the way production runs it, signing in with Firebase against the emulator
func firebaseServer(t *testing.T) string {
	authEmulator(t)
	clearEmulatorAccounts(t)
	return e2eServer(t, func(cfg *Config) {
		cfg.DevMode = false
		cfg.Firebase = auth.Config{
			ProjectID:    emulatorProject,
			APIKey:       "fake-api-key",
			AuthDomain:   emulatorProject + ".firebaseapp.com",
			EmulatorHost: emulatorHost,
		}
	})
}

// verifyEmail follows the verification link the emulator "sent" to the address
func verifyEmail(t *testing.T, email string) {
	resp, err := http.Get("http://" + emulatorHost + "/emulator/v1/projects/" + emulatorProject + "/oobCodes")
	require.NoError(t, err)
	defer resp.Body.Close()

	var codes struct {
		OOBCodes []struct {
			Email       string `json:"email"`
			RequestType string `json:"requestType"`
			OOBLink     string `json:"oobLink"`
		} `json:"oobCodes"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&codes))

	for _, code := range codes.OOBCodes {
		if code.Email == email && code.RequestType == "VERIFY_EMAIL" {
			link, err := http.Get(code.OOBLink)
			require.NoError(t, err)
			_ = link.Body.Close()
			return
		}
	}
	require.Failf(t, "no verification mail", "the emulator sent none to %s", email)
}

// submitEmailForm fills the email form of the login page and submits it
func (p *e2ePage) submitEmailForm(email, password string) {
	require.NoError(p.t, p.Locator("#login-email").Fill(email))
	require.NoError(p.t, p.Locator("#login-password").Fill(password))
	require.NoError(p.t, p.Locator("#email-submit").Click())
}

func (p *e2ePage) requireMessage(contains string) {
	require.NoError(p.t, p.Locator("#login-message", playwright.PageLocatorOptions{HasText: contains}).WaitFor(), contains)
}

func TestE2E_FirebaseEmailSignUpNeedsAVerifiedEmail(t *testing.T) {
	base := firebaseServer(t)
	page := newE2EPage(t, e2eBrowser(t))
	email := fmt.Sprintf("jane-%d@example.com", time.Now().UnixNano())

	_, err := page.Goto(base + "/auth/login")
	require.NoError(t, err)

	// Sign up
	require.NoError(t, page.Locator("#toggle-signup").Click())
	require.NoError(t, page.Locator("#login-name").Fill("Jane Doe"))
	page.submitEmailForm(email, "a-good-password")
	page.requireMessage("bekræftelsesmail")

	t.Run("an unverified email cannot sign in", func(t *testing.T) {
		page.submitEmailForm(email, "a-good-password")

		page.requireMessage("ikke bekræftet")
		require.Contains(t, page.URL(), "/auth/login")
		resp, err := page.Goto(base + "/activities/list")
		require.NoError(t, err)
		require.Contains(t, resp.URL(), "/auth/login", "no session was started")
	})

	t.Run("a wrong password is refused", func(t *testing.T) {
		_, err := page.Goto(base + "/auth/login")
		require.NoError(t, err)

		page.submitEmailForm(email, "not-the-password")

		page.requireMessage("Forkert e-mail eller adgangskode")
	})

	t.Run("a verified email signs in", func(t *testing.T) {
		verifyEmail(t, email)
		_, err := page.Goto(base + "/auth/login")
		require.NoError(t, err)

		page.submitEmailForm(email, "a-good-password")

		require.NoError(t, page.WaitForURL(base+"/"))
		require.NoError(t, page.Locator("[data-post='/auth/logout']").WaitFor(), "the user is logged in")
	})

	t.Run("logging out ends the session", func(t *testing.T) {
		require.NoError(t, page.Locator("[data-post='/auth/logout']").Click())
		require.NoError(t, page.WaitForURL(base+"/"))
		require.NoError(t, page.Locator("a[href='/auth/login']").First().WaitFor(), "the front page offers to log in")
	})

	page.requireNoProblems("firebase email login")
}

func TestE2E_FirebaseCancelledGoogleLoginGoesToTheFrontPage(t *testing.T) {
	base := firebaseServer(t)
	page := newE2EPage(t, e2eBrowser(t))
	_, err := page.Goto(base + "/auth/login")
	require.NoError(t, err)

	popup, err := page.ExpectPopup(func() error {
		return page.Locator("#google-login").Click()
	})
	require.NoError(t, err)
	require.NoError(t, popup.Close())

	require.NoError(t, page.WaitForURL(base+"/"))
	require.NoError(t, page.Locator("a[href='/auth/login']").First().WaitFor(), "the front page offers to log in")
	require.False(t, strings.Contains(page.URL(), "/auth/login"))
	page.requireNoProblems("cancelled google login")
}

func TestE2E_FirebaseGoogleLogin(t *testing.T) {
	base := firebaseServer(t)
	page := newE2EPage(t, e2eBrowser(t))
	_, err := page.Goto(base + "/auth/login")
	require.NoError(t, err)

	// The emulator's popup offers to make up a Google account
	popup, err := page.ExpectPopup(func() error {
		return page.Locator("#google-login").Click()
	})
	require.NoError(t, err)
	// The popup can be shown before its script listens for clicks, so the click is repeated until the form opens
	autogen := popup.Locator("#autogen-button")
	for attempt := 0; ; attempt++ {
		require.NoError(t, popup.Locator("#add-account-button").Click())
		err := autogen.WaitFor(playwright.LocatorWaitForOptions{Timeout: playwright.Float(3000)})
		if err == nil {
			break
		}
		require.Less(t, attempt, 5, "the emulator popup never opened its form: %v", err)
	}
	require.NoError(t, autogen.Click())
	require.NoError(t, popup.Locator("#sign-in").Click())

	require.NoError(t, page.WaitForURL(base+"/"))
	require.NoError(t, page.Locator("[data-post='/auth/logout']").WaitFor(), "the user is logged in")
	page.requireNoProblems("google login")
}
