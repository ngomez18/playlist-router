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
	user, err := authService.findUserBySpotifyID(context.Background(), spotifyID)

	// Assert
	assert.NoError(err)
	assert.Equal(expectedUser, user)
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
	user, err := authService.findUserBySpotifyID(context.Background(), spotifyID)

	// Assert - should return nil, nil when not found (not an error)
	assert.NoError(err)
	assert.Nil(user)
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
	user, err := authService.findUserBySpotifyID(context.Background(), spotifyID)

	// Assert
	assert.Error(err)
	assert.Nil(user)
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
	user, err := authService.findUserBySpotifyID(context.Background(), spotifyID)

	// Assert
	assert.Error(err)
	assert.Nil(user)
	assert.Contains(err.Error(), "failed to retrieve user")
}

func TestAuthService_CreateNewUser_Success(t *testing.T) {
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

	// Test data
	profile := &models.SpotifyUserProfile{
		ID:    "spotify_user_123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	tokens := &models.SpotifyTokenResponse{
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		Scope:        "user-read-private user-read-email",
	}

	expectedUser := &models.User{
		ID:      "user123",
		Email:   profile.Email,
		Name:    profile.Name,
		Created: time.Now(),
		Updated: time.Now(),
	}

	expectedIntegration := &models.SpotifyIntegration{
		ID:           "integration123",
		UserID:       expectedUser.ID,
		SpotifyID:    profile.ID,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresAt:    time.Now().Add(time.Hour),
		Scope:        tokens.Scope,
		DisplayName:  profile.Name,
		Created:      time.Now(),
		Updated:      time.Now(),
	}

	// Setup mock expectations
	mockUserRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, user *models.User) (*models.User, error) {
			// Verify input user fields
			assert.Equal(profile.Email, user.Email)
			assert.Equal(profile.Name, user.Name)
			return expectedUser, nil
		}).
		Times(1)

	mockSpotifyIntegrationRepo.EXPECT().
		CreateOrUpdate(gomock.Any(), expectedUser.ID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, userID string, integration *models.SpotifyIntegration) (*models.SpotifyIntegration, error) {
			// Verify input integration fields
			assert.Equal(profile.ID, integration.SpotifyID)
			assert.Equal(tokens.AccessToken, integration.AccessToken)
			assert.Equal(tokens.RefreshToken, integration.RefreshToken)
			assert.Equal(tokens.TokenType, integration.TokenType)
			assert.Equal(tokens.Scope, integration.Scope)
			assert.Equal(profile.Name, integration.DisplayName)
			// Verify expiration is approximately correct (within 5 seconds)
			expectedExpiry := time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
			assert.WithinDuration(expectedExpiry, integration.ExpiresAt, 5*time.Second)
			return expectedIntegration, nil
		}).
		Times(1)

	// Execute
	result, err := authService.createNewUser(context.Background(), profile, tokens)

	// Assert
	assert.NoError(err)
	assert.NotNil(result)
	assert.Equal(expectedUser.ID, result.ID)
	assert.Equal(expectedUser.Email, result.Email)
	assert.Equal(expectedUser.Name, result.Name)
	assert.Equal(expectedIntegration.SpotifyID, result.SpotifyID)
}

func TestAuthService_CreateNewUser_UserCreationError(t *testing.T) {
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

	profile := &models.SpotifyUserProfile{
		ID:    "spotify_user_123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	tokens := &models.SpotifyTokenResponse{
		AccessToken: "access_token_123",
		ExpiresIn:   3600,
	}

	// Setup mock expectations
	mockUserRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(nil, repositories.ErrDatabaseOperation).
		Times(1)

	// Execute
	result, err := authService.createNewUser(context.Background(), profile, tokens)

	// Assert
	assert.Error(err)
	assert.Nil(result)
	assert.Contains(err.Error(), "unable to complete db operation")
}

func TestAuthService_CreateNewUser_IntegrationCreationError(t *testing.T) {
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

	profile := &models.SpotifyUserProfile{
		ID:    "spotify_user_123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	tokens := &models.SpotifyTokenResponse{
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		Scope:        "user-read-private",
	}

	expectedUser := &models.User{
		ID:    "user123",
		Email: profile.Email,
		Name:  profile.Name,
	}

	// Setup mock expectations
	mockUserRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(expectedUser, nil).
		Times(1)

	mockSpotifyIntegrationRepo.EXPECT().
		CreateOrUpdate(gomock.Any(), expectedUser.ID, gomock.Any()).
		Return(nil, repositories.ErrDatabaseOperation).
		Times(1)

	// Execute
	result, err := authService.createNewUser(context.Background(), profile, tokens)

	// Assert
	assert.Error(err)
	assert.Nil(result)
	assert.Contains(err.Error(), "unable to complete db operation")
}

func TestAuthService_UpdateExistingUser_Success_NoUserChanges(t *testing.T) {
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

	// Test data - user profile matches existing user
	existingUser := &models.User{
		ID:    "user123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	profile := &models.SpotifyUserProfile{
		ID:    "spotify_user_123",
		Email: "test@example.com", // Same as existing
		Name:  "Test User",        // Same as existing
	}
	tokens := &models.SpotifyTokenResponse{
		AccessToken:  "new_access_token",
		RefreshToken: "new_refresh_token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		Scope:        "user-read-private user-read-email",
	}

	expectedUpdatedIntegration := &models.SpotifyIntegration{
		ID:           "integration123",
		UserID:       "user123",
		SpotifyID:    "spotify_user_123",
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresAt:    time.Now().Add(time.Hour),
		Scope:        tokens.Scope,
		DisplayName:  profile.Name,
		Created:      time.Now().Add(-24 * time.Hour),
		Updated:      time.Now(),
	}

	// Setup mock expectations - no user update since profile unchanged
	mockSpotifyIntegrationRepo.EXPECT().
		CreateOrUpdate(gomock.Any(), existingUser.ID, gomock.Any()).
		DoAndReturn(func(ctx context.Context, userID string, integration *models.SpotifyIntegration) (*models.SpotifyIntegration, error) {
			// Verify integration fields
			assert.Equal(profile.ID, integration.SpotifyID)
			assert.Equal(tokens.AccessToken, integration.AccessToken)
			assert.Equal(tokens.RefreshToken, integration.RefreshToken)
			assert.Equal(tokens.TokenType, integration.TokenType)
			assert.Equal(tokens.Scope, integration.Scope)
			assert.Equal(profile.Name, integration.DisplayName)
			// Verify expiration is approximately correct
			expectedExpiry := time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
			assert.WithinDuration(expectedExpiry, integration.ExpiresAt, 5*time.Second)
			return expectedUpdatedIntegration, nil
		}).
		Times(1)

	// Execute
	result, err := authService.updateExistingUser(context.Background(), existingUser, profile, tokens)

	// Assert
	assert.NoError(err)
	assert.NotNil(result)
	assert.Equal(existingUser.ID, result.ID)
	assert.Equal(existingUser.Email, result.Email)
	assert.Equal(existingUser.Name, result.Name)
	assert.Equal(expectedUpdatedIntegration.SpotifyID, result.SpotifyID)
}

func TestAuthService_UpdateExistingUser_Success_WithUserChanges(t *testing.T) {
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

	// Test data - user profile has changes
	existingUser := &models.User{
		ID:    "user123",
		Email: "old@example.com",
		Name:  "Old Name",
	}
	profile := &models.SpotifyUserProfile{
		ID:    "spotify_user_123",
		Email: "new@example.com", // Changed
		Name:  "New Name",        // Changed
	}
	tokens := &models.SpotifyTokenResponse{
		AccessToken:  "new_access_token",
		RefreshToken: "new_refresh_token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		Scope:        "user-read-private user-read-email",
	}

	expectedUpdatedUser := &models.User{
		ID:      "user123",
		Email:   profile.Email,
		Name:    profile.Name,
		Created: time.Now().Add(-24 * time.Hour),
		Updated: time.Now(),
	}

	expectedUpdatedIntegration := &models.SpotifyIntegration{
		ID:           "integration123",
		UserID:       "user123",
		SpotifyID:    "spotify_user_123",
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    tokens.TokenType,
		ExpiresAt:    time.Now().Add(time.Hour),
		Scope:        tokens.Scope,
		DisplayName:  profile.Name,
		Created:      time.Now().Add(-24 * time.Hour),
		Updated:      time.Now(),
	}

	// Setup mock expectations
	mockUserRepo.EXPECT().
		Update(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, user *models.User) (*models.User, error) {
			// Verify input user fields
			assert.Equal(existingUser.ID, user.ID)
			assert.Equal(profile.Email, user.Email)
			assert.Equal(profile.Name, user.Name)
			return expectedUpdatedUser, nil
		}).
		Times(1)

	mockSpotifyIntegrationRepo.EXPECT().
		CreateOrUpdate(gomock.Any(), expectedUpdatedUser.ID, gomock.Any()).
		Return(expectedUpdatedIntegration, nil).
		Times(1)

	// Execute
	result, err := authService.updateExistingUser(context.Background(), existingUser, profile, tokens)

	// Assert
	assert.NoError(err)
	assert.NotNil(result)
	assert.Equal(expectedUpdatedUser.ID, result.ID)
	assert.Equal(expectedUpdatedUser.Email, result.Email)
	assert.Equal(expectedUpdatedUser.Name, result.Name)
	assert.Equal(expectedUpdatedIntegration.SpotifyID, result.SpotifyID)
}

func TestAuthService_UpdateExistingUser_UserUpdateError(t *testing.T) {
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

	existingUser := &models.User{
		ID:    "user123",
		Email: "old@example.com",
		Name:  "Old Name",
	}
	profile := &models.SpotifyUserProfile{
		ID:    "spotify_user_123",
		Email: "new@example.com", // Changed
		Name:  "New Name",        // Changed
	}
	tokens := &models.SpotifyTokenResponse{
		AccessToken: "new_access_token",
		ExpiresIn:   3600,
	}

	// Setup mock expectations
	mockUserRepo.EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Return(nil, repositories.ErrDatabaseOperation).
		Times(1)

	// Execute
	result, err := authService.updateExistingUser(context.Background(), existingUser, profile, tokens)

	// Assert
	assert.Error(err)
	assert.Nil(result)
	assert.Contains(err.Error(), "unable to complete db operation")
}

func TestAuthService_UpdateExistingUser_IntegrationUpdateError(t *testing.T) {
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

	existingUser := &models.User{
		ID:    "user123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	profile := &models.SpotifyUserProfile{
		ID:    "spotify_user_123",
		Email: "test@example.com", // No change
		Name:  "Test User",        // No change
	}
	tokens := &models.SpotifyTokenResponse{
		AccessToken:  "new_access_token",
		RefreshToken: "new_refresh_token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		Scope:        "user-read-private",
	}

	// Setup mock expectations - no user update needed, but integration update fails
	mockSpotifyIntegrationRepo.EXPECT().
		CreateOrUpdate(gomock.Any(), existingUser.ID, gomock.Any()).
		Return(nil, repositories.ErrDatabaseOperation).
		Times(1)

	// Execute
	result, err := authService.updateExistingUser(context.Background(), existingUser, profile, tokens)

	// Assert
	assert.Error(err)
	assert.Nil(result)
	assert.Contains(err.Error(), "unable to complete db operation")
}
