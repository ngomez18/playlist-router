package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	spotifyclient "github.com/ngomez18/playlist-router/internal/clients/spotify"
	spotifyClientMocks "github.com/ngomez18/playlist-router/internal/clients/spotify/mocks"
	"github.com/ngomez18/playlist-router/internal/models"
	repositoryMocks "github.com/ngomez18/playlist-router/internal/repositories/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewSpotifyAPIService(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSpotifyClient := spotifyClientMocks.NewMockSpotifyAPI(ctrl)
	mockBasePlaylistRepo := repositoryMocks.NewMockBasePlaylistRepository(ctrl)
	mockChildPlaylistRepo := repositoryMocks.NewMockChildPlaylistRepository(ctrl)
	logger := createTestLogger()

	service := NewSpotifyAPIService(mockSpotifyClient, mockBasePlaylistRepo, mockChildPlaylistRepo, logger)

	assert.NotNil(service)
	assert.Equal(mockSpotifyClient, service.spotifyClient)
	assert.Equal(mockBasePlaylistRepo, service.basePlaylistRepo)
	assert.Equal(mockChildPlaylistRepo, service.childPlaylistRepo)
	assert.NotNil(service.logger)
}

func TestSpotifyAPIService_GetFilteredUserPlaylists_Success(t *testing.T) {
	tests := []struct {
		name              string
		userID            string
		spotifyPlaylists  []*spotifyclient.SpotifyPlaylist
		basePlaylists     []*models.BasePlaylist
		childPlaylistsMap map[string][]*models.ChildPlaylist
		expectedPlaylists []*models.SpotifyPlaylist
	}{
		{
			name:   "successful fetch with no existing playlists",
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
			basePlaylists:     []*models.BasePlaylist{},
			childPlaylistsMap: map[string][]*models.ChildPlaylist{},
			expectedPlaylists: []*models.SpotifyPlaylist{
				{ID: "playlist1", Name: "My Rock Playlist", Tracks: 25},
				{ID: "playlist2", Name: "Jazz Favorites", Tracks: 18},
			},
		},
		{
			name:   "successful fetch with base playlist filtered out",
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
			basePlaylists: []*models.BasePlaylist{
				{ID: "base1", SpotifyPlaylistID: "playlist1", UserID: "user123", Name: "My Rock Playlist"},
			},
			childPlaylistsMap: map[string][]*models.ChildPlaylist{
				"base1": {},
			},
			expectedPlaylists: []*models.SpotifyPlaylist{
				{ID: "playlist2", Name: "Jazz Favorites", Tracks: 18},
			},
		},
		{
			name:   "successful fetch with child playlist filtered out",
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
			basePlaylists: []*models.BasePlaylist{
				{ID: "base1", SpotifyPlaylistID: "playlist1", UserID: "user123", Name: "My Rock Playlist"},
			},
			childPlaylistsMap: map[string][]*models.ChildPlaylist{
				"base1": {
					{ID: "child1", SpotifyPlaylistID: "playlist2", UserID: "user123", BasePlaylistID: "base1", Name: "Jazz Child"},
				},
			},
			expectedPlaylists: []*models.SpotifyPlaylist{},
		},
		{
			name:              "successful fetch with empty playlists",
			userID:            "user456",
			spotifyPlaylists:  []*spotifyclient.SpotifyPlaylist{},
			basePlaylists:     []*models.BasePlaylist{},
			childPlaylistsMap: map[string][]*models.ChildPlaylist{},
			expectedPlaylists: []*models.SpotifyPlaylist{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSpotifyClient := spotifyClientMocks.NewMockSpotifyAPI(ctrl)
			mockBasePlaylistRepo := repositoryMocks.NewMockBasePlaylistRepository(ctrl)
			mockChildPlaylistRepo := repositoryMocks.NewMockChildPlaylistRepository(ctrl)
			logger := createTestLogger()

			service := NewSpotifyAPIService(mockSpotifyClient, mockBasePlaylistRepo, mockChildPlaylistRepo, logger)

			mockSpotifyClient.EXPECT().
				GetAllUserPlaylists(gomock.Any()).
				Return(tt.spotifyPlaylists, nil).
				Times(1)

			mockBasePlaylistRepo.EXPECT().
				GetByUserID(gomock.Any(), tt.userID).
				Return(tt.basePlaylists, nil).
				Times(1)

			for _, basePlaylist := range tt.basePlaylists {
				childPlaylists := tt.childPlaylistsMap[basePlaylist.ID]
				mockChildPlaylistRepo.EXPECT().
					GetByBasePlaylistID(gomock.Any(), basePlaylist.ID, tt.userID).
					Return(childPlaylists, nil).
					Times(1)
			}

			result, err := service.GetFilteredUserPlaylists(context.Background(), tt.userID)

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

func TestSpotifyAPIService_GetFilteredUserPlaylists_Errors(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		spotifyErr    error
		baseErr       error
		childErr      error
		expectedError string
	}{
		{
			name:          "spotify API network error",
			userID:        "user456",
			spotifyErr:    errors.New("network timeout"),
			expectedError: "network timeout",
		},
		{
			name:          "base playlist repository error",
			userID:        "user789",
			baseErr:       errors.New("database connection failed"),
			expectedError: "failed to fetch base playlists",
		},
		{
			name:          "child playlist repository error",
			userID:        "user101",
			childErr:      errors.New("database query failed"),
			expectedError: "failed to fetch child playlists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSpotifyClient := spotifyClientMocks.NewMockSpotifyAPI(ctrl)
			mockBasePlaylistRepo := repositoryMocks.NewMockBasePlaylistRepository(ctrl)
			mockChildPlaylistRepo := repositoryMocks.NewMockChildPlaylistRepository(ctrl)
			logger := createTestLogger()

			service := NewSpotifyAPIService(mockSpotifyClient, mockBasePlaylistRepo, mockChildPlaylistRepo, logger)

			if tt.spotifyErr != nil {
				// Mock Spotify client error
				mockSpotifyClient.EXPECT().
					GetAllUserPlaylists(gomock.Any()).
					Return(nil, tt.spotifyErr).
					Times(1)
			} else if tt.baseErr != nil {
				mockSpotifyClient.EXPECT().
					GetAllUserPlaylists(gomock.Any()).
					Return([]*spotifyclient.SpotifyPlaylist{}, nil).
					Times(1)
				
				mockBasePlaylistRepo.EXPECT().
					GetByUserID(gomock.Any(), tt.userID).
					Return(nil, tt.baseErr).
					Times(1)
			} else if tt.childErr != nil {
				basePlaylists := []*models.BasePlaylist{
					{ID: "base1", SpotifyPlaylistID: "playlist1", UserID: tt.userID, Name: "Test Base"},
				}
				
				mockSpotifyClient.EXPECT().
					GetAllUserPlaylists(gomock.Any()).
					Return([]*spotifyclient.SpotifyPlaylist{}, nil).
					Times(1)
				
				mockBasePlaylistRepo.EXPECT().
					GetByUserID(gomock.Any(), tt.userID).
					Return(basePlaylists, nil).
					Times(1)
				
				mockChildPlaylistRepo.EXPECT().
					GetByBasePlaylistID(gomock.Any(), "base1", tt.userID).
					Return(nil, tt.childErr).
					Times(1)
			}

			result, err := service.GetFilteredUserPlaylists(context.Background(), tt.userID)

			assert.Error(err)
			assert.Nil(result)
			assert.Contains(err.Error(), tt.expectedError)
		})
	}
}
