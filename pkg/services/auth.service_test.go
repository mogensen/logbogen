package services

import (
	"testing"

	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/mocks"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestAuthService_upsertUser(t *testing.T) {
	t.Run("returns existing user when the Auth0 sub is known", func(t *testing.T) {
		userDalMock := &mocks.UserDalMock{}
		existing := &dal.User{
			Model: gorm.Model{ID: 7},
			Name:  "Existing User",
			Email: "existing@example.com",
		}
		userDalMock.On("FindUserByAuth0Sub", "auth0|123").Return(existing, nil)

		service := NewAuthService(userDalMock, nil, true)
		user, err := service.upsertUser("auth0|123", "Existing User", "existing@example.com")

		assert.NoError(t, err)
		assert.Equal(t, existing, user)
		// No user is created when one already exists.
		userDalMock.AssertNotCalled(t, "CreateUser", mock.Anything)
		userDalMock.AssertExpectations(t)
	})

	t.Run("creates a new user when the Auth0 sub is unknown", func(t *testing.T) {
		userDalMock := &mocks.UserDalMock{}
		userDalMock.On("FindUserByAuth0Sub", "auth0|new").Return(nil, gorm.ErrRecordNotFound)
		userDalMock.On("CreateUser", mock.Anything).Return(&gorm.DB{Error: nil})

		service := NewAuthService(userDalMock, nil, true)
		user, err := service.upsertUser("auth0|new", "New User", "new@example.com")

		assert.NoError(t, err)
		assert.Equal(t, "New User", user.Name)
		assert.Equal(t, "new@example.com", user.Email)
		assert.NotNil(t, user.Auth0Sub)
		assert.Equal(t, "auth0|new", *user.Auth0Sub)
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

			service := NewAuthService(userDalMock, nil, true)
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
	authService := NewAuthService(userDalMock, nil, true)

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
