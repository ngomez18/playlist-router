package services

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ngomez18/playlist-router/internal/clients/mocks"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/pocketbase/pocketbase"
	"github.com/stretchr/testify/require"
)

func TestAuthService_GenerateSpotifyAuthURL(t *testing.T) {
	assert := require.New(t)
	
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpotifyClient := mocks.NewMockSpotifyAPI(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	app := &pocketbase.PocketBase{} // Mock or minimal instance

	authService := NewAuthService(app, mockSpotifyClient, logger)

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

func TestAuthService_HandleSpotifyCallback(t *testing.T) {
	tests := []struct {
		name                  string
		code                  string
		state                 string
		mockTokenResponse     *models.SpotifyTokenResponse
		tokenExchangeError    error
		mockProfileResponse   *models.SpotifyUserProfile
		profileFetchError     error
		expectError           bool
		expectedErrorContains string
		expectedSpotifyID     string
		expectedEmail         string
	}{
		{
			name:  "successful callback handling",
			code:  "auth_code_123",
			state: "state_123",
			mockTokenResponse: &models.SpotifyTokenResponse{
				AccessToken:  "access_token_123",
				TokenType:    "Bearer",
				Scope:        "user-read-email",
				ExpiresIn:    3600,
				RefreshToken: "refresh_token_123",
			},
			tokenExchangeError: nil,
			mockProfileResponse: &models.SpotifyUserProfile{
				ID:    "spotify_user_123",
				Email: "test@example.com",
				Name:  "Test User",
			},
			profileFetchError: nil,
			expectError:       false,
			expectedSpotifyID: "spotify_user_123",
			expectedEmail:     "test@example.com",
		},
		{
			name:                  "token exchange failure",
			code:                  "invalid_code",
			state:                 "state_123",
			mockTokenResponse:     nil,
			tokenExchangeError:    errors.New("invalid authorization code"),
			expectError:           true,
			expectedErrorContains: "failed to exchange code for tokens",
		},
		{
			name:  "profile fetch failure",
			code:  "auth_code_123",
			state: "state_123",
			mockTokenResponse: &models.SpotifyTokenResponse{
				AccessToken:  "access_token_123",
				TokenType:    "Bearer",
				Scope:        "user-read-email",
				ExpiresIn:    3600,
				RefreshToken: "refresh_token_123",
			},
			tokenExchangeError:    nil,
			mockProfileResponse:   nil,
			profileFetchError:     errors.New("invalid access token"),
			expectError:           true,
			expectedErrorContains: "failed to get user profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)
			
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSpotifyClient := mocks.NewMockSpotifyAPI(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			app := &pocketbase.PocketBase{} // Mock or minimal instance

			authService := NewAuthService(app, mockSpotifyClient, logger)
			ctx := context.Background()

			// Setup mock expectations for token exchange
			mockSpotifyClient.EXPECT().
				ExchangeCodeForTokens(ctx, tt.code).
				Return(tt.mockTokenResponse, tt.tokenExchangeError).
				Times(1)

			// Setup mock expectations for profile fetch (only if token exchange succeeds)
			if tt.tokenExchangeError == nil {
				mockSpotifyClient.EXPECT().
					GetUserProfile(ctx, tt.mockTokenResponse.AccessToken).
					Return(tt.mockProfileResponse, tt.profileFetchError).
					Times(1)
			}

			// Execute
			result, err := authService.HandleSpotifyCallback(ctx, tt.code, tt.state)

			// Assert
			if tt.expectError {
				assert.Error(err)
				assert.Nil(result)
				if tt.expectedErrorContains != "" {
					assert.Contains(err.Error(), tt.expectedErrorContains)
				}
			} else {
				assert.NoError(err)
				assert.NotNil(result)
				assert.NotNil(result.User)
				assert.Equal(tt.expectedSpotifyID, result.User.SpotifyID)
				assert.Equal(tt.expectedEmail, result.User.Email)
				assert.Equal("Test User", result.User.Name)
				assert.Equal("temp_id", result.User.ID)  // Current placeholder
				assert.Equal("temp_token", result.Token) // Current placeholder
				assert.Equal("", result.RefreshToken)    // PocketBase handles its own refresh
			}
		})
	}
}

func TestAuthService_createOrUpdateUser(t *testing.T) {
	assert := require.New(t)
	
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpotifyClient := mocks.NewMockSpotifyAPI(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	app := &pocketbase.PocketBase{} // Mock or minimal instance

	authService := NewAuthService(app, mockSpotifyClient, logger)
	ctx := context.Background()

	profile := &models.SpotifyUserProfile{
		ID:    "spotify_user_123",
		Email: "test@example.com",
		Name:  "Test User",
	}

	tokens := &models.SpotifyTokenResponse{
		AccessToken:  "access_token_123",
		TokenType:    "Bearer",
		Scope:        "user-read-email",
		ExpiresIn:    3600,
		RefreshToken: "refresh_token_123",
	}

	// Execute
	user, pbToken, err := authService.createOrUpdateUser(ctx, profile, tokens)

	// Assert - Currently returns placeholder values
	assert.NoError(err)
	assert.NotNil(user)
	assert.Equal("temp_id", user.ID)
	assert.Equal("test@example.com", user.Email)
	assert.Equal("Test User", user.Name)
	assert.Equal("spotify_user_123", user.SpotifyID)
	assert.Equal("temp_token", pbToken)
}

func TestNewAuthService(t *testing.T) {
	assert := require.New(t)
	
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpotifyClient := mocks.NewMockSpotifyAPI(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	app := &pocketbase.PocketBase{}

	// Execute
	authService := NewAuthService(app, mockSpotifyClient, logger)

	// Assert
	assert.NotNil(authService)
	assert.Equal(app, authService.app)
	assert.Equal(mockSpotifyClient, authService.spotifyClient)
	assert.NotNil(authService.logger)
}

// Test helper to verify the service properly uses context cancellation
func TestAuthService_HandleSpotifyCallback_ContextCancellation(t *testing.T) {
	assert := require.New(t)
	
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpotifyClient := mocks.NewMockSpotifyAPI(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	app := &pocketbase.PocketBase{}

	authService := NewAuthService(app, mockSpotifyClient, logger)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Setup mock to return context cancellation error
	mockSpotifyClient.EXPECT().
		ExchangeCodeForTokens(ctx, "test_code").
		Return(nil, context.Canceled).
		Times(1)

	// Execute
	result, err := authService.HandleSpotifyCallback(ctx, "test_code", "test_state")

	// Assert
	assert.Error(err)
	assert.Nil(result)
	assert.Contains(err.Error(), "failed to exchange code for tokens")
}