package routes

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// DevRoutes exposes endpoints that only exist to exercise the app during development. Register in dev mode only.
func DevRoutes(app fiber.Router) {
	// /errors/:code previews the error page for any HTTP status without reproducing a real failure
	app.Get("/errors/:code", func(c *fiber.Ctx) error {
		code, err := strconv.Atoi(c.Params("code"))
		if err != nil || code < 400 || code > 599 {
			return fiber.NewError(fiber.StatusBadRequest, "Fejlkoden skal være mellem 400 og 599")
		}
		return fiber.NewError(code, "Testfejl udløst manuelt via /errors/"+c.Params("code"))
	})
}
