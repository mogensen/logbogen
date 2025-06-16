package main

import (
	"github.com/gofiber/fiber/v2"
)

func indexPage(c *fiber.Ctx) error {
	return c.Render("home/index", fiber.Map{})
}
