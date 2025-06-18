package services

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/database"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/mogensen/logbook/pkg/utils"
	"github.com/mogensen/logbook/pkg/utils/password"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

var ErrNotLoggedIn = fmt.Errorf("User is not logged in")

// AuthService handles authentication related operations
type AuthService struct {
	userDal dal.UserDal
}

// NewAuthService creates a new instance of AuthService
func NewAuthService(userDal dal.UserDal) *AuthService {
	return &AuthService{
		userDal: userDal,
	}
}

// LoginRequest represents the login request data
type LoginRequest struct {
	Email    string
	Password string
}

// LoginResponse represents the login response data
type LoginResponse struct {
	UserID   uint64
	Email    string
	LoggedIn bool
}

// Login attempts to log in a user
func (s *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
	u, err := s.userDal.FindUserByEmail(req.Email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("invalid email or password")
	}

	if err := password.Verify(u.Password, req.Password); err != nil {
		return nil, errors.New("invalid email or password")
	}

	return &LoginResponse{
		UserID:   u.ID,
		Email:    u.Email,
		LoggedIn: true,
	}, nil
}

// SignupRequest represents the signup request data
type SignupRequest struct {
	Name     string
	Email    string
	Password string
}

// SignupResponse represents the signup response data
type SignupResponse struct {
	Success bool
	Message string
}

// Signup attempts to create a new user
func (s *AuthService) Signup(req SignupRequest) (*SignupResponse, error) {
	_, err := s.userDal.FindUserByEmail(req.Email)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return &SignupResponse{
			Success: false,
			Message: "Der er already en bruger med denne email",
		}, nil
	}

	user := &dal.User{
		Name:     req.Name,
		Password: password.Generate(req.Password),
		Email:    req.Email,
	}

	if err := s.userDal.CreateUser(user); err.Error != nil {
		return nil, err.Error
	}

	return &SignupResponse{
		Success: true,
		Message: "Brugeren er oprettet, du kan nu logge ind",
	}, nil
}

// GetUsersResponse represents the get users response data
type GetUsersResponse struct {
	Users []*types.UserForLogin
}

// GetUsers returns all users except the current user
func (s *AuthService) GetUsers(currentUserID uint64) (*GetUsersResponse, error) {
	users, err := s.userDal.FindUsers()
	if err != nil {
		return nil, err
	}

	res := make([]*types.UserForLogin, 0, len(users))
	for _, v := range users {
		user := v
		if v.ID == currentUserID {
			continue // Skip the current user
		}

		res = append(res, &types.UserForLogin{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		})
	}

	return &GetUsersResponse{
		Users: res,
	}, nil
}

// GetUserRequest represents the get user request data
type GetUserRequest struct {
	UserID uint64
}

// GetUserResponse represents the get user response data
type GetUserResponse struct {
	User *types.User
}

// GetUser returns a specific user by ID
func (s *AuthService) GetUser(req GetUserRequest) (*GetUserResponse, error) {
	user, err := s.userDal.FindUserById(req.UserID)
	if err != nil {
		return nil, err
	}

	return &GetUserResponse{
		User: types.UserFromDal(user),
	}, nil
}

// GetUserByID retrieves a user by their ID
func (s *AuthService) GetUserByID(userID uint64) (*types.User, error) {
	user, err := s.userDal.FindUserById(userID)
	if err != nil {
		return nil, err
	}

	return types.UserFromDal(user), nil
}

// HTTP Handlers

// LoginHandler handles the login HTTP request
func (s *AuthService) LoginHandler(ctx *fiber.Ctx) error {
	b := new(types.LoginDTO)

	if err := utils.ParseBodyAndValidate(ctx, b); err != nil {
		return ctx.Render("auth/login", fiber.Map{
			"error": err.Message,
		})
	}

	resp, err := s.Login(LoginRequest{
		Email:    b.Email,
		Password: b.Password,
	})
	if err != nil {
		return ctx.Render("auth/login", fiber.Map{
			"error": err.Error(),
		})
	}

	// Set a session variable to mark the user as logged in
	session, err := database.SessionStore.Get(ctx)
	if err != nil {
		slog.Error("Failed to get session store", "error", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	if err := session.Reset(); err != nil {
		slog.Error("Failed to reset session", "error", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}
	session.Set("username", resp.Email)
	session.Set("userID", resp.UserID)
	session.Set("loggedIn", resp.LoggedIn)

	if err := session.Save(); err != nil {
		slog.Error("Failed to save session", "error", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	slog.Info("Logged in", "userID", resp.UserID, "email", resp.Email)
	// Redirect to the home page
	return ctx.Redirect("/")
}

// LogoutHandler handles the logout HTTP request
func (s *AuthService) LogoutHandler(ctx *fiber.Ctx) error {
	session, err := database.SessionStore.Get(ctx)
	if err != nil {
		slog.Error("Failed to get session store", "error", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	if err := session.Destroy(); err != nil {
		slog.Error("Failed to destroy session", "error", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	return ctx.Redirect("/")
}

// SignupPageHandler renders the registration page
func (s *AuthService) SignupPageHandler(ctx *fiber.Ctx) error {
	return ctx.Render("users/register", fiber.Map{
		"Title": "Register",
	})
}

// SignupHandler handles the signup HTTP request
func (s *AuthService) SignupHandler(ctx *fiber.Ctx) error {
	b := new(types.SignupDTO)

	if err := utils.ParseBodyAndValidate(ctx, b); err != nil {
		return err
	}

	resp, err := s.Signup(SignupRequest{
		Name:     b.Name,
		Email:    b.Email,
		Password: b.Password,
	})
	if err != nil {
		return err
	}

	if !resp.Success {
		return ctx.Render("users/register", fiber.Map{
			"error": resp.Message,
		})
	}

	return ctx.Render("auth/login", fiber.Map{
		"info": resp.Message,
	})
}

// LoginPageHandler renders the login page
func (s *AuthService) LoginPageHandler(ctx *fiber.Ctx) error {
	return ctx.Render("auth/login", fiber.Map{
		"Title": "Login",
	})
}
