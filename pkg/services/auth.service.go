package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"

	"github.com/mogensen/logbook/pkg/auth"
	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/database"
	"github.com/mogensen/logbook/pkg/types"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

var ErrNotLoggedIn = fmt.Errorf("User is not logged in")

// AuthService handles authentication related operations
type AuthService struct {
	userDal       dal.UserDal
	authenticator *auth.Authenticator // nil in dev mode
	devMode       bool
}

// NewAuthService creates a new instance of AuthService. authenticator may be
// nil when devMode is true (the local dev-login bypass is used instead).
func NewAuthService(userDal dal.UserDal, authenticator *auth.Authenticator, devMode bool) *AuthService {
	return &AuthService{
		userDal:       userDal,
		authenticator: authenticator,
		devMode:       devMode,
	}
}

// upsertUser finds an existing user by Auth0 subject or creates a new one.
// This is the single identity entry point shared by the Auth0 callback and the
// dev-login bypass.
func (s *AuthService) upsertUser(sub, name, email string) (*dal.User, error) {
	user, err := s.userDal.FindUserByAuth0Sub(sub)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	subCopy := sub
	newUser := &dal.User{
		Name:     name,
		Email:    email,
		Auth0Sub: &subCopy,
	}
	if res := s.userDal.CreateUser(newUser); res.Error != nil {
		return nil, res.Error
	}
	return newUser, nil
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
		if uint64(v.ID) == currentUserID {
			continue // Skip the current user
		}

		res = append(res, &types.UserForLogin{
			ID:    uint64(user.ID),
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

// LoginPageHandler starts a login. In dev mode it renders the local dev-login
// form; otherwise it redirects to Auth0 Universal Login.
func (s *AuthService) LoginPageHandler(ctx *fiber.Ctx) error {
	if s.devMode {
		return ctx.Render("auth/login", fiber.Map{
			"Title":   "Login",
			"DevMode": true,
		})
	}

	state, err := randomToken()
	if err != nil {
		slog.Error("Failed to generate state", "error", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}
	nonce, err := randomToken()
	if err != nil {
		slog.Error("Failed to generate nonce", "error", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	session, err := database.SessionStore.Get(ctx)
	if err != nil {
		slog.Error("Failed to get session store", "error", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}
	session.Set("oauth_state", state)
	session.Set("oauth_nonce", nonce)
	if err := session.Save(); err != nil {
		slog.Error("Failed to save session", "error", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	return ctx.Redirect(s.authenticator.AuthCodeURL(state, nonce))
}

// CallbackHandler completes the Auth0 Authorization Code flow: it verifies the
// state, exchanges the code, verifies the ID token, upserts the user and
// establishes the local session.
func (s *AuthService) CallbackHandler(ctx *fiber.Ctx) error {
	session, err := database.SessionStore.Get(ctx)
	if err != nil {
		slog.Error("Failed to get session store", "error", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	expectedState, _ := session.Get("oauth_state").(string)
	expectedNonce, _ := session.Get("oauth_nonce").(string)
	if expectedState == "" || ctx.Query("state") != expectedState {
		slog.Warn("Invalid OAuth state on callback")
		return ctx.Status(fiber.StatusBadRequest).SendString("Invalid state")
	}

	token, err := s.authenticator.Exchange(ctx.Context(), ctx.Query("code"))
	if err != nil {
		slog.Error("Failed to exchange auth code", "error", err)
		return ctx.Status(fiber.StatusUnauthorized).SendString("Authentication failed")
	}

	claims, err := s.authenticator.VerifyIDToken(ctx.Context(), token, expectedNonce)
	if err != nil {
		slog.Error("Failed to verify ID token", "error", err)
		return ctx.Status(fiber.StatusUnauthorized).SendString("Authentication failed")
	}

	user, err := s.upsertUser(claims.Sub, claims.Name, claims.Email)
	if err != nil {
		slog.Error("Failed to upsert user", "error", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	if err := s.establishSession(ctx, user); err != nil {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	slog.Info("Logged in via Auth0", "userID", user.ID, "email", user.Email)
	return ctx.Redirect("/")
}

// DevLoginHandler is the local dev-login bypass (only registered when devMode
// is enabled). It upserts a user keyed by a synthetic "dev|<email>" subject and
// establishes the session, mirroring the Auth0 callback.
func (s *AuthService) DevLoginHandler(ctx *fiber.Ctx) error {
	if !s.devMode {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	email := ctx.FormValue("email")
	name := ctx.FormValue("name")
	if email == "" {
		return ctx.Render("auth/login", fiber.Map{
			"Title":   "Login",
			"DevMode": true,
			"error":   "Email is required",
		})
	}

	user, err := s.upsertUser("dev|"+email, name, email)
	if err != nil {
		slog.Error("Failed to upsert dev user", "error", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	if err := s.establishSession(ctx, user); err != nil {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	slog.Info("Logged in via dev bypass", "userID", user.ID, "email", user.Email)
	return ctx.Redirect("/")
}

// LogoutHandler destroys the local session. With Auth0 it also redirects to the
// RP-initiated logout endpoint so the Auth0 session is cleared too.
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

	if !s.devMode && s.authenticator != nil {
		return ctx.Redirect(s.authenticator.LogoutURL())
	}
	return ctx.Redirect("/")
}

// establishSession resets the session and marks the user as logged in, using
// the same session keys the auth middleware reads.
func (s *AuthService) establishSession(ctx *fiber.Ctx, user *dal.User) error {
	session, err := database.SessionStore.Get(ctx)
	if err != nil {
		slog.Error("Failed to get session store", "error", err)
		return err
	}

	if err := session.Reset(); err != nil {
		slog.Error("Failed to reset session", "error", err)
		return err
	}
	session.Set("username", user.Email)
	session.Set("userID", uint64(user.ID))
	session.Set("loggedIn", true)

	if err := session.Save(); err != nil {
		slog.Error("Failed to save session", "error", err)
		return err
	}
	return nil
}

// randomToken returns a URL-safe random token for OAuth state/nonce values.
func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
