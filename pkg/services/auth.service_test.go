package services

import (
	"errors"
	"testing"

	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/mocks"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/mogensen/logbook/pkg/utils/password"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestAuthService_Signup(t *testing.T) {
	tests := []struct {
		testName    string
		req         SignupRequest
		mockSetup   func(*mocks.UserDalMock)
		expectedErr error
		expected    *SignupResponse
	}{
		{
			testName: "successful signup",
			req: SignupRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func(m *mocks.UserDalMock) {
				m.On("FindUserByEmail", "test@example.com").Return(nil, gorm.ErrRecordNotFound)
				m.On("CreateUser", mock.Anything).Return(&gorm.DB{Error: nil})
			},
			expectedErr: nil,
			expected: &SignupResponse{
				Success: true,
				Message: "Brugeren er oprettet, du kan nu logge ind",
			},
		},
		{
			testName: "email already exists",
			req: SignupRequest{
				Name:     "Test User",
				Email:    "existing@example.com",
				Password: "password123",
			},
			mockSetup: func(m *mocks.UserDalMock) {
				m.On("FindUserByEmail", "existing@example.com").Return(&dal.User{
					Model:    gorm.Model{ID: 1},
					Name:     "Test User",
					Email:    "existing@example.com",
					Password: "password123",
				}, nil)
			},
			expectedErr: nil,
			expected: &SignupResponse{
				Success: false,
				Message: "Der er already en bruger med denne email",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			userDalMock := &mocks.UserDalMock{}
			tt.mockSetup(userDalMock)

			service := NewAuthService(userDalMock)
			resp, err := service.Signup(tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, resp)
			}

			userDalMock.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	tests := []struct {
		testName    string
		req         LoginRequest
		mockSetup   func(*mocks.UserDalMock)
		expectedErr error
		expected    *LoginResponse
	}{
		{
			testName: "successful login",
			req: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func(m *mocks.UserDalMock) {
				m.On("FindUserByEmail", "test@example.com").Return(&dal.User{
					Model:    gorm.Model{ID: 1},
					Name:     "Test User",
					Email:    "test@example.com",
					Password: password.Generate("password123"),
				}, nil)
			},
			expectedErr: nil,
			expected: &LoginResponse{
				UserID:   1,
				Email:    "test@example.com",
				LoggedIn: true,
			},
		},
		{
			testName: "wrong password",
			req: LoginRequest{
				Email:    "bad-password@example.com",
				Password: "not-the-password",
			},
			mockSetup: func(m *mocks.UserDalMock) {
				m.On("FindUserByEmail", "bad-password@example.com").Return(&dal.User{
					Model:    gorm.Model{ID: 1},
					Name:     "Test User",
					Email:    "bad-password@example.com",
					Password: password.Generate("password123"),
				}, nil)
			},
			expectedErr: errors.New("invalid email or password"),
			expected:    nil,
		},
		{
			testName: "user not found",
			req: LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			mockSetup: func(m *mocks.UserDalMock) {
				m.On("FindUserByEmail", "nonexistent@example.com").Return(nil, gorm.ErrRecordNotFound)
			},
			expectedErr: errors.New("invalid email or password"),
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			userDalMock := &mocks.UserDalMock{}
			tt.mockSetup(userDalMock)

			service := NewAuthService(userDalMock)
			resp, err := service.Login(tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, resp)
			}

			userDalMock.AssertExpectations(t)
		})
	}
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

			service := NewAuthService(userDalMock)
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
	authService := NewAuthService(userDalMock)

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
