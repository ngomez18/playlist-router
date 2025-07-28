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
	CreateBasePlaylist(ctx context.Context, userId string, input *models.CreateBasePlaylistRequest) (*models.BasePlaylist, error)
	DeleteBasePlaylist(ctx context.Context, id, userId string) error
	GetBasePlaylist(ctx context.Context, id, userId string) (*models.BasePlaylist, error)
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

func (bpService *BasePlaylistService) CreateBasePlaylist(ctx context.Context, userId string, input *models.CreateBasePlaylistRequest) (*models.BasePlaylist, error) {
	bpService.logger.InfoContext(ctx, "creating base playlist", "user_id", userId, "input", input)

	 // TODO: Check SpotifyPlaylistID
	 // If empty, should create playlist in Spotify and get the ID
	 // If not empty, should validate Spotify playlist exists and is accessible

	playlist, err := bpService.basePlaylistRepo.Create(ctx, userId, input.Name, input.SpotifyPlaylistID)
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
