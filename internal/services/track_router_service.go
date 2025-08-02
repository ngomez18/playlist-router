package services

import (
	"context"

	"github.com/ngomez18/playlist-router/internal/models"
)

//go:generate mockgen -source=track_router_service.go -destination=mocks/mock_track_router_service.go -package=mocks

type TrackRouterServicer interface {
	// RouteTracksForBasePlaylist performs the complete track routing process for a base playlist
	// This includes:
	// 1. Creating a sync event to track the operation
	// 2. Fetching tracks from the base playlist via Spotify API
	// 3. Getting track metadata (artists, audio features, etc.)
	// 4. Applying filters for each child playlist to determine matching tracks
	// 5. Updating child playlists with matching tracks
	// 6. Completing the sync event with success/failure status
	RouteTracksForBasePlaylist(ctx context.Context, userID, basePlaylistID string) (*models.SyncEvent, error)
}