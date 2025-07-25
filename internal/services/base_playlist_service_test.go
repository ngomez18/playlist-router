package services

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
	"github.com/ngomez18/playlist-router/internal/repositories/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewBasePlaylistService(t *testing.T) {
	assert := assert.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockBasePlaylistRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	service := NewBasePlaylistService(mockRepo, logger)

	assert.NotNil(service)
	assert.Equal(mockRepo, service.basePlaylistRepo)
	assert.NotNil(service.logger)
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
			assert := assert.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockBasePlaylistRepository(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
			service := NewBasePlaylistService(mockRepo, logger)

			ctx := context.Background()

			// Set expectations
			mockRepo.EXPECT().
				Create(ctx, tt.userId, tt.input.Name, tt.input.SpotifyPlaylistID).
				Return(tt.expected, nil).
				Times(1)

			// Execute
			result, err := service.CreateBasePlaylist(ctx, tt.userId, tt.input)

			// Verify
			assert.NoError(err)
			assert.NotNil(result)
			assert.Equal(tt.expected.ID, result.ID)
			assert.Equal(tt.expected.UserID, result.UserID)
			assert.Equal(tt.expected.Name, result.Name)
			assert.Equal(tt.expected.SpotifyPlaylistID, result.SpotifyPlaylistID)
			assert.Equal(tt.expected.IsActive, result.IsActive)
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
			name: "repository validation error",
			input: &models.CreateBasePlaylistRequest{
				Name:              "",
				SpotifyPlaylistID: "spotify123",
			},
			repositoryErr: errors.New("validation failed: name cannot be blank"),
			expectedErr:   "failed to create playlist: validation failed: name cannot be blank",
		},
		{
			name: "repository database error",
			input: &models.CreateBasePlaylistRequest{
				Name:              "Test Playlist",
				SpotifyPlaylistID: "spotify123",
			},
			repositoryErr: errors.New("database connection failed"),
			expectedErr:   "failed to create playlist: database connection failed",
		},
		{
			name: "repository duplicate error",
			input: &models.CreateBasePlaylistRequest{
				Name:              "Existing Playlist",
				SpotifyPlaylistID: "spotify123",
			},
			repositoryErr: errors.New("playlist with spotify_playlist_id already exists"),
			expectedErr:   "failed to create playlist: playlist with spotify_playlist_id already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockBasePlaylistRepository(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
			service := NewBasePlaylistService(mockRepo, logger)

			ctx := context.Background()

			// Set expectations
			mockRepo.EXPECT().
				Create(ctx, "placeholder_user_id", tt.input.Name, tt.input.SpotifyPlaylistID).
				Return(nil, tt.repositoryErr).
				Times(1)

			// Execute
			result, err := service.CreateBasePlaylist(ctx, "placeholder_user_id", tt.input)

			// Verify
			assert.Error(err)
			assert.Nil(result)
			assert.Contains(err.Error(), tt.expectedErr)
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
			assert := assert.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockBasePlaylistRepository(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
			service := NewBasePlaylistService(mockRepo, logger)

			ctx := context.Background()

			// Set expectations
			mockRepo.EXPECT().
				Delete(ctx, tt.id, tt.userId).
				Return(nil).
				Times(1)

			// Execute
			err := service.DeleteBasePlaylist(ctx, tt.id, tt.userId)

			// Verify
			assert.NoError(err)
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
			name:          "playlist not found",
			id:            "nonexistent123",
			userId:        "user123",
			repositoryErr: repositories.ErrBasePlaylistNotFound,
			expectedErr:   "failed to delete playlist: base playlist not found",
		},
		{
			name:          "unauthorized access",
			id:            "playlist123",
			userId:        "user456",
			repositoryErr: repositories.ErrUnauthorized,
			expectedErr:   "failed to delete playlist: user can not access this resource",
		},
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
			assert := assert.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockBasePlaylistRepository(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
			service := NewBasePlaylistService(mockRepo, logger)

			ctx := context.Background()

			// Set expectations
			mockRepo.EXPECT().
				Delete(ctx, tt.id, tt.userId).
				Return(tt.repositoryErr).
				Times(1)

			// Execute
			err := service.DeleteBasePlaylist(ctx, tt.id, tt.userId)

			// Verify
			assert.Error(err)
			assert.Contains(err.Error(), tt.expectedErr)
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
			assert := assert.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockBasePlaylistRepository(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
			service := NewBasePlaylistService(mockRepo, logger)

			ctx := context.Background()

			// Set expectations
			mockRepo.EXPECT().
				GetByID(ctx, tt.id, tt.userId).
				Return(tt.expected, nil).
				Times(1)

			// Execute
			result, err := service.GetBasePlaylist(ctx, tt.id, tt.userId)

			// Verify
			assert.NoError(err)
			assert.NotNil(result)
			assert.Equal(tt.expected.ID, result.ID)
			assert.Equal(tt.expected.UserID, result.UserID)
			assert.Equal(tt.expected.Name, result.Name)
			assert.Equal(tt.expected.SpotifyPlaylistID, result.SpotifyPlaylistID)
			assert.Equal(tt.expected.IsActive, result.IsActive)
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
			name:          "playlist not found",
			id:            "nonexistent123",
			userId:        "user123",
			repositoryErr: repositories.ErrBasePlaylistNotFound,
			expectedErr:   "failed to retrieve playlist: base playlist not found",
		},
		{
			name:          "unauthorized access",
			id:            "playlist123",
			userId:        "user456",
			repositoryErr: repositories.ErrUnauthorized,
			expectedErr:   "failed to retrieve playlist: user can not access this resource",
		},
		{
			name:          "database error",
			id:            "playlist123",
			userId:        "user123",
			repositoryErr: repositories.ErrDatabaseOperation,
			expectedErr:   "failed to retrieve playlist: unable to complete db operation",
		},
		{
			name:          "collection not found",
			id:            "playlist123",
			userId:        "user123",
			repositoryErr: repositories.ErrCollectionNotFound,
			expectedErr:   "failed to retrieve playlist: collection not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockBasePlaylistRepository(ctrl)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
			service := NewBasePlaylistService(mockRepo, logger)

			ctx := context.Background()

			// Set expectations
			mockRepo.EXPECT().
				GetByID(ctx, tt.id, tt.userId).
				Return(nil, tt.repositoryErr).
				Times(1)

			// Execute
			result, err := service.GetBasePlaylist(ctx, tt.id, tt.userId)

			// Verify
			assert.Error(err)
			assert.Nil(result)
			assert.Contains(err.Error(), tt.expectedErr)
		})
	}
}
