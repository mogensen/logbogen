package services

import (
	"errors"

	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/database"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/mogensen/logbook/pkg/utils"
	"github.com/mogensen/logbook/pkg/utils/password"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Login service logs in a user
func Login(ctx *fiber.Ctx) error {
	// Retrieve the submitted form data
	uName := ctx.FormValue("username")
	uPass := ctx.FormValue("password")

	u := &types.UserResponse{}

	err := dal.FindUserByEmail(u, uName).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ctx.Render("index", fiber.Map{
			"Title": "Login",
			"csrf":  utils.GetCsrf(ctx),
			"error": "Invalid email or password",
		})
	}

	if err := password.Verify(u.Password, uPass); err != nil {
		return ctx.Render("index", fiber.Map{
			"Title": "Login",
			"csrf":  utils.GetCsrf(ctx),
			"error": "Invalid email or password",
		})
	}

	// Set a session variable to mark the user as logged in
	session, err := database.SessionStore.Get(ctx)
	if err != nil {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}
	if err := session.Reset(); err != nil {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}
	session.Set("username", u.Email)
	session.Set("userID", u.ID)
	session.Set("loggedIn", true)

	if err := session.Save(); err != nil {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	// Redirect to the home page
	return ctx.Redirect("/")
}

func Logout(ctx *fiber.Ctx) error {
	// Retrieve the session
	session, err := database.SessionStore.Get(ctx)
	if err != nil {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	// Revoke users authentication
	if err := session.Destroy(); err != nil {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	// Redirect to the home page
	return ctx.Redirect("/")
}

func SignupPage(ctx *fiber.Ctx) error {
	// Render the registration page
	return ctx.Render("register", fiber.Map{
		"Title": "Register",
		"csrf":  utils.GetCsrf(ctx),
	})
}

// Signup service creates a user
func Signup(ctx *fiber.Ctx) error {
	// Retrieve the submitted form data
	uEmail := ctx.FormValue("username")
	uPass := ctx.FormValue("password")
	uName := ctx.FormValue("name")

	err := dal.FindUserByEmail(&struct{ ID string }{}, uEmail).Error

	// If email already exists, return
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fiber.NewError(fiber.StatusConflict, "Email already exists")
	}

	user := &dal.User{
		Name:     uName,
		Password: password.Generate(uPass),
		Email:    uEmail,
	}

	// Create a user, if error return
	if err := dal.CreateUser(user); err.Error != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error.Error())
	}

	return ctx.Render("login", fiber.Map{
		"Title": "Login",
		"csrf":  utils.GetCsrf(ctx),
		"error": "User created, please login",
	})
}
