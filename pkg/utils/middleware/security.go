package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
)

// ContentSecurityPolicy allows the CDNs, map tiles and avatar hosts (Gravatar redirects through *.wp.com) in use.
// Scripts are all external files, only inline styles are still used by the templates.
//
// A non empty firebaseAuthDomain also allows the Firebase sign-in: its SDK, the Google identity endpoints
// and the auth handler frame. An emulatorHost allows the local Firebase Auth emulator instead of Google.
func ContentSecurityPolicy(firebaseAuthDomain, emulatorHost string) string {
	scriptSrc := []string{"script-src 'self' https://code.jquery.com https://unpkg.com https://cdn.jsdelivr.net"}
	connectSrc := []string{"connect-src 'self' https://nominatim.openstreetmap.org"}
	frameSrc := []string{"frame-src 'self'"}
	if firebaseAuthDomain != "" {
		scriptSrc = append(scriptSrc, "https://www.gstatic.com", "https://apis.google.com")
		connectSrc = append(connectSrc, "https://identitytoolkit.googleapis.com", "https://securetoken.googleapis.com",
			"https://www.googleapis.com")
		frameSrc = append(frameSrc, "https://"+firebaseAuthDomain, "https://accounts.google.com")
	}

	if emulatorHost != "" {
		connectSrc = append(connectSrc, "http://"+emulatorHost)
		frameSrc = append(frameSrc, "http://"+emulatorHost)
	}

	return strings.Join([]string{
		"default-src 'self'",
		strings.Join(scriptSrc, " "),
		"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://unpkg.com https://cdn.jsdelivr.net https://cdnjs.cloudflare.com",
		"font-src 'self' https://fonts.gstatic.com https://cdnjs.cloudflare.com",
		"img-src 'self' data: https://*.tile.openstreetmap.org https://unpkg.com https://*.gravatar.com https://ui-avatars.com https://*.wp.com https://openweathermap.org",
		strings.Join(connectSrc, " "),
		strings.Join(frameSrc, " "),
		"object-src 'none'",
		"base-uri 'self'",
		"form-action 'self'",
		"frame-ancestors 'none'",
	}, "; ")
}

// SecurityHeaders sets clickjacking, content type, referrer, CSP and HSTS headers on every response.
func SecurityHeaders(firebaseAuthDomain, emulatorHost string) fiber.Handler {
	return helmet.New(helmet.Config{
		XFrameOptions:         "DENY",
		ContentSecurityPolicy: ContentSecurityPolicy(firebaseAuthDomain, emulatorHost),
		HSTSMaxAge:            31536000,
		// OpenStreetMap tile servers require a referer, and the CDN assets are loaded cross-origin
		ReferrerPolicy:            "strict-origin-when-cross-origin",
		CrossOriginEmbedderPolicy: "unsafe-none",
		// The default same-origin cuts the Google sign-in popup off from the page, so it cannot report back
		CrossOriginOpenerPolicy: "same-origin-allow-popups",
	})
}
