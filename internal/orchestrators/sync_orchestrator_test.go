package orchestrators

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
	spotifyclient "github.com/ngomez18/playlist-router/internal/clients/spotify"
	clientmocks "github.com/ngomez18/playlist-router/internal/clients/spotify/mocks"
	"github.com/ngomez18/playlist-router/internal/models"
	servicemocks "github.com/ngomez18/playlist-router/internal/services/mocks"
	"github.com/stretchr/testify/require"
)

func TestNewDefaultSyncOrchestrator(t *testing.T) {
	assert := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockTrackAggregator := servicemocks.NewMockTrackAggregatorServicer(ctrl)
	mockTrackRouter := servicemocks.NewMockTrackRouterServicer(ctrl)
	mockChildPlaylistService := servicemocks.NewMockChildPlaylistServicer(ctrl)
	mockBasePlaylistService := servicemocks.NewMockBasePlaylistServicer(ctrl)
	mockSyncEventService := servicemocks.NewMockSyncEventServicer(ctrl)
	mockSpotifyClient := clientmocks.NewMockSpotifyAPI(ctrl)
	logger := createTestLogger()

	orchestrator := NewDefaultSyncOrchestrator(
		mockTrackAggregator,
		mockTrackRouter,
		mockChildPlaylistService,
		mockBasePlaylistService,
		mockSyncEventService,
		mockSpotifyClient,
		logger,
	)

	assert.NotNil(orchestrator)
	assert.Equal(mockTrackAggregator, orchestrator.trackAggregator)
	assert.Equal(mockTrackRouter, orchestrator.trackRouter)
	assert.Equal(mockChildPlaylistService, orchestrator.childPlaylistService)
	assert.Equal(mockSyncEventService, orchestrator.syncEventService)
	assert.Equal(mockSpotifyClient, orchestrator.spotifyClient)
	assert.NotNil(orchestrator.logger)
}

func TestDefaultSyncOrchestrator_SyncBasePlaylist_Success(t *testing.T) {
	assert := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup test data
	userID := "user123"
	basePlaylistID := "base456"

	childPlaylists := []*models.ChildPlaylist{
		{
			ID:                "child1",
			UserID:            userID,
			SpotifyPlaylistID: "spotify1",
			Name:              "Child 1",
			Description:       "Description 1",
			IsActive:          true,
		},
		{
			ID:                "child2",
			UserID:            userID,
			SpotifyPlaylistID: "spotify2",
			Name:              "Child 2",
			Description:       "Description 2",
			IsActive:          true,
		},
	}

	trackData := &models.PlaylistTracksInfo{
		PlaylistID:   basePlaylistID,
		APICallCount: 5,
		Tracks: []models.TrackInfo{
			{URI: "spotify:track:1", Name: "Track 1"},
			{URI: "spotify:track:2", Name: "Track 2"},
		},
	}

	routing := map[string][]string{
		"spotify1": {"spotify:track:1"},
		"spotify2": {"spotify:track:2"},
	}

	createdSyncEvent := &models.SyncEvent{
		ID:             "sync123",
		UserID:         userID,
		BasePlaylistID: basePlaylistID,
		Status:         models.SyncStatusInProgress,
	}

	// Setup mocks
	mocks := createMockServices(ctrl)
	orchestrator := createTestOrchestrator(mocks)

	// Mock expectations
	mocks.syncEventService.EXPECT().HasActiveSyncForBasePlaylist(gomock.Any(), userID, basePlaylistID).Return(false, nil)
	mocks.syncEventService.EXPECT().CreateSyncEvent(gomock.Any(), gomock.Any()).Return(createdSyncEvent, nil)
	mocks.basePlaylistService.EXPECT().GetBasePlaylist(gomock.Any(), basePlaylistID, userID).Return(&models.BasePlaylist{
		ID:     basePlaylistID,
		UserID: userID,
		Name:   "Test Base Playlist",
	}, nil)
	mocks.childPlaylistService.EXPECT().GetChildPlaylistsByBasePlaylistID(gomock.Any(), basePlaylistID, userID).Return(childPlaylists, nil)
	mocks.trackAggregator.EXPECT().AggregatePlaylistData(gomock.Any(), userID, basePlaylistID).Return(trackData, nil)
	mocks.trackRouter.EXPECT().RouteTracksToChildren(gomock.Any(), trackData, childPlaylists).Return(routing, nil)

	// Mock Spotify operations - use MinTimes/MaxTimes to handle non-deterministic map iteration order
	mocks.spotifyClient.EXPECT().DeletePlaylist(gomock.Any(), gomock.Any()).Return(nil).Times(2)
	mocks.spotifyClient.EXPECT().CreatePlaylist(gomock.Any(), gomock.Any(), gomock.Any(), false).DoAndReturn(
		func(ctx context.Context, name, desc string, private bool) (*spotifyclient.SpotifyPlaylist, error) {
			// Return different IDs based on the formatted name to ensure correct mapping
			switch name {
			case "[Test Base Playlist] > Child 1":
				return &spotifyclient.SpotifyPlaylist{ID: "new_spotify1", Name: name}, nil
			case "[Test Base Playlist] > Child 2":
				return &spotifyclient.SpotifyPlaylist{ID: "new_spotify2", Name: name}, nil
			default:
				return &spotifyclient.SpotifyPlaylist{ID: "unknown", Name: name}, nil
			}
		}).Times(2)

	// Mock child playlist updates - expect each exactly once but in any order
	mocks.childPlaylistService.EXPECT().UpdateChildPlaylistSpotifyID(gomock.Any(), "child1", userID, "new_spotify1").Return(childPlaylists[0], nil).Times(1)
	mocks.childPlaylistService.EXPECT().UpdateChildPlaylistSpotifyID(gomock.Any(), "child2", userID, "new_spotify2").Return(childPlaylists[1], nil).Times(1)

	// Mock track addition - expect each exactly once but in any order
	mocks.spotifyClient.EXPECT().AddTracksToPlaylist(gomock.Any(), "new_spotify1", []string{"spotify:track:1"}).Return(nil).Times(1)
	mocks.spotifyClient.EXPECT().AddTracksToPlaylist(gomock.Any(), "new_spotify2", []string{"spotify:track:2"}).Return(nil).Times(1)

	mocks.syncEventService.EXPECT().UpdateSyncEvent(gomock.Any(), createdSyncEvent.ID, gomock.Any()).Return(createdSyncEvent, nil)

	// Execute
	result, err := orchestrator.SyncBasePlaylist(context.Background(), userID, basePlaylistID)

	// Assert
	assert.NoError(err)
	assert.NotNil(result)
	assert.Equal(createdSyncEvent.ID, result.ID)
}

func TestDefaultSyncOrchestrator_SyncBasePlaylist_ActiveSyncInProgress(t *testing.T) {
	assert := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := "user123"
	basePlaylistID := "base456"

	mocks := createMockServices(ctrl)
	orchestrator := createTestOrchestrator(mocks)

	mocks.syncEventService.EXPECT().HasActiveSyncForBasePlaylist(gomock.Any(), userID, basePlaylistID).Return(true, nil)

	result, err := orchestrator.SyncBasePlaylist(context.Background(), userID, basePlaylistID)

	assert.Error(err)
	assert.Nil(result)
	assert.Contains(err.Error(), "sync already in progress")
}

func TestDefaultSyncOrchestrator_SyncBasePlaylist_NoChildPlaylists(t *testing.T) {
	assert := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := "user123"
	basePlaylistID := "base456"

	createdSyncEvent := &models.SyncEvent{
		ID:             "sync123",
		UserID:         userID,
		BasePlaylistID: basePlaylistID,
		Status:         models.SyncStatusInProgress,
	}

	mocks := createMockServices(ctrl)
	orchestrator := createTestOrchestrator(mocks)

	mocks.syncEventService.EXPECT().HasActiveSyncForBasePlaylist(gomock.Any(), userID, basePlaylistID).Return(false, nil)
	mocks.syncEventService.EXPECT().CreateSyncEvent(gomock.Any(), gomock.Any()).Return(createdSyncEvent, nil)
	mocks.basePlaylistService.EXPECT().GetBasePlaylist(gomock.Any(), basePlaylistID, userID).Return(&models.BasePlaylist{
		ID:     basePlaylistID,
		UserID: userID,
		Name:   "Test Base Playlist",
	}, nil)
	mocks.childPlaylistService.EXPECT().GetChildPlaylistsByBasePlaylistID(gomock.Any(), basePlaylistID, userID).Return([]*models.ChildPlaylist{}, nil)
	mocks.syncEventService.EXPECT().UpdateSyncEvent(gomock.Any(), createdSyncEvent.ID, gomock.Any()).Return(createdSyncEvent, nil)

	result, err := orchestrator.SyncBasePlaylist(context.Background(), userID, basePlaylistID)

	assert.NoError(err)
	assert.NotNil(result)
	assert.Equal(models.SyncStatusCompleted, result.Status)
}

func TestDefaultSyncOrchestrator_SyncBasePlaylist_TrackAggregationError(t *testing.T) {
	assert := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := "user123"
	basePlaylistID := "base456"

	childPlaylists := []*models.ChildPlaylist{
		{ID: "child1", UserID: userID, IsActive: true},
	}

	createdSyncEvent := &models.SyncEvent{
		ID:             "sync123",
		UserID:         userID,
		BasePlaylistID: basePlaylistID,
		Status:         models.SyncStatusInProgress,
	}

	mocks := createMockServices(ctrl)
	orchestrator := createTestOrchestrator(mocks)

	mocks.syncEventService.EXPECT().HasActiveSyncForBasePlaylist(gomock.Any(), userID, basePlaylistID).Return(false, nil)
	mocks.syncEventService.EXPECT().CreateSyncEvent(gomock.Any(), gomock.Any()).Return(createdSyncEvent, nil)
	mocks.basePlaylistService.EXPECT().GetBasePlaylist(gomock.Any(), basePlaylistID, userID).Return(&models.BasePlaylist{
		ID:     basePlaylistID,
		UserID: userID,
		Name:   "Test Base Playlist",
	}, nil)
	mocks.childPlaylistService.EXPECT().GetChildPlaylistsByBasePlaylistID(gomock.Any(), basePlaylistID, userID).Return(childPlaylists, nil)
	mocks.trackAggregator.EXPECT().AggregatePlaylistData(gomock.Any(), userID, basePlaylistID).Return(nil, errors.New("aggregation failed"))
	mocks.syncEventService.EXPECT().UpdateSyncEvent(gomock.Any(), createdSyncEvent.ID, gomock.Any()).Return(createdSyncEvent, nil)

	result, err := orchestrator.SyncBasePlaylist(context.Background(), userID, basePlaylistID)

	assert.Error(err)
	assert.NotNil(result)
	assert.Equal(models.SyncStatusFailed, result.Status)
	assert.Contains(err.Error(), "failed to aggregate track data")
}

func TestDefaultSyncOrchestrator_SyncChildPlaylist_Success(t *testing.T) {
	assert := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	basePlaylist := &models.BasePlaylist{
		ID:     "base1",
		UserID: "user123",
		Name:   "Base Playlist",
	}

	childPlaylist := models.ChildPlaylist{
		ID:                "child1",
		UserID:            "user123",
		SpotifyPlaylistID: "old_spotify1",
		Name:              "Child Playlist",
		Description:       "Test Description",
	}

	trackURIs := []string{"spotify:track:1", "spotify:track:2"}
	syncEvent := &models.SyncEvent{ID: "sync123"}

	// Expected formatted names
	expectedName := models.BuildChildPlaylistName(basePlaylist.Name, childPlaylist.Name)
	expectedDescription := models.BuildChildPlaylistDescription(childPlaylist.Description)

	newPlaylist := &spotifyclient.SpotifyPlaylist{
		ID:   "new_spotify1",
		Name: expectedName,
	}

	mocks := createMockServices(ctrl)
	orchestrator := createTestOrchestrator(mocks)

	// Mock expectations
	mocks.spotifyClient.EXPECT().DeletePlaylist(gomock.Any(), "old_spotify1").Return(nil)
	mocks.spotifyClient.EXPECT().CreatePlaylist(gomock.Any(), expectedName, expectedDescription, false).Return(newPlaylist, nil)
	mocks.childPlaylistService.EXPECT().UpdateChildPlaylistSpotifyID(gomock.Any(), childPlaylist.ID, childPlaylist.UserID, newPlaylist.ID).Return(&childPlaylist, nil)
	mocks.spotifyClient.EXPECT().AddTracksToPlaylist(gomock.Any(), newPlaylist.ID, trackURIs).Return(nil)

	// Execute
	apiRequestCount, err := orchestrator.syncChildPlaylist(context.Background(), basePlaylist, childPlaylist, "old_spotify1", trackURIs, syncEvent)

	// Assert
	assert.NoError(err)
	assert.Equal(3, apiRequestCount) // delete + create + add tracks
}

func TestDefaultSyncOrchestrator_SyncChildPlaylist_DeletePlaylistError(t *testing.T) {
	assert := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	basePlaylist := &models.BasePlaylist{
		ID:     "base1",
		UserID: "user123",
		Name:   "Base Playlist",
	}

	childPlaylist := models.ChildPlaylist{
		ID:                "child1",
		SpotifyPlaylistID: "old_spotify1",
	}

	trackURIs := []string{"spotify:track:1"}
	syncEvent := &models.SyncEvent{ID: "sync123"}

	mocks := createMockServices(ctrl)
	orchestrator := createTestOrchestrator(mocks)

	mocks.spotifyClient.EXPECT().DeletePlaylist(gomock.Any(), "old_spotify1").Return(errors.New("delete failed"))

	apiRequestCount, err := orchestrator.syncChildPlaylist(context.Background(), basePlaylist, childPlaylist, "old_spotify1", trackURIs, syncEvent)

	assert.Error(err)
	assert.Equal(0, apiRequestCount)
	assert.Contains(err.Error(), "failed to delete playlist")
}

func TestDefaultSyncOrchestrator_AddTracksInBatches_Success(t *testing.T) {
	assert := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test with 150 tracks (should create 2 batches of 100 and 50)
	trackURIs := make([]string, 150)
	for i := 0; i < 150; i++ {
		trackURIs[i] = "spotify:track:" + string(rune(i))
	}

	mocks := createMockServices(ctrl)
	orchestrator := createTestOrchestrator(mocks)

	// Expect 2 batch calls
	mocks.spotifyClient.EXPECT().AddTracksToPlaylist(gomock.Any(), "playlist123", trackURIs[0:100]).Return(nil)
	mocks.spotifyClient.EXPECT().AddTracksToPlaylist(gomock.Any(), "playlist123", trackURIs[100:150]).Return(nil)

	batchCount, err := orchestrator.addTracksInBatches(context.Background(), "sync123", "playlist123", trackURIs)

	assert.NoError(err)
	assert.Equal(2, batchCount)
}

func TestDefaultSyncOrchestrator_AddTracksInBatches_EmptyTracks(t *testing.T) {
	assert := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mocks := createMockServices(ctrl)
	orchestrator := createTestOrchestrator(mocks)

	// No mock expectations since no calls should be made

	batchCount, err := orchestrator.addTracksInBatches(context.Background(), "sync123", "playlist123", []string{})

	assert.NoError(err)
	assert.Equal(0, batchCount)
}

func TestDefaultSyncOrchestrator_AddTracksInBatches_BatchError(t *testing.T) {
	assert := require.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	trackURIs := []string{"spotify:track:1", "spotify:track:2"}

	mocks := createMockServices(ctrl)
	orchestrator := createTestOrchestrator(mocks)

	mocks.spotifyClient.EXPECT().AddTracksToPlaylist(gomock.Any(), "playlist123", trackURIs).Return(errors.New("batch failed"))

	batchCount, err := orchestrator.addTracksInBatches(context.Background(), "sync123", "playlist123", trackURIs)

	assert.Error(err)
	assert.Equal(0, batchCount)
	assert.Contains(err.Error(), "failed to add tracks batch")
}

// Helper structs and functions

type mockServices struct {
	trackAggregator      *servicemocks.MockTrackAggregatorServicer
	trackRouter          *servicemocks.MockTrackRouterServicer
	childPlaylistService *servicemocks.MockChildPlaylistServicer
	basePlaylistService  *servicemocks.MockBasePlaylistServicer
	syncEventService     *servicemocks.MockSyncEventServicer
	spotifyClient        *clientmocks.MockSpotifyAPI
}

func createMockServices(ctrl *gomock.Controller) mockServices {
	return mockServices{
		trackAggregator:      servicemocks.NewMockTrackAggregatorServicer(ctrl),
		trackRouter:          servicemocks.NewMockTrackRouterServicer(ctrl),
		childPlaylistService: servicemocks.NewMockChildPlaylistServicer(ctrl),
		basePlaylistService:  servicemocks.NewMockBasePlaylistServicer(ctrl),
		syncEventService:     servicemocks.NewMockSyncEventServicer(ctrl),
		spotifyClient:        clientmocks.NewMockSpotifyAPI(ctrl),
	}
}

func createTestOrchestrator(mocks mockServices) *DefaultSyncOrchestrator {
	return NewDefaultSyncOrchestrator(
		mocks.trackAggregator,
		mocks.trackRouter,
		mocks.childPlaylistService,
		mocks.basePlaylistService,
		mocks.syncEventService,
		mocks.spotifyClient,
		createTestLogger(),
	)
}

func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
}
