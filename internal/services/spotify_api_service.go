package services

import (
	"context"
	"fmt"
	"log/slog"

	spotifyclient "github.com/ngomez18/playlist-router/internal/clients/spotify"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
)

//go:generate mockgen -source=spotify_api_service.go -destination=mocks/mock_spotify_api_service.go -package=mocks

type SpotifyAPIServicer interface {
	GetFilteredUserPlaylists(ctx context.Context, userID string) ([]*models.SpotifyPlaylist, error)
}

type SpotifyAPIService struct {
	spotifyClient     spotifyclient.SpotifyAPI
	basePlaylistRepo  repositories.BasePlaylistRepository
	childPlaylistRepo repositories.ChildPlaylistRepository
	logger            *slog.Logger
}

func NewSpotifyAPIService(
	spotifyClient spotifyclient.SpotifyAPI,
	basePlaylistRepo repositories.BasePlaylistRepository,
	childPlaylistRepo repositories.ChildPlaylistRepository,
	logger *slog.Logger,
) *SpotifyAPIService {
	return &SpotifyAPIService{
		spotifyClient:     spotifyClient,
		basePlaylistRepo:  basePlaylistRepo,
		childPlaylistRepo: childPlaylistRepo,
		logger:            logger.With("component", "SpotifyAPIService"),
	}
}

func (sas *SpotifyAPIService) GetFilteredUserPlaylists(ctx context.Context, userID string) ([]*models.SpotifyPlaylist, error) {
	sas.logger.InfoContext(ctx, "fetching filtered user playlists from spotify", "user_id", userID)

	allPlaylists, err := sas.spotifyClient.GetAllUserPlaylists(ctx)
	if err != nil {
		return nil, err
	}

	// Get existing base playlists to exclude their Spotify IDs
	basePlaylists, err := sas.basePlaylistRepo.GetByUserID(ctx, userID)
	if err != nil {
		sas.logger.ErrorContext(ctx, "failed to fetch base playlists", "user_id", userID, "error", err.Error())
		return nil, fmt.Errorf("failed to fetch base playlists: %w", err)
	}

	// Get existing child playlists to exclude their Spotify IDs
	var childPlaylists []*models.ChildPlaylist
	for _, basePlaylist := range basePlaylists {
		baseChildPlaylists, err := sas.childPlaylistRepo.GetByBasePlaylistID(ctx, basePlaylist.ID, userID)
		if err != nil {
			sas.logger.ErrorContext(ctx, "failed to fetch child playlists for base playlist",
				"user_id", userID, "base_playlist_id", basePlaylist.ID, "error", err.Error())

			return nil, fmt.Errorf("failed to fetch child playlists: %w", err)
		}
		childPlaylists = append(childPlaylists, baseChildPlaylists...)
	}

	usedPlaylistIDs := make(map[string]bool)
	for _, basePlaylist := range basePlaylists {
		usedPlaylistIDs[basePlaylist.SpotifyPlaylistID] = true
	}
	for _, childPlaylist := range childPlaylists {
		usedPlaylistIDs[childPlaylist.SpotifyPlaylistID] = true
	}

	filteredPlaylists := make([]*models.SpotifyPlaylist, 0)
	for _, playlist := range allPlaylists {
		if !usedPlaylistIDs[playlist.ID] {
			filteredPlaylists = append(filteredPlaylists, spotifyclient.ParseSpotifyPlaylist(playlist))
		}
	}

	sas.logger.InfoContext(ctx, "successfully filtered user playlists",
		"user_id", userID,
		"total_playlists", len(allPlaylists),
		"filtered_playlists", len(filteredPlaylists),
		"excluded_count", len(usedPlaylistIDs))

	return filteredPlaylists, nil
}
