package services

import (
	"log/slog"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ngomez18/playlist-router/internal/clients/mocks"
	"github.com/pocketbase/pocketbase"
	"github.com/stretchr/testify/require"
)

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
