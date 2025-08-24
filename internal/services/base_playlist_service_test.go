package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	spotifyMocks "github.com/ngomez18/playlist-router/internal/clients/spotify/mocks"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
	"github.com/ngomez18/playlist-router/internal/repositories/mocks"
	"github.com/stretchr/testify/require"
)

func TestNewBasePlaylistService(t *testing.T) {
	require := require.New(t)

	ctrl := setupMockController(t)

	mockRepo := mocks.NewMockBasePlaylistRepository(ctrl)
	mockSpotifyIntegrationRepo := mocks.NewMockSpotifyIntegrationRepository(ctrl)
	mockSpotifyClient := spotifyMocks.NewMockSpotifyAPI(ctrl)
	logger := createTestLogger()

	service := NewBasePlaylistService(mockRepo, mockSpotifyIntegrationRepo, mockSpotifyClient, logger)

	require.NotNil(service)
	require.Equal(mockRepo, service.basePlaylistRepo)
	require.Equal(mockSpotifyIntegrationRepo, service.spotifyIntegrationRepo)
	require.Equal(mockSpotifyClient, service.spotifyClient)
	require.NotNil(service.logger)
}

func TestBasePlaylistService_CreateBasePlaylist_Success(t *testing.T) {
	tests := []struct {
		name     string
		userId   string
		input    *models.CreateBasePlaylistRequest
		expected *models.BasePlaylist
	}{
		{
			name:   "successful creation with valid input",
			userId: "user123",
			input: &models.CreateBasePlaylistRequest{
				Name:              "My Test Playlist",
				SpotifyPlaylistID: "spotify123",
			},
			expected: &models.BasePlaylist{
				ID:                "playlist123",
				UserID:            "user123",
				Name:              "My Test Playlist",
				SpotifyPlaylistID: "spotify123",
				IsActive:          true,
			},
		},
		{
			name:   "successful creation with minimum valid name",
			userId: "user456",
			input: &models.CreateBasePlaylistRequest{
				Name:              "A",
				SpotifyPlaylistID: "spotify456",
			},
			expected: &models.BasePlaylist{
				ID:                "playlist456",
				UserID:            "user456",
				Name:              "A",
				SpotifyPlaylistID: "spotify456",
				IsActive:          true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockBasePlaylistRepository(ctrl)
			mockSpotifyIntegrationRepo := mocks.NewMockSpotifyIntegrationRepository(ctrl)
			mockSpotifyClient := spotifyMocks.NewMockSpotifyAPI(ctrl)
			logger := createTestLogger()
			service := NewBasePlaylistService(mockRepo, mockSpotifyIntegrationRepo, mockSpotifyClient, logger)

			ctx := context.Background()

			// Set expectations
			mockRepo.EXPECT().
				Create(ctx, tt.userId, tt.input.Name, tt.input.SpotifyPlaylistID).
				Return(tt.expected, nil).
				Times(1)

			// Execute
			result, err := service.CreateBasePlaylist(ctx, tt.userId, tt.input)

			// Verify
			require.NoError(err)
			require.NotNil(result)
			require.Equal(tt.expected.ID, result.ID)
			require.Equal(tt.expected.UserID, result.UserID)
			require.Equal(tt.expected.Name, result.Name)
			require.Equal(tt.expected.SpotifyPlaylistID, result.SpotifyPlaylistID)
			require.Equal(tt.expected.IsActive, result.IsActive)
		})
	}
}

func TestBasePlaylistService_CreateBasePlaylist_RepositoryError(t *testing.T) {
	tests := []struct {
		name          string
		input         *models.CreateBasePlaylistRequest
		repositoryErr error
		expectedErr   string
	}{
		{
			name: "repository database error",
			input: &models.CreateBasePlaylistRequest{
				Name:              "Test Playlist",
				SpotifyPlaylistID: "spotify123",
			},
			repositoryErr: errors.New("database connection failed"),
			expectedErr:   "failed to create playlist: database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockBasePlaylistRepository(ctrl)
			mockSpotifyIntegrationRepo := mocks.NewMockSpotifyIntegrationRepository(ctrl)
			mockSpotifyClient := spotifyMocks.NewMockSpotifyAPI(ctrl)
			logger := createTestLogger()
			service := NewBasePlaylistService(mockRepo, mockSpotifyIntegrationRepo, mockSpotifyClient, logger)

			ctx := context.Background()

			// Set expectations
			mockRepo.EXPECT().
				Create(ctx, "placeholder_user_id", tt.input.Name, tt.input.SpotifyPlaylistID).
				Return(nil, tt.repositoryErr).
				Times(1)

			// Execute
			result, err := service.CreateBasePlaylist(ctx, "placeholder_user_id", tt.input)

			// Verify
			require.Error(err)
			require.Nil(result)
			require.Contains(err.Error(), tt.expectedErr)
		})
	}
}

func TestBasePlaylistService_DeleteBasePlaylist_Success(t *testing.T) {
	tests := []struct {
		name   string
		id     string
		userId string
	}{
		{
			name:   "successful deletion with valid id",
			id:     "playlist123",
			userId: "user123",
		},
		{
			name:   "successful deletion with different id format",
			id:     "pl_abc123def456",
			userId: "user456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockBasePlaylistRepository(ctrl)
			mockSpotifyIntegrationRepo := mocks.NewMockSpotifyIntegrationRepository(ctrl)
			mockSpotifyClient := spotifyMocks.NewMockSpotifyAPI(ctrl)
			logger := createTestLogger()
			service := NewBasePlaylistService(mockRepo, mockSpotifyIntegrationRepo, mockSpotifyClient, logger)

			ctx := context.Background()

			// Set expectations
			mockRepo.EXPECT().
				Delete(ctx, tt.id, tt.userId).
				Return(nil).
				Times(1)

			// Execute
			err := service.DeleteBasePlaylist(ctx, tt.id, tt.userId)

			// Verify
			require.NoError(err)
		})
	}
}

func TestBasePlaylistService_DeleteBasePlaylist_RepositoryErrors(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		userId        string
		repositoryErr error
		expectedErr   string
	}{
		{
			name:          "database error",
			id:            "playlist123",
			userId:        "user123",
			repositoryErr: repositories.ErrDatabaseOperation,
			expectedErr:   "failed to delete playlist: unable to complete db operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockBasePlaylistRepository(ctrl)
			mockSpotifyIntegrationRepo := mocks.NewMockSpotifyIntegrationRepository(ctrl)
			mockSpotifyClient := spotifyMocks.NewMockSpotifyAPI(ctrl)
			logger := createTestLogger()
			service := NewBasePlaylistService(mockRepo, mockSpotifyIntegrationRepo, mockSpotifyClient, logger)

			ctx := context.Background()

			// Set expectations
			mockRepo.EXPECT().
				Delete(ctx, tt.id, tt.userId).
				Return(tt.repositoryErr).
				Times(1)

			// Execute
			err := service.DeleteBasePlaylist(ctx, tt.id, tt.userId)

			// Verify
			require.Error(err)
			require.Contains(err.Error(), tt.expectedErr)
		})
	}
}

func TestBasePlaylistService_GetBasePlaylist_Success(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		userId   string
		expected *models.BasePlaylist
	}{
		{
			name:   "successful retrieval with valid id",
			id:     "playlist123",
			userId: "user123",
			expected: &models.BasePlaylist{
				ID:                "playlist123",
				UserID:            "user123",
				Name:              "My Test Playlist",
				SpotifyPlaylistID: "spotify123",
				IsActive:          true,
			},
		},
		{
			name:   "successful retrieval with different user",
			id:     "playlist456",
			userId: "user456",
			expected: &models.BasePlaylist{
				ID:                "playlist456",
				UserID:            "user456",
				Name:              "Another Playlist",
				SpotifyPlaylistID: "spotify456",
				IsActive:          false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockBasePlaylistRepository(ctrl)
			mockSpotifyIntegrationRepo := mocks.NewMockSpotifyIntegrationRepository(ctrl)
			mockSpotifyClient := spotifyMocks.NewMockSpotifyAPI(ctrl)
			logger := createTestLogger()
			service := NewBasePlaylistService(mockRepo, mockSpotifyIntegrationRepo, mockSpotifyClient, logger)

			ctx := context.Background()

			// Set expectations
			mockRepo.EXPECT().
				GetByID(ctx, tt.id, tt.userId).
				Return(tt.expected, nil).
				Times(1)

			// Execute
			result, err := service.GetBasePlaylist(ctx, tt.id, tt.userId)

			// Verify
			require.NoError(err)
			require.NotNil(result)
			require.Equal(tt.expected.ID, result.ID)
			require.Equal(tt.expected.UserID, result.UserID)
			require.Equal(tt.expected.Name, result.Name)
			require.Equal(tt.expected.SpotifyPlaylistID, result.SpotifyPlaylistID)
			require.Equal(tt.expected.IsActive, result.IsActive)
		})
	}
}

func TestBasePlaylistService_GetBasePlaylist_RepositoryErrors(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		userId        string
		repositoryErr error
		expectedErr   string
	}{
		{
			name:          "database error",
			id:            "playlist123",
			userId:        "user123",
			repositoryErr: repositories.ErrDatabaseOperation,
			expectedErr:   "failed to retrieve playlist: unable to complete db operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockBasePlaylistRepository(ctrl)
			mockSpotifyIntegrationRepo := mocks.NewMockSpotifyIntegrationRepository(ctrl)
			mockSpotifyClient := spotifyMocks.NewMockSpotifyAPI(ctrl)
			logger := createTestLogger()
			service := NewBasePlaylistService(mockRepo, mockSpotifyIntegrationRepo, mockSpotifyClient, logger)

			ctx := context.Background()

			// Set expectations
			mockRepo.EXPECT().
				GetByID(ctx, tt.id, tt.userId).
				Return(nil, tt.repositoryErr).
				Times(1)

			// Execute
			result, err := service.GetBasePlaylist(ctx, tt.id, tt.userId)

			// Verify
			require.Error(err)
			require.Nil(result)
			require.Contains(err.Error(), tt.expectedErr)
		})
	}
}

func TestBasePlaylistService_GetBasePlaylistsByUserID_Success(t *testing.T) {
	tests := []struct {
		name          string
		userId        string
		mockPlaylists []*models.BasePlaylist
		expectedCount int
	}{
		{
			name:   "user with multiple playlists",
			userId: "user123",
			mockPlaylists: []*models.BasePlaylist{
				{
					ID:                "playlist1",
					UserID:            "user123",
					Name:              "First Playlist",
					SpotifyPlaylistID: "spotify1",
					IsActive:          true,
				},
				{
					ID:                "playlist2",
					UserID:            "user123",
					Name:              "Second Playlist",
					SpotifyPlaylistID: "spotify2",
					IsActive:          true,
				},
				{
					ID:                "playlist3",
					UserID:            "user123",
					Name:              "Third Playlist",
					SpotifyPlaylistID: "spotify3",
					IsActive:          true,
				},
			},
			expectedCount: 3,
		},
		{
			name:   "user with single playlist",
			userId: "user456",
			mockPlaylists: []*models.BasePlaylist{
				{
					ID:                "playlist4",
					UserID:            "user456",
					Name:              "Only Playlist",
					SpotifyPlaylistID: "spotify4",
					IsActive:          true,
				},
			},
			expectedCount: 1,
		},
		{
			name:          "user with no playlists",
			userId:        "user789",
			mockPlaylists: []*models.BasePlaylist{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockBasePlaylistRepository(ctrl)
			mockSpotifyIntegrationRepo := mocks.NewMockSpotifyIntegrationRepository(ctrl)
			mockSpotifyClient := spotifyMocks.NewMockSpotifyAPI(ctrl)
			logger := createTestLogger()
			service := NewBasePlaylistService(mockRepo, mockSpotifyIntegrationRepo, mockSpotifyClient, logger)

			ctx := context.Background()

			// Set expectations
			mockRepo.EXPECT().
				GetByUserID(ctx, tt.userId).
				Return(tt.mockPlaylists, nil).
				Times(1)

			// Execute
			result, err := service.GetBasePlaylistsByUserID(ctx, tt.userId)

			// Verify
			require.NoError(err)
			require.NotNil(result)
			require.Len(result, tt.expectedCount)

			// Verify each playlist matches expected data
			for i, playlist := range result {
				require.Equal(tt.mockPlaylists[i].ID, playlist.ID)
				require.Equal(tt.mockPlaylists[i].UserID, playlist.UserID)
				require.Equal(tt.mockPlaylists[i].Name, playlist.Name)
				require.Equal(tt.mockPlaylists[i].SpotifyPlaylistID, playlist.SpotifyPlaylistID)
				require.Equal(tt.mockPlaylists[i].IsActive, playlist.IsActive)
			}
		})
	}
}

func TestBasePlaylistService_GetBasePlaylistsByUserID_RepositoryErrors(t *testing.T) {
	tests := []struct {
		name          string
		userId        string
		repositoryErr error
		expectedErr   string
	}{
		{
			name:          "database operation error",
			userId:        "user123",
			repositoryErr: repositories.ErrDatabaseOperation,
			expectedErr:   "failed to retrieve playlists: unable to complete db operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockBasePlaylistRepository(ctrl)
			mockSpotifyIntegrationRepo := mocks.NewMockSpotifyIntegrationRepository(ctrl)
			mockSpotifyClient := spotifyMocks.NewMockSpotifyAPI(ctrl)
			logger := createTestLogger()
			service := NewBasePlaylistService(mockRepo, mockSpotifyIntegrationRepo, mockSpotifyClient, logger)

			ctx := context.Background()

			// Set expectations
			mockRepo.EXPECT().
				GetByUserID(ctx, tt.userId).
				Return(nil, tt.repositoryErr).
				Times(1)

			// Execute
			result, err := service.GetBasePlaylistsByUserID(ctx, tt.userId)

			// Verify
			require.Error(err)
			require.Nil(result)
			require.Contains(err.Error(), tt.expectedErr)
		})
	}
}
