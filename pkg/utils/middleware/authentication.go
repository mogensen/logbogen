package middleware

import (
	"github.com/mogensen/logbook/pkg/database"

	"github.com/gofiber/fiber/v2"
)

// Auth is the authentication middleware
func Auth(c *fiber.Ctx) error {
	session, err := database.SessionStore.Get(c)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	loggedIn, _ := session.Get("loggedIn").(bool)
	if !loggedIn {
		// User is not authenticated, redirect to the login page
		return c.Redirect("/login")
	}

	username, _ := session.Get("username").(string)
	userID, _ := session.Get("userID").(uint)

	c.Locals("LoggedIn", loggedIn)
	c.Locals("UserName", username)
	c.Locals("USER", userID)

	return c.Next()
}

// User adds the user information to the context
func User(c *fiber.Ctx) error {
	session, err := database.SessionStore.Get(c)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	loggedIn, _ := session.Get("loggedIn").(bool)
	if !loggedIn {
		return c.Next()
	}

	username, _ := session.Get("username").(string)
	userID, _ := session.Get("userID").(uint)

	c.Locals("LoggedIn", loggedIn)
	c.Locals("UserName", username)
	c.Locals("USER", userID)

	return c.Next()
}
