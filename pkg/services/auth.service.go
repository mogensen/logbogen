package services

import (
	"errors"
	"strconv"

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
	b := new(types.LoginDTO)

	if err := utils.ParseBodyAndValidate(ctx, b); err != nil {
		return ctx.Render("index", fiber.Map{
			"csrf":  utils.GetCsrf(ctx),
			"error": err.Message,
		})
	}

	u := &types.UserResponse{}

	err := dal.FindUserByEmail(u, b.Email).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ctx.Render("index", fiber.Map{
			"csrf":  utils.GetCsrf(ctx),
			"error": "Invalid email or password",
		})
	}

	if err := password.Verify(u.Password, b.Password); err != nil {
		return ctx.Render("index", fiber.Map{
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
	return ctx.Render("users/register", fiber.Map{
		"Title": "Register",
		"csrf":  utils.GetCsrf(ctx),
	})
}

// Signup service creates a user
func Signup(ctx *fiber.Ctx) error {
	b := new(types.SignupDTO)

	if err := utils.ParseBodyAndValidate(ctx, b); err != nil {
		return err
	}

	err := dal.FindUserByEmail(&struct{ ID string }{}, b.Email).Error

	// If email already exists, return
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fiber.NewError(fiber.StatusConflict, "Email already exists")
	}

	user := &dal.User{
		Name:     b.Name,
		Password: password.Generate(b.Password),
		Email:    b.Email,
	}

	// Create a user, if error return
	if err := dal.CreateUser(user); err.Error != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error.Error())
	}

	return ctx.Render("login", fiber.Map{
		"csrf":  utils.GetCsrf(ctx),
		"error": "User created, please login",
	})
}

func GetUsers(ctx *fiber.Ctx) error {
	users := &[]types.UserResponse{}

	err := dal.FindUsers(users).Error
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return ctx.JSON(users)
}

func GetUser(ctx *fiber.Ctx) error {
	userIdParam := ctx.Params("UserID")

	if userIdParam == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid user")
	}

	// Parse the user ID from the URL parameter
	userId, err := strconv.ParseUint(userIdParam, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "Invalid user")
	}

	user, err := GetUserByID(userId)
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return ctx.Render("users/user", fiber.Map{
		"User": user,
	})
}

func GetUserByID(userId uint64) (*types.User, error) {
	user := &dal.User{}

	err := dal.FindUserById(user, userId).Error
	if err != nil {
		return nil, fiber.NewError(fiber.StatusConflict, err.Error())
	}

	activies, err := GetActivitiesForUser(user.ID)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusConflict, err.Error())
	}

	return types.UserFromDal(user, Archivements(activies)), nil
}
