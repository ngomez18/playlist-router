package services

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
	"github.com/ngomez18/playlist-router/internal/repositories/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSpotifyIntegrationService(t *testing.T) {
	require := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSpotifyIntegrationRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	service := NewSpotifyIntegrationService(mockRepo, logger)

	require.NotNil(service)
	require.Equal(mockRepo, service.integrationRepo)
	require.NotNil(service.logger)
}

func TestSpotifyIntegrationService_CreateOrUpdateIntegration_Success(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		input    *models.SpotifyIntegration
		expected *models.SpotifyIntegration
	}{
		{
			name:   "successful creation with complete data",
			userID: "user123",
			input: &models.SpotifyIntegration{
				SpotifyID:    "spotify_user_123",
				AccessToken:  "access_token_123",
				RefreshToken: "refresh_token_123",
				TokenType:    "Bearer",
				ExpiresAt:    time.Now().Add(time.Hour),
				Scope:        "user-read-private user-read-email",
				DisplayName:  "Test User",
			},
			expected: &models.SpotifyIntegration{
				ID:           "integration123",
				UserID:       "user123",
				SpotifyID:    "spotify_user_123",
				AccessToken:  "access_token_123",
				RefreshToken: "refresh_token_123",
				TokenType:    "Bearer",
				ExpiresAt:    time.Now().Add(time.Hour),
				Scope:        "user-read-private user-read-email",
				DisplayName:  "Test User",
				Created:      time.Now(),
				Updated:      time.Now(),
			},
		},
		{
			name:   "successful update of existing integration",
			userID: "user456",
			input: &models.SpotifyIntegration{
				SpotifyID:    "spotify_user_456",
				AccessToken:  "new_access_token",
				RefreshToken: "new_refresh_token",
				TokenType:    "Bearer",
				ExpiresAt:    time.Now().Add(time.Hour),
			},
			expected: &models.SpotifyIntegration{
				ID:           "integration456",
				UserID:       "user456",
				SpotifyID:    "spotify_user_456",
				AccessToken:  "new_access_token",
				RefreshToken: "new_refresh_token",
				TokenType:    "Bearer",
				ExpiresAt:    time.Now().Add(time.Hour),
				Created:      time.Now().Add(-24 * time.Hour),
				Updated:      time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockSpotifyIntegrationRepository(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
			service := NewSpotifyIntegrationService(mockRepo, logger)

			mockRepo.EXPECT().
				CreateOrUpdate(gomock.Any(), tt.userID, tt.input).
				Return(tt.expected, nil).
				Times(1)

			result, err := service.CreateOrUpdateIntegration(context.Background(), tt.userID, tt.input)

			assert.NoError(err)
			assert.Equal(tt.expected, result)
		})
	}
}

func TestSpotifyIntegrationService_CreateOrUpdateIntegration_Error(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		input       *models.SpotifyIntegration
		repoError   error
		expectedErr string
	}{
		{
			name:   "database operation error",
			userID: "user123",
			input: &models.SpotifyIntegration{
				SpotifyID:   "spotify_user_123",
				AccessToken: "access_token_123",
			},
			repoError:   repositories.ErrDatabaseOperation,
			expectedErr: "unable to complete db operation",
		},
		{
			name:   "generic repository error",
			userID: "user456",
			input: &models.SpotifyIntegration{
				SpotifyID:   "spotify_user_456",
				AccessToken: "access_token_456",
			},
			repoError:   errors.New("connection timeout"),
			expectedErr: "connection timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockSpotifyIntegrationRepository(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
			service := NewSpotifyIntegrationService(mockRepo, logger)

			mockRepo.EXPECT().
				CreateOrUpdate(gomock.Any(), tt.userID, tt.input).
				Return(nil, tt.repoError).
				Times(1)

			result, err := service.CreateOrUpdateIntegration(context.Background(), tt.userID, tt.input)

			assert.Error(err)
			assert.Nil(result)
			assert.Contains(err.Error(), tt.expectedErr)
		})
	}
}

func TestSpotifyIntegrationService_GetIntegrationByUserID_Success(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSpotifyIntegrationRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	service := NewSpotifyIntegrationService(mockRepo, logger)

	userID := "user123"
	expected := &models.SpotifyIntegration{
		ID:           "integration123",
		UserID:       userID,
		SpotifyID:    "spotify_user_123",
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(time.Hour),
		Created:      time.Now().Add(-24 * time.Hour),
		Updated:      time.Now(),
	}

	mockRepo.EXPECT().
		GetByUserID(gomock.Any(), userID).
		Return(expected, nil).
		Times(1)

	result, err := service.GetIntegrationByUserID(context.Background(), userID)

	assert.NoError(err)
	assert.Equal(expected, result)
}

func TestSpotifyIntegrationService_GetIntegrationByUserID_Error(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		repoError   error
		expectedErr string
	}{
		{
			name:        "integration not found error",
			userID:      "nonexistent",
			repoError:   repositories.ErrSpotifyIntegrationNotFound,
			expectedErr: "spotify integration not found",
		},
		{
			name:        "database operation error",
			userID:      "user123",
			repoError:   repositories.ErrDatabaseOperation,
			expectedErr: "unable to complete db operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockSpotifyIntegrationRepository(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
			service := NewSpotifyIntegrationService(mockRepo, logger)

			mockRepo.EXPECT().
				GetByUserID(gomock.Any(), tt.userID).
				Return(nil, tt.repoError).
				Times(1)

			result, err := service.GetIntegrationByUserID(context.Background(), tt.userID)

			assert.Error(err)
			assert.Nil(result)
			assert.Contains(err.Error(), tt.expectedErr)
		})
	}
}

func TestSpotifyIntegrationService_GetIntegrationBySpotifyID_Success(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSpotifyIntegrationRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	service := NewSpotifyIntegrationService(mockRepo, logger)

	spotifyID := "spotify_user_123"
	expected := &models.SpotifyIntegration{
		ID:           "integration123",
		UserID:       "user123",
		SpotifyID:    spotifyID,
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(time.Hour),
		Created:      time.Now().Add(-24 * time.Hour),
		Updated:      time.Now(),
	}

	mockRepo.EXPECT().
		GetBySpotifyID(gomock.Any(), spotifyID).
		Return(expected, nil).
		Times(1)

	result, err := service.GetIntegrationBySpotifyID(context.Background(), spotifyID)

	assert.NoError(err)
	assert.Equal(expected, result)
}

func TestSpotifyIntegrationService_GetIntegrationBySpotifyID_Error(t *testing.T) {
	tests := []struct {
		name        string
		spotifyID   string
		repoError   error
		expectedErr string
	}{
		{
			name:        "integration not found error",
			spotifyID:   "nonexistent",
			repoError:   repositories.ErrSpotifyIntegrationNotFound,
			expectedErr: "spotify integration not found",
		},
		{
			name:        "database operation error",
			spotifyID:   "spotify_user_123",
			repoError:   repositories.ErrDatabaseOperation,
			expectedErr: "unable to complete db operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockSpotifyIntegrationRepository(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
			service := NewSpotifyIntegrationService(mockRepo, logger)

			mockRepo.EXPECT().
				GetBySpotifyID(gomock.Any(), tt.spotifyID).
				Return(nil, tt.repoError).
				Times(1)

			result, err := service.GetIntegrationBySpotifyID(context.Background(), tt.spotifyID)

			assert.Error(err)
			assert.Nil(result)
			assert.Contains(err.Error(), tt.expectedErr)
		})
	}
}

func TestSpotifyIntegrationService_UpdateTokens_Success(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSpotifyIntegrationRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	service := NewSpotifyIntegrationService(mockRepo, logger)

	integrationID := "integration123"
	tokens := &models.SpotifyIntegrationTokenRefresh{
		AccessToken:  "new_access_token",
		RefreshToken: "new_refresh_token",
		ExpiresIn:    3600,
	}

	mockRepo.EXPECT().
		UpdateTokens(gomock.Any(), integrationID, tokens).
		Return(nil).
		Times(1)

	err := service.UpdateTokens(context.Background(), integrationID, tokens)

	assert.NoError(err)
}

func TestSpotifyIntegrationService_UpdateTokens_Error(t *testing.T) {
	tests := []struct {
		name          string
		integrationID string
		tokens        *models.SpotifyIntegrationTokenRefresh
		repoError     error
		expectedErr   string
	}{
		{
			name:          "integration not found error",
			integrationID: "nonexistent",
			tokens: &models.SpotifyIntegrationTokenRefresh{
				AccessToken: "access_token",
			},
			repoError:   repositories.ErrSpotifyIntegrationNotFound,
			expectedErr: "spotify integration not found",
		},
		{
			name:          "database operation error",
			integrationID: "integration123",
			tokens: &models.SpotifyIntegrationTokenRefresh{
				AccessToken: "access_token",
			},
			repoError:   repositories.ErrDatabaseOperation,
			expectedErr: "unable to complete db operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockSpotifyIntegrationRepository(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
			service := NewSpotifyIntegrationService(mockRepo, logger)

			mockRepo.EXPECT().
				UpdateTokens(gomock.Any(), tt.integrationID, tt.tokens).
				Return(tt.repoError).
				Times(1)

			err := service.UpdateTokens(context.Background(), tt.integrationID, tt.tokens)

			assert.Error(err)
			assert.Contains(err.Error(), tt.expectedErr)
		})
	}
}

func TestSpotifyIntegrationService_DeleteIntegration_Success(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockSpotifyIntegrationRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	service := NewSpotifyIntegrationService(mockRepo, logger)

	userID := "user123"

	mockRepo.EXPECT().
		Delete(gomock.Any(), userID).
		Return(nil).
		Times(1)

	err := service.DeleteIntegration(context.Background(), userID)

	assert.NoError(err)
}

func TestSpotifyIntegrationService_DeleteIntegration_Error(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		repoError   error
		expectedErr string
	}{
		{
			name:        "integration not found error",
			userID:      "nonexistent",
			repoError:   repositories.ErrSpotifyIntegrationNotFound,
			expectedErr: "spotify integration not found",
		},
		{
			name:        "database operation error",
			userID:      "user123",
			repoError:   repositories.ErrDatabaseOperation,
			expectedErr: "unable to complete db operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockSpotifyIntegrationRepository(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
			service := NewSpotifyIntegrationService(mockRepo, logger)

			mockRepo.EXPECT().
				Delete(gomock.Any(), tt.userID).
				Return(tt.repoError).
				Times(1)

			err := service.DeleteIntegration(context.Background(), tt.userID)

			assert.Error(err)
			assert.Contains(err.Error(), tt.expectedErr)
		})
	}
}
