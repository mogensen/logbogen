package middleware

import (
	"fmt"

	"github.com/mogensen/logbook/pkg/database"
	"github.com/mogensen/logbook/pkg/services"
	"github.com/mogensen/logbook/pkg/types"

	"github.com/gofiber/fiber/v2"
)

var ErrNotLoggedIn = fmt.Errorf("User is not logged in")

// Auth is the authentication middleware
func Auth(c *fiber.Ctx) error {
	user, err := getCurrentUser(c)
	if err != nil {
		if err == ErrNotLoggedIn {
			return c.Redirect("/")
		}
		return err
	}

	c.Locals("LoggedIn", true)
	c.Locals("UserName", user.Email)
	c.Locals("USER", user)

	return c.Next()
}

// User adds the user information to the context
func User(c *fiber.Ctx) error {
	user, err := getCurrentUser(c)
	if err != nil {
		return c.Next() // User is not logged in, continue to next middleware
	}

	c.Locals("LoggedIn", true)
	c.Locals("UserName", user.Email)
	c.Locals("USER", user)

	return c.Next()
}

func getCurrentUser(c *fiber.Ctx) (*types.User, error) {
	session, err := database.SessionStore.Get(c)
	if err != nil {
		return nil, fiber.ErrInternalServerError
	}

	loggedIn, _ := session.Get("loggedIn").(bool)
	if !loggedIn {
		// User is not authenticated, redirect to the login page
		return nil, ErrNotLoggedIn
	}

	userID, _ := session.Get("userID").(uint64)

	return services.GetUserByID(userID)
}
