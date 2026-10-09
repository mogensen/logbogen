package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
)

// ContentSecurityPolicy allows the CDNs, map tiles and avatar hosts (Gravatar redirects through *.wp.com) in use.
// Inline scripts and styles are used by the templates, so 'unsafe-inline' is still required.
//
// Browsers also check the redirects after a form post against form-action, so formActionOrigins must list
// the origins a form post may redirect to, like the Auth0 logout that follows the logout form.
func ContentSecurityPolicy(formActionOrigins ...string) string {
	return strings.Join([]string{
		"default-src 'self'",
		"script-src 'self' 'unsafe-inline' https://code.jquery.com https://unpkg.com https://cdn.jsdelivr.net",
		"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://unpkg.com https://cdn.jsdelivr.net https://cdnjs.cloudflare.com",
		"font-src 'self' https://fonts.gstatic.com https://cdnjs.cloudflare.com",
		"img-src 'self' data: https://*.tile.openstreetmap.org https://unpkg.com https://*.gravatar.com https://ui-avatars.com https://*.wp.com https://openweathermap.org",
		"connect-src 'self' https://nominatim.openstreetmap.org",
		"object-src 'none'",
		"base-uri 'self'",
		strings.Join(append([]string{"form-action 'self'"}, formActionOrigins...), " "),
		"frame-ancestors 'none'",
	}, "; ")
}

// SecurityHeaders sets clickjacking, content type, referrer, CSP and HSTS headers on every response.
func SecurityHeaders(formActionOrigins ...string) fiber.Handler {
	return helmet.New(helmet.Config{
		XFrameOptions:         "DENY",
		ContentSecurityPolicy: ContentSecurityPolicy(formActionOrigins...),
		HSTSMaxAge:            31536000,
		// OpenStreetMap tile servers require a referer, and the CDN assets are loaded cross-origin
		ReferrerPolicy:            "strict-origin-when-cross-origin",
		CrossOriginEmbedderPolicy: "unsafe-none",
	})
}
