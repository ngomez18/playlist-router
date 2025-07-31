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
	requestcontext "github.com/ngomez18/playlist-router/internal/context"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
	repoMocks "github.com/ngomez18/playlist-router/internal/repositories/mocks"
	"github.com/stretchr/testify/assert"
)

// Helper functions for common test setups
func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func createTestService(
	childRepo repositories.ChildPlaylistRepository,
	baseRepo repositories.BasePlaylistRepository,
	spotifyIntegrationRepo repositories.SpotifyIntegrationRepository,
	spotifyClient spotifyclient.SpotifyAPI,
) *ChildPlaylistService {
	return NewChildPlaylistService(childRepo, baseRepo, spotifyIntegrationRepo, spotifyClient, createTestLogger())
}

func contextWithSpotifyIntegration() context.Context {
	integration := &models.SpotifyIntegration{
		AccessToken: "test_token",
		SpotifyID:   "test_spotify_id",
	}
	return requestcontext.ContextWithSpotifyAuth(context.Background(), integration)
}

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
	integration := &models.SpotifyIntegration{
		AccessToken: "spotify_access_token",
		SpotifyID:   "spotify_user_id",
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
		integration.AccessToken,
		integration.SpotifyID,
		expectedPlaylistName,
		expectedDescription,
		false,
	).Return(spotifyPlaylist, nil)
	mockChildRepo.EXPECT().Create(
		gomock.Any(),
		userID,
		basePlaylistID,
		input.Name,
		input.Description,
		spotifyPlaylist.ID,
		input.FilterRules,
	).Return(expectedChildPlaylist, nil)

	// Execution
	ctx := requestcontext.ContextWithSpotifyAuth(context.Background(), integration)
	result, err := service.CreateChildPlaylist(ctx, userID, basePlaylistID, input)

	// Assertions
	assert.NoError(err)
	assert.NotNil(result)
	assert.Equal(expectedChildPlaylist, result)
}

func TestChildPlaylistService_CreateChildPlaylist_MissingIntegration(t *testing.T) {
	assert := assert.New(t)
	service := createTestService(nil, nil, nil, nil)

	_, err := service.CreateChildPlaylist(context.Background(), "uid", "bpid", &models.CreateChildPlaylistRequest{Name: "Test"})

	assert.Error(err)
	assert.Contains(err.Error(), "failed to get spotify integration")
}

func TestChildPlaylistService_CreateChildPlaylist_GetBasePlaylistError(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBaseRepo := repoMocks.NewMockBasePlaylistRepository(ctrl)
	mockBaseRepo.EXPECT().GetByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
	service := createTestService(nil, mockBaseRepo, nil, nil)

	ctx := contextWithSpotifyIntegration()
	_, err := service.CreateChildPlaylist(ctx, "uid", "bpid", &models.CreateChildPlaylistRequest{Name: "Test"})

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
	mockSpotifyClient.EXPECT().CreatePlaylist(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("spotify api error"))
	service := createTestService(nil, mockBaseRepo, nil, mockSpotifyClient)

	ctx := contextWithSpotifyIntegration()
	_, err := service.CreateChildPlaylist(ctx, "uid", "bpid", &models.CreateChildPlaylistRequest{Name: "Test"})

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
	mockSpotifyClient.EXPECT().CreatePlaylist(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&spotifyclient.SpotifyPlaylist{ID: "sp_id"}, nil)
	mockChildRepo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
	service := createTestService(mockChildRepo, mockBaseRepo, nil, mockSpotifyClient)

	ctx := contextWithSpotifyIntegration()
	_, err := service.CreateChildPlaylist(ctx, "uid", "bpid", &models.CreateChildPlaylistRequest{Name: "Test"})

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
	integration := &models.SpotifyIntegration{
		AccessToken: "spotify_access_token",
		SpotifyID:   "spotify_user_id",
	}
	childPlaylist := &models.ChildPlaylist{
		ID:                childPlaylistID,
		SpotifyPlaylistID: "spotify_playlist_to_delete",
	}

	// Mock Calls
	mockChildRepo.EXPECT().GetByID(gomock.Any(), childPlaylistID, userID).Return(childPlaylist, nil)
	mockSpotifyClient.EXPECT().DeletePlaylist(gomock.Any(), integration.AccessToken, integration.SpotifyID, childPlaylist.SpotifyPlaylistID).Return(nil)
	mockChildRepo.EXPECT().Delete(gomock.Any(), childPlaylistID, userID).Return(nil)

	// Execution
	ctx := requestcontext.ContextWithSpotifyAuth(context.Background(), integration)
	err := service.DeleteChildPlaylist(ctx, childPlaylistID, userID)

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

	ctx := contextWithSpotifyIntegration()
	err := service.DeleteChildPlaylist(ctx, "cpid", "uid")

	assert.Error(err)
	assert.Contains(err.Error(), "failed to get child playlist")
}

func TestChildPlaylistService_DeleteChildPlaylist_MissingIntegration(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChildRepo := repoMocks.NewMockChildPlaylistRepository(ctrl)
	mockChildRepo.EXPECT().GetByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.ChildPlaylist{}, nil)
	service := createTestService(mockChildRepo, nil, nil, nil)

	err := service.DeleteChildPlaylist(context.Background(), "cpid", "uid")

	assert.Error(err)
	assert.Contains(err.Error(), "failed to get spotify integration")
}

func TestChildPlaylistService_DeleteChildPlaylist_SpotifyError(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChildRepo := repoMocks.NewMockChildPlaylistRepository(ctrl)
	mockSpotifyClient := spotifyMocks.NewMockSpotifyAPI(ctrl)
	mockChildRepo.EXPECT().GetByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.ChildPlaylist{}, nil)
	mockSpotifyClient.EXPECT().DeletePlaylist(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("spotify api error"))
	service := createTestService(mockChildRepo, nil, nil, mockSpotifyClient)

	ctx := contextWithSpotifyIntegration()
	err := service.DeleteChildPlaylist(ctx, "cpid", "uid")

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
	mockSpotifyClient.EXPECT().DeletePlaylist(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockChildRepo.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("db error"))
	service := createTestService(mockChildRepo, nil, nil, mockSpotifyClient)

	ctx := contextWithSpotifyIntegration()
	err := service.DeleteChildPlaylist(ctx, "cpid", "uid")

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
