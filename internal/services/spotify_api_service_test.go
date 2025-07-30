package services

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	spotifyclient "github.com/ngomez18/playlist-router/internal/clients/spotify"
	spotifyClientMocks "github.com/ngomez18/playlist-router/internal/clients/spotify/mocks"
	"github.com/ngomez18/playlist-router/internal/models"
	repoMocks "github.com/ngomez18/playlist-router/internal/repositories/mocks"
	"github.com/stretchr/testify/assert"
)

func TestSpotifyAPIService_GetUserPlaylists_Success(t *testing.T) {
	tests := []struct {
		name              string
		userID            string
		integration       *models.SpotifyIntegration
		spotifyPlaylists  []*spotifyclient.SpotifyPlaylist
		expectedPlaylists []*models.SpotifyPlaylist
	}{
		{
			name:   "successful fetch with playlists",
			userID: "user123",
			integration: &models.SpotifyIntegration{
				ID:           "integration123",
				UserID:       "user123",
				SpotifyID:    "spotify_user_123",
				AccessToken:  "valid_access_token",
				RefreshToken: "refresh_token",
				TokenType:    "Bearer",
				ExpiresAt:    time.Now().Add(time.Hour),
			},
			spotifyPlaylists: []*spotifyclient.SpotifyPlaylist{
				{
					ID:     "playlist1",
					Name:   "My Rock Playlist",
					Tracks: &spotifyclient.SpotifyPlaylistTracks{Total: 25},
				},
				{
					ID:     "playlist2",
					Name:   "Jazz Favorites",
					Tracks: &spotifyclient.SpotifyPlaylistTracks{Total: 18},
				},
			},
			expectedPlaylists: []*models.SpotifyPlaylist{
				{ID: "playlist1", Name: "My Rock Playlist", Tracks: 25},
				{ID: "playlist2", Name: "Jazz Favorites", Tracks: 18},
			},
		},
		{
			name:   "successful fetch with empty playlists",
			userID: "user456",
			integration: &models.SpotifyIntegration{
				ID:           "integration456",
				UserID:       "user456",
				SpotifyID:    "spotify_user_456",
				AccessToken:  "valid_access_token",
				RefreshToken: "refresh_token",
				TokenType:    "Bearer",
				ExpiresAt:    time.Now().Add(time.Hour),
			},
			spotifyPlaylists:  []*spotifyclient.SpotifyPlaylist{},
			expectedPlaylists: []*models.SpotifyPlaylist{},
		},
		{
			name:   "successful fetch with nil tracks",
			userID: "user789",
			integration: &models.SpotifyIntegration{
				ID:           "integration789",
				UserID:       "user789",
				SpotifyID:    "spotify_user_789",
				AccessToken:  "valid_access_token",
				RefreshToken: "refresh_token",
				TokenType:    "Bearer",
				ExpiresAt:    time.Now().Add(time.Hour),
			},
			spotifyPlaylists: []*spotifyclient.SpotifyPlaylist{
				{
					ID:     "playlist_no_tracks",
					Name:   "Playlist Without Track Info",
					Tracks: nil,
				},
			},
			expectedPlaylists: []*models.SpotifyPlaylist{
				{ID: "playlist_no_tracks", Name: "Playlist Without Track Info", Tracks: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockIntegrationRepo := repoMocks.NewMockSpotifyIntegrationRepository(ctrl)
			mockSpotifyClient := spotifyClientMocks.NewMockSpotifyAPI(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			service := NewSpotifyAPIService(mockIntegrationRepo, mockSpotifyClient, logger)

			// Mock repository call
			mockIntegrationRepo.EXPECT().
				GetByUserID(gomock.Any(), tt.userID).
				Return(tt.integration, nil).
				Times(1)

			// Mock Spotify client call
			mockSpotifyClient.EXPECT().
				GetAllUserPlaylists(gomock.Any(), tt.integration.AccessToken).
				Return(tt.spotifyPlaylists, nil).
				Times(1)

			ctx := context.Background()
			result, err := service.GetUserPlaylists(ctx, tt.userID)

			assert.NoError(err)
			assert.NotNil(result)
			assert.Equal(len(tt.expectedPlaylists), len(result))

			for i, expected := range tt.expectedPlaylists {
				assert.Equal(expected.ID, result[i].ID)
				assert.Equal(expected.Name, result[i].Name)
				assert.Equal(expected.Tracks, result[i].Tracks)
			}
		})
	}
}

func TestSpotifyAPIService_GetUserPlaylists_RepositoryError(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		repositoryErr error
	}{
		{
			name:          "integration not found",
			userID:        "nonexistent_user",
			repositoryErr: errors.New("spotify integration not found"),
		},
		{
			name:          "database connection error",
			userID:        "user123",
			repositoryErr: errors.New("database connection failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockIntegrationRepo := repoMocks.NewMockSpotifyIntegrationRepository(ctrl)
			mockSpotifyClient := spotifyClientMocks.NewMockSpotifyAPI(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			service := NewSpotifyAPIService(mockIntegrationRepo, mockSpotifyClient, logger)

			// Mock repository error
			mockIntegrationRepo.EXPECT().
				GetByUserID(gomock.Any(), tt.userID).
				Return(nil, tt.repositoryErr).
				Times(1)

			// Spotify client should not be called
			mockSpotifyClient.EXPECT().
				GetAllUserPlaylists(gomock.Any(), gomock.Any()).
				Times(0)

			ctx := context.Background()
			result, err := service.GetUserPlaylists(ctx, tt.userID)

			assert.Error(err)
			assert.Nil(result)
			assert.Contains(err.Error(), "failed to get spotify integration")
		})
	}
}

func TestSpotifyAPIService_GetUserPlaylists_SpotifyClientError(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		integration *models.SpotifyIntegration
		clientErr   error
	}{
		{
			name:   "spotify API unauthorized",
			userID: "user123",
			integration: &models.SpotifyIntegration{
				ID:           "integration123",
				UserID:       "user123",
				AccessToken:  "expired_token",
				RefreshToken: "refresh_token",
				TokenType:    "Bearer",
				ExpiresAt:    time.Now().Add(-time.Hour), // expired
			},
			clientErr: errors.New("spotify API unauthorized (401)"),
		},
		{
			name:   "spotify API network error",
			userID: "user456",
			integration: &models.SpotifyIntegration{
				ID:           "integration456",
				UserID:       "user456",
				AccessToken:  "valid_token",
				RefreshToken: "refresh_token",
				TokenType:    "Bearer",
				ExpiresAt:    time.Now().Add(time.Hour),
			},
			clientErr: errors.New("network timeout"),
		},
		{
			name:   "spotify API rate limit",
			userID: "user789",
			integration: &models.SpotifyIntegration{
				ID:           "integration789",
				UserID:       "user789",
				AccessToken:  "valid_token",
				RefreshToken: "refresh_token",
				TokenType:    "Bearer",
				ExpiresAt:    time.Now().Add(time.Hour),
			},
			clientErr: errors.New("spotify API rate limit exceeded (429)"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockIntegrationRepo := repoMocks.NewMockSpotifyIntegrationRepository(ctrl)
			mockSpotifyClient := spotifyClientMocks.NewMockSpotifyAPI(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			service := NewSpotifyAPIService(mockIntegrationRepo, mockSpotifyClient, logger)

			// Mock successful repository call
			mockIntegrationRepo.EXPECT().
				GetByUserID(gomock.Any(), tt.userID).
				Return(tt.integration, nil).
				Times(1)

			// Mock Spotify client error
			mockSpotifyClient.EXPECT().
				GetAllUserPlaylists(gomock.Any(), tt.integration.AccessToken).
				Return(nil, tt.clientErr).
				Times(1)

			ctx := context.Background()
			result, err := service.GetUserPlaylists(ctx, tt.userID)

			assert.Error(err)
			assert.Nil(result)
			assert.Contains(err.Error(), "failed to fetch user playlists")
		})
	}
}

func TestNewSpotifyAPIService(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockIntegrationRepo := repoMocks.NewMockSpotifyIntegrationRepository(ctrl)
	mockSpotifyClient := spotifyClientMocks.NewMockSpotifyAPI(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := NewSpotifyAPIService(mockIntegrationRepo, mockSpotifyClient, logger)

	assert.NotNil(service)
	assert.Equal(mockIntegrationRepo, service.integrationRepo)
	assert.Equal(mockSpotifyClient, service.spotifyClient)
	assert.NotNil(service.logger)
}
