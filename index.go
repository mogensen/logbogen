package main

import (
	"github.com/gofiber/fiber/v2"
)

func indexPage(c *fiber.Ctx) error {

	// render the root page as HTML
	return c.Render("index", fiber.Map{
		"Title": "Index",
	})
}
