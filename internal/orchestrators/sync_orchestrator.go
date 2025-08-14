package orchestrators

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	spotifyclient "github.com/ngomez18/playlist-router/internal/clients/spotify"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/services"
)

const (
	MAX_PLAYLIST_TRACKS = 100
)

//go:generate mockgen -source=sync_orchestrator.go -destination=mocks/mock_sync_orchestrator.go -package=mocks

type SyncOrchestrator interface {
	SyncBasePlaylist(ctx context.Context, userID, basePlaylistID string) (*models.SyncEvent, error)
}

type DefaultSyncOrchestrator struct {
	trackAggregator      services.TrackAggregatorServicer
	trackRouter          services.TrackRouterServicer
	childPlaylistService services.ChildPlaylistServicer
	syncEventService     services.SyncEventServicer
	spotifyClient        spotifyclient.SpotifyAPI

	logger *slog.Logger
}

func NewDefaultSyncOrchestrator(
	trackAggregator services.TrackAggregatorServicer,
	trackRouter services.TrackRouterServicer,
	childPlaylistService services.ChildPlaylistServicer,
	syncEventService services.SyncEventServicer,
	spotifyClient spotifyclient.SpotifyAPI,
	logger *slog.Logger,
) *DefaultSyncOrchestrator {
	return &DefaultSyncOrchestrator{
		trackAggregator:      trackAggregator,
		trackRouter:          trackRouter,
		childPlaylistService: childPlaylistService,
		syncEventService:     syncEventService,
		spotifyClient:        spotifyClient,
		logger:               logger.With("component", "DefaultSyncOrchestrator"),
	}
}

func (s *DefaultSyncOrchestrator) SyncBasePlaylist(ctx context.Context, userID, basePlaylistID string) (*models.SyncEvent, error) {
	s.logger.InfoContext(ctx, "starting playlist sync orchestration",
		"user_id", userID,
		"base_playlist_id", basePlaylistID,
	)

	// Check for existing active sync
	hasActiveSync, err := s.syncEventService.HasActiveSyncForBasePlaylist(ctx, userID, basePlaylistID)
	if err != nil {
		return nil, fmt.Errorf("failed to check for active sync: %w", err)
	}
	if hasActiveSync {
		return nil, fmt.Errorf("sync already in progress for base playlist %s", basePlaylistID)
	}

	syncEvent := &models.SyncEvent{
		UserID:         userID,
		BasePlaylistID: basePlaylistID,
		Status:         models.SyncStatusInProgress,
		StartedAt:      time.Now(),
	}

	syncEvent, err = s.syncEventService.CreateSyncEvent(ctx, syncEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to create sync event: %w", err)
	}

	// Execute sync and handle completion/failure
	if syncErr := s.executeSyncFlow(ctx, syncEvent); syncErr != nil {
		s.completeSyncWithError(ctx, syncEvent, syncErr)
		return syncEvent, syncErr
	}

	s.completeSyncWithSuccess(ctx, syncEvent)
	return syncEvent, nil
}

func (s *DefaultSyncOrchestrator) executeSyncFlow(ctx context.Context, syncEvent *models.SyncEvent) error {
	// Get child playlists
	s.logger.InfoContext(ctx, "step 1: fetching child playlists", "sync_event_id", syncEvent.ID)

	childPlaylists, err := s.childPlaylistService.GetChildPlaylistsByBasePlaylistID(ctx, syncEvent.BasePlaylistID, syncEvent.UserID)
	if err != nil {
		return fmt.Errorf("failed to get child playlists: %w", err)
	}

	if len(childPlaylists) == 0 {
		s.logger.InfoContext(ctx, "no child playlists found, skipping sync", "sync_event_id", syncEvent.ID)
		return nil
	}

	childPlaylistIDs := make([]string, len(childPlaylists))
	for i, child := range childPlaylists {
		childPlaylistIDs[i] = child.ID
	}
	syncEvent.ChildPlaylistIDs = childPlaylistIDs

	s.logger.InfoContext(ctx, "found child playlists",
		"sync_event_id", syncEvent.ID,
		"child_playlist_count", len(childPlaylists),
	)

	// Aggregate track data
	s.logger.InfoContext(ctx, "step 2: aggregating track data", "sync_event_id", syncEvent.ID)

	trackData, err := s.trackAggregator.AggregatePlaylistData(ctx, syncEvent.UserID, syncEvent.BasePlaylistID)
	if err != nil {
		return fmt.Errorf("failed to aggregate track data: %w", err)
	}

	syncEvent.TracksProcessed = len(trackData.Tracks)
	syncEvent.TotalAPIRequests += trackData.APICallCount

	s.logger.InfoContext(ctx, "track aggregation completed",
		"sync_event_id", syncEvent.ID,
		"tracks_processed", syncEvent.TracksProcessed,
		"api_requests", trackData.APICallCount,
	)

	// Route tracks to child playlists
	s.logger.InfoContext(ctx, "step 3: routing tracks", "sync_event_id", syncEvent.ID)

	routing, err := s.trackRouter.RouteTracksToChildren(ctx, trackData, childPlaylists)
	if err != nil {
		return fmt.Errorf("failed to route tracks: %w", err)
	}

	totalRoutedTracks := 0
	for _, trackURIs := range routing {
		totalRoutedTracks += len(trackURIs)
	}

	s.logger.InfoContext(ctx, "track routing completed",
		"sync_event_id", syncEvent.ID,
		"child_playlists_with_tracks", len(routing),
		"total_routed_tracks", totalRoutedTracks,
	)

	// Update Spotify playlists (delete/recreate)
	s.logger.InfoContext(ctx, "step 4: updating spotify playlists", "sync_event_id", syncEvent.ID)

	if err := s.updateSpotifyPlaylists(ctx, syncEvent, childPlaylists, routing); err != nil {
		return fmt.Errorf("failed to update spotify playlists: %w", err)
	}

	s.logger.InfoContext(ctx, "spotify playlist updates completed", "sync_event_id", syncEvent.ID)
	return nil
}

func (s *DefaultSyncOrchestrator) updateSpotifyPlaylists(
	ctx context.Context,
	syncEvent *models.SyncEvent,
	childPlaylists []*models.ChildPlaylist,
	routing map[string][]string,
) error {
	playlistLookup := make(map[string]*models.ChildPlaylist)
	for _, child := range childPlaylists {
		playlistLookup[child.SpotifyPlaylistID] = child
	}

	for spotifyPlaylistID, trackURIs := range routing {
		childPlaylist, exists := playlistLookup[spotifyPlaylistID]
		if !exists {
			s.logger.WarnContext(ctx, "child playlist not found for spotify playlist",
				"spotify_playlist_id", spotifyPlaylistID,
				"sync_event_id", syncEvent.ID,
			)
			continue
		}

		apiRequestCount, err := s.syncChildPlaylist(ctx, *childPlaylist, spotifyPlaylistID, trackURIs, syncEvent)
		if err != nil {
			return err
		}

		syncEvent.TotalAPIRequests += apiRequestCount
	}

	return nil
}

func (s *DefaultSyncOrchestrator) syncChildPlaylist(
	ctx context.Context,
	childPlaylist models.ChildPlaylist,
	spotifyPlaylistID string,
	trackURIs []string,
	syncEvent *models.SyncEvent,
) (int, error) {
	apiRequestCount := 0

	s.logger.InfoContext(ctx, "recreating spotify playlist",
		"sync_event_id", syncEvent.ID,
		"child_playlist_id", childPlaylist.ID,
		"spotify_playlist_id", spotifyPlaylistID,
		"track_count", len(trackURIs),
	)

	if err := s.spotifyClient.DeletePlaylist(ctx, spotifyPlaylistID); err != nil {
		return apiRequestCount, fmt.Errorf("failed to delete playlist %s: %w", spotifyPlaylistID, err)
	}
	apiRequestCount++

	newPlaylist, err := s.spotifyClient.CreatePlaylist(ctx, childPlaylist.Name, childPlaylist.Description, false)
	if err != nil {
		return apiRequestCount, fmt.Errorf("failed to create new playlist for %s: %w", childPlaylist.Name, err)
	}
	apiRequestCount++

	s.logger.InfoContext(ctx, "created new playlist",
		"sync_event_id", syncEvent.ID,
		"old_spotify_playlist_id", spotifyPlaylistID,
		"new_spotify_playlist_id", newPlaylist.ID,
		"playlist_name", newPlaylist.Name,
	)

	_, err = s.childPlaylistService.UpdateChildPlaylistSpotifyID(ctx, childPlaylist.ID, childPlaylist.UserID, newPlaylist.ID)
	if err != nil {
		return apiRequestCount, fmt.Errorf("failed to update child playlist %s: %w", childPlaylist.Name, err)
	}

	batchCount, err := s.addTracksInBatches(ctx, syncEvent.ID, newPlaylist.ID, trackURIs)
	if err != nil {
		return apiRequestCount, fmt.Errorf("failed to add tracks to playlist %s: %w", newPlaylist.ID, err)
	}
	apiRequestCount += batchCount

	s.logger.InfoContext(ctx, "added tracks to new playlist",
		"sync_event_id", syncEvent.ID,
		"spotify_playlist_id", newPlaylist.ID,
		"tracks_added", len(trackURIs),
		"batch_count", batchCount,
	)

	return apiRequestCount, nil
}

func (s *DefaultSyncOrchestrator) addTracksInBatches(ctx context.Context, syncEventID, playlistID string, trackURIs []string) (int, error) {
	batchCount := 0

	for i := 0; i < len(trackURIs); i += MAX_PLAYLIST_TRACKS {
		end := min(i+MAX_PLAYLIST_TRACKS, len(trackURIs))

		batch := trackURIs[i:end]
		if err := s.spotifyClient.AddTracksToPlaylist(ctx, playlistID, batch); err != nil {
			return batchCount, fmt.Errorf("failed to add tracks batch %d-%d: %w", i, end, err)
		}

		batchCount++

		s.logger.InfoContext(ctx, "added track batch",
			"sync_event_id", syncEventID,
			"playlist_id", playlistID,
			"batch_start", i,
			"batch_end", end,
			"batch_size", len(batch),
		)
	}

	return batchCount, nil
}

func (s *DefaultSyncOrchestrator) completeSyncWithSuccess(ctx context.Context, syncEvent *models.SyncEvent) {
	now := time.Now()
	syncEvent.Status = models.SyncStatusCompleted
	syncEvent.CompletedAt = &now

	if _, err := s.syncEventService.UpdateSyncEvent(ctx, syncEvent.ID, syncEvent); err != nil {
		s.logger.ErrorContext(ctx, "failed to update sync event on success",
			"sync_event_id", syncEvent.ID,
			"error", err.Error(),
		)
	}

	s.logger.InfoContext(ctx, "playlist sync completed successfully",
		"sync_event_id", syncEvent.ID,
		"tracks_processed", syncEvent.TracksProcessed,
		"total_api_requests", syncEvent.TotalAPIRequests,
	)
}

func (s *DefaultSyncOrchestrator) completeSyncWithError(ctx context.Context, syncEvent *models.SyncEvent, syncErr error) {
	now := time.Now()
	errorMessage := syncErr.Error()
	syncEvent.Status = models.SyncStatusFailed
	syncEvent.CompletedAt = &now
	syncEvent.ErrorMessage = &errorMessage

	if _, err := s.syncEventService.UpdateSyncEvent(ctx, syncEvent.ID, syncEvent); err != nil {
		s.logger.ErrorContext(ctx, "failed to update sync event on error",
			"sync_event_id", syncEvent.ID,
			"error", err.Error(),
		)
	}

	s.logger.ErrorContext(ctx, "playlist sync failed",
		"sync_event_id", syncEvent.ID,
		"error", syncErr.Error(),
		"tracks_processed", syncEvent.TracksProcessed,
		"total_api_requests", syncEvent.TotalAPIRequests,
	)
}
