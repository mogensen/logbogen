package services

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/mogensen/logbook/pkg/auth"
	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/database"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/mogensen/logbook/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

var ErrNotLoggedIn = fmt.Errorf("user is not logged in")

// TokenVerifier validates a Firebase ID token and returns the identity it carries.
type TokenVerifier interface {
	Verify(ctx context.Context, rawToken string) (*auth.Identity, error)
}

// AuthService handles authentication related operations
type AuthService struct {
	userDal  dal.UserDal
	verifier TokenVerifier // nil in dev mode
	firebase auth.Config
	devMode  bool
}

// NewAuthService creates a new instance of AuthService. verifier may be nil when devMode is true
// (the local dev-login bypass is used instead).
func NewAuthService(userDal dal.UserDal, verifier TokenVerifier, firebase auth.Config, devMode bool) *AuthService {
	return &AuthService{
		userDal:  userDal,
		verifier: verifier,
		firebase: firebase,
		devMode:  devMode,
	}
}

// upsertUser finds an existing user by Firebase UID or creates a new one.
// This is the single identity entry point shared by the Firebase session
// endpoint and the dev-login bypass.
func (s *AuthService) upsertUser(sub, name, email string) (*dal.User, error) {
	user, err := s.userDal.FindUserByFirebaseUID(sub)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	subCopy := sub
	newUser := &dal.User{
		Name:        name,
		Email:       email,
		FirebaseUID: &subCopy,
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

// LoginPageHandler renders the login page. In dev mode it shows the local dev-login form; otherwise it
// hands the public Firebase settings to the page, which signs in with the Firebase JS SDK.
func (s *AuthService) LoginPageHandler(ctx *fiber.Ctx) error {
	return ctx.Render("auth/login", fiber.Map{
		"Title":    "Login",
		"DevMode":  s.devMode,
		"Firebase": s.firebase,
	})
}

// sessionRequest is the body the login page posts after signing in with Firebase.
type sessionRequest struct {
	IDToken string `json:"idToken"`
}

// SessionHandler turns a verified Firebase ID token into a local session: it verifies the token, requires
// a verified email, upserts the user and establishes the session.
func (s *AuthService) SessionHandler(ctx *fiber.Ctx) error {
	if s.devMode || s.verifier == nil {
		return ctx.SendStatus(fiber.StatusNotFound)
	}

	var req sessionRequest
	if err := ctx.BodyParser(&req); err != nil || req.IDToken == "" {
		return sessionError(ctx, fiber.StatusBadRequest, "invalid_request")
	}

	identity, err := s.verifier.Verify(ctx.Context(), req.IDToken)
	if errors.Is(err, auth.ErrStaleLogin) {
		return sessionError(ctx, fiber.StatusUnauthorized, "stale_login")
	}
	if err != nil {
		utils.Log(ctx).Warn("Failed to verify Firebase ID token", "error", err)
		return sessionError(ctx, fiber.StatusUnauthorized, "invalid_token")
	}

	if !identity.EmailVerified {
		return sessionError(ctx, fiber.StatusForbidden, "email_not_verified")
	}

	user, err := s.upsertUser(identity.UID, displayName(identity), identity.Email)
	if err != nil {
		utils.Log(ctx).Error("Failed to upsert user", "error", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	if err := s.establishSession(ctx, user); err != nil {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	utils.LogEvent(ctx, "user_logged_in", map[string]any{"user_id": user.ID, "login_method": "firebase"})
	return ctx.JSON(fiber.Map{"redirect": "/"})
}

// sessionError answers the login page with a code it maps to a message for the user.
func sessionError(ctx *fiber.Ctx, status int, code string) error {
	return ctx.Status(status).JSON(fiber.Map{"error": code})
}

// displayName falls back to the part of the email before the @ for accounts without a name.
func displayName(identity *auth.Identity) string {
	if name := strings.TrimSpace(identity.Name); name != "" {
		return name
	}
	if addr, err := mail.ParseAddress(identity.Email); err == nil {
		return strings.SplitN(addr.Address, "@", 2)[0]
	}
	return identity.Email
}

// DevLoginHandler is the local dev-login bypass (only registered when devMode
// is enabled). It upserts a user keyed by a synthetic "dev|<email>" subject and
// establishes the session, mirroring the Firebase session endpoint.
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
		utils.Log(ctx).Error("Failed to upsert dev user", "error", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	if err := s.establishSession(ctx, user); err != nil {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	utils.LogEvent(ctx, "user_logged_in", map[string]any{"user_id": user.ID, "login_method": "dev"})
	return ctx.Redirect("/")
}

// LogoutHandler destroys the local session and returns the user to the front page.
func (s *AuthService) LogoutHandler(ctx *fiber.Ctx) error {
	session, err := database.SessionStore.Get(ctx)
	if err != nil {
		utils.Log(ctx).Error("Failed to get session store", "error", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	if err := session.Destroy(); err != nil {
		utils.Log(ctx).Error("Failed to destroy session", "error", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	utils.LogEvent(ctx, "user_logged_out", map[string]any{})

	return ctx.Redirect("/")
}

// establishSession resets the session and marks the user as logged in, using
// the same session keys the auth middleware reads.
func (s *AuthService) establishSession(ctx *fiber.Ctx, user *dal.User) error {
	session, err := database.SessionStore.Get(ctx)
	if err != nil {
		utils.Log(ctx).Error("Failed to get session store", "error", err)
		return err
	}

	if err := session.Reset(); err != nil {
		utils.Log(ctx).Error("Failed to reset session", "error", err)
		return err
	}
	session.Set("username", user.Email)
	session.Set("userID", uint64(user.ID))
	session.Set("loggedIn", true)

	if err := session.Save(); err != nil {
		utils.Log(ctx).Error("Failed to save session", "error", err)
		return err
	}
	return nil
}
