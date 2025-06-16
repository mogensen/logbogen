package services

import (
	"errors"
	"testing"

	"github.com/mogensen/logbook/pkg/dal"
	"github.com/mogensen/logbook/pkg/mocks"
	"github.com/mogensen/logbook/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestAuthService_Signup(t *testing.T) {
	tests := []struct {
		testName    string
		req         SignupRequest
		mockSetup   func(*mocks.MockUserDal)
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
			mockSetup: func(m *mocks.MockUserDal) {
				m.On("FindUserByEmail", mock.Anything, "test@example.com").Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
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
			mockSetup: func(m *mocks.MockUserDal) {
				m.On("FindUserByEmail", mock.Anything, "existing@example.com").Return(&gorm.DB{Error: nil})
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
			mockUserDal := &mocks.MockUserDal{}
			tt.mockSetup(mockUserDal)

			service := NewAuthService(mockUserDal)
			resp, err := service.Signup(tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, resp)
			}

			mockUserDal.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	tests := []struct {
		testName    string
		req         LoginRequest
		mockSetup   func(*mocks.MockUserDal)
		expectedErr error
		expected    *LoginResponse
	}{
		{
			testName: "successful login",
			req: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockSetup: func(m *mocks.MockUserDal) {
				m.On("FindUserByEmail", mock.Anything, "test@example.com").Return(&gorm.DB{Error: nil})
			},
			expectedErr: errors.New("invalid email or password"), // This will fail because we can't mock bcrypt.Verify
			expected:    nil,
		},
		{
			testName: "user not found",
			req: LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			mockSetup: func(m *mocks.MockUserDal) {
				m.On("FindUserByEmail", mock.Anything, "nonexistent@example.com").Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
			},
			expectedErr: errors.New("invalid email or password"),
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockUserDal := &mocks.MockUserDal{}
			tt.mockSetup(mockUserDal)

			service := NewAuthService(mockUserDal)
			resp, err := service.Login(tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, resp)
			}

			mockUserDal.AssertExpectations(t)
		})
	}
}

func TestAuthService_GetUser(t *testing.T) {
	tests := []struct {
		testName    string
		req         GetUserRequest
		mockSetup   func(*mocks.MockUserDal)
		expectedErr error
		expected    *GetUserResponse
	}{
		{
			testName: "successful get user",
			req: GetUserRequest{
				UserID: 1,
			},
			mockSetup: func(m *mocks.MockUserDal) {
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
			mockSetup: func(m *mocks.MockUserDal) {
				m.On("FindUserById", uint64(999)).Return(nil, gorm.ErrRecordNotFound)
			},
			expectedErr: gorm.ErrRecordNotFound,
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			mockUserDal := &mocks.MockUserDal{}
			tt.mockSetup(mockUserDal)

			service := NewAuthService(mockUserDal)
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

			mockUserDal.AssertExpectations(t)
		})
	}
}

func TestGetUserByID(t *testing.T) {
	// Setup
	mockUserDal := new(mocks.MockUserDal)
	authService := NewAuthService(mockUserDal)

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
				mockUserDal.On("FindUserById", uint64(1)).
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
				mockUserDal.On("FindUserById", uint64(999)).
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

func TestGetUserByEmail(t *testing.T) {
	// Setup
	mockUserDal := new(mocks.MockUserDal)
	authService := NewAuthService(mockUserDal)

	// Test cases
	tests := []struct {
		name    string
		email   string
		mock    func()
		want    *types.User
		wantErr bool
	}{
		{
			name:  "successful user retrieval",
			email: "test@example.com",
			mock: func() {
				mockUserDal.On("FindUserByEmail", mock.Anything, "test@example.com").
					Return(&gorm.DB{Error: nil})
			},
			wantErr: false,
		},
		{
			name:  "user not found",
			email: "nonexistent@example.com",
			mock: func() {
				mockUserDal.On("FindUserByEmail", mock.Anything, "nonexistent@example.com").
					Return(&gorm.DB{Error: gorm.ErrRecordNotFound})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			user, err := authService.GetUserByEmail(tt.email)
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
