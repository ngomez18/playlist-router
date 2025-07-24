package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
)

//go:generate mockgen -source=base_playlist_service.go -destination=mocks/mock_base_playlist_service.go -package=mocks

type BasePlaylistServicer interface {
	CreateBasePlaylist(ctx context.Context, input *models.CreateBasePlaylistRequest) (*models.BasePlaylist, error)
}

type BasePlaylistService struct {
	basePlaylistRepo repositories.BasePlaylistRepository
	logger           *slog.Logger
}

func NewBasePlaylistService(basePlaylistRepo repositories.BasePlaylistRepository, logger *slog.Logger) *BasePlaylistService {
	return &BasePlaylistService{
		basePlaylistRepo: basePlaylistRepo,
		logger:           logger.With("component", "BasePlaylistService"),
	}
}

func (bpService *BasePlaylistService) CreateBasePlaylist(ctx context.Context, input *models.CreateBasePlaylistRequest) (*models.BasePlaylist, error) {
	bpService.logger.InfoContext(ctx, "creating base playlist", 
		slog.String("name", input.Name),
		slog.String("spotify_id", input.SpotifyPlaylistID),
	)

	// TODO: Extract user ID from context or authentication
	// For now, using placeholder - this should come from JWT token or auth context
	userID := "placeholder_user_id"
	
	playlist, err := bpService.basePlaylistRepo.Create(ctx, userID, input.Name, input.SpotifyPlaylistID)
	if err != nil {
		bpService.logger.ErrorContext(ctx, "failed to create base playlist", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to create playlist: %w", err)
	}

	bpService.logger.InfoContext(ctx, "base playlist created successfully", 
		slog.String("id", playlist.ID),
		slog.String("user_id", playlist.UserID),
		slog.String("name", playlist.Name),
	)

	return playlist, nil
}
