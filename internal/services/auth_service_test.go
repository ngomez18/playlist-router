package services

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	clientMocks "github.com/ngomez18/playlist-router/internal/clients/mocks"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
	repoMocks "github.com/ngomez18/playlist-router/internal/repositories/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuthService(t *testing.T) {
	assert := require.New(t)

	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create real services with mock repositories for testing
	mockUserRepo := repoMocks.NewMockUserRepository(ctrl)
	mockSpotifyIntegrationRepo := repoMocks.NewMockSpotifyIntegrationRepository(ctrl)
	mockSpotifyClient := clientMocks.NewMockSpotifyAPI(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	userService := NewUserService(mockUserRepo, logger)
	spotifyIntegrationService := NewSpotifyIntegrationService(mockSpotifyIntegrationRepo, logger)

	// Execute
	authService := NewAuthService(userService, spotifyIntegrationService, mockSpotifyClient, logger)

	// Assert
	assert.NotNil(authService)
	assert.Equal(userService, authService.userService)
	assert.Equal(spotifyIntegrationService, authService.spotifyIntegrationService)
	assert.Equal(mockSpotifyClient, authService.spotifyClient)
	assert.NotNil(authService.logger)
}

func TestAuthService_GenerateSpotifyAuthURL(t *testing.T) {
	assert := require.New(t)

	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := repoMocks.NewMockUserRepository(ctrl)
	mockSpotifyIntegrationRepo := repoMocks.NewMockSpotifyIntegrationRepository(ctrl)
	mockSpotifyClient := clientMocks.NewMockSpotifyAPI(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	userService := NewUserService(mockUserRepo, logger)
	spotifyIntegrationService := NewSpotifyIntegrationService(mockSpotifyIntegrationRepo, logger)
	authService := NewAuthService(userService, spotifyIntegrationService, mockSpotifyClient, logger)

	state := "test_state"
	expectedURL := "https://accounts.spotify.com/authorize?client_id=test&state=test_state"

	// Setup mock expectations
	mockSpotifyClient.EXPECT().
		GenerateAuthURL(state).
		Return(expectedURL).
		Times(1)

	// Execute
	actualURL := authService.GenerateSpotifyAuthURL(state)

	// Assert
	assert.Equal(expectedURL, actualURL)
}

func TestAuthService_FindUserBySpotifyID_Success(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := repoMocks.NewMockUserRepository(ctrl)
	mockSpotifyIntegrationRepo := repoMocks.NewMockSpotifyIntegrationRepository(ctrl)
	mockSpotifyClient := clientMocks.NewMockSpotifyAPI(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	userService := NewUserService(mockUserRepo, logger)
	spotifyIntegrationService := NewSpotifyIntegrationService(mockSpotifyIntegrationRepo, logger)
	authService := NewAuthService(userService, spotifyIntegrationService, mockSpotifyClient, logger)

	spotifyID := "spotify_user_123"
	expectedIntegration := &models.SpotifyIntegration{
		ID:           "integration123",
		UserID:       "user123",
		SpotifyID:    spotifyID,
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(time.Hour),
	}
	expectedUser := &models.User{
		ID:       "user123",
		Email:    "test@example.com",
		Username: "testuser",
		Name:     "Test User",
		Created:  time.Now().Add(-24 * time.Hour),
		Updated:  time.Now(),
	}

	// Setup mock expectations
	mockSpotifyIntegrationRepo.EXPECT().
		GetBySpotifyID(gomock.Any(), spotifyID).
		Return(expectedIntegration, nil).
		Times(1)

	mockUserRepo.EXPECT().
		GetByID(gomock.Any(), expectedIntegration.UserID).
		Return(expectedUser, nil).
		Times(1)

	// Execute
	user, integration, err := authService.findUserBySpotifyID(context.Background(), spotifyID)

	// Assert
	assert.NoError(err)
	assert.Equal(expectedUser, user)
	assert.Equal(expectedIntegration, integration)
}

func TestAuthService_FindUserBySpotifyID_IntegrationNotFound(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := repoMocks.NewMockUserRepository(ctrl)
	mockSpotifyIntegrationRepo := repoMocks.NewMockSpotifyIntegrationRepository(ctrl)
	mockSpotifyClient := clientMocks.NewMockSpotifyAPI(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	userService := NewUserService(mockUserRepo, logger)
	spotifyIntegrationService := NewSpotifyIntegrationService(mockSpotifyIntegrationRepo, logger)
	authService := NewAuthService(userService, spotifyIntegrationService, mockSpotifyClient, logger)

	spotifyID := "nonexistent_spotify_user"

	// Setup mock expectations
	mockSpotifyIntegrationRepo.EXPECT().
		GetBySpotifyID(gomock.Any(), spotifyID).
		Return(nil, repositories.ErrSpotifyIntegrationNotFound).
		Times(1)

	// Execute
	user, integration, err := authService.findUserBySpotifyID(context.Background(), spotifyID)

	// Assert - should return nil, nil, nil when not found (not an error)
	assert.NoError(err)
	assert.Nil(user)
	assert.Nil(integration)
}

func TestAuthService_FindUserBySpotifyID_IntegrationError(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := repoMocks.NewMockUserRepository(ctrl)
	mockSpotifyIntegrationRepo := repoMocks.NewMockSpotifyIntegrationRepository(ctrl)
	mockSpotifyClient := clientMocks.NewMockSpotifyAPI(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	userService := NewUserService(mockUserRepo, logger)
	spotifyIntegrationService := NewSpotifyIntegrationService(mockSpotifyIntegrationRepo, logger)
	authService := NewAuthService(userService, spotifyIntegrationService, mockSpotifyClient, logger)

	spotifyID := "spotify_user_123"

	// Setup mock expectations
	mockSpotifyIntegrationRepo.EXPECT().
		GetBySpotifyID(gomock.Any(), spotifyID).
		Return(nil, repositories.ErrDatabaseOperation).
		Times(1)

	// Execute
	user, integration, err := authService.findUserBySpotifyID(context.Background(), spotifyID)

	// Assert
	assert.Error(err)
	assert.Nil(user)
	assert.Nil(integration)
	assert.Contains(err.Error(), "unable to complete db operation")
}

func TestAuthService_FindUserBySpotifyID_UserError(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := repoMocks.NewMockUserRepository(ctrl)
	mockSpotifyIntegrationRepo := repoMocks.NewMockSpotifyIntegrationRepository(ctrl)
	mockSpotifyClient := clientMocks.NewMockSpotifyAPI(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	userService := NewUserService(mockUserRepo, logger)
	spotifyIntegrationService := NewSpotifyIntegrationService(mockSpotifyIntegrationRepo, logger)
	authService := NewAuthService(userService, spotifyIntegrationService, mockSpotifyClient, logger)

	spotifyID := "spotify_user_123"
	integration := &models.SpotifyIntegration{
		ID:        "integration123",
		UserID:    "user123",
		SpotifyID: spotifyID,
	}

	// Setup mock expectations
	mockSpotifyIntegrationRepo.EXPECT().
		GetBySpotifyID(gomock.Any(), spotifyID).
		Return(integration, nil).
		Times(1)

	mockUserRepo.EXPECT().
		GetByID(gomock.Any(), integration.UserID).
		Return(nil, repositories.ErrUseNotFound).
		Times(1)

	// Execute
	user, returnedIntegration, err := authService.findUserBySpotifyID(context.Background(), spotifyID)

	// Assert
	assert.Error(err)
	assert.Nil(user)
	assert.Nil(returnedIntegration)
	assert.Contains(err.Error(), "failed to retrieve user")
}
