package services

import (
	"context"
	"fmt"
	"log/slog"

	spotifyclient "github.com/ngomez18/playlist-router/internal/clients/spotify"
	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
)

//go:generate mockgen -source=base_playlist_service.go -destination=mocks/mock_base_playlist_service.go -package=mocks

type BasePlaylistServicer interface {
	CreateBasePlaylist(ctx context.Context, userId string, input *models.CreateBasePlaylistRequest) (*models.BasePlaylist, error)
	DeleteBasePlaylist(ctx context.Context, id, userId string) error
	GetBasePlaylist(ctx context.Context, id, userId string) (*models.BasePlaylist, error)
	GetBasePlaylistsByUserID(ctx context.Context, userId string) ([]*models.BasePlaylist, error)
}

type BasePlaylistService struct {
	basePlaylistRepo       repositories.BasePlaylistRepository
	spotifyIntegrationRepo repositories.SpotifyIntegrationRepository
	spotifyClient          spotifyclient.SpotifyAPI
	logger                 *slog.Logger
}

func NewBasePlaylistService(
	basePlaylistRepo repositories.BasePlaylistRepository,
	spotifyIntegrationRepo repositories.SpotifyIntegrationRepository,
	spotifyClient spotifyclient.SpotifyAPI,
	logger *slog.Logger,
) *BasePlaylistService {
	return &BasePlaylistService{
		basePlaylistRepo:       basePlaylistRepo,
		spotifyIntegrationRepo: spotifyIntegrationRepo,
		spotifyClient:          spotifyClient,
		logger:                 logger.With("component", "BasePlaylistService"),
	}
}

func (bpService *BasePlaylistService) CreateBasePlaylist(ctx context.Context, userId string, input *models.CreateBasePlaylistRequest) (*models.BasePlaylist, error) {
	bpService.logger.InfoContext(ctx, "creating base playlist", "user_id", userId, "input", input)

	spotifyPlaylistID := input.SpotifyPlaylistID

	// If no Spotify playlist ID provided, create a new playlist in Spotify
	if spotifyPlaylistID == "" {
		bpService.logger.InfoContext(ctx, "spotify playlist ID empty, creating new playlist in Spotify", "name", input.Name)

		// Get user's Spotify integration to access tokens
		integration, err := bpService.spotifyIntegrationRepo.GetByUserID(ctx, userId)
		if err != nil {
			bpService.logger.ErrorContext(ctx, "failed to get spotify integration", "user_id", userId, "error", err.Error())
			return nil, fmt.Errorf("failed to get spotify integration: %w", err)
		}

		// Create playlist in Spotify
		spotifyPlaylist, err := bpService.spotifyClient.CreatePlaylist(
			ctx,
			integration.AccessToken,
			integration.SpotifyID,
			input.Name,
			"",    // empty description for now
			false, // private by default
		)
		if err != nil {
			bpService.logger.ErrorContext(ctx, "failed to create playlist in spotify", "error", err.Error())
			return nil, fmt.Errorf("failed to create spotify playlist: %w", err)
		}

		spotifyPlaylistID = spotifyPlaylist.ID
		bpService.logger.InfoContext(ctx, "successfully created spotify playlist", "spotify_playlist_id", spotifyPlaylistID, "name", spotifyPlaylist.Name)
	} else {
		// TODO: Validate that the provided Spotify playlist exists and is accessible
		bpService.logger.InfoContext(ctx, "using provided spotify playlist ID", "spotify_playlist_id", spotifyPlaylistID)
	}

	// Create the base playlist record in our database
	playlist, err := bpService.basePlaylistRepo.Create(ctx, userId, input.Name, spotifyPlaylistID)
	if err != nil {
		bpService.logger.ErrorContext(ctx, "failed to create base playlist", "error", err.Error())
		return nil, fmt.Errorf("failed to create playlist: %w", err)
	}

	bpService.logger.InfoContext(ctx, "base playlist created successfully", "base_playlist", playlist)
	return playlist, nil
}

func (bpService *BasePlaylistService) DeleteBasePlaylist(ctx context.Context, id, userId string) error {
	bpService.logger.InfoContext(ctx, "deleting base playlist", "id", id)

	err := bpService.basePlaylistRepo.Delete(ctx, id, userId)
	if err != nil {
		bpService.logger.ErrorContext(ctx, "failed to delete base playlist", "id", id, "error", err.Error())
		return fmt.Errorf("failed to delete playlist: %w", err)
	}

	bpService.logger.InfoContext(ctx, "base playlist deleted successfully", "id", id)
	return nil
}

func (bpService *BasePlaylistService) GetBasePlaylist(ctx context.Context, id, userId string) (*models.BasePlaylist, error) {
	bpService.logger.InfoContext(ctx, "retrieving base playlist", "id", id)

	playlist, err := bpService.basePlaylistRepo.GetByID(ctx, id, userId)
	if err != nil {
		bpService.logger.ErrorContext(ctx, "failed to retrieve base playlist", "id", id, "error", err.Error())
		return nil, fmt.Errorf("failed to retrieve playlist: %w", err)
	}

	bpService.logger.InfoContext(ctx, "base playlist retrieved successfully", "base_playlist", playlist)
	return playlist, nil
}

func (bpService *BasePlaylistService) GetBasePlaylistsByUserID(ctx context.Context, userId string) ([]*models.BasePlaylist, error) {
	bpService.logger.InfoContext(ctx, "retrieving base playlists for user", "user_id", userId)

	playlists, err := bpService.basePlaylistRepo.GetByUserID(ctx, userId)
	if err != nil {
		bpService.logger.ErrorContext(ctx, "failed to retrieve base playlists for user", "user_id", userId, "error", err.Error())
		return nil, fmt.Errorf("failed to retrieve playlists: %w", err)
	}

	bpService.logger.InfoContext(ctx, "base playlists retrieved successfully", "user_id", userId, "count", len(playlists))
	return playlists, nil
}
