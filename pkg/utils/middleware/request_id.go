package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mogensen/logbook/pkg/utils"
)

// RequestID gives every request an id, available to the logs through utils.Log and sent back to the
// client as the X-Request-Id header. It runs first so every line of a request, even a panic, carries it.
func RequestID(c *fiber.Ctx) error {
	id := utils.NewRequestID()
	c.Locals(utils.RequestIDLocalsKey, id)
	c.Set("X-Request-Id", id)
	return c.Next()
}
