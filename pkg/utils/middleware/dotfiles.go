package middleware

import (
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// BlockDotfiles answers 404 for any path with a segment starting with a dot (.DS_Store, .env, .git/...).
// It checks the decoded path too, so %2e cannot be used to get past it.
func BlockDotfiles(c *fiber.Ctx) error {
	path := c.Path()
	if decoded, err := url.PathUnescape(path); err == nil {
		path = decoded
	}
	for segment := range strings.SplitSeq(path, "/") {
		if strings.HasPrefix(segment, ".") {
			return fiber.NewError(fiber.StatusNotFound, "Not found")
		}
	}
	return c.Next()
}
