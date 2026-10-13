package services

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/mogensen/logbook/pkg/auth"
	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/mocks"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAuthService_upsertUser(t *testing.T) {
	t.Run("returns existing user when the Firebase UID is known", func(t *testing.T) {
		userDalMock := &mocks.UserDalMock{}
		existing := &dal.User{
			Model: gorm.Model{ID: 7},
			Name:  "Existing User",
			Email: "existing@example.com",
		}
		userDalMock.On("FindUserByFirebaseUID", "uid-123").Return(existing, nil)

		service := NewAuthService(userDalMock, nil, auth.Config{}, true)
		user, err := service.upsertUser("uid-123", "Existing User", "existing@example.com")

		assert.NoError(t, err)
		assert.Equal(t, existing, user)
		// No user is created when one already exists.
		userDalMock.AssertNotCalled(t, "CreateUser", mock.Anything)
		userDalMock.AssertExpectations(t)
	})

	t.Run("creates a new user when the Firebase UID is unknown", func(t *testing.T) {
		userDalMock := &mocks.UserDalMock{}
		userDalMock.On("FindUserByFirebaseUID", "uid-new").Return(nil, gorm.ErrRecordNotFound)
		userDalMock.On("CreateUser", mock.Anything).Return(&gorm.DB{Error: nil})

		service := NewAuthService(userDalMock, nil, auth.Config{}, true)
		user, err := service.upsertUser("uid-new", "New User", "new@example.com")

		assert.NoError(t, err)
		assert.Equal(t, "New User", user.Name)
		assert.Equal(t, "new@example.com", user.Email)
		assert.NotNil(t, user.FirebaseUID)
		assert.Equal(t, "uid-new", *user.FirebaseUID)
		userDalMock.AssertExpectations(t)
	})
}

func TestAuthService_GetUser(t *testing.T) {
	tests := []struct {
		testName    string
		req         GetUserRequest
		mockSetup   func(*mocks.UserDalMock)
		expectedErr error
		expected    *GetUserResponse
	}{
		{
			testName: "successful get user",
			req: GetUserRequest{
				UserID: 1,
			},
			mockSetup: func(m *mocks.UserDalMock) {
				m.On("FindUserById", uint64(1)).Return(&dal.User{
					Model: gorm.Model{ID: 1},
					Name:  "Test User",
					Email: "test@example.com",
				}, nil)
			},
			expectedErr: nil,
			expected: &GetUserResponse{
				User: &types.User{
					ID:    1,
					Name:  "Test User",
					Email: "test@example.com",
				},
			},
		},
		{
			testName: "user not found",
			req: GetUserRequest{
				UserID: 999,
			},
			mockSetup: func(m *mocks.UserDalMock) {
				m.On("FindUserById", uint64(999)).Return(nil, gorm.ErrRecordNotFound)
			},
			expectedErr: gorm.ErrRecordNotFound,
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			userDalMock := &mocks.UserDalMock{}
			tt.mockSetup(userDalMock)

			service := NewAuthService(userDalMock, nil, auth.Config{}, true)
			resp, err := service.GetUser(tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tt.expected.User.ID, resp.User.ID)
				assert.EqualValues(t, tt.expected.User.Name, resp.User.Name)
				assert.EqualValues(t, tt.expected.User.Email, resp.User.Email)
			}

			userDalMock.AssertExpectations(t)
		})
	}
}

func TestGetUserByID(t *testing.T) {
	// Setup
	userDalMock := new(mocks.UserDalMock)
	authService := NewAuthService(userDalMock, nil, auth.Config{}, true)

	// Test cases
	tests := []struct {
		name    string
		userID  uint64
		mock    func()
		want    *types.User
		wantErr bool
	}{
		{
			name:   "successful user retrieval",
			userID: 1,
			mock: func() {
				userDalMock.On("FindUserById", uint64(1)).
					Return(&dal.User{
						Model: gorm.Model{ID: 1},
						Name:  "Test User",
						Email: "test@example.com",
					}, nil)
			},
			wantErr: false,
		},
		{
			name:   "user not found",
			userID: 999,
			mock: func() {
				userDalMock.On("FindUserById", uint64(999)).
					Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			user, err := authService.GetUserByID(tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
			}
		})
	}
}

// fakeVerifier answers Verify with a fixed identity or error
type fakeVerifier struct {
	identity *auth.Identity
	err      error
}

func (f fakeVerifier) Verify(context.Context, string) (*auth.Identity, error) {
	return f.identity, f.err
}

// postSession posts an ID token to the session endpoint of a service built for production
func postSession(t *testing.T, service *AuthService, body string, cookies []*http.Cookie) *http.Response {
	app := fiber.New()
	app.Post("/auth/session", service.SessionHandler)

	req := httptest.NewRequest("POST", "/auth/session", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	return resp
}

func errorCode(t *testing.T, resp *http.Response) string {
	var body struct {
		Error string `json:"error"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	return body.Error
}

func TestAuthService_SessionHandler(t *testing.T) {
	verified := &auth.Identity{UID: "uid-1", Email: "jane@example.com", Name: "Jane Doe", EmailVerified: true}
	token := `{"idToken":"signed-by-firebase"}`

	t.Run("a verified login creates the user and the session", func(t *testing.T) {
		userDalMock := &mocks.UserDalMock{}
		userDalMock.On("FindUserByFirebaseUID", "uid-1").Return(nil, gorm.ErrRecordNotFound)
		userDalMock.On("CreateUser", mock.MatchedBy(func(u *dal.User) bool {
			return u.Name == "Jane Doe" && u.Email == "jane@example.com" && *u.FirebaseUID == "uid-1"
		})).Return(&gorm.DB{})

		resp := postSession(t, NewAuthService(userDalMock, fakeVerifier{identity: verified}, auth.Config{}, false), token, nil)

		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		var body struct {
			Redirect string `json:"redirect"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		require.Equal(t, "/", body.Redirect)
		require.NotEmpty(t, resp.Cookies(), "the session cookie is set")
		userDalMock.AssertExpectations(t)
	})

	t.Run("an existing user is reused", func(t *testing.T) {
		userDalMock := &mocks.UserDalMock{}
		userDalMock.On("FindUserByFirebaseUID", "uid-1").Return(&dal.User{Model: gorm.Model{ID: 3}, Email: "jane@example.com"}, nil)

		resp := postSession(t, NewAuthService(userDalMock, fakeVerifier{identity: verified}, auth.Config{}, false), token, nil)

		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		userDalMock.AssertNotCalled(t, "CreateUser", mock.Anything)
	})

	t.Run("an unverified email is refused without creating a user", func(t *testing.T) {
		userDalMock := &mocks.UserDalMock{}
		unverified := &auth.Identity{UID: "uid-2", Email: "x@example.com"}

		resp := postSession(t, NewAuthService(userDalMock, fakeVerifier{identity: unverified}, auth.Config{}, false), token, nil)

		require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
		require.Equal(t, "email_not_verified", errorCode(t, resp))
		require.Empty(t, resp.Cookies())
		userDalMock.AssertNotCalled(t, "FindUserByFirebaseUID", mock.Anything)
	})

	t.Run("a token that does not verify is refused", func(t *testing.T) {
		service := NewAuthService(&mocks.UserDalMock{}, fakeVerifier{err: errors.New("bad signature")}, auth.Config{}, false)

		resp := postSession(t, service, token, nil)

		require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
		require.Equal(t, "invalid_token", errorCode(t, resp))
		require.Empty(t, resp.Cookies())
	})

	t.Run("a stale login asks the user to sign in again", func(t *testing.T) {
		service := NewAuthService(&mocks.UserDalMock{}, fakeVerifier{err: auth.ErrStaleLogin}, auth.Config{}, false)

		resp := postSession(t, service, token, nil)

		require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
		require.Equal(t, "stale_login", errorCode(t, resp))
	})

	t.Run("a request without a token is refused", func(t *testing.T) {
		service := NewAuthService(&mocks.UserDalMock{}, fakeVerifier{identity: verified}, auth.Config{}, false)

		for _, body := range []string{`{}`, `{"idToken":""}`, `not json`} {
			resp := postSession(t, service, body, nil)
			require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, body)
			require.Equal(t, "invalid_request", errorCode(t, resp), body)
		}
	})

	t.Run("dev mode has no session endpoint", func(t *testing.T) {
		service := NewAuthService(&mocks.UserDalMock{}, nil, auth.Config{}, true)

		resp := postSession(t, service, token, nil)

		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

func TestDisplayName(t *testing.T) {
	require.Equal(t, "Jane Doe", displayName(&auth.Identity{Name: " Jane Doe ", Email: "jane@example.com"}))
	require.Equal(t, "jane", displayName(&auth.Identity{Email: "jane@example.com"}))
	require.Equal(t, "not-an-email", displayName(&auth.Identity{Email: "not-an-email"}))
}

func TestAuthService_LogoutGoesToTheFrontPage(t *testing.T) {
	app := fiber.New()
	service := NewAuthService(&mocks.UserDalMock{}, fakeVerifier{}, auth.Config{}, false)
	app.Post("/auth/logout", service.LogoutHandler)

	resp, err := app.Test(httptest.NewRequest("POST", "/auth/logout", nil), -1)
	require.NoError(t, err)

	require.Equal(t, fiber.StatusFound, resp.StatusCode)
	require.Equal(t, "/", resp.Header.Get("Location"))
}

// loginPage renders the real login view for a service
func loginPage(t *testing.T, service *AuthService) string {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "auth"), 0o755))
	view, err := os.ReadFile("../../views/auth/login.html")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "auth", "login.html"), view, 0o644))

	app := fiber.New(fiber.Config{Views: html.New(dir, ".html")})
	app.Get("/auth/login", service.LoginPageHandler)

	resp, err := app.Test(httptest.NewRequest("GET", "/auth/login", nil), -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(body)
}

func TestAuthService_LoginPage(t *testing.T) {
	firebase := auth.Config{ProjectID: "logbogen", APIKey: "key-123", AuthDomain: "logbogen.firebaseapp.com"}

	t.Run("hands the public Firebase settings to the sign-in script", func(t *testing.T) {
		page := loginPage(t, NewAuthService(&mocks.UserDalMock{}, fakeVerifier{}, firebase, false))

		require.Contains(t, page, `data-api-key="key-123"`)
		require.Contains(t, page, `data-auth-domain="logbogen.firebaseapp.com"`)
		require.Contains(t, page, `data-project-id="logbogen"`)
		require.Contains(t, page, `data-emulator-host=""`)
		require.Contains(t, page, `src="/js/firebase-login.js"`)
		require.NotContains(t, page, "/auth/dev-login")
	})

	t.Run("dev mode shows the dev login and no Firebase", func(t *testing.T) {
		page := loginPage(t, NewAuthService(&mocks.UserDalMock{}, nil, auth.Config{}, true))

		require.Contains(t, page, "/auth/dev-login")
		require.NotContains(t, page, "firebase-login")
	})
}
