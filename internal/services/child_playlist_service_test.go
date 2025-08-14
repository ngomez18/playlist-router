package services

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	spotifyclient "github.com/ngomez18/playlist-router/internal/clients/spotify"
	spotifyMocks "github.com/ngomez18/playlist-router/internal/clients/spotify/mocks"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
	repoMocks "github.com/ngomez18/playlist-router/internal/repositories/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewChildPlaylistService(t *testing.T) {
	assert := assert.New(t)

	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockChildRepo := repoMocks.NewMockChildPlaylistRepository(ctrl)
	mockBaseRepo := repoMocks.NewMockBasePlaylistRepository(ctrl)
	mockSpotifyIntegrationRepo := repoMocks.NewMockSpotifyIntegrationRepository(ctrl)
	mockSpotifyClient := spotifyMocks.NewMockSpotifyAPI(ctrl)

	// Execute
	service := createTestService(mockChildRepo, mockBaseRepo, mockSpotifyIntegrationRepo, mockSpotifyClient)

	// Assert
	assert.NotNil(service)
	assert.Equal(mockChildRepo, service.childPlaylistRepo)
	assert.Equal(mockBaseRepo, service.basePlaylistRepo)
	assert.Equal(mockSpotifyIntegrationRepo, service.spotifyIntegrationRepo)
	assert.Equal(mockSpotifyClient, service.spotifyClient)
	assert.NotNil(service.logger)
}

func TestChildPlaylistService_CreateChildPlaylist_Success(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mocks
	mockChildRepo := repoMocks.NewMockChildPlaylistRepository(ctrl)
	mockBaseRepo := repoMocks.NewMockBasePlaylistRepository(ctrl)
	mockSpotifyIntegrationRepo := repoMocks.NewMockSpotifyIntegrationRepository(ctrl)
	mockSpotifyClient := spotifyMocks.NewMockSpotifyAPI(ctrl)
	service := createTestService(mockChildRepo, mockBaseRepo, mockSpotifyIntegrationRepo, mockSpotifyClient)

	// Test Data
	userID := "user123"
	basePlaylistID := "basePlaylist456"
	input := &models.CreateChildPlaylistRequest{
		Name:        "Child Playlist Name",
		Description: "Child playlist description.",
		FilterRules: &models.AudioFeatureFilters{},
	}
	basePlaylist := &models.BasePlaylist{
		ID:   basePlaylistID,
		Name: "Base Playlist Name",
	}
	spotifyPlaylist := &spotifyclient.SpotifyPlaylist{
		ID:   "new_spotify_playlist_id",
		Name: models.BuildChildPlaylistName(basePlaylist.Name, input.Name),
	}
	expectedChildPlaylist := &models.ChildPlaylist{
		ID:                "childPlaylist789",
		UserID:            userID,
		BasePlaylistID:    basePlaylistID,
		Name:              input.Name,
		Description:       input.Description,
		SpotifyPlaylistID: spotifyPlaylist.ID,
	}

	// Mock Calls
	mockBaseRepo.EXPECT().GetByID(gomock.Any(), basePlaylistID, userID).Return(basePlaylist, nil)
	expectedPlaylistName := models.BuildChildPlaylistName(basePlaylist.Name, input.Name)
	expectedDescription := models.BuildChildPlaylistDescription(input.Description)
	mockSpotifyClient.EXPECT().CreatePlaylist(
		gomock.Any(),
		expectedPlaylistName,
		expectedDescription,
		false,
	).Return(spotifyPlaylist, nil)
	mockChildRepo.EXPECT().Create(
		gomock.Any(),
		repositories.CreateChildPlaylistFields{
			UserID:            userID,
			BasePlaylistID:    basePlaylistID,
			Name:              input.Name,
			Description:       input.Description,
			SpotifyPlaylistID: spotifyPlaylist.ID,
			FilterRules:       input.FilterRules,
			IsActive:          true,
		},
	).Return(expectedChildPlaylist, nil)

	// Execution
	result, err := service.CreateChildPlaylist(context.Background(), userID, basePlaylistID, input)

	// Assertions
	assert.NoError(err)
	assert.NotNil(result)
	assert.Equal(expectedChildPlaylist, result)
}

func TestChildPlaylistService_CreateChildPlaylist_GetBasePlaylistError(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBaseRepo := repoMocks.NewMockBasePlaylistRepository(ctrl)
	mockBaseRepo.EXPECT().GetByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
	service := createTestService(nil, mockBaseRepo, nil, nil)

	_, err := service.CreateChildPlaylist(context.Background(), "uid", "bpid", &models.CreateChildPlaylistRequest{Name: "Test"})

	assert.Error(err)
	assert.Contains(err.Error(), "failed to get base playlist")
}

func TestChildPlaylistService_CreateChildPlaylist_SpotifyError(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBaseRepo := repoMocks.NewMockBasePlaylistRepository(ctrl)
	mockSpotifyClient := spotifyMocks.NewMockSpotifyAPI(ctrl)
	mockBaseRepo.EXPECT().GetByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.BasePlaylist{Name: "Base"}, nil)
	mockSpotifyClient.EXPECT().CreatePlaylist(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("spotify api error"))
	service := createTestService(nil, mockBaseRepo, nil, mockSpotifyClient)

	_, err := service.CreateChildPlaylist(context.Background(), "uid", "bpid", &models.CreateChildPlaylistRequest{Name: "Test"})

	assert.Error(err)
	assert.Contains(err.Error(), "failed to create spotify playlist")
}

func TestChildPlaylistService_CreateChildPlaylist_RepoError(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChildRepo := repoMocks.NewMockChildPlaylistRepository(ctrl)
	mockBaseRepo := repoMocks.NewMockBasePlaylistRepository(ctrl)
	mockSpotifyClient := spotifyMocks.NewMockSpotifyAPI(ctrl)
	mockBaseRepo.EXPECT().GetByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.BasePlaylist{Name: "Base"}, nil)
	mockSpotifyClient.EXPECT().CreatePlaylist(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&spotifyclient.SpotifyPlaylist{ID: "sp_id"}, nil)
	mockChildRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
	service := createTestService(mockChildRepo, mockBaseRepo, nil, mockSpotifyClient)

	_, err := service.CreateChildPlaylist(context.Background(), "uid", "bpid", &models.CreateChildPlaylistRequest{Name: "Test"})

	assert.Error(err)
	assert.Contains(err.Error(), "failed to create child playlist")
}

func TestChildPlaylistService_DeleteChildPlaylist_Success(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mocks
	mockChildRepo := repoMocks.NewMockChildPlaylistRepository(ctrl)
	mockSpotifyClient := spotifyMocks.NewMockSpotifyAPI(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewChildPlaylistService(mockChildRepo, nil, nil, mockSpotifyClient, logger)

	// Test Data
	userID := "user123"
	childPlaylistID := "childPlaylist789"
	childPlaylist := &models.ChildPlaylist{
		ID:                childPlaylistID,
		SpotifyPlaylistID: "spotify_playlist_to_delete",
	}

	// Mock Calls
	mockChildRepo.EXPECT().GetByID(gomock.Any(), childPlaylistID, userID).Return(childPlaylist, nil)
	mockSpotifyClient.EXPECT().DeletePlaylist(gomock.Any(), childPlaylist.SpotifyPlaylistID).Return(nil)
	mockChildRepo.EXPECT().Delete(gomock.Any(), childPlaylistID, userID).Return(nil)

	// Execution
	err := service.DeleteChildPlaylist(context.Background(), childPlaylistID, userID)

	// Assertions
	assert.NoError(err)
}

func TestChildPlaylistService_DeleteChildPlaylist_GetByIDError(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChildRepo := repoMocks.NewMockChildPlaylistRepository(ctrl)
	mockChildRepo.EXPECT().GetByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
	service := createTestService(mockChildRepo, nil, nil, nil)

	err := service.DeleteChildPlaylist(context.Background(), "cpid", "uid")

	assert.Error(err)
	assert.Contains(err.Error(), "failed to get child playlist")
}

func TestChildPlaylistService_DeleteChildPlaylist_SpotifyError(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChildRepo := repoMocks.NewMockChildPlaylistRepository(ctrl)
	mockSpotifyClient := spotifyMocks.NewMockSpotifyAPI(ctrl)
	mockChildRepo.EXPECT().GetByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.ChildPlaylist{}, nil)
	mockSpotifyClient.EXPECT().DeletePlaylist(gomock.Any(), gomock.Any()).Return(errors.New("spotify api error"))
	service := createTestService(mockChildRepo, nil, nil, mockSpotifyClient)

	err := service.DeleteChildPlaylist(context.Background(), "cpid", "uid")

	assert.Error(err)
	assert.Contains(err.Error(), "failed to delete spotify playlist")
}

func TestChildPlaylistService_DeleteChildPlaylist_RepoError(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChildRepo := repoMocks.NewMockChildPlaylistRepository(ctrl)
	mockSpotifyClient := spotifyMocks.NewMockSpotifyAPI(ctrl)
	mockChildRepo.EXPECT().GetByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.ChildPlaylist{}, nil)
	mockSpotifyClient.EXPECT().DeletePlaylist(gomock.Any(), gomock.Any()).Return(nil)
	mockChildRepo.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("db error"))
	service := createTestService(mockChildRepo, nil, nil, mockSpotifyClient)

	err := service.DeleteChildPlaylist(context.Background(), "cpid", "uid")

	assert.Error(err)
	assert.Contains(err.Error(), "failed to delete child playlist")
}

func TestChildPlaylistService_GetChildPlaylist_Success(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChildRepo := repoMocks.NewMockChildPlaylistRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewChildPlaylistService(mockChildRepo, nil, nil, nil, logger)

	expectedPlaylist := &models.ChildPlaylist{ID: "cp123", Name: "Test"}
	mockChildRepo.EXPECT().GetByID(gomock.Any(), "cp123", "user123").Return(expectedPlaylist, nil)

	result, err := service.GetChildPlaylist(context.Background(), "cp123", "user123")

	assert.NoError(err)
	assert.Equal(expectedPlaylist, result)
}

func TestChildPlaylistService_GetChildPlaylist_Error(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChildRepo := repoMocks.NewMockChildPlaylistRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewChildPlaylistService(mockChildRepo, nil, nil, nil, logger)

	mockChildRepo.EXPECT().GetByID(gomock.Any(), "cp123", "user123").Return(nil, repositories.ErrChildPlaylistNotFound)

	result, err := service.GetChildPlaylist(context.Background(), "cp123", "user123")

	assert.Error(err)
	assert.Nil(result)
	assert.ErrorIs(err, repositories.ErrChildPlaylistNotFound)
}

func TestChildPlaylistService_GetChildPlaylistsByBasePlaylistID_Success(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChildRepo := repoMocks.NewMockChildPlaylistRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewChildPlaylistService(mockChildRepo, nil, nil, nil, logger)

	expectedPlaylists := []*models.ChildPlaylist{
		{ID: "cp1", Name: "Child 1"},
		{ID: "cp2", Name: "Child 2"},
	}
	mockChildRepo.EXPECT().GetByBasePlaylistID(gomock.Any(), "bp123", "user123").Return(expectedPlaylists, nil)

	result, err := service.GetChildPlaylistsByBasePlaylistID(context.Background(), "bp123", "user123")

	assert.NoError(err)
	assert.Equal(expectedPlaylists, result)
}

func TestChildPlaylistService_GetChildPlaylistsByBasePlaylistID_Error(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChildRepo := repoMocks.NewMockChildPlaylistRepository(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	service := NewChildPlaylistService(mockChildRepo, nil, nil, nil, logger)

	mockChildRepo.EXPECT().GetByBasePlaylistID(gomock.Any(), "bp123", "user123").Return(nil, repositories.ErrDatabaseOperation)

	result, err := service.GetChildPlaylistsByBasePlaylistID(context.Background(), "bp123", "user123")

	assert.Error(err)
	assert.Nil(result)
	assert.ErrorIs(err, repositories.ErrDatabaseOperation)
}

func TestChildPlaylistService_UpdateChildPlaylist_Success(t *testing.T) {
	tests := []struct {
		name                  string
		input                 *models.UpdateChildPlaylistRequest
		updatedChildPlaylist  *models.ChildPlaylist
		basePlaylist          *models.BasePlaylist
		needsBasePlaylistCall bool
		needsSpotifyCall      bool
		expectedSpotifyName   string
		expectedSpotifyDesc   string
	}{
		{
			name: "update both name and description",
			input: &models.UpdateChildPlaylistRequest{
				Name:        stringToPointer("Updated Child Name"),
				Description: stringToPointer("Updated description"),
			},
			updatedChildPlaylist: &models.ChildPlaylist{
				ID:                "cp789",
				BasePlaylistID:    "bp456",
				SpotifyPlaylistID: "sp_id",
				Name:              "Updated Child Name",
				Description:       "Updated description",
			},
			basePlaylist: &models.BasePlaylist{
				ID:   "bp456",
				Name: "Base Playlist Name",
			},
			needsBasePlaylistCall: true,
			needsSpotifyCall:      true,
			expectedSpotifyName:   "[Base Playlist Name] > Updated Child Name",
			expectedSpotifyDesc:   "[PLAYLIST GENERATED AND MANAGEED BY PlaylistRouter] Updated description",
		},
		{
			name: "update name only",
			input: &models.UpdateChildPlaylistRequest{
				Name: stringToPointer("Updated Name Only"),
			},
			updatedChildPlaylist: &models.ChildPlaylist{
				ID:                "cp789",
				BasePlaylistID:    "bp456",
				SpotifyPlaylistID: "sp_id",
				Name:              "Updated Name Only",
			},
			basePlaylist: &models.BasePlaylist{
				ID:   "bp456",
				Name: "Base Name",
			},
			needsBasePlaylistCall: true,
			needsSpotifyCall:      true,
			expectedSpotifyName:   "[Base Name] > Updated Name Only",
			expectedSpotifyDesc:   "",
		},
		{
			name: "update description only",
			input: &models.UpdateChildPlaylistRequest{
				Description: stringToPointer("Updated Description Only"),
			},
			updatedChildPlaylist: &models.ChildPlaylist{
				ID:                "cp789",
				SpotifyPlaylistID: "sp_id",
				Description:       "Updated Description Only",
			},
			needsBasePlaylistCall: false,
			needsSpotifyCall:      true,
			expectedSpotifyName:   "",
			expectedSpotifyDesc:   "[PLAYLIST GENERATED AND MANAGEED BY PlaylistRouter] Updated Description Only",
		},
		{
			name: "update IsActive only (no Spotify update)",
			input: &models.UpdateChildPlaylistRequest{
				IsActive: boolToPointer(true),
			},
			updatedChildPlaylist: &models.ChildPlaylist{
				ID:       "cp789",
				IsActive: true,
			},
			needsBasePlaylistCall: false,
			needsSpotifyCall:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Setup mocks
			mockChildRepo := repoMocks.NewMockChildPlaylistRepository(ctrl)
			var mockBaseRepo *repoMocks.MockBasePlaylistRepository
			var mockSpotifyClient *spotifyMocks.MockSpotifyAPI

			if tt.needsBasePlaylistCall {
				mockBaseRepo = repoMocks.NewMockBasePlaylistRepository(ctrl)
			}
			if tt.needsSpotifyCall {
				mockSpotifyClient = spotifyMocks.NewMockSpotifyAPI(ctrl)
			}

			service := createTestService(mockChildRepo, mockBaseRepo, nil, mockSpotifyClient)

			// Mock expectations
			expectedUpdateFields := repositories.UpdateChildPlaylistFields{
				Name:        tt.input.Name,
				Description: tt.input.Description,
				IsActive:    tt.input.IsActive,
				FilterRules: tt.input.FilterRules,
			}
			mockChildRepo.EXPECT().Update(gomock.Any(), "cp789", "user123", expectedUpdateFields).Return(tt.updatedChildPlaylist, nil)

			if tt.needsBasePlaylistCall {
				mockBaseRepo.EXPECT().GetByID(gomock.Any(), tt.basePlaylist.ID, "user123").Return(tt.basePlaylist, nil)
			}

			if tt.needsSpotifyCall {
				mockSpotifyClient.EXPECT().UpdatePlaylist(
					gomock.Any(),
					tt.updatedChildPlaylist.SpotifyPlaylistID,
					tt.expectedSpotifyName,
					tt.expectedSpotifyDesc,
				).Return(nil)
			}

			// Execute
			result, err := service.UpdateChildPlaylist(context.Background(), "cp789", "user123", tt.input)

			// Assert
			assert.NoError(err)
			assert.Equal(tt.updatedChildPlaylist, result)
		})
	}
}

func TestChildPlaylistService_UpdateChildPlaylist_RepoError(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChildRepo := repoMocks.NewMockChildPlaylistRepository(ctrl)
	mockChildRepo.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
	service := createTestService(mockChildRepo, nil, nil, nil)

	_, err := service.UpdateChildPlaylist(context.Background(), "cp789", "user123", &models.UpdateChildPlaylistRequest{})

	assert.Error(err)
	assert.Contains(err.Error(), "failed to update child playlist")
}

func TestChildPlaylistService_UpdateChildPlaylist_GetBasePlaylistError(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChildRepo := repoMocks.NewMockChildPlaylistRepository(ctrl)
	mockBaseRepo := repoMocks.NewMockBasePlaylistRepository(ctrl)
	service := createTestService(mockChildRepo, mockBaseRepo, nil, nil)

	newName := "New Name"
	input := &models.UpdateChildPlaylistRequest{Name: &newName}
	updatedChildPlaylist := &models.ChildPlaylist{BasePlaylistID: "bp456"}

	mockChildRepo.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(updatedChildPlaylist, nil)
	mockBaseRepo.EXPECT().GetByID(gomock.Any(), "bp456", gomock.Any()).Return(nil, errors.New("base playlist not found"))

	_, err := service.UpdateChildPlaylist(context.Background(), "cp789", "user123", input)

	assert.Error(err)
	assert.Contains(err.Error(), "failed to get base playlist")
}

func TestChildPlaylistService_UpdateChildPlaylist_SpotifyError(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChildRepo := repoMocks.NewMockChildPlaylistRepository(ctrl)
	mockBaseRepo := repoMocks.NewMockBasePlaylistRepository(ctrl)
	mockSpotifyClient := spotifyMocks.NewMockSpotifyAPI(ctrl)
	service := createTestService(mockChildRepo, mockBaseRepo, nil, mockSpotifyClient)

	newName := "New Name"
	input := &models.UpdateChildPlaylistRequest{Name: &newName}
	basePlaylist := &models.BasePlaylist{ID: "bp456", Name: "Base"}
	updatedChildPlaylist := &models.ChildPlaylist{
		BasePlaylistID:    "bp456",
		SpotifyPlaylistID: "sp_id",
	}

	mockChildRepo.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(updatedChildPlaylist, nil)
	mockBaseRepo.EXPECT().GetByID(gomock.Any(), "bp456", gomock.Any()).Return(basePlaylist, nil)
	mockSpotifyClient.EXPECT().UpdatePlaylist(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("spotify api error"))

	_, err := service.UpdateChildPlaylist(context.Background(), "cp789", "user123", input)

	assert.Error(err)
	assert.Contains(err.Error(), "failed to update spotify playlist")
}

// Helper functions for common test setups
func createTestService(
	childRepo repositories.ChildPlaylistRepository,
	baseRepo repositories.BasePlaylistRepository,
	spotifyIntegrationRepo repositories.SpotifyIntegrationRepository,
	spotifyClient spotifyclient.SpotifyAPI,
) *ChildPlaylistService {
	return NewChildPlaylistService(childRepo, baseRepo, spotifyIntegrationRepo, spotifyClient, createTestLogger())
}

func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func stringToPointer(s string) *string {
	return &s
}

func boolToPointer(b bool) *bool {
	return &b
}
