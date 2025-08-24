package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
	"github.com/ngomez18/playlist-router/internal/repositories/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUserService(t *testing.T) {
	require := require.New(t)

	ctrl := setupMockController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)
	logger := createTestLogger()

	service := NewUserService(mockRepo, logger)

	require.NotNil(service)
	require.Equal(mockRepo, service.userRepo)
	require.NotNil(service.logger)
}

func TestUserService_CreateUser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    *models.User
		expected *models.User
	}{
		{
			name: "successful creation with complete user data",
			input: &models.User{
				Email:    "test@example.com",
				Username: "testuser",
				Name:     "Test User",
			},
			expected: &models.User{
				ID:       "user123",
				Email:    "test@example.com",
				Username: "testuser",
				Name:     "Test User",
				Created:  time.Now(),
				Updated:  time.Now(),
			},
		},
		{
			name: "successful creation with minimal user data",
			input: &models.User{
				Email: "minimal@example.com",
				Name:  "Minimal User",
			},
			expected: &models.User{
				ID:      "user456",
				Email:   "minimal@example.com",
				Name:    "Minimal User",
				Created: time.Now(),
				Updated: time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			ctrl := setupMockController(t)

			mockRepo := mocks.NewMockUserRepository(ctrl)
			logger := createTestLogger()
			service := NewUserService(mockRepo, logger)

			mockRepo.EXPECT().
				Create(gomock.Any(), tt.input).
				Return(tt.expected, nil).
				Times(1)

			result, err := service.CreateUser(context.Background(), tt.input)

			assert.NoError(err)
			assert.Equal(tt.expected, result)
		})
	}
}

func TestUserService_CreateUser_Error(t *testing.T) {
	tests := []struct {
		name        string
		input       *models.User
		repoError   error
		expectedErr string
	}{
		{
			name: "database operation error",
			input: &models.User{
				Email: "test@example.com",
				Name:  "Test User",
			},
			repoError:   repositories.ErrDatabaseOperation,
			expectedErr: "failed to create user: unable to complete db operation",
		},
		{
			name: "generic repository error",
			input: &models.User{
				Email: "error@example.com",
				Name:  "Error User",
			},
			repoError:   errors.New("connection timeout"),
			expectedErr: "failed to create user: connection timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			ctrl := setupMockController(t)

			mockRepo := mocks.NewMockUserRepository(ctrl)
			logger := createTestLogger()
			service := NewUserService(mockRepo, logger)

			mockRepo.EXPECT().
				Create(gomock.Any(), tt.input).
				Return(nil, tt.repoError).
				Times(1)

			result, err := service.CreateUser(context.Background(), tt.input)

			assert.Error(err)
			assert.Nil(result)
			assert.Contains(err.Error(), tt.expectedErr)
		})
	}
}

func TestUserService_UpdateUser_Success(t *testing.T) {
	assert := assert.New(t)
	ctrl := setupMockController(t)

	mockRepo := mocks.NewMockUserRepository(ctrl)
	logger := createTestLogger()
	service := NewUserService(mockRepo, logger)

	input := &models.User{
		ID:       "user123",
		Email:    "updated@example.com",
		Username: "updateduser",
		Name:     "Updated User",
	}

	expected := &models.User{
		ID:       "user123",
		Email:    "updated@example.com",
		Username: "updateduser",
		Name:     "Updated User",
		Created:  time.Now().Add(-24 * time.Hour),
		Updated:  time.Now(),
	}

	mockRepo.EXPECT().
		Update(gomock.Any(), input).
		Return(expected, nil).
		Times(1)

	result, err := service.UpdateUser(context.Background(), input)

	assert.NoError(err)
	assert.Equal(expected, result)
}

func TestUserService_UpdateUser_Error(t *testing.T) {
	tests := []struct {
		name        string
		input       *models.User
		repoError   error
		expectedErr string
	}{
		{
			name: "user not found error",
			input: &models.User{
				ID:    "nonexistent",
				Email: "test@example.com",
				Name:  "Test User",
			},
			repoError:   repositories.ErrUseNotFound,
			expectedErr: "failed to update user: user not found",
		},
		{
			name: "database operation error",
			input: &models.User{
				ID:    "user123",
				Email: "test@example.com",
				Name:  "Test User",
			},
			repoError:   repositories.ErrDatabaseOperation,
			expectedErr: "failed to update user: unable to complete db operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			ctrl := setupMockController(t)

			mockRepo := mocks.NewMockUserRepository(ctrl)
			logger := createTestLogger()
			service := NewUserService(mockRepo, logger)

			mockRepo.EXPECT().
				Update(gomock.Any(), tt.input).
				Return(nil, tt.repoError).
				Times(1)

			result, err := service.UpdateUser(context.Background(), tt.input)

			assert.Error(err)
			assert.Nil(result)
			assert.Contains(err.Error(), tt.expectedErr)
		})
	}
}

func TestUserService_GetUserByID_Success(t *testing.T) {
	assert := assert.New(t)
	ctrl := setupMockController(t)

	mockRepo := mocks.NewMockUserRepository(ctrl)
	logger := createTestLogger()
	service := NewUserService(mockRepo, logger)

	userID := "user123"
	expected := &models.User{
		ID:       userID,
		Email:    "test@example.com",
		Username: "testuser",
		Name:     "Test User",
		Created:  time.Now().Add(-24 * time.Hour),
		Updated:  time.Now(),
	}

	mockRepo.EXPECT().
		GetByID(gomock.Any(), userID).
		Return(expected, nil).
		Times(1)

	result, err := service.GetUserByID(context.Background(), userID)

	assert.NoError(err)
	assert.Equal(expected, result)
}

func TestUserService_GetUserByID_Error(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		repoError   error
		expectedErr string
	}{
		{
			name:        "user not found error",
			userID:      "nonexistent",
			repoError:   repositories.ErrUseNotFound,
			expectedErr: "failed to retrieve user: user not found",
		},
		{
			name:        "database operation error",
			userID:      "user123",
			repoError:   repositories.ErrDatabaseOperation,
			expectedErr: "failed to retrieve user: unable to complete db operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			ctrl := setupMockController(t)

			mockRepo := mocks.NewMockUserRepository(ctrl)
			logger := createTestLogger()
			service := NewUserService(mockRepo, logger)

			mockRepo.EXPECT().
				GetByID(gomock.Any(), tt.userID).
				Return(nil, tt.repoError).
				Times(1)

			result, err := service.GetUserByID(context.Background(), tt.userID)

			assert.Error(err)
			assert.Nil(result)
			assert.Contains(err.Error(), tt.expectedErr)
		})
	}
}

func TestUserService_DeleteUser_Success(t *testing.T) {
	assert := assert.New(t)
	ctrl := setupMockController(t)

	mockRepo := mocks.NewMockUserRepository(ctrl)
	logger := createTestLogger()
	service := NewUserService(mockRepo, logger)

	userID := "user123"

	mockRepo.EXPECT().
		Delete(gomock.Any(), userID).
		Return(nil).
		Times(1)

	err := service.DeleteUser(context.Background(), userID)

	assert.NoError(err)
}

func TestUserService_DeleteUser_Error(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		repoError   error
		expectedErr string
	}{
		{
			name:        "user not found error",
			userID:      "nonexistent",
			repoError:   repositories.ErrUseNotFound,
			expectedErr: "failed to delete user: user not found",
		},
		{
			name:        "database operation error",
			userID:      "user123",
			repoError:   repositories.ErrDatabaseOperation,
			expectedErr: "failed to delete user: unable to complete db operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			ctrl := setupMockController(t)

			mockRepo := mocks.NewMockUserRepository(ctrl)
			logger := createTestLogger()
			service := NewUserService(mockRepo, logger)

			mockRepo.EXPECT().
				Delete(gomock.Any(), tt.userID).
				Return(tt.repoError).
				Times(1)

			err := service.DeleteUser(context.Background(), tt.userID)

			assert.Error(err)
			assert.Contains(err.Error(), tt.expectedErr)
		})
	}
}

func TestUserService_GenerateAuthToken_Success(t *testing.T) {
	assert := assert.New(t)
	ctrl := setupMockController(t)

	mockRepo := mocks.NewMockUserRepository(ctrl)
	logger := createTestLogger()
	service := NewUserService(mockRepo, logger)

	userID := "user123"

	mockRepo.EXPECT().
		GenerateAuthToken(gomock.Any(), userID).
		Return("token", nil).
		Times(1)

	token, err := service.GenerateAuthToken(context.Background(), userID)
	assert.NoError(err)
	assert.Equal("token", token)
}

func TestUserService_GenerateAuthToken_DatabaseError(t *testing.T) {
	assert := assert.New(t)
	ctrl := setupMockController(t)

	mockRepo := mocks.NewMockUserRepository(ctrl)
	logger := createTestLogger()
	service := NewUserService(mockRepo, logger)

	userID := "user123"

	mockRepo.EXPECT().
		GenerateAuthToken(gomock.Any(), userID).
		Return("", repositories.ErrDatabaseOperation).
		Times(1)

	_, err := service.GenerateAuthToken(context.Background(), userID)
	assert.ErrorIs(err, repositories.ErrDatabaseOperation)
}

func TestUserService_ValidateAuthToken_Success(t *testing.T) {
	assert := assert.New(t)
	ctrl := setupMockController(t)

	mockRepo := mocks.NewMockUserRepository(ctrl)
	logger := createTestLogger()
	service := NewUserService(mockRepo, logger)

	token := "valid_token_123"
	expectedUser := &models.User{
		ID:    "user123",
		Email: "test@example.com",
		Name:  "Test User",
	}

	mockRepo.EXPECT().
		ValidateAuthToken(gomock.Any(), token).
		Return(expectedUser, nil).
		Times(1)

	user, err := service.ValidateAuthToken(context.Background(), token)
	assert.NoError(err)
	assert.Equal(expectedUser.ID, user.ID)
	assert.Equal(expectedUser.Email, user.Email)
	assert.Equal(expectedUser.Name, user.Name)
}

func TestUserService_ValidateAuthToken_Error(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		repositoryErr error
		expectedErr   error
	}{
		{
			name:          "generic repository error",
			token:         "some_token",
			repositoryErr: errors.New("repository failure"),
			expectedErr:   errors.New("repository failure"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			ctrl := setupMockController(t)

			mockRepo := mocks.NewMockUserRepository(ctrl)
			logger := createTestLogger()
			service := NewUserService(mockRepo, logger)

			mockRepo.EXPECT().
				ValidateAuthToken(gomock.Any(), tt.token).
				Return(nil, tt.repositoryErr).
				Times(1)

			user, err := service.ValidateAuthToken(context.Background(), tt.token)
			assert.Error(err)
			assert.Nil(user)
			assert.Contains(err.Error(), "failed to validate auth token")
		})
	}
}
