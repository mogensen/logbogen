package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/utils"
)

// ThemeMiddleware injects the user's theme preference into the context
func ThemeMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if user is logged in and has a theme preference
		if user := utils.GetUser(c); user != nil {
			// If ThemePref is empty, default to "auto"
			theme := user.ThemePref
			if theme == "" {
				theme = "auto"
			}
			c.Locals("ThemePref", theme)
		} else {
			// Not logged in, use auto (system preference)
			c.Locals("ThemePref", "auto")
		}
		return c.Next()
	}
}
