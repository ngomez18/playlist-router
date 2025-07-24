package repositories

import (
	"context"

	"github.com/ngomez18/playlist-router/internal/models"
)

//go:generate mockgen -source=base_playlist_repository.go -destination=mocks/mock_base_playlist_repository.go -package=mocks

type BasePlaylistRepository interface {
	Create(ctx context.Context, userId, name, spotifyPlaylistId string) (*models.BasePlaylist, error)
	// GetByID(ctx context.Context, id string) (*models.BasePlaylist, error)
	// Delete(ctx context.Context, id string) error
}
