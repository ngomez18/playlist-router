package services

import (
	"context"

	"github.com/ngomez18/playlist-router/internal/models"
	"github.com/ngomez18/playlist-router/internal/repositories"
)

type BasePlaylistServicer interface {
	CreateBasePlaylist(ctx context.Context, input *models.CreateBasePlaylistRequest) (*models.BasePlaylist, error)
}

type BasePlaylistService struct {
	basePlaylistRepo repositories.BasePlaylistRepository
}

func NewBasePlaylistService(basePlaylistRepo repositories.BasePlaylistRepository) *BasePlaylistService {
	return &BasePlaylistService{
		basePlaylistRepo: basePlaylistRepo,
	}
}

func (bpService *BasePlaylistService) CreateBasePlaylist(ctx context.Context, input *models.CreateBasePlaylistRequest) (*models.BasePlaylist, error) {
	// TODO: Extract user ID from context or authentication
	// For now, using placeholder - this should come from JWT token or auth context
	userID := "placeholder_user_id"
	return bpService.basePlaylistRepo.Create(ctx, userID, input.Name, input.SpotifyPlaylistID)
}