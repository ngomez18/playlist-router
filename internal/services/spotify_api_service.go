package services

import (
	"context"
	"fmt"
	"log/slog"

	spotifyclient "github.com/ngomez18/playlist-router/internal/clients/spotify"
	requestcontext "github.com/ngomez18/playlist-router/internal/context"
	"github.com/ngomez18/playlist-router/internal/models"
)

//go:generate mockgen -source=spotify_api_service.go -destination=mocks/mock_spotify_api_service.go -package=mocks

type SpotifyAPIServicer interface {
	GetUserPlaylists(ctx context.Context, userID string) ([]*models.SpotifyPlaylist, error)
}

type SpotifyAPIService struct {
	spotifyClient spotifyclient.SpotifyAPI
	logger        *slog.Logger
}

func NewSpotifyAPIService(
	spotifyClient spotifyclient.SpotifyAPI,
	logger *slog.Logger,
) *SpotifyAPIService {
	return &SpotifyAPIService{
		spotifyClient: spotifyClient,
		logger:        logger.With("component", "SpotifyAPIService"),
	}
}

func (sas *SpotifyAPIService) GetUserPlaylists(ctx context.Context, userID string) ([]*models.SpotifyPlaylist, error) {
	sas.logger.InfoContext(ctx, "fetching user playlists from spotify", "user_id", userID)

	// Get the user's Spotify integration to get the access token
	integration, ok := requestcontext.GetSpotifyAuthFromContext(ctx)
	if !ok {
		sas.logger.ErrorContext(ctx, "failed to get spotify integration", "user_id", userID)
		return nil, fmt.Errorf("failed to get spotify integration")
	}

	// Fetch all playlists from Spotify API with pagination handling
	playlists, err := sas.spotifyClient.GetAllUserPlaylists(ctx, integration.AccessToken)
	if err != nil {
		sas.logger.ErrorContext(ctx, "failed to fetch user playlists from spotify", "user_id", userID, "error", err.Error())
		return nil, fmt.Errorf("failed to fetch user playlists: %w", err)
	}

	sas.logger.InfoContext(ctx, "successfully fetched user playlists", "user_id", userID, "playlist_count", len(playlists))
	return spotifyclient.ParseManySpotifyPlaylist(playlists), nil
}
