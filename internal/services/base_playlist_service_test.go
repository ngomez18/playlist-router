package services

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories/mocks"
	"github.com/stretchr/testify/assert"
)

func TestBasePlaylistService_CreateBasePlaylist_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    *models.CreateBasePlaylistRequest
		expected *models.BasePlaylist
	}{
		{
			name: "successful creation with valid input",
			input: &models.CreateBasePlaylistRequest{
				Name:              "My Test Playlist",
				SpotifyPlaylistID: "spotify123",
			},
			expected: &models.BasePlaylist{
				ID:                "playlist123",
				UserID:            "placeholder_user_id",
				Name:              "My Test Playlist",
				SpotifyPlaylistID: "spotify123",
				IsActive:          true,
			},
		},
		{
			name: "successful creation with minimum valid name",
			input: &models.CreateBasePlaylistRequest{
				Name:              "A",
				SpotifyPlaylistID: "spotify456",
			},
			expected: &models.BasePlaylist{
				ID:                "playlist456",
				UserID:            "placeholder_user_id",
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
				Create(ctx, "placeholder_user_id", tt.input.Name, tt.input.SpotifyPlaylistID).
				Return(tt.expected, nil).
				Times(1)

			// Execute
			result, err := service.CreateBasePlaylist(ctx, tt.input)

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
			result, err := service.CreateBasePlaylist(ctx, tt.input)

			// Verify
			assert.Error(err)
			assert.Nil(result)
			assert.Contains(err.Error(), tt.expectedErr)
		})
	}
}

func TestBasePlaylistService_CreateBasePlaylist_NilInput(t *testing.T) {
	assert := assert.New(t)

	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockBasePlaylistRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewBasePlaylistService(mockRepo, logger)

	ctx := context.Background()

	// This test should panic or handle nil input gracefully
	// Since the current implementation doesn't check for nil, this will panic
	// In a real application, you might want to add nil input validation
	assert.Panics(func() {
		_, _ = service.CreateBasePlaylist(ctx, nil)
	})
}

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