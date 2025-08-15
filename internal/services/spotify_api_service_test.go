package services

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	spotifyclient "github.com/ngomez18/playlist-router/internal/clients/spotify"
	spotifyClientMocks "github.com/ngomez18/playlist-router/internal/clients/spotify/mocks"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewSpotifyAPIService(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpotifyClient := spotifyClientMocks.NewMockSpotifyAPI(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := NewSpotifyAPIService(mockSpotifyClient, logger)

	assert.NotNil(service)
	assert.Equal(mockSpotifyClient, service.spotifyClient)
	assert.NotNil(service.logger)
}

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
			name:              "successful fetch with empty playlists",
			userID:            "user456",
			spotifyPlaylists:  []*spotifyclient.SpotifyPlaylist{},
			expectedPlaylists: []*models.SpotifyPlaylist{},
		},
		{
			name:   "successful fetch with nil tracks",
			userID: "user789",
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

			mockSpotifyClient := spotifyClientMocks.NewMockSpotifyAPI(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			service := NewSpotifyAPIService(mockSpotifyClient, logger)

			// Mock Spotify client call
			mockSpotifyClient.EXPECT().
				GetAllUserPlaylists(gomock.Any()).
				Return(tt.spotifyPlaylists, nil).
				Times(1)

			result, err := service.GetUserPlaylists(context.Background(), tt.userID)

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

func TestSpotifyAPIService_GetUserPlaylists_SpotifyClientError(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		integration *models.SpotifyIntegration
		clientErr   error
	}{
		{
			name:      "spotify API network error",
			userID:    "user456",
			clientErr: errors.New("network timeout"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSpotifyClient := spotifyClientMocks.NewMockSpotifyAPI(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			service := NewSpotifyAPIService(mockSpotifyClient, logger)

			// Mock Spotify client error
			mockSpotifyClient.EXPECT().
				GetAllUserPlaylists(gomock.Any()).
				Return(nil, tt.clientErr).
				Times(1)

			result, err := service.GetUserPlaylists(context.Background(), tt.userID)

			assert.Error(err)
			assert.Nil(result)
			assert.Contains(err.Error(), "failed to fetch user playlists")
		})
	}
}
