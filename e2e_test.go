//go:build e2e

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mogensen/logbook/pkg/types"
	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

// These tests drive a real Chrome against the real app, so they catch what the browser enforces and the
// Go tests cannot: content security policy violations, script errors and the JavaScript behavior.
// Run them with `make test-e2e`. They use the installed Chrome and the public CDNs the pages load from.

const e2eTimeout = 20 * time.Second

// e2eServer starts the app on a local port
func e2eServer(t *testing.T, opts ...func(*Config)) string {
	app := newTestApp(t, opts...)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	go func() { _ = app.Listener(listener) }()
	t.Cleanup(func() { _ = app.Shutdown() })

	return "http://" + listener.Addr().String()
}

// e2eBrowser starts Chrome, installing the Playwright driver on first use
func e2eBrowser(t *testing.T) playwright.Browser {
	// The default driver download hosts no longer serve the driver, this one does, but only for the driver of
	// the pinned playwright-go (v0.5001.0). Newer drivers answer 400, so do not upgrade it without checking.
	if os.Getenv("PLAYWRIGHT_DOWNLOAD_HOST") == "" {
		t.Setenv("PLAYWRIGHT_DOWNLOAD_HOST", "https://cdn.playwright.dev/dbazure/download/playwright")
	}

	pw, err := playwright.Run(&playwright.RunOptions{SkipInstallBrowsers: true})
	require.NoError(t, err, "the driver is installed on first use, it needs network access")
	t.Cleanup(func() { _ = pw.Stop() })

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Channel:  playwright.String("chrome"),
		Headless: playwright.Bool(true),
	})
	require.NoError(t, err, "Google Chrome must be installed")
	t.Cleanup(func() { _ = browser.Close() })
	return browser
}

// e2ePage is a browser tab that records what went wrong in it
type e2ePage struct {
	playwright.Page
	t *testing.T

	mu       sync.Mutex
	problems []string
}

func newE2EPage(t *testing.T, browser playwright.Browser) *e2ePage {
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{Width: 1280, Height: 900},
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = context.Close() })

	page, err := context.NewPage()
	require.NoError(t, err)
	page.SetDefaultTimeout(float64(e2eTimeout.Milliseconds()))

	p := &e2ePage{Page: page, t: t}

	// Every content security policy violation is reported, including the ones on redirects
	require.NoError(t, page.AddInitScript(playwright.Script{Content: playwright.String(`
		document.addEventListener('securitypolicyviolation', (e) => {
			console.error('CSP violation: ' + e.violatedDirective + ' blocked ' + e.blockedURI);
		});`)}))
	page.On("console", func(message playwright.ConsoleMessage) {
		text := message.Text()
		if (message.Type() == "error" && strings.Contains(text, "Content Security Policy")) || strings.HasPrefix(text, "CSP violation") {
			p.addProblem("console: " + text)
		}
	})
	page.On("pageerror", func(err error) { p.addProblem("script error: " + err.Error()) })
	return p
}

func (p *e2ePage) addProblem(problem string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.problems = append(p.problems, problem)
}

// requireNoProblems fails when a content security policy violation or script error was seen
func (p *e2ePage) requireNoProblems(what string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	require.Empty(p.t, p.problems, what)
}

// settle waits for the page and its images to finish loading. A request that never finishes is not an
// error here, since the more useful failure is the policy violation or script error that caused it.
func (p *e2ePage) settle() {
	err := p.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		p.t.Logf("the page did not become idle: %v", err)
	}
}

// login signs in through the dev login form
func (p *e2ePage) login(base, name, email string) {
	_, err := p.Goto(base + "/auth/login")
	require.NoError(p.t, err)
	require.NoError(p.t, p.Locator("#name").Fill(name))
	require.NoError(p.t, p.Locator("#email").Fill(email))
	require.NoError(p.t, p.Locator("button.login-btn").Click())
	require.NoError(p.t, p.WaitForURL(base+"/"))
	require.NoError(p.t, p.Locator("[data-post='/auth/logout']").WaitFor())
}

// createActivity logs an activity from the page itself, with the CSRF token the page carries
func (p *e2ePage) createActivity(category, activityType string, daysAgo int) {
	activity := types.Activity{
		Date:       types.Date(time.Now().AddDate(0, 0, -daysAgo)),
		Lat:        55.676098,
		Lng:        12.568337,
		CategoryID: category,
		TypeID:     activityType,
		Role:       "Deltager",
		Comment:    "End to end",
	}
	body, err := json.Marshal(activity)
	require.NoError(p.t, err)

	status, err := p.Evaluate(`async (body) => (await fetch('/activities/create', {
		method: 'POST',
		redirect: 'manual',
		headers: {'Content-Type': 'application/json', 'X-Csrf-Token': window.csrfToken()},
		body,
	})).type`, string(body))
	require.NoError(p.t, err)
	// A created activity redirects to its page, which a manual redirect reports as opaque
	require.Equal(p.t, "opaqueredirect", status)
}

// theme sets the light or dark theme the way the toggle does
func (p *e2ePage) theme(theme string) {
	_, err := p.Evaluate(`(theme) => document.documentElement.setAttribute('data-bs-theme', theme)`, theme)
	require.NoError(p.t, err)
}

func (p *e2ePage) style(selector, property string) string {
	value, err := p.Locator(selector).First().Evaluate(
		`(element, property) => getComputedStyle(element).getPropertyValue(property)`, property)
	require.NoError(p.t, err)
	return fmt.Sprint(value)
}

func TestE2E_PagesLoadWithoutPolicyViolationsOrScriptErrors(t *testing.T) {
	base := e2eServer(t)
	page := newE2EPage(t, e2eBrowser(t))

	page.login(base, "Browser User", "browser@example.com")
	page.createActivity("climbing", "boulder", 3)
	page.createActivity("sailing", "kayak", 10)

	_, err := page.Goto(base + "/activities/list")
	require.NoError(t, err)
	page.settle()
	href, err := page.Locator("ul.contacts li a").First().GetAttribute("href")
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(href, "/activities/"), href)

	for _, path := range []string{
		"/",
		"/activities/list",
		href,
		href + "/edit",
		"/activities/create",
		"/scoreboard",
		"/users/1",
		"/certifications/list",
		"/certifications/create",
	} {
		t.Run(path, func(t *testing.T) {
			_, err := page.Goto(base + path)
			require.NoError(t, err)
			page.settle()

			page.requireNoProblems(path)
		})
	}
}

func TestE2E_LogoutReturnsToTheFrontPage(t *testing.T) {
	base := e2eServer(t)
	page := newE2EPage(t, e2eBrowser(t))
	page.login(base, "Logout User", "logout@example.com")

	require.NoError(t, page.Locator("[data-post='/auth/logout']").Click())

	require.NoError(t, page.WaitForURL(base+"/"))
	require.NoError(t, page.Locator("a[href='/auth/login']").First().WaitFor(), "the front page asks to log in again")
	_, err := page.Goto(base + "/activities/list")
	require.NoError(t, err)
	require.Equal(t, base+"/auth/login", page.URL(), "the session is gone")
	page.requireNoProblems("logout")
}

func TestE2E_CategoryFilter(t *testing.T) {
	base := e2eServer(t)
	page := newE2EPage(t, e2eBrowser(t))
	page.login(base, "Filter User", "filter@example.com")
	page.createActivity("climbing", "boulder", 3)
	page.createActivity("climbing", "rock", 4)
	page.createActivity("sailing", "kayak", 5)

	_, err := page.Goto(base + "/activities/list")
	require.NoError(t, err)
	cards := page.Locator("ul.contacts > li")
	count, err := cards.Count()
	require.NoError(t, err)
	require.Equal(t, 3, count)

	require.NoError(t, page.Locator(".category-filter a", playwright.PageLocatorOptions{HasText: "Klatring"}).Click())

	require.NoError(t, page.WaitForURL(base+"/activities/list?category=climbing"))
	count, err = cards.Count()
	require.NoError(t, err)
	require.Equal(t, 2, count)
	icon, err := page.Locator(".category-filter a.active img").GetAttribute("src")
	require.NoError(t, err)
	require.Equal(t, "/images/categories/climbing.png", icon)

	require.NoError(t, page.Locator(".category-filter a", playwright.PageLocatorOptions{HasText: "Alle"}).Click())
	require.NoError(t, page.WaitForURL(base+"/activities/list"))
	count, err = cards.Count()
	require.NoError(t, err)
	require.Equal(t, 3, count)
}

func TestE2E_Scoreboard(t *testing.T) {
	base := e2eServer(t)
	browser := e2eBrowser(t)

	first := newE2EPage(t, browser)
	first.login(base, "First Place", "first@example.com")
	first.createActivity("climbing", "boulder", 3)
	first.createActivity("climbing", "rock", 4)
	second := newE2EPage(t, browser)
	second.login(base, "Second Place", "second@example.com")
	second.createActivity("sailing", "kayak", 5)

	_, err := first.Goto(base + "/scoreboard")
	require.NoError(t, err)
	first.settle()

	t.Run("places start at one, most points on top", func(t *testing.T) {
		rows, err := first.Locator("tbody tr").AllInnerTexts()
		require.NoError(t, err)
		require.Len(t, rows, 2)
		require.True(t, strings.HasPrefix(rows[0], "1"), rows[0])
		require.Contains(t, rows[0], "First Place")
		require.True(t, strings.HasPrefix(rows[1], "2"), rows[1])
	})

	t.Run("the header and the hovered row follow the theme", func(t *testing.T) {
		const white = "rgb(255, 255, 255)"

		first.theme("dark")
		require.NotEqual(t, white, first.style("thead th", "background-color"), "the dark mode header is not white")
		require.NoError(t, first.Locator("tbody tr").Nth(1).Hover())
		hovered := first.style("tbody tr:nth-child(2) td", "box-shadow")
		resting := first.style("tbody tr:nth-child(1) td", "box-shadow")
		require.NotEqual(t, resting, hovered, "the hover is visible")
		require.NotContains(t, hovered, white, "the dark mode hover is not white")

		first.theme("light")
		require.NoError(t, first.Locator("tbody tr").Nth(1).Hover())
		require.NotEqual(t, white, first.style("tbody tr:nth-child(2) td", "box-shadow"), "the light mode hover is a tint")
		require.NotEqual(t,
			first.style("tbody tr:nth-child(1) td", "box-shadow"), first.style("tbody tr:nth-child(2) td", "box-shadow"),
			"the light mode hover is visible")
	})

	t.Run("clicking anywhere in a row opens the profile", func(t *testing.T) {
		require.NoError(t, first.Locator("tbody tr").Nth(1).Locator("td").Nth(2).Click())

		require.NoError(t, first.WaitForURL(base+"/users/2/"))
		require.Contains(t, mustText(t, first), "Second Place")
	})

	first.requireNoProblems("scoreboard")
}

func TestE2E_ThemeToggleIsRemembered(t *testing.T) {
	base := e2eServer(t)
	page := newE2EPage(t, e2eBrowser(t))
	page.login(base, "Theme User", "theme@example.com")

	before, err := page.Locator("html").GetAttribute("data-bs-theme")
	require.NoError(t, err)
	require.NoError(t, page.Locator("#theme-toggle").Click())
	require.NoError(t, page.Locator("html:not([data-bs-theme='"+before+"'])").WaitFor())
	after, err := page.Locator("html").GetAttribute("data-bs-theme")
	require.NoError(t, err)
	require.NotEqual(t, before, after)

	// The choice is saved on the user, so a new page load keeps it
	_, err = page.Goto(base + "/")
	require.NoError(t, err)
	saved, err := page.Locator("html").GetAttribute("data-bs-theme")
	require.NoError(t, err)
	require.Equal(t, after, saved)
	page.requireNoProblems("theme toggle")
}

func mustText(t *testing.T, page *e2ePage) string {
	text, err := page.Locator("body").InnerText()
	require.NoError(t, err)
	return text
}

// fillActivityForm picks a category and type in the create form
func (p *e2ePage) chooseTypeInTheForm(category, activityType string) {
	require.NoError(p.t, p.Locator(".act-edit-pill", playwright.PageLocatorOptions{
		Has: p.Locator("input[name='category'][value='" + category + "']"),
	}).WaitFor())
	require.NoError(p.t, p.Locator(".act-edit-pill", playwright.PageLocatorOptions{
		Has: p.Locator("input[name='category'][value='" + category + "']"),
	}).Click())
	typePill := p.Locator(".act-edit-pill", playwright.PageLocatorOptions{
		Has: p.Locator("input[name='type'][value='" + activityType + "']"),
	})
	require.NoError(p.t, typePill.WaitFor())
	require.NoError(p.t, typePill.Click())
}

func TestE2E_ActivityFormStopsTypingAtTheLimitAndSaves(t *testing.T) {
	base := e2eServer(t)
	page := newE2EPage(t, e2eBrowser(t))
	page.login(base, "Limit User", "limit@example.com")

	_, err := page.Goto(base + "/activities/create")
	require.NoError(t, err)
	comment := page.Locator("#activity-comment")
	counter := page.Locator("#comment-count")
	require.NoError(t, comment.WaitFor())

	t.Run("the counter follows the typing", func(t *testing.T) {
		require.Equal(t, "0", mustInnerText(t, counter))

		require.NoError(t, comment.PressSequentially("Hej verden"))

		require.Equal(t, "10", mustInnerText(t, counter))
	})

	t.Run("typing stops at the limit and the counter warns", func(t *testing.T) {
		require.NoError(t, comment.Fill(""))
		require.NoError(t, comment.PressSequentially(strings.Repeat("a", 2100)))

		value, err := comment.InputValue()
		require.NoError(t, err)
		require.Len(t, value, 2000, "the browser stops at the maxlength")
		require.Equal(t, "2000", mustInnerText(t, counter))
		class, err := page.Locator(".act-edit-count").GetAttribute("class")
		require.NoError(t, err)
		require.Contains(t, class, "act-edit-count--near")
	})

	t.Run("the other fields stop at their limits too", func(t *testing.T) {
		require.NoError(t, page.Locator("#activity-role").PressSequentially(strings.Repeat("r", 150)))

		value, err := page.Locator("#activity-role").InputValue()
		require.NoError(t, err)
		require.Len(t, value, 100)
	})

	t.Run("clicking a repeated copy of the world still gives a longitude the server accepts", func(t *testing.T) {
		// Zoomed out the world repeats, and the right edge shows the copy east of 180 degrees. The map ignores a
		// zoom click during the animation, so each step waits for the tiles of its new level and for the animation to end.
		for level := 6; level >= 1; level-- {
			require.NoError(t, page.Locator(".leaflet-control-zoom-out").Click())
			require.NoError(t, page.Locator(fmt.Sprintf("#activity-form-map img.leaflet-tile[src*='tile.openstreetmap.org/%d/']", level)).First().WaitFor(
				playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateAttached}))
			require.NoError(t, page.Locator(".leaflet-map-pane:not(.leaflet-zoom-anim)").WaitFor(playwright.LocatorWaitForOptions{
				State: playwright.WaitForSelectorStateAttached,
			}))
		}
		box, err := page.Locator("#activity-form-map").BoundingBox()
		require.NoError(t, err)
		require.NoError(t, page.Locator("#activity-form-map").Click(playwright.LocatorClickOptions{
			Position: &playwright.Position{X: box.Width * 0.95, Y: box.Height / 2},
		}))

		lng, err := page.Locator("#activity-lng").InputValue()
		require.NoError(t, err)
		value, err := strconv.ParseFloat(lng, 64)
		require.NoError(t, err)
		require.GreaterOrEqual(t, value, -180.0, lng)
		require.LessOrEqual(t, value, 180.0, lng)
	})

	t.Run("a comment of many lines fits, and is saved, though the browser submits two characters per line break", func(t *testing.T) {
		require.NoError(t, comment.Fill(""))
		for i := 0; i < 105; i++ {
			require.NoError(t, comment.PressSequentially(strings.Repeat("a", 18)))
			require.NoError(t, page.Keyboard().Press("Enter"))
		}
		value, err := comment.InputValue()
		require.NoError(t, err)
		require.Len(t, value, 1995, "within the limit as the browser counts it")

		page.chooseTypeInTheForm("climbing", "boulder")
		require.NoError(t, page.Locator(".act-edit-header-actions .act-edit-save").Click())

		require.NoError(t, page.WaitForURL(regexp.MustCompile(`/activities/[0-9a-f-]{36}$`)), "saved, not an error page: "+page.URL())
		body, err := page.Locator("body").InnerText()
		require.NoError(t, err)
		require.NotContains(t, body, "Der er sket en fejl")
	})
	page.requireNoProblems("activity form")
}

func mustInnerText(t *testing.T, locator playwright.Locator) string {
	text, err := locator.InnerText()
	require.NoError(t, err)
	return text
}
