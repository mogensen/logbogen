package middleware

import (
	"fmt"
	"log/slog"

	"github.com/mogensen/logbook/pkg/database"
	"github.com/mogensen/logbook/pkg/services"
	"github.com/mogensen/logbook/pkg/types"

	"github.com/gofiber/fiber/v2"
)

var ErrNotLoggedIn = fmt.Errorf("User is not logged in")

type AuthMiddleware struct {
	authService *services.AuthService
}

func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

// Auth is the authentication middleware
func (a *AuthMiddleware) Auth(c *fiber.Ctx) error {
	user, err := a.getCurrentUser(c, a.authService)
	if err != nil {
		if err == ErrNotLoggedIn {
			return c.Redirect("/auth/login")
		}
		return err
	}

	c.Locals("LoggedIn", true)
	c.Locals("UserName", user.Email)
	c.Locals("USER", user)

	return c.Next()
}

// User adds the user information to the context
func (a *AuthMiddleware) User(c *fiber.Ctx) error {
	user, err := a.getCurrentUser(c, a.authService)
	if err != nil {
		return c.Next() // User is not logged in, continue to next middleware
	}

	c.Locals("LoggedIn", true)
	c.Locals("UserName", user.Email)
	c.Locals("USER", user)

	return c.Next()
}

func (a *AuthMiddleware) getCurrentUser(c *fiber.Ctx, authService *services.AuthService) (*types.User, error) {
	session, err := database.SessionStore.Get(c)
	if err != nil {
		slog.Error("Failed to get session store", "error", err)
		return nil, fiber.ErrInternalServerError
	}

	loggedIn, _ := session.Get("loggedIn").(bool)
	if !loggedIn {
		// User is not authenticated, redirect to the login page
		return nil, ErrNotLoggedIn
	}

	userID, _ := session.Get("userID").(uint64)

	return authService.GetUserByID(userID)
}
