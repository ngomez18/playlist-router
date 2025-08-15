package services

import (
	"context"
	"log/slog"

	"github.com/ngomez18/playlist-router/internal/filters"
	"github.com/ngomez18/playlist-router/internal/models"
)

//go:generate mockgen -source=track_router_service.go -destination=mocks/mock_track_router_service.go -package=mocks

type TrackRouterServicer interface {
	RouteTracksToChildren(ctx context.Context, tracks *models.PlaylistTracksInfo, childPlaylists []*models.ChildPlaylist) (map[string][]string, error)
}

type TrackRouterService struct {
	logger *slog.Logger
}

func NewTrackRouterService(logger *slog.Logger) *TrackRouterService {
	return &TrackRouterService{
		logger: logger.With("component", "TrackRouterService"),
	}
}

func (r *TrackRouterService) RouteTracksToChildren(ctx context.Context, tracks *models.PlaylistTracksInfo, childPlaylists []*models.ChildPlaylist) (map[string][]string, error) {
	r.logger.InfoContext(ctx, "routing tracks to child playlists",
		"total_tracks", len(tracks.Tracks),
		"child_playlists", len(childPlaylists),
		"base_playlist", tracks.PlaylistID,
	)

	filterEngines := map[string]filters.FilterEngine{}

	for _, child := range childPlaylists {
		if !child.IsActive {
			continue
		}

		filterEngines[child.SpotifyPlaylistID] = *filters.NewFilterEngine(child)
	}

	routing := make(map[string][]string)

	for _, track := range tracks.Tracks {
		for childPlaylistId, filterEngine := range filterEngines {
			if filterEngine.MatchTrack(track) {
				routing[childPlaylistId] = append(routing[childPlaylistId], track.URI)
			}
		}
	}

	totalRouted := 0
	for playlistID, trackIDs := range routing {
		totalRouted += len(trackIDs)
		r.logger.InfoContext(ctx, "routing result",
			"child_playlist_id", playlistID,
			"matched_tracks", len(trackIDs),
		)
	}

	r.logger.InfoContext(ctx, "routing completed",
		"total_tracks_routed", totalRouted,
		"child_playlists_with_matches", len(routing),
	)

	return routing, nil
}
